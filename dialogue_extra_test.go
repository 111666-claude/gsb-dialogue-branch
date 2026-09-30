package dialogue

import (
	"math/rand"
	"testing"
)

func TestRequireBlocksEntryAndState(t *testing.T) {
	instance := session()
	if instance.Enter("b") {
		t.Fatal("没有 vip 标记不该进得去 b")
	}
	if instance.Current() != "a" {
		t.Fatalf("前置不满足不该改当前节点，得到 %s", instance.Current())
	}
	if instance.Reward() != 0 {
		t.Fatal("前置不满足不该发奖励")
	}
	if instance.Back() {
		t.Fatal("前置不满足不该压栈")
	}

	instance.SetFlag("vip", false)
	if instance.Enter("b") {
		t.Fatal("vip=false 时也不该进得去 b")
	}
	instance.SetFlag("vip", true)
	if !instance.Enter("b") {
		t.Fatal("vip=true 后应该进得去 b")
	}
}

func TestRewardOnlyOnFirstEnter(t *testing.T) {
	instance := session()
	instance.SetFlag("vip", true)
	if !instance.Enter("b") || instance.Reward() != 5 {
		t.Fatalf("首次进入 b 应加 5 奖励，reward=%d", instance.Reward())
	}
	if !instance.Enter("b") || instance.Reward() != 5 {
		t.Fatalf("重复进入 b 不该再加奖励，reward=%d", instance.Reward())
	}
	if !instance.Back() || instance.Current() != "a" {
		t.Fatal("回到 a 后才能验证再次进入 b 的幂等")
	}
	if !instance.Enter("b") || instance.Reward() != 5 {
		t.Fatalf("离开后再次进入 b 仍不该加奖励，reward=%d", instance.Reward())
	}
}

func TestReEnterCurrentDoesNotPush(t *testing.T) {
	instance := session()
	instance.SetFlag("vip", true)
	instance.Enter("b")
	instance.Enter("b")
	instance.Enter("b")
	if !instance.Back() || instance.Current() != "a" {
		t.Fatal("当前节点重复进入不该压栈，Back 应直接回到 a")
	}
	if instance.Back() {
		t.Fatal("栈已空，Back 应返回 false")
	}
}

func TestBackRestoresStackOrder(t *testing.T) {
	instance := NewSession([]Node{
		{ID: "a", Goto: map[string]string{"go": "b"}},
		{ID: "b", Goto: map[string]string{"go": "c"}, Require: map[string]bool{"vip": true}},
		{ID: "c", Require: map[string]bool{"vip": true}},
	}, "a")
	if instance.Back() {
		t.Fatal("初始栈为空，Back 应失败")
	}
	instance.SetFlag("vip", true)
	instance.Choose("go")
	instance.Choose("go")
	if instance.Current() != "c" {
		t.Fatalf("应已到达 c，得到 %s", instance.Current())
	}
	if !instance.Back() || instance.Current() != "b" {
		t.Fatalf("第一次 Back 应回到 b，得到 %s", instance.Current())
	}
	if !instance.Back() || instance.Current() != "a" {
		t.Fatalf("第二次 Back 应回到 a，得到 %s", instance.Current())
	}
	if instance.Back() {
		t.Fatal("回到起点后栈应为空")
	}
}

func TestChooseUsesOnlyCurrentNode(t *testing.T) {
	instance := NewSession([]Node{
		{ID: "a", Goto: map[string]string{"ask": "b"}},
		{ID: "b", Require: map[string]bool{"vip": true}},
	}, "a")
	if instance.Choose("missing") {
		t.Fatal("当前节点没有该分支，应失败")
	}
	if instance.Scanned() != 0 {
		t.Fatal("分支不存在时不该产生扫描计数")
	}
	if instance.Choose("ask") {
		t.Fatal("目标 b 前置不满足，Choose 应失败")
	}
	if instance.Current() != "a" {
		t.Fatal("目标前置不满足时当前节点不该变")
	}
	instance.SetFlag("vip", true)
	if !instance.Choose("ask") || instance.Current() != "b" {
		t.Fatal("满足前置后 Choose 应到达 b")
	}
	if scanned := instance.Scanned(); scanned != 2 {
		t.Fatalf("两次定位当前节点分支应计数 2，得到 %d", scanned)
	}
}

func TestScannedIndependentOfNodeCount(t *testing.T) {
	const total = 3000
	nodes := make([]Node, 0, total)
	for i := 0; i < total; i++ {
		node := Node{ID: nodeName(i), Reward: i}
		if i+1 < total {
			node.Goto = map[string]string{"next": nodeName(i + 1)}
		}
		nodes = append(nodes, node)
	}
	instance := NewSession(nodes, nodeName(0))
	for i := 0; i < total-1; i++ {
		if !instance.Choose("next") {
			t.Fatalf("第 %d 次选择应成功", i)
		}
	}
	if scanned := instance.Scanned(); scanned >= total {
		t.Fatalf("Scanned 不应随节点总数增长，得到 %d（节点 %d）", scanned, total)
	}
	if instance.Current() != nodeName(total-1) {
		t.Fatalf("应到达末节点，得到 %s", instance.Current())
	}
	want := total * (total - 1) / 2
	if instance.Reward() != want {
		t.Fatalf("奖励应等于所有首次进入节点奖励之和 %d，得到 %d", want, instance.Reward())
	}
}

func nodeName(i int) string {
	return "n-" + itoa(i)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

// ---- 参考实现：条件校验 + 首次奖励 + 栈回退（与产品实现独立书写） ----

type refNode struct {
	require map[string]bool
	reward  int
	gotoTo  map[string]string
}

type refSession struct {
	nodes   map[string]refNode
	flags   map[string]bool
	entered map[string]bool
	current string
	stack   []string
	reward  int
	scanned int
}

func newRefSession(nodes map[string]refNode, start string) *refSession {
	return &refSession{nodes: nodes, flags: map[string]bool{}, entered: map[string]bool{}, current: start}
}

func (r *refSession) setFlag(name string, value bool) { r.flags[name] = value }

func (r *refSession) enter(id string) bool {
	node, ok := r.nodes[id]
	if !ok {
		return false
	}
	for flag, want := range node.require {
		if r.flags[flag] != want {
			return false
		}
	}
	if r.current != id {
		r.stack = append(r.stack, r.current)
	}
	r.current = id
	if !r.entered[id] {
		r.entered[id] = true
		r.reward += node.reward
	}
	return true
}

func (r *refSession) choose(branch string) bool {
	next, ok := r.nodes[r.current].gotoTo[branch]
	if !ok {
		return false
	}
	r.scanned++
	return r.enter(next)
}

func (r *refSession) back() bool {
	n := len(r.stack)
	if n == 0 {
		return false
	}
	r.current = r.stack[n-1]
	r.stack = r.stack[:n-1]
	return true
}

// TestRandomDifferential 用 200 组随机「进入 + 选分支 + 回退」序列
// （含条件不满足与重复进入）对照参考实现，逐次比对结果、当前节点、奖励与扫描计数。
func TestRandomDifferential(t *testing.T) {
	const groups = 200
	const ops = 240
	flags := []string{"vip", "member", "beta"}
	branches := []string{"go", "ask", "next"}

	type step struct {
		kind  int
		arg   string
		flag  string
		value bool
	}

	for group := 0; group < groups; group++ {
		rng := rand.New(rand.NewSource(int64(group + 1)))
		nodeCount := 1 + rng.Intn(12)
		nodeIDs := make([]string, nodeCount)
		productNodes := make([]Node, 0, nodeCount)
		refNodes := map[string]refNode{}
		for i := 0; i < nodeCount; i++ {
			nodeIDs[i] = "n" + itoa(i)
			var require map[string]bool
			if rng.Intn(2) == 0 {
				require = map[string]bool{flags[rng.Intn(len(flags))]: rng.Intn(2) == 0}
			}
			reward := 0
			if rng.Intn(2) == 0 {
				reward = rng.Intn(11)
			}
			var gotoTo map[string]string
			if rng.Intn(2) == 0 {
				gotoTo = map[string]string{branches[rng.Intn(len(branches))]: nodeIDs[rng.Intn(nodeCount)]}
				if rng.Intn(2) == 0 {
					gotoTo[branches[rng.Intn(len(branches))]] = nodeIDs[rng.Intn(nodeCount)]
				}
			}
			productNodes = append(productNodes, Node{ID: nodeIDs[i], Require: require, Reward: reward, Goto: gotoTo})
			refNodes[nodeIDs[i]] = refNode{require: require, reward: reward, gotoTo: gotoTo}
		}

		steps := make([]step, ops)
		for op := range steps {
			steps[op] = step{
				kind:  rng.Intn(3),
				flag:  flags[rng.Intn(len(flags))],
				value: rng.Intn(2) == 0,
			}
			switch steps[op].kind {
			case 0:
				if rng.Intn(4) == 0 {
					steps[op].arg = "missing" + itoa(op) // 不存在的节点
				} else {
					steps[op].arg = nodeIDs[rng.Intn(nodeCount)]
				}
			case 1:
				if rng.Intn(4) == 0 {
					steps[op].arg = "nope" + itoa(op) // 不存在的分支
				} else {
					steps[op].arg = branches[rng.Intn(len(branches))]
				}
			}
		}

		run := func() (results []bool, currents []string, rewards, scanned []int) {
			product := NewSession(productNodes, nodeIDs[0])
			ref := newRefSession(refNodes, nodeIDs[0])
			results = make([]bool, ops)
			currents = make([]string, ops)
			rewards = make([]int, ops)
			scanned = make([]int, ops)
			for op, st := range steps {
				product.SetFlag(st.flag, st.value)
				ref.setFlag(st.flag, st.value)
				var got bool
				switch st.kind {
				case 0:
					got = product.Enter(st.arg)
					if want := ref.enter(st.arg); got != want {
						t.Fatalf("group %d op %d Enter(%s) 结果 %v，参考 %v", group, op, st.arg, got, want)
					}
				case 1:
					got = product.Choose(st.arg)
					if want := ref.choose(st.arg); got != want {
						t.Fatalf("group %d op %d Choose(%s) 结果 %v，参考 %v", group, op, st.arg, got, want)
					}
				case 2:
					got = product.Back()
					if want := ref.back(); got != want {
						t.Fatalf("group %d op %d Back 结果 %v，参考 %v", group, op, got, want)
					}
				}
				results[op] = got
				currents[op] = product.Current()
				rewards[op] = product.Reward()
				scanned[op] = product.Scanned()
				if product.Current() != ref.current {
					t.Fatalf("group %d op %d 当前节点 %s，参考 %s", group, op, product.Current(), ref.current)
				}
				if product.Reward() != ref.reward {
					t.Fatalf("group %d op %d 奖励 %d，参考 %d", group, op, product.Reward(), ref.reward)
				}
				if product.Scanned() != ref.scanned {
					t.Fatalf("group %d op %d 扫描 %d，参考 %d", group, op, product.Scanned(), ref.scanned)
				}
			}
			return results, currents, rewards, scanned
		}

		r1, c1, w1, s1 := run()
		r2, c2, w2, s2 := run()
		for op := 0; op < ops; op++ {
			if r1[op] != r2[op] || c1[op] != c2[op] || w1[op] != w2[op] || s1[op] != s2[op] {
				t.Fatalf("group %d op %d 同一串操作重复执行结果不一致", group, op)
			}
		}
	}
}

func rngn(rng *rand.Rand, n int) int { return rng.Intn(n) }
