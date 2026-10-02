import { apiDelete, apiGet, apiPost } from "@/services/api/request";
import type { CustomWebDAVStorageProvider } from "@/services/image-storage";

export type RegisteredStorageObject = {
    url: string;
    storageKey: string;
    bytes: number;
    mimeType: string;
};

export type StorageObjectInfo = {
    id: string;
    objectKey: string;
    publicUrl: string;
    mimeType: string;
    bytes: number;
    direct: boolean;
};

export function getStorageObjectInfo(id: string) {
    return apiGet<StorageObjectInfo>(`/api/files/${encodeURIComponent(id)}`);
}

export function registerDirectStorageObject(
    payload: { provider: CustomWebDAVStorageProvider; objectKey: string; mimeType: string; bytes: number },
) {
    return apiPost<RegisteredStorageObject>("/api/v1/files/direct", payload);
}

export function deleteDirectStorageObjectRecord(id: string) {
    return apiDelete<boolean>(`/api/v1/files/${encodeURIComponent(id)}/record`);
}
