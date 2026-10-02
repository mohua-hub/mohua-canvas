"use client";

import { useEffect, type ReactNode } from "react";
import { AppTopNav } from "@/components/layout/app-top-nav";
import { fetchWorkspaceConfig } from "@/services/api/workspace-config";
import { useBackendStore } from "@/stores/use-backend-store";

export default function AppLayout({ children }: { children: ReactNode }) {
    const available = useBackendStore((state) => state.available);
    useEffect(() => {
        if (!available) return;
        let canceled = false;
        void fetchWorkspaceConfig().then(async (config) => {
            const syncEnabled = config.syncCapabilities?.workspaceData === true;
            const [{ useCanvasStore }, { useAssetStore }] = await Promise.all([
                import("@/app/(user)/canvas/stores/use-canvas-store"),
                import("@/stores/use-asset-store"),
            ]);
            if (canceled) return;
            useCanvasStore.getState().setSyncEnabled(syncEnabled);
            void useAssetStore.getState().hydrateWorkspaceAssets(syncEnabled);
        }).catch(() => {});
        return () => { canceled = true; };
    }, [available]);
    return <div className="flex h-dvh flex-col overflow-hidden bg-background text-foreground"><AppTopNav /><div className="min-h-0 flex-1 overflow-hidden">{children}</div></div>;
}
