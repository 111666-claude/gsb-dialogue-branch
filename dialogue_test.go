package dialogue

import (
	"fmt"
	"math/rand"
	"testing"
)

func session() *Session {
	instance := NewSession([]Node{
		{ID: "a", Reward: 10, Goto: map[string]string{"ask": "b"}},
		{ID: "b", Reward: 5, Require: map[string]bool{"vip": true}},
	}, "a")
	instance.SetFlag("vip", true)
	return instance
}

func TestStartsAtStartNode(t *testing.T) {
	if session().Current() != "a" {
		t.Fatal("会话应该从起点开始")
	}
}

func TestEnterKnownNode(t *testing.T) {
	if !session().Enter("b") {
		t.Fatal("进入已知节点应该成功")
	}
}

func TestEnterUnknownNodeFails(t *testing.T) {
	if session().Enter("missing") {
		t.Fatal("不存在的节点不该进得去")
	}
}

func TestRewardAfterFirstEnter(t *testing.T) {
	instance := session()
	instance.Enter("b")
	if instance.Reward() <= 0 {
		t.Fatal("进入带奖励的节点应该加奖励")
	}
}

func TestEnterBlockedWithoutRequiredFlag(t *testing.T) {
	instance := NewSession([]Node{
		{ID: "a", Goto: map[string]string{"ask": "b"}},
		{ID: "b", Reward: 5, Require: map[string]bool{"vip": true}},
	}, "a")
	if instance.Enter("b") {
		t.Fatal("缺少前置标记不该进得去")
	}
	if instance.Current() != "a" {
		t.Fatal("进入失败不该改当前节点")
	}
	if instance.Reward() != 0 {
		t.Fatal("进入失败不该发奖励")
	}
	if instance.Back() {
		t.Fatal("进入失败不该压栈")
	}
}

func TestEnterBlockedWhenFlagMismatch(t *testing.T) {
	instance := NewSession([]Node{
		{ID: "a"},
		{ID: "b", Require: map[string]bool{"vip": true}},
		{ID: "c", Require: map[string]bool{"banned": false}},
	}, "a")
	instance.SetFlag("vip", false)
	if instance.Enter("b") {
		t.Fatal("标记为假不该满足 Require")
	}
	instance.SetFlag("banned", true)
	if instance.Enter("c") {
		t.Fatal("Require 要求为假的标记被置真后不该进得去")
	}
	instance.SetFlag("vip", true)
	instance.SetFlag("banned", false)
	if !instance.Enter("b") || !instance.Enter("c") {
		t.Fatal("标记全部满足后应该进得去")
	}
}

func TestRewardGrantedOncePerNode(t *testing.T) {
	instance := session()
	instance.Enter("b")
	instance.Enter("b")
	if instance.Reward() != 5 {
		t.Fatalf("同一节点进两次应该只发一次奖励，得到 %d", instance.Reward())
	}
	instance.Back()
	instance.Enter("b")
	if instance.Reward() != 5 {
		t.Fatalf("回退后再进同一节点不该重复发奖励，得到 %d", instance.Reward())
	}
}

func TestBackPopsPreviousNode(t *testing.T) {
	instance := session()
	instance.Enter("b")
	if !instance.Back() {
		t.Fatal("有上一个节点应该回退成功")
	}
	if instance.Current() != "a" {
		t.Fatal("回退后应该回到上一个节点")
	}
	if instance.Back() {
		t.Fatal("栈空应该回退失败")
	}
}

func TestBackOnEmptyStackFails(t *testing.T) {
	instance := session()
	if instance.Back() {
		t.Fatal("空栈回退应该失败")
	}
	if instance.Current() != "a" {
		t.Fatal("回退失败不该改当前节点")
	}
}

func TestChooseFollowsCurrentNodeGoto(t *testing.T) {
	instance := NewSession([]Node{
		{ID: "a", Goto: map[string]string{"ask": "b"}},
		{ID: "b", Goto: map[string]string{"ask": "c"}},
		{ID: "c", Goto: map[string]string{"ask": "a"}},
	}, "a")
	if !instance.Choose("ask") || instance.Current() != "b" {
		t.Fatal("应该按当前节点的 Goto 定位到 b")
	}
	if !instance.Choose("ask") || instance.Current() != "c" {
		t.Fatal("移动后应该按新当前节点的 Goto 定位到 c")
	}
	if instance.Choose("missing") {
		t.Fatal("当前节点没有的分支不该选得动")
	}
	if instance.Current() != "c" {
		t.Fatal("选分支失败不该改当前节点")
	}
}

func TestChooseCostDoesNotScaleWithTreeSize(t *testing.T) {
	const size = 3000
	table := make([]Node, 0, size)
	for index := 0; index < size; index++ {
		table = append(table, Node{ID: fmt.Sprintf("n-%d", index)})
	}
	instance := NewSession(table, "n-0")
	const queries = 3000
	for query := 0; query < queries; query++ {
		instance.Choose("ask")
	}
	if instance.Scanned() > queries {
		t.Fatalf("选 %d 次分支扫了 %d 个节点，单次应该是 O(1)", queries, instance.Scanned())
	}
}

// refSession 是参考实现：条件校验 + 首次奖励 + 栈回退。
type refSession struct {
	nodes   map[string]Node
	current string
	stack   []string
	flags   map[string]bool
	reward  int
	entered map[string]bool
}

func newRefSession(nodes []Node, start string) *refSession {
	table := map[string]Node{}
	for _, node := range nodes {
		table[node.ID] = node
	}
	return &refSession{nodes: table, current: start, flags: map[string]bool{}, entered: map[string]bool{}}
}

func (r *refSession) setFlag(name string, value bool) { r.flags[name] = value }

func (r *refSession) allowed(node Node) bool {
	for name, want := range node.Require {
		if r.flags[name] != want {
			return false
		}
	}
	return true
}

func (r *refSession) enter(id string) bool {
	node, ok := r.nodes[id]
	if !ok || !r.allowed(node) {
		return false
	}
	if r.current != id {
		r.stack = append(r.stack, r.current)
	}
	r.current = id
	if !r.entered[id] {
		r.entered[id] = true
		r.reward += node.Reward
	}
	return true
}

func (r *refSession) choose(branch string) bool {
	node, ok := r.nodes[r.current]
	if !ok {
		return false
	}
	next, ok := node.Goto[branch]
	if !ok {
		return false
	}
	return r.enter(next)
}

func (r *refSession) back() bool {
	if len(r.stack) == 0 {
		return false
	}
	last := len(r.stack) - 1
	r.current = r.stack[last]
	r.stack = r.stack[:last]
	return true
}

func randomTree(rng *rand.Rand) ([]Node, string) {
	count := 2 + rng.Intn(8)
	nodes := make([]Node, 0, count)
	for index := 0; index < count; index++ {
		node := Node{ID: fmt.Sprintf("n%d", index), Reward: rng.Intn(10)}
		switch rng.Intn(3) {
		case 0:
			node.Require = map[string]bool{fmt.Sprintf("f%d", rng.Intn(3)): true}
		case 1:
			node.Require = map[string]bool{fmt.Sprintf("f%d", rng.Intn(3)): false}
		}
		if rng.Intn(2) == 0 {
			node.Goto = map[string]string{}
			for branch := 0; branch < 3; branch++ {
				if rng.Intn(2) == 0 {
					node.Goto[fmt.Sprintf("b%d", branch)] = fmt.Sprintf("n%d", rng.Intn(count))
				}
			}
		}
		nodes = append(nodes, node)
	}
	return nodes, "n0"
}

func TestRandomizedMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(20260930))
	for trial := 0; trial < 200; trial++ {
		nodes, start := randomTree(rng)
		actual := NewSession(nodes, start)
		reference := newRefSession(nodes, start)
		steps := 1 + rng.Intn(50)
		for step := 0; step < steps; step++ {
			context := fmt.Sprintf("trial=%d step=%d", trial, step)
			switch rng.Intn(4) {
			case 0:
				name := fmt.Sprintf("f%d", rng.Intn(3))
				value := rng.Intn(2) == 0
				actual.SetFlag(name, value)
				reference.setFlag(name, value)
			case 1:
				id := fmt.Sprintf("n%d", rng.Intn(len(nodes)+1))
				if got, want := actual.Enter(id), reference.enter(id); got != want {
					t.Fatalf("%s Enter(%s)=%v，参考为 %v", context, id, got, want)
				}
			case 2:
				branch := fmt.Sprintf("b%d", rng.Intn(4))
				if got, want := actual.Choose(branch), reference.choose(branch); got != want {
					t.Fatalf("%s Choose(%s)=%v，参考为 %v", context, branch, got, want)
				}
			case 3:
				if got, want := actual.Back(), reference.back(); got != want {
					t.Fatalf("%s Back()=%v，参考为 %v", context, got, want)
				}
			}
			if actual.Current() != reference.current {
				t.Fatalf("%s 当前节点 %s，参考为 %s", context, actual.Current(), reference.current)
			}
			if actual.Reward() != reference.reward {
				t.Fatalf("%s 奖励 %d，参考为 %d", context, actual.Reward(), reference.reward)
			}
		}
	}
}
