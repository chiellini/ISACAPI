# Git 同步与发布流程

本项目的 `main` 是 fork 自有的集成分支。`origin` 指向 ISAC AI 的 fork，`upstream` 指向上游仓库。所有需要同步主线的 Git 操作，统一按下面顺序执行：

## 标准流程

在仓库根目录执行：

```bash
# 1. 确认没有未处理的本地改动
git status --short --branch

# 2. 更新两个远端引用
git fetch origin
git fetch upstream

# 3. 切换到本地集成分支
git switch main

# 4. 先合并 fork 远端，再合并上游远端
git merge --no-edit origin/main
git merge --no-edit upstream/main

# 5. 解决冲突、检查并测试后提交
git diff --check
git status --short
git add <已确认的文件>
git commit -m "type(scope): description"

# 6. 只推送到 fork 的 main
git push origin main

# 7. 验证本地与 fork 远端一致
git status --short --branch
git log -1 --oneline --decorate
```

如果第 4 步显示 `Already up to date`，不要为了凑流程创建空提交。若合并已经自动生成 merge commit，直接检查、测试并推送即可。

## 冲突处理规则

- 不使用 `git reset --hard`、`git checkout --` 或强制推送来掩盖冲突。
- `README.md`、`README_CN.md`、`README_JA.md` 是 fork 自有文件；保留 `https://isacai.space` 和 `https://isacai.cn`，拒绝上游的赞助、广告和推广内容。
- 业务代码冲突要结合两边改动逐段合并，并保留本项目的支付、Codex 教程及其他 fork 定制功能。
- 本次同步使用了 `git merge --no-edit -X ours upstream/main` 作为冲突较多时的保守策略：它只在冲突块优先保留本地版本，仍会纳入 upstream 的非冲突提交。今后只有在确认冲突区域属于 fork 定制、且完成差异检查和测试后，才可使用该选项；不能把它当作无审查的自动解决方案。
- 合并完成后至少运行受影响模块的测试；涉及前端时运行 lint、类型检查和相关 Vitest；涉及后端时运行对应的 `go test`。

## 远端约定

```bash
git remote -v
# origin   = fork（推送目标）
# upstream = 上游（只用于获取和合并）
```

不要直接向 `upstream/main` 推送，也不要在 `main` 上使用 rebase 后强制推送。
