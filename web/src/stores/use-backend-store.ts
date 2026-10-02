"use client";

import { create } from "zustand";

type BackendStore = {
    available: boolean;
    isReady: boolean;
    initialize: () => Promise<void>;
};

let pending: Promise<void> | undefined;

export const useBackendStore = create<BackendStore>((set) => ({
    available: false,
    isReady: false,
    initialize: () => pending ??= fetch("/api/health", { cache: "no-store", signal: AbortSignal.timeout(5_000) })
        .then((response) => { set({ available: response.ok, isReady: true }); })
        .catch(() => { set({ available: false, isReady: true }); }),
}));
