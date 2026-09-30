// Command dialogue-branch 跑对话树样例。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/dialogue"
)

func nodes() []dialogue.Node {
	return []dialogue.Node{
		{ID: "a", Reward: 10, Goto: map[string]string{"ask": "b"}},
		{ID: "b", Reward: 5, Require: map[string]bool{"vip": true}},
	}
}

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("dialogue-branch", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sample := flags.String("sample", "require", "require / reward / back / scan")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	switch *sample {
	case "require":
		// 没有 vip 标记，b 节点不该进得去。
		instance := dialogue.NewSession(nodes(), "a")
		entered := instance.Enter("b")
		fmt.Fprintf(stdout, "entered=%v current=%s\n", entered, instance.Current())
	case "reward":
		// 同一个节点进两次。
		instance := dialogue.NewSession(nodes(), "a")
		instance.Enter("b")
		instance.Enter("b")
		fmt.Fprintf(stdout, "reward=%d\n", instance.Reward())
	case "back":
		instance := dialogue.NewSession(nodes(), "a")
		instance.Enter("b")
		ok := instance.Back()
		fmt.Fprintf(stdout, "ok=%v current=%s\n", ok, instance.Current())
	case "scan":
		table := make([]dialogue.Node, 0, 3000)
		for index := 0; index < 3000; index++ {
			table = append(table, dialogue.Node{ID: fmt.Sprintf("n-%d", index)})
		}
		instance := dialogue.NewSession(table, "n-0")
		for query := 0; query < 3000; query++ {
			instance.Choose("ask")
		}
		fmt.Fprintf(stdout, "scanned=%d\n", instance.Scanned())
	default:
		fmt.Fprintln(stderr, "需要 --sample require|reward|back|scan")
		return 2
	}
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
