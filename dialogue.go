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
	table := map[string]Node{}
	for _, node := range nodes {
		table[node.ID] = node
	}
	return &Session{nodes: table, current: start, flags: map[string]bool{}, entered: map[string]bool{}}
}

// SetFlag 设置一个剧情标记。
func (s *Session) SetFlag(name string, value bool) { s.flags[name] = value }

// allowed 检查节点的全部前置标记是否满足。
func (s *Session) allowed(node Node) bool {
	for name, want := range node.Require {
		if s.flags[name] != want {
			return false
		}
	}
	return true
}

// Enter 进入一个节点：Require 里所有标记满足才进入，
// 否则返回 false 且不改当前节点、不压栈、不发奖励；只有第一次进入才发奖励。
func (s *Session) Enter(nodeID string) bool {
	node, ok := s.nodes[nodeID]
	if !ok || !s.allowed(node) {
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

// Choose 按分支名选下一个节点。
// 只看当前节点的 Goto，代价 O(1)，与节点总数无关。
func (s *Session) Choose(branch string) bool {
	node, ok := s.nodes[s.current]
	if !ok {
		return false
	}
	s.scanned++
	next, ok := node.Goto[branch]
	if !ok {
		return false
	}
	return s.Enter(next)
}

// Back 弹出上一个节点成为当前节点，栈空时返回 false。
func (s *Session) Back() bool {
	if len(s.stack) == 0 {
		return false
	}
	last := len(s.stack) - 1
	s.current = s.stack[last]
	s.stack = s.stack[:last]
	return true
}

// Current 是当前节点 ID。
func (s *Session) Current() string { return s.current }

// Reward 是累计奖励。
func (s *Session) Reward() int { return s.reward }

// Scanned 是累计扫描节点数（规模观测）。
func (s *Session) Scanned() int { return s.scanned }
