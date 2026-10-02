import type { WorkflowEntry, WorkflowFieldMapping, WorkflowRef } from "@/lib/workflow-channel";
import type { CanvasNodeMetadata } from "../types";
import type { ReferenceImage } from "@/types/image";
import type { ReferenceAudio, ReferenceVideo } from "@/types/media";
import { workflowMediaSource } from "@/services/api/workflow-generation";
import { normalizeLocalChannels, type AiConfig } from "@/stores/use-config-store";

export function canvasDefaultWorkflowRef(config: AiConfig, ref?: WorkflowRef) {
    const channels = ref?.scope === "system" ? config.publicChannels : normalizeLocalChannels(config);
    return channels.find((channel) => channel.id === ref?.channelId)?.protocol === "comfyui" ? ref : undefined;
}

export const runningHubFieldKey = (field: WorkflowFieldMapping) => `field:${field.nodeId}:${field.fieldName}`;

export function runningHubMediaKind(field: WorkflowFieldMapping): "image" | "video" | "audio" | undefined {
    if (field.source === "referenceImage" || field.source === "mask") return "image";
    if (field.source === "referenceVideo") return "video";
    if (field.source === "referenceAudio") return "audio";
    if (field.sourceFromUpstream && ["IMAGE", "VIDEO", "AUDIO"].includes(String(field.fieldType).toUpperCase())) return String(field.fieldType).toLowerCase() as "image" | "video" | "audio";
}

export function runningHubMediaIndex(field: WorkflowFieldMapping) {
    return field.source === "referenceImage" && field.imageOrder ? field.imageOrder - 1 : field.sourceIndex || 0;
}

export function runningHubInitialValues(entry: WorkflowEntry) {
    return Object.fromEntries(entry.fields.filter((field) => !runningHubMediaKind(field) && !field.randomEnabled && field.fieldValue !== undefined).map((field) => [runningHubFieldKey(field), field.fieldValue]));
}

export async function runningHubCanvasInputs(metadata: CanvasNodeMetadata, images: ReferenceImage[], videos: ReferenceVideo[], audios: ReferenceAudio[]) {
    const referenceImages: string[] = [];
    const referenceVideos: string[] = [];
    const referenceAudios: string[] = [];
    let mask = "";
    const sources = { image: images.map((item) => ({ id: item.id, url: item.dataUrl })), video: videos.map((item) => ({ id: item.id, url: item.url })), audio: audios.map((item) => ({ id: item.id, url: item.url })) };
    const cache = new Map<string, Promise<string>>();
    await Promise.all((metadata.runningHubEntry?.fields || []).map(async (field) => {
        const kind = runningHubMediaKind(field);
        if (!kind) return;
        const key = runningHubFieldKey(field);
        const selectedId = metadata.runningHubMedia?.[key];
        const index = runningHubMediaIndex(field);
        const selected = selectedId === undefined ? sources[kind][index] : sources[kind].find((item) => item.id === selectedId);
        if (!selected && field.required) throw new Error(`请选择“${field.label || field.fieldName}”的参考素材`);
        if (!selected) return;
        if (!cache.has(selected.id)) cache.set(selected.id, workflowMediaSource(selected.url));
        const source = await cache.get(selected.id)!;
        if (field.source === "mask") mask = source;
        else ({ image: referenceImages, video: referenceVideos, audio: referenceAudios })[kind][index] = source;
    }));
    return { referenceImages: Array.from(referenceImages, (value) => value || ""), referenceVideos: Array.from(referenceVideos, (value) => value || ""), referenceAudios: Array.from(referenceAudios, (value) => value || ""), mask, fieldValues: metadata.runningHubFieldValues };
}
