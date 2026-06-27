package service

import (
	"context"
	"sort"
)

// 公开服务状态聚合层：把内部的 per-channel 监控视图聚合成按 provider → model 的
// 视图，供无鉴权的 /status/public 端点使用。
// 永远不输出渠道名(Name)等内部供给信息；分组名(GroupName)会作为标签随模型暴露，
// 方便用户在公开状态页看到自己所属的分组。

// Public-facing status values. Intentionally a small, stable set decoupled from
// the internal MonitorStatus* enum so the public contract never leaks finer
// internal states (failed/error are both surfaced as "down").
const (
	PublicStatusValueOperational = "operational"
	PublicStatusValueDegraded    = "degraded"
	PublicStatusValueDown        = "down"
)

// PublicModelStatus 单个模型的公开可用状态。
type PublicModelStatus struct {
	Model           string
	Status          string
	Availability7d  float64
	HasAvailability bool
	Groups          []string // 该模型所属的分组名集合（已去重排序），用于公开页展示分组标签
}

// PublicProviderStatus 单个 provider（平台）的公开可用状态，含其下模型。
type PublicProviderStatus struct {
	Provider        string
	Status          string
	Availability7d  float64
	HasAvailability bool
	Models          []PublicModelStatus
}

// PublicServiceStatus 公开服务状态总览，按 provider 分组。
type PublicServiceStatus struct {
	OverallStatus string
	Providers     []PublicProviderStatus
}

// providerDisplayOrder 控制公开页 provider 的展示顺序；未列出的排在后面按字母序。
// 用监控侧的 provider 字符串常量（与 ChannelMonitor.Provider 取值一致）。
var providerDisplayOrder = map[string]int{
	MonitorProviderAnthropic: 0,
	MonitorProviderOpenAI:    1,
	MonitorProviderGemini:    2,
}

// modelStatusAccumulator 聚合同一 (provider, model) 跨多个渠道的状态与可用率。
type modelStatusAccumulator struct {
	operational int                 // 状态为 operational 的渠道数
	known       int                 // 有明确状态（非空）的渠道数
	availSum    float64             // 可用率之和（仅统计有可用率的渠道）
	availCount  int                 // 有可用率的渠道数
	groups      map[string]struct{} // 该模型出现过的分组名集合（去重）
}

// PublicStatusView 构造无鉴权公开服务状态。
// 复用 ListUserView（已是批量查询，无 N+1），在内存里按 provider → model 聚合，
// 剥离所有内部命名。失败时向上传递错误，由 handler 决定如何降级。
func (s *ChannelMonitorService) PublicStatusView(ctx context.Context) (*PublicServiceStatus, error) {
	views, err := s.ListUserView(ctx)
	if err != nil {
		return nil, err
	}

	// provider -> model -> 累积器
	byProvider := make(map[string]map[string]*modelStatusAccumulator)
	accFor := func(provider, model string) *modelStatusAccumulator {
		models, ok := byProvider[provider]
		if !ok {
			models = make(map[string]*modelStatusAccumulator)
			byProvider[provider] = models
		}
		acc, ok := models[model]
		if !ok {
			acc = &modelStatusAccumulator{}
			models[model] = acc
		}
		return acc
	}

	for _, v := range views {
		if v.Provider == "" {
			continue
		}
		// 主模型：带状态 + 7d 可用率。
		if v.PrimaryModel != "" {
			acc := accFor(v.Provider, v.PrimaryModel)
			acc.addStatus(v.PrimaryStatus)
			acc.addGroup(v.GroupName)
			acc.availSum += v.Availability7d
			acc.availCount++
		}
		// 额外模型：仅有最新状态，无 7d 可用率。
		for _, e := range v.ExtraModels {
			if e.Model == "" {
				continue
			}
			acc := accFor(v.Provider, e.Model)
			acc.addStatus(e.Status)
			acc.addGroup(v.GroupName)
		}
	}

	providers := make([]PublicProviderStatus, 0, len(byProvider))
	for provider, models := range byProvider {
		ps := buildProviderStatus(provider, models)
		if len(ps.Models) == 0 {
			continue // 该 provider 下所有模型都没有可用状态数据，跳过
		}
		providers = append(providers, ps)
	}

	sortProviders(providers)

	return &PublicServiceStatus{
		OverallStatus: aggregateOverallStatus(providers),
		Providers:     providers,
	}, nil
}

// addStatus 累加一个渠道对某模型的最新状态。空状态视为"无数据"，不计入 known。
func (a *modelStatusAccumulator) addStatus(status string) {
	if status == "" {
		return
	}
	a.known++
	if status == MonitorStatusOperational {
		a.operational++
	}
}

// addGroup 记录该模型出现过的分组名（去重）。空分组名忽略。
func (a *modelStatusAccumulator) addGroup(group string) {
	if group == "" {
		return
	}
	if a.groups == nil {
		a.groups = make(map[string]struct{})
	}
	a.groups[group] = struct{}{}
}

// sortedGroups 把分组名集合转成去重、按字母序排序的切片。空集合返回 nil。
func sortedGroups(set map[string]struct{}) []string {
	if len(set) == 0 {
		return nil
	}
	groups := make([]string, 0, len(set))
	for g := range set {
		groups = append(groups, g)
	}
	sort.Strings(groups)
	return groups
}

// buildProviderStatus 由该 provider 下各模型的累积器构造 PublicProviderStatus。
func buildProviderStatus(provider string, models map[string]*modelStatusAccumulator) PublicProviderStatus {
	modelStatuses := make([]PublicModelStatus, 0, len(models))
	var availSum float64
	var availCount int
	var opModels, downModels, knownModels int

	for model, acc := range models {
		if acc.known == 0 {
			continue // 无任何明确状态，跳过这个模型
		}
		status := aggregateStatus(acc.operational, acc.known)
		ms := PublicModelStatus{Model: model, Status: status, Groups: sortedGroups(acc.groups)}
		if acc.availCount > 0 {
			ms.Availability7d = acc.availSum / float64(acc.availCount)
			ms.HasAvailability = true
			availSum += ms.Availability7d
			availCount++
		}
		modelStatuses = append(modelStatuses, ms)

		knownModels++
		switch status {
		case PublicStatusValueOperational:
			opModels++
		case PublicStatusValueDown:
			downModels++
		}
	}

	sort.Slice(modelStatuses, func(i, j int) bool {
		return modelStatuses[i].Model < modelStatuses[j].Model
	})

	ps := PublicProviderStatus{
		Provider: provider,
		Status:   rollupStatus(knownModels, opModels, downModels),
		Models:   modelStatuses,
	}
	if availCount > 0 {
		ps.Availability7d = availSum / float64(availCount)
		ps.HasAvailability = true
	}
	return ps
}

// aggregateStatus 由"操作正常数 / 已知数"映射为模型级状态。
func aggregateStatus(operational, known int) string {
	switch {
	case known == 0:
		return PublicStatusValueDown
	case operational == known:
		return PublicStatusValueOperational
	case operational == 0:
		return PublicStatusValueDown
	default:
		return PublicStatusValueDegraded
	}
}

// rollupStatus 把一组子项（模型/provider）的统计汇总成上一级状态：
// 全部正常 → operational；全部宕机 → down；其余（含部分降级）→ degraded。
func rollupStatus(known, operational, down int) string {
	switch {
	case known == 0:
		return PublicStatusValueDown
	case operational == known:
		return PublicStatusValueOperational
	case down == known:
		return PublicStatusValueDown
	default:
		return PublicStatusValueDegraded
	}
}

// aggregateOverallStatus 汇总所有 provider 的总体状态。
func aggregateOverallStatus(providers []PublicProviderStatus) string {
	var known, op, down int
	for _, p := range providers {
		known++
		switch p.Status {
		case PublicStatusValueOperational:
			op++
		case PublicStatusValueDown:
			down++
		}
	}
	return rollupStatus(known, op, down)
}

// sortProviders 按 providerDisplayOrder 排序，未列出的排在后面按字母序。
func sortProviders(providers []PublicProviderStatus) {
	sort.Slice(providers, func(i, j int) bool {
		oi, iok := providerDisplayOrder[providers[i].Provider]
		oj, jok := providerDisplayOrder[providers[j].Provider]
		if iok && jok {
			return oi < oj
		}
		if iok != jok {
			return iok // 已列出的排在前
		}
		return providers[i].Provider < providers[j].Provider
	})
}
