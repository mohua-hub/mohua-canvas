import { useCallback, useEffect, useRef, useState } from "react";
import type { Update } from "@tauri-apps/plugin-updater";

type DownloadProgress = { downloaded: number; total?: number };

const updaterBundleTypes = new Set(["nsis"]);

export function useDesktopUpdater() {
    const updateRef = useRef<Update | null>(null);
    const checkRef = useRef<Promise<void> | null>(null);
    const bundleTypeRef = useRef("");
    const [isDesktop, setIsDesktop] = useState(false);
    const [supported, setSupported] = useState(false);
    const [checking, setChecking] = useState(false);
    const [installing, setInstalling] = useState(false);
    const [update, setUpdate] = useState<Update | null>(null);
    const [progress, setProgress] = useState<DownloadProgress | null>(null);
    const [error, setError] = useState("");

    const checkForUpdate = useCallback(() => {
        if (checkRef.current) return checkRef.current;
        const operation = (async () => {
            setChecking(true);
            setError("");
            try {
                const { isTauri } = await import("@tauri-apps/api/core");
                const desktop = isTauri();
                setIsDesktop(desktop);
                if (!desktop) {
                    setSupported(false);
                    return;
                }
                const { getBundleType } = await import("@tauri-apps/api/app");
                const bundleType = await getBundleType();
                bundleTypeRef.current = bundleType;
                const canUpdate = updaterBundleTypes.has(bundleType);
                setSupported(canUpdate);
                if (!canUpdate) {
                    const previous = updateRef.current;
                    updateRef.current = null;
                    setUpdate(null);
                    if (previous) await previous.close().catch(() => {});
                    return;
                }

                const { check } = await import("@tauri-apps/plugin-updater");
                const next = await check();
                const previous = updateRef.current;
                if (previous && previous !== next) await previous.close().catch(() => {});
                updateRef.current = next;
                setUpdate(next);
            } catch (reason) {
                setError(reason instanceof Error ? reason.message : typeof reason === "string" ? reason : "自动更新检查失败");
            } finally {
                setChecking(false);
            }
        })();
        checkRef.current = operation;
        void operation.finally(() => {
            if (checkRef.current === operation) checkRef.current = null;
        });
        return operation;
    }, []);

    const installUpdate = useCallback(async () => {
        const current = updateRef.current;
        if (!current || installing) return;
        setInstalling(true);
        setProgress({ downloaded: 0 });
        setError("");
        try {
            await current.downloadAndInstall((event) => {
                if (event.event === "Started") setProgress({ downloaded: 0, total: event.data.contentLength });
                if (event.event === "Progress") setProgress((value) => ({ downloaded: (value?.downloaded || 0) + event.data.chunkLength, total: value?.total }));
            });
            if (bundleTypeRef.current !== "nsis") {
                const { relaunch } = await import("@tauri-apps/plugin-process");
                await relaunch();
            }
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : typeof reason === "string" ? reason : "自动更新安装失败");
            setInstalling(false);
        }
    }, [installing]);

    useEffect(() => {
        void checkForUpdate();
        return () => {
            const current = updateRef.current;
            updateRef.current = null;
            if (current) void current.close().catch(() => {});
        };
    }, [checkForUpdate]);

    return { isDesktop, supported, checking, installing, update, progress, error, checkForUpdate, installUpdate };
}
