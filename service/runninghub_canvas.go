package service

import (
	"context"
	"encoding/json"
	"errors"
	"slices"

	"github.com/tigerowo/infinite-canvas/model"
	"github.com/tigerowo/infinite-canvas/repository"
)

type RunningHubCollectionItem struct {
	Ref         WorkflowRef           `json:"ref"`
	ChannelName string                `json:"channelName"`
	Entry       model.WorkflowSummary `json:"entry"`
}

func ListRunningHubCollection(workspaceID string, scope string) ([]RunningHubCollectionItem, error) {
	if scope == "all" {
		items, err := ListRunningHubCollection(workspaceID, "personal")
		if err != nil {
			return items, err
		}
		system, err := ListRunningHubCollection(workspaceID, "system")
		if err != nil {
			return nil, err
		}
		return append(items, system...), nil
	}
	var channels []model.ModelChannel
	var allowed []string
	switch scope {
	case "system":
		settings, err := repository.GetSettings()
		if err != nil {
			return nil, err
		}
		channels = settings.Private.Channels
		allowed = settings.Public.ModelChannel.AvailableWorkflows
	case "personal":
		settings, err := repository.GetSettings()
		if err != nil {
			return nil, err
		}
		personalWorkflows, err := personalWorkflowChannels(workspaceID)
		if err != nil {
			return nil, err
		}
		channels = settings.Private.Channels
		for i := range channels {
			channels[i].Workflows = nil
			for _, saved := range personalWorkflows {
				if saved.ChannelID == channels[i].ID && saved.Protocol == channels[i].Protocol {
					channels[i].Workflows = saved.Workflows
					break
				}
			}
		}
	default:
		return nil, errors.New("RunningHub 集合归属无效")
	}
	items := []RunningHubCollectionItem{}
	for _, channel := range channels {
		if channel.Protocol != "runninghub" {
			continue
		}
		for _, entry := range channel.Workflows {
			ref := WorkflowRef{Scope: scope, ChannelID: channel.ID, Kind: entry.Kind, WorkflowID: entry.WorkflowID}
			if entry.Provider != "runninghub" || !entry.Enabled || entry.Kind != "app" && entry.Kind != "workflow" || len(allowed) > 0 && !slices.Contains(allowed, workflowEntryName(ref)) {
				continue
			}
			items = append(items, RunningHubCollectionItem{Ref: ref, ChannelName: channel.Name, Entry: model.WorkflowSummary{Provider: entry.Provider, Kind: entry.Kind, WorkflowID: entry.WorkflowID, Title: entry.Title, Capability: entry.Capability, Enabled: true}})
		}
	}
	return items, nil
}

func RunningHubCanvasEntry(workspaceID string, ref WorkflowRef) (model.WorkflowEntry, error) {
	resolved, err := ResolveWorkflowForWorkspace(workspaceID, ref)
	if err != nil {
		return model.WorkflowEntry{}, err
	}
	if resolved.Channel.Protocol != "runninghub" {
		return model.WorkflowEntry{}, errors.New("请选择 RunningHub 应用或工作流")
	}
	entry := resolved.Entry
	entry.WorkflowJSON, entry.WorkflowGraph = nil, nil
	entry.Fields = []model.WorkflowFieldMapping{}
	for _, field := range resolved.Entry.Fields {
		if field.Enabled != nil && !*field.Enabled || field.SafeToOverride != nil && !*field.SafeToOverride {
			continue
		}
		entry.Fields = append(entry.Fields, field)
	}
	return entry, nil
}

func CreateRunningHubCanvasTask(ctx context.Context, workspaceID string, input WorkflowRunInput) (WorkflowTaskResult, error) {
	if _, err := RunningHubCanvasEntry(workspaceID, input.Ref); err != nil {
		return WorkflowTaskResult{}, err
	}
	return CreateWorkflowTask(ctx, workspaceID, input)
}

func GetRunningHubCanvasTask(ctx context.Context, workspaceID string, taskID string) (WorkflowTaskResult, error) {
	snapshot, found, err := workflowTaskSnapshot(workspaceID, taskID)
	if err != nil {
		return WorkflowTaskResult{}, err
	}
	if !found {
		return WorkflowTaskResult{}, ErrWorkflowTaskNotFound
	}
	var ref WorkflowRef
	if err := json.Unmarshal([]byte(snapshot.ref), &ref); err != nil {
		return WorkflowTaskResult{}, err
	}
	resolved, err := resolveWorkflowForWorkspace(workspaceID, ref, false)
	if err != nil {
		return WorkflowTaskResult{}, err
	}
	if resolved.Channel.Protocol != "runninghub" {
		return WorkflowTaskResult{}, errors.New("任务不属于 RunningHub")
	}
	return GetWorkflowTask(ctx, workspaceID, taskID)
}
