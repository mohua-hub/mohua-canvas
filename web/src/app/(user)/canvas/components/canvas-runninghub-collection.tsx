"use client";

import { useEffect, useRef, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { App, Button, Empty, Input, Modal, Segmented, Spin } from "antd";
import { ArrowUpRight, RefreshCw, Search, Workflow } from "lucide-react";
import { canvasThemes } from "@/lib/canvas-theme";
import type { WorkflowEntry, WorkflowRef } from "@/lib/workflow-channel";
import { getRunningHubCollection, getRunningHubEntry, type RunningHubCollectionItem } from "@/services/api/runninghub";
import { useConfigStore } from "@/stores/use-config-store";
import { useThemeStore } from "@/stores/use-theme-store";
import { useBackendStore } from "@/stores/use-backend-store";

export default function CanvasRunningHubCollection({ onClose, onSelect }: { onClose: () => void; onSelect: (ref: WorkflowRef, entry: WorkflowEntry) => void }) {
    const { message } = App.useApp();
    const backendConnected = useBackendStore((state) => state.available);
    const openConfigDialog = useConfigStore((state) => state.openConfigDialog);
    const theme = canvasThemes[useThemeStore((state) => state.theme)];
    const [scope, setScope] = useState<WorkflowRef["scope"] | "all">("all");
    const [kind, setKind] = useState("all");
    const [search, setSearch] = useState("");
    const [selecting, setSelecting] = useState("");
    const active = useRef(true);
    useEffect(() => { active.current = true; return () => { active.current = false; }; }, []);
    const collection = useQuery({ queryKey: ["runninghub-collection", backendConnected, scope], queryFn: () => getRunningHubCollection(scope), enabled: Boolean(backendConnected), retry: false });
    const items = (collection.data || []).filter(({ entry, channelName }) => (kind === "all" || entry.kind === kind) && `${entry.title} ${entry.workflowId} ${channelName}`.toLowerCase().includes(search.trim().toLowerCase()));

    const select = async (item: RunningHubCollectionItem) => {
        setSelecting(JSON.stringify(item.ref));
        try {
            const entry = await getRunningHubEntry(item.ref);
            if (!active.current) return;
            if (useBackendStore.getState().available !== backendConnected) throw new Error("后端连接已变化，请重新选择");
            onSelect(item.ref, entry);
        } catch (error) {
            if (active.current) message.error(error instanceof Error ? error.message : "读取 RunningHub 字段失败");
        } finally {
            if (active.current) setSelecting("");
        }
    };

    return <Modal open centered title="RunningHub 集合" width={760} onCancel={onClose} footer={null} styles={{ container: { background: theme.toolbar.panel, color: theme.node.text } }}>
        <div className="space-y-4 pt-2">
            <div className="flex flex-wrap items-center justify-between gap-2">
                <Segmented value={scope} onChange={(value) => setScope(value as typeof scope)} options={[{ label: "全部集合", value: "all" }, { label: "本地集合", value: "personal" }, { label: "系统集合", value: "system" }]} />
                <div className="flex items-center gap-1">
                    <Button type="text" icon={<RefreshCw className="size-4" />} loading={collection.isFetching} disabled={!backendConnected} onClick={() => void collection.refetch()}>刷新</Button>
                    <Button type="text" onClick={() => { onClose(); openConfigDialog(true); }}>配置渠道</Button>
                </div>
            </div>
            <Input prefix={<Search className="size-4 opacity-50" />} placeholder="搜索名称、ID 或渠道" value={search} onChange={(event) => setSearch(event.target.value)} allowClear />
            <div className="flex items-center justify-between gap-3">
                <Segmented value={kind} onChange={(value) => setKind(String(value))} options={[{ label: "全部", value: "all" }, { label: "AI 应用", value: "app" }, { label: "工作流", value: "workflow" }]} />
                <span className="text-xs opacity-50">{items.length} 个条目</span>
            </div>
            <div className="thin-scrollbar max-h-[55vh] min-h-48 overflow-y-auto">
                {!backendConnected ? <Empty description="请先连接后端服务，然后配置 RunningHub 渠道及应用或工作流" /> : collection.isPending ? <div className="flex h-48 items-center justify-center"><Spin /></div> : collection.error ? <div className="py-12 text-center text-sm" role="alert">{collection.error.message}</div> : !items.length ? <Empty description="暂无匹配条目，请在渠道配置中添加并启用应用或工作流" /> : <div className="grid gap-2 sm:grid-cols-2">
                    {items.map((item) => {
                        const key = JSON.stringify(item.ref);
                        return <button key={key} type="button" disabled={Boolean(selecting)} className="group flex min-w-0 cursor-pointer items-start gap-3 rounded-xl border p-4 text-left transition hover:opacity-80 disabled:cursor-wait" style={{ background: theme.node.fill, borderColor: theme.node.stroke, color: theme.node.text }} onClick={() => void select(item)}>
                            <Workflow className="mt-0.5 size-5 shrink-0 opacity-60" />
                            <div className="min-w-0 flex-1">
                                <div className="truncate text-sm font-medium">{item.entry.title || item.entry.workflowId}</div>
                                <div className="mt-1 text-xs opacity-60">{item.ref.scope === "personal" ? "本地" : "系统"} · {item.channelName} · {item.entry.kind === "app" ? "AI 应用" : "工作流"} · {{ image: "图片", video: "视频", audio: "音频" }[item.entry.capability]}</div>
                                <div className="mt-2 truncate font-mono text-[11px] opacity-40">{item.entry.workflowId}</div>
                            </div>
                            {selecting === key ? <Spin size="small" /> : <ArrowUpRight className="size-4 shrink-0 opacity-40" />}
                        </button>;
                    })}
                </div>}
            </div>
            <div className="text-xs opacity-50">选择后插入画布，节点显示渠道中已启用的字段配置。</div>
        </div>
    </Modal>;
}
