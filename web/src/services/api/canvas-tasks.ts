import type { CanvasProject } from "@/app/(user)/canvas/stores/use-canvas-store";
import { apiGet, apiPost } from "@/services/api/request";
import { useBackendStore } from "@/stores/use-backend-store";

export async function listCanvasProjects() {
    return apiGet<CanvasProject[]>("/api/v1/canvas/projects", undefined);
}

export async function saveCanvasProject(
    project: CanvasProject,
) {
    return apiPost<CanvasProject>(
        "/api/v1/canvas/projects",
        { data: project },
    );
}

export async function syncCanvasProjects(
    projects: CanvasProject[],
) {
    return apiPost<CanvasProject[]>(
        "/api/v1/canvas/projects/sync",
        { projects },
    );
}

export async function deleteCanvasTasks(sourceId: string, nodeIds: string[] = []) {
    const backendConnected = useBackendStore.getState().available;
    const source = sourceId.trim();
    if (!backendConnected || !source) return;
    return apiPost<{ deleted: boolean }>(
        "/api/v1/canvas/tasks/delete",
        {
            source_id: source,
            node_ids: Array.from(new Set(nodeIds.map((id) => id.trim()).filter(Boolean))),
        },
    );
}

export async function deleteCanvasProjects(ids: string[]) {
    const backendConnected = useBackendStore.getState().available;
    const projectIds = Array.from(
        new Set(ids.map((id) => id.trim()).filter(Boolean)),
    );
    if (!backendConnected || !projectIds.length) return;
    return apiPost<{ deleted: boolean }>(
        "/api/v1/canvas/projects/delete",
        { ids: projectIds },
    );
}
