import type { WorkflowCapability, WorkflowEntry, WorkflowKind, WorkflowRef, WorkflowSummary } from "@/lib/workflow-channel";
import type { AdminModelChannel } from "@/services/api/admin";
import { apiGet, apiPost } from "@/services/api/request";
import { workflowRequest, type WorkflowGenerationTask, type WorkflowRunInput } from "@/services/api/workflow-generation";

export type RunningHubCollectionItem = { ref: WorkflowRef; channelName: string; entry: WorkflowSummary };

export function getRunningHubCollection(scope: WorkflowRef["scope"] | "all") {
    return apiGet<RunningHubCollectionItem[]>("/api/v1/runninghub/collection", { scope });
}

export function getRunningHubEntry(ref: WorkflowRef) {
    return apiPost<WorkflowEntry>("/api/v1/runninghub/entry", ref);
}

export function submitRunningHubTask(input: WorkflowRunInput) {
    return workflowRequest<WorkflowGenerationTask>("/api/v1/runninghub/tasks", undefined, input);
}

export function getRunningHubTask(id: string) {
    return workflowRequest<WorkflowGenerationTask>(`/api/v1/runninghub/tasks/${encodeURIComponent(id)}`);
}

export type RunningHubInspectInput = {
    baseUrl: string;
    apiKey: string;
    kind: WorkflowKind;
    workflowId: string;
    title: string;
    capability: WorkflowCapability;
};

export function inspectRunningHub(input: RunningHubInspectInput, admin?: { index?: number; channel: AdminModelChannel }) {
    return admin
        ? apiPost<WorkflowEntry>("/api/settings/workflow-providers/runninghub/inspect", { ...admin, input })
        : apiPost<WorkflowEntry>("/api/v1/workflow-providers/runninghub/inspect", input);
}
