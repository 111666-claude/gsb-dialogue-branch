// Package dialogue 是对话树：按前置标记选节点，会话可回退，重复进入不重复发奖励。
package dialogue

// Node 是一个对话节点。
type Node struct {
	ID      string
	Require map[string]bool
	Reward  int
	Goto    map[string]string
}

// Session 是一次对话会话。
type Session struct {
	nodes   map[string]Node
	current string
	stack   []string
	flags   map[string]bool
	reward  int
	entered map[string]bool
	scanned int
}

// NewSession 用节点表建会话，起点是 start。
func NewSession(nodes []Node, start string) *Session {
	table := make(map[string]Node, len(nodes))
	for _, node := range nodes {
		table[node.ID] = node
	}
	return &Session{
		nodes:   table,
		current: start,
		flags:   map[string]bool{},
		entered: map[string]bool{},
	}
}

// SetFlag 设置一个剧情标记。
func (s *Session) SetFlag(name string, value bool) { s.flags[name] = value }

// Enter 进入一个节点：Require 中所有标记都为真才能进入；
// 条件不满足或节点不存在时返回 false，且不改当前节点、不压栈、不发奖励。
// 同一节点只有第一次进入时加 Reward。
func (s *Session) Enter(nodeID string) bool {
	node, ok := s.nodes[nodeID]
	if !ok || !s.satisfies(node.Require) {
		return false
	}
	if s.current != nodeID {
		s.stack = append(s.stack, s.current)
	}
	s.current = nodeID
	if !s.entered[nodeID] {
		s.entered[nodeID] = true
		s.reward += node.Reward
	}
	return true
}

// satisfies 判断当前标记是否满足节点的全部前置条件，代价只与 Require 条目数有关。
func (s *Session) satisfies(require map[string]bool) bool {
	for flag, want := range require {
		if s.flags[flag] != want {
			return false
		}
	}
	return true
}

// Choose 按分支名选下一个节点：只查当前节点的 Goto，O(1) 定位，
// Scanned 每次成功定位至多记 1，与节点总数无关。目标节点仍需通过 Enter 的前置校验。
func (s *Session) Choose(branch string) bool {
	node, ok := s.nodes[s.current]
	if !ok {
		return false
	}
	target, ok := node.Goto[branch]
	if !ok {
		return false
	}
	s.scanned++
	return s.Enter(target)
}

// Back 弹出栈顶节点成为当前节点；栈为空时返回 false。
func (s *Session) Back() bool {
	n := len(s.stack)
	if n == 0 {
		return false
	}
	s.current = s.stack[n-1]
	s.stack = s.stack[:n-1]
	return true
}

// Current 是当前节点 ID。
func (s *Session) Current() string { return s.current }

// Reward 是累计奖励。
func (s *Session) Reward() int { return s.reward }

// Scanned 是累计扫描节点数（规模观测）。
func (s *Session) Scanned() int { return s.scanned }
