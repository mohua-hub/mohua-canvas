import { apiDelete, apiGet, apiPost } from "@/services/api/request";

export async function fetchImageGenerationLogs<T>() {
    return apiGet<T[]>("/api/v1/generation-logs/images", undefined);
}

export async function saveImageGenerationLogs<T>(logs: T[]) {
    return apiPost<T[]>("/api/v1/generation-logs/images", { logs });
}

export async function deleteImageGenerationLogs(ids: string[]) {
    return apiPost<{ deleted: boolean }>("/api/v1/generation-logs/images/delete", { ids });
}


