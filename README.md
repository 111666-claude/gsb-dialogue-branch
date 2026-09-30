# dialogue-branch

对话树：按前置标记选节点，会话可回退，重复进入不重复发奖励，只用 Go 标准库。

```
go test ./...
go run ./cmd/dialogue-branch --sample require
go run ./cmd/dialogue-branch --sample reward
go run ./cmd/dialogue-branch --sample back
```

## 口径（README 为准）

- **前置标记**：`Require` 里所有标记都为真才能进入；条件不满足时返回 `false` 且不改当前节点、
  不压栈、不发奖励。
- **奖励只发一次**：同一个节点第一次进入才加 `Reward`，重复进入不再加。
- **选分支只看当前节点**：`Choose` 用当前节点的 `Goto` 直接定位，代价 O(1)，
  `Scanned` 不随节点总数增长。
- **回退**：`Back` 弹出上一个节点成为当前节点，栈空时返回 `false`。
- **不变量**：`Reward` 等于所有首次进入节点的奖励之和；同一串操作重复执行结果相同。
- **规模**：单会话 3000 个节点、每秒 10 万次选择，单次 O(1)，内存 O(节点数)。

## 输出契约（不改格式）

```
entered=false current=a
reward=5
ok=true current=a
scanned<=3000
```

## 目录

```
dialogue.go             对话节点与会话
cmd/dialogue-branch     命令行入口
dialogue_test.go        go test 用例
```
