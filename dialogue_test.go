package dialogue

import "testing"

func session() *Session {
	return NewSession([]Node{
		{ID: "a", Reward: 10, Goto: map[string]string{"ask": "b"}},
		{ID: "b", Reward: 5, Require: map[string]bool{"vip": true}},
	}, "a")
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
