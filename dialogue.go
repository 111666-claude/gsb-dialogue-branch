// Package dialogue 是对话树：按前置标记选节点，会话可回退，重复进入不重复发奖励。
// 缺陷：不校验前置标记、重复进入会重复发奖励、选分支全扫所有节点、回退栈永远是空的。
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

// Enter 进入一个节点。
// 缺陷：完全不校验 Require，而且每次进入都重复发奖励。
func (s *Session) Enter(nodeID string) bool {
	node, ok := s.nodes[nodeID]
	if !ok {
		return false
	}
	if s.current != nodeID {
		s.stack = append(s.stack, s.current)
	}
	s.current = nodeID
	s.entered[nodeID] = true
	s.reward += node.Reward
	return true
}

// Choose 按分支名选下一个节点。
// 缺陷：遍历所有节点找 Goto（而不是只看当前节点），代价随节点数增长。
func (s *Session) Choose(branch string) bool {
	target := ""
	for _, node := range s.nodes {
		s.scanned++
		if next, ok := node.Goto[branch]; ok && node.ID == s.current {
			target = next
		}
	}
	if target == "" {
		return false
	}
	return s.Enter(target)
}

// Back 回到上一个节点。缺陷：栈从来没人写过，永远失败。
func (s *Session) Back() bool {
	return false
}

// Current 是当前节点 ID。
func (s *Session) Current() string { return s.current }

// Reward 是累计奖励。
func (s *Session) Reward() int { return s.reward }

// Scanned 是累计扫描节点数（规模观测）。
func (s *Session) Scanned() int { return s.scanned }
