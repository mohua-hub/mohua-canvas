import { apiPost } from "@/services/api/request";
import { resolveMediaUrl, uploadRemoteMediaToServer } from "@/services/file-storage";
import { resolveImageUrl } from "@/services/image-storage";
import { getStorageObjectInfo } from "@/services/api/storage";
import type { ReferenceImage } from "@/types/image";
import type { ReferenceAudio, ReferenceVideo } from "@/types/media";

type AutoDLReference = ReferenceImage | ReferenceVideo | ReferenceAudio;

export async function autoDLReferenceURL(reference: AutoDLReference) {
    for (const value of [reference.url, "dataUrl" in reference ? reference.dataUrl : ""]) {
        const url = publicReferenceURL(value);
        if (url) return url;
    }
    const storedUrl = await storedReferenceURL(reference.storageKey);
    if (storedUrl) return storedUrl;
    if (reference.storageKey?.startsWith("server:")) throw new Error("AutoDL 参考素材需要云存储提供可公开访问的地址");
    const source = "dataUrl" in reference
        ? await resolveImageUrl(reference.storageKey, reference.dataUrl || reference.url || "")
        : await resolveMediaUrl(reference.storageKey, reference.url);
    if (!source) throw new Error("参考素材不可用");
    const uploaded = await uploadRemoteMediaToServer(source, reference.name || "reference");
    const url = publicReferenceURL(uploaded.url) || await storedReferenceURL(uploaded.storageKey);
    if (!url) throw new Error("AutoDL 参考素材需要云存储提供可公开访问的地址");
    return url;
}

async function storedReferenceURL(storageKey?: string) {
    if (!storageKey?.startsWith("server:") || storageKey.startsWith("server:webdav:")) return "";
    const info = await getStorageObjectInfo(storageKey.slice("server:".length));
    return publicReferenceURL(info.publicUrl);
}

function publicReferenceURL(value?: string) {
    if (!value || !/^https?:\/\//i.test(value)) return "";
    try {
        const url = new URL(value);
        return ["localhost", "127.0.0.1", "[::1]"].includes(url.hostname) ? "" : url.href;
    } catch {
        return "";
    }
}

export type AutoDLInputRule = {
    type: string;
    required?: boolean;
    default?: string | number;
    min?: number;
    max?: number;
};

export type AutoDLWorkflow = {
    uuid: string;
    name: string;
    kind: "video" | "audio" | "unsupported";
    input_rules?: Record<string, AutoDLInputRule>;
};

export function fetchAutoDLWorkflows(baseUrl: string) {
    return apiPost<AutoDLWorkflow[]>("/api/ai/autodl/workflows", { baseUrl });
}

export function fetchAutoDLWorkflow(baseUrl: string, workflowId: string) {
    return apiPost<AutoDLWorkflow>("/api/ai/autodl/workflows", { baseUrl, workflowId });
}
