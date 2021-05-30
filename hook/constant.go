package hook

const (
	branchExist    = "当前 PR 合并后，将创建同步 PR"
	branchNonExist = "目标分支不存在，忽略处理"
	createdPR      = "创建同步 PR"
	syncFailed     = "同步失败：请手动创建 PR 进行同步，我们会继续完善分支之间同步操作，尽量避免同步失败的情况"
)
