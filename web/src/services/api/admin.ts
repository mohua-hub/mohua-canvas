import type { ModelChannelProtocol } from "@/lib/model-channel";
import type { WorkflowEntry, WorkflowSummary } from "@/lib/workflow-channel";
import { apiDelete, apiGet, apiPost, compactApiParams } from "@/services/api/request";
import type { Prompt, PromptListResponse } from "@/services/api/prompts";
import type { AgentSkill, AgentSkillFile } from "@/services/api/agent-skills";

export type AdminPromptCategory = {
    category: string;
    name: string;
    description: string;
    file: string;
    githubUrl: string;
    remote: boolean;
};

export type AdminListQuery = { keyword?: string; page?: number; pageSize?: number };

export async function fetchAdminPromptCategories() {
    return apiGet<AdminPromptCategory[]>("/api/settings/prompt-categories", undefined);
}

export async function syncAdminPromptCategory(category: string) {
    return apiPost<AdminPromptCategory[]>("/api/settings/prompt-categories/sync", { category });
}

export async function syncAdminPromptCategoriesAll() {
    return apiPost<AdminPromptCategory[]>("/api/settings/prompt-categories/sync-all", {});
}

export type AdminPromptQuery = {
    keyword?: string;
    category?: string;
    tag?: string[];
    page?: number;
    pageSize?: number;
};

export type AdminAsset = {
    id: string;
    title: string;
    type: "text" | "image" | "video" | "audio";
    coverUrl: string;
    tags: string[];
    category: string;
    description: string;
    content: string;
    url: string;
    createdAt: string;
    updatedAt: string;
};

export type AdminAssetListResponse = {
    items: AdminAsset[];
    tags: string[];
    total: number;
};

export async function fetchAdminPrompts(query: AdminPromptQuery = {}) {
    return apiGet<PromptListResponse>("/api/settings/prompts", compactApiParams(query));
}

export async function saveAdminPrompt(prompt: Partial<Prompt>) {
    return apiPost<Prompt>("/api/settings/prompts", prompt);
}

export async function deleteAdminPrompt(id: string) {
    return apiDelete<boolean>(`/api/settings/prompts/${encodeURIComponent(id)}`);
}

export async function deleteAdminPrompts(ids: string[]) {
    return apiPost<boolean>("/api/settings/prompts/batch-delete", { ids });
}

export function fetchAdminAgentSkills() {
    return apiGet<AgentSkill[]>("/api/settings/agent-skills", undefined);
}

export function saveAdminAgentSkill(skill: Partial<AgentSkill>) {
    return apiPost<AgentSkill>("/api/settings/agent-skills", skill);
}

export function fetchAdminAgentSkillFiles(id: string) {
    return apiGet<AgentSkillFile[]>(`/api/settings/agent-skills/${encodeURIComponent(id)}/files`, undefined);
}

export function deleteAdminAgentSkill(id: string) {
    return apiDelete<boolean>(`/api/settings/agent-skills/${encodeURIComponent(id)}`);
}

export type AdminAssetQuery = {
    keyword?: string;
    type?: string;
    tag?: string[];
    page?: number;
    pageSize?: number;
};

export async function fetchAdminAssets(query: AdminAssetQuery = {}) {
    return apiGet<AdminAssetListResponse>("/api/settings/assets", compactApiParams(query));
}

export async function saveAdminAsset(asset: Partial<AdminAsset>) {
    return apiPost<AdminAsset>("/api/settings/assets", asset);
}

export async function deleteAdminAsset(id: string) {
    return apiDelete<boolean>(`/api/settings/assets/${encodeURIComponent(id)}`);
}

export type AdminModelChannel = {
    id: string;
    protocol: ModelChannelProtocol;
    name: string;
    baseUrl: string;
    apiKey: string;
    models: string[];
    weight: number;
    timeout: number;
    enabled: boolean;
    remark: string;
    bridgeId?: string;
    comfyUrl?: string;
    workflowDir?: string;
    workflows?: WorkflowEntry[];
};

export type AdminPublicModelChannelSettings = {
    availableModels: string[];
    availableWorkflows: string[];
    channels: AdminPublicModelChannelInfo[];
    defaultModel: string;
    defaultImageModel: string;
    defaultVideoModel: string;
    defaultTextModel: string;
    systemPrompt: string;
    systemPrompts: {
        image: string;
        video: string;
        text: string;
        workflow: string;
        workflowAgent: string;
    };
    allowCustomChannel: boolean;
};

export type AdminPublicModelChannelInfo = {
    id: string;
    protocol: AdminModelChannel["protocol"];
    name: string;
    baseUrl: string;
    models: string[];
    weight: number;
    timeout: number;
    enabled: boolean;
    remark: string;
    workflows?: WorkflowSummary[];
};

export type AdminPublicSettings = {
    modelChannel: AdminPublicModelChannelSettings;
    storage: {
        mode: string;
        allowCustomProvider: boolean;
    };
};

export type AdminStorageProvider = {
    id: string;
    name: string;
    type: "s3" | "webdav";
    endpoint: string;
    region: string;
    bucket: string;
    accessKeyId: string;
    secretAccessKey: string;
    publicBaseUrl: string;
    pathPrefix: string;
    username: string;
    password: string;
    weight: number;
    enabled: boolean;
    ownerWorkspaceId: string;
    capacityBytes: number;
    capacityCheckedAt: string;
    capacityExceeded: boolean;
};

export type AdminPrivateSettings = {
    channels: AdminModelChannel[];
    promptSync: {
        enabled: boolean;
        cron: string;
    };
    aiLog: {
        localDirectReportEnabled: boolean;
        cleanup: {
            enabled: boolean;
            retentionDays: number;
            cron: string;
        };
    };
    storage: {
        mode: string;
        allowCustomProvider: boolean;
        useGlobalProvider: boolean;
        autoSyncAllAssets: boolean;
        providers: AdminStorageProvider[];
        roundRobinCursor: number;
        capacityCheck: {
            enabled: boolean;
            cron: string;
        };
        capacityLimitBytes: number;
    };
};

export type AdminAICallLog = {
    id: string;
    workspaceId: string;
    endpoint: string;
    method: string;
    model: string;
    channelId: string;
    channelName: string;
    status: number;
    durationMs: number;
    requestBody: string;
    responseBody: string;
    error: string;
    createdAt: string;
};

export type AdminAICallLogListResponse = {
    items: AdminAICallLog[];
    total: number;
};

export async function fetchAdminAICallLogs(query: AdminListQuery = {}) {
    return apiGet<AdminAICallLogListResponse>("/api/settings/ai-logs", compactApiParams(query));
}

export async function deleteAdminAICallLogs(olderThanDays = 7) {
    return apiDelete<{ removedFiles: number }>(`/api/settings/ai-logs?olderThanDays=${encodeURIComponent(String(olderThanDays))}`);
}

export type AdminSettings = {
    public: AdminPublicSettings;
    private: AdminPrivateSettings;
};

export async function fetchAdminSettings() {
    return apiGet<AdminSettings>("/api/settings/config", undefined);
}

export async function saveAdminSettings(settings: AdminSettings) {
    return apiPost<AdminSettings>("/api/settings/config", settings);
}

export type AdminChannelActionRequest = {
    index?: number;
    channel: AdminModelChannel;
    model?: string;
};

export async function fetchChannelModels(payload: AdminChannelActionRequest) {
    return apiPost<string[]>("/api/settings/config/channel-models", payload);
}

export async function testChannelModel(payload: AdminChannelActionRequest) {
    return apiPost<string>("/api/settings/config/channel-test", payload);
}

export type StorageCapacityResult = {
    bytes: number;
    limitBytes: number;
    overLimit: boolean;
    checkedAt: string;
    providerName: string;
};

export async function measureAdminStorageProvider(payload: { index: number; provider: AdminStorageProvider }) {
    return apiPost<StorageCapacityResult>("/api/settings/storage/measure", payload);
}
