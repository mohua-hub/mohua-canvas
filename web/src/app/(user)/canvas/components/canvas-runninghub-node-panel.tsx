"use client";

import { useState } from "react";
import { App, Button, Input, InputNumber, Select, Switch, Tooltip } from "antd";
import { LoaderCircle, Play, RefreshCw, Workflow } from "lucide-react";
import { canvasThemes } from "@/lib/canvas-theme";
import { workflowFieldChoiceValues, workflowFieldNumberBounds, workflowOptionValue } from "@/lib/workflow-field-options";
import { getRunningHubEntry } from "@/services/api/runninghub";
import { useThemeStore } from "@/stores/use-theme-store";
import { useBackendStore } from "@/stores/use-backend-store";
import type { CanvasNodeData, CanvasNodeMetadata } from "../types";
import type { NodeGenerationInput } from "./canvas-node-generation";
import { runningHubFieldKey, runningHubInitialValues, runningHubMediaIndex, runningHubMediaKind } from "../utils/runninghub-fields";

export function CanvasRunningHubNodePanel({ node, inputs, isRunning, onConfigChange, onGenerate, onChoose, embedded = true }: { node: CanvasNodeData; inputs: NodeGenerationInput[]; isRunning: boolean; onConfigChange: (id: string, patch: Partial<CanvasNodeMetadata>) => void; onGenerate: () => void; onChoose: () => void; embedded?: boolean }) {
    const { message } = App.useApp();
    const theme = canvasThemes[useThemeStore((state) => state.theme)];
    const [refreshing, setRefreshing] = useState(false);
    const entry = node.metadata?.runningHubEntry;
    const values = node.metadata?.runningHubFieldValues || {};
    const media = node.metadata?.runningHubMedia || {};
    if (!entry) return null;
    const updateValue = (key: string, value: unknown) => onConfigChange(node.id, { runningHubFieldValues: { ...values, [key]: value } });
    const refresh = async () => {
        const backendConnected = useBackendStore.getState().available;
        if (!backendConnected || !node.metadata?.workflowRef) return message.error("请先连接后端服务");
        setRefreshing(true);
        try {
            const next = await getRunningHubEntry(node.metadata.workflowRef);
            if (next.capability !== entry.capability) throw new Error("条目用途已变化，请重新从集合插入节点");
            const keys = new Set(next.fields.map(runningHubFieldKey));
            const oldKeys = new Set(entry.fields.map(runningHubFieldKey));
            const defaults = Object.fromEntries(Object.entries(runningHubInitialValues(next)).filter(([key]) => !oldKeys.has(key)));
            onConfigChange(node.id, { runningHubEntry: next, runningHubFieldValues: { ...defaults, ...Object.fromEntries(Object.entries(values).filter(([key]) => keys.has(key))) }, runningHubMedia: Object.fromEntries(Object.entries(media).filter(([key]) => keys.has(key))) });
            message.success("参数配置已刷新");
        } catch (error) {
            message.error(error instanceof Error ? error.message : "刷新参数失败");
        } finally {
            setRefreshing(false);
        }
    };

    return <div data-canvas-no-zoom className={`flex min-h-0 flex-col gap-3 p-3 ${embedded ? "h-full pt-7" : "max-h-[520px] rounded-2xl border"}`} style={{ background: theme.toolbar.panel, borderColor: theme.toolbar.border, color: theme.node.text }} onWheel={(event) => event.stopPropagation()}>
        <div className="flex items-center justify-between gap-2">
            <div className="flex min-w-0 items-center gap-2"><Workflow className="size-4 shrink-0" /><span className="text-sm font-semibold">RunningHub</span><span className="text-[11px] opacity-50">{entry.kind === "app" ? "AI 应用" : "工作流"}</span></div>
            <div className="flex cursor-default items-center" onMouseDown={(event) => event.stopPropagation()}>
                <Tooltip title="刷新字段配置"><Button type="text" size="small" aria-label="刷新字段配置" loading={refreshing} disabled={isRunning} icon={<RefreshCw className="size-3.5" />} onClick={() => void refresh()} /></Tooltip>
                <Button type="text" size="small" disabled={isRunning || refreshing} onClick={onChoose}>更换</Button>
            </div>
        </div>
        <div className="truncate text-xs opacity-50" title={entry.workflowId}>{entry.title} · {entry.workflowId}</div>
        <div className="thin-scrollbar min-h-0 flex-1 space-y-3 overflow-y-auto cursor-default" onMouseDown={(event) => event.stopPropagation()} onPointerDown={(event) => event.stopPropagation()}>
            {entry.fields.length ? entry.fields.map((field) => {
                const key = runningHubFieldKey(field);
                const kind = runningHubMediaKind(field);
                const candidates = kind ? inputs.filter((input) => input.type === kind) : [];
                const selectedId = media[key] ?? candidates[runningHubMediaIndex(field)]?.nodeId;
                const choices = workflowFieldChoiceValues(field);
                const value = Object.prototype.hasOwnProperty.call(values, key) ? values[key] : field.fieldValue ?? field.defaultValue ?? field.default ?? field.value;
                const type = String(field.fieldType || "").toUpperCase();
                const upstreamPrompt = field.source === "prompt" && !Object.prototype.hasOwnProperty.call(values, key);
                return <div key={key} className="space-y-1.5">
                    <div className="flex items-center justify-between gap-2 text-xs">
                        <span>{field.label || field.fieldName}{field.required ? <span className="ml-1 opacity-50">*</span> : null}</span>
                        {field.source === "prompt" && !field.randomEnabled ? <label className="flex items-center gap-1.5 text-[10px] opacity-60">上游文本<Switch size="small" disabled={isRunning} checked={upstreamPrompt} onChange={(checked) => { const next = { ...values }; if (checked) delete next[key]; else next[key] = field.fieldValue ?? ""; onConfigChange(node.id, { runningHubFieldValues: next }); }} /></label> : null}
                    </div>
                    {kind ? (
                        <Select className="w-full" disabled={isRunning} allowClear value={candidates.some((input) => input.nodeId === selectedId) ? selectedId : undefined} placeholder={`连接${{ image: "图片", video: "视频", audio: "音频" }[kind]}节点后选择`} options={candidates.map((input) => ({ label: input.title, value: input.nodeId }))} onChange={(id) => onConfigChange(node.id, { runningHubMedia: { ...media, [key]: id || "" } })} />
                    ) : field.randomEnabled ? (
                        <div className="rounded-lg px-2 py-2 text-xs opacity-60" style={{ background: theme.node.fill }}>每次运行随机生成{field.min !== undefined || field.max !== undefined ? `（${field.min ?? "自动"} ～ ${field.max ?? "自动"}）` : ""}</div>
                    ) : upstreamPrompt ? (
                        <div className="py-2 text-xs opacity-50">使用已连接文本和节点提示词</div>
                    ) : choices.length ? (
                        <Select className="w-full" disabled={isRunning} value={value === undefined ? undefined : workflowOptionValue(value)} options={choices.map((option) => ({ value: workflowOptionValue(option), label: option && typeof option === "object" ? String((option as Record<string, unknown>).label ?? workflowOptionValue(option)) : String(option) }))} onChange={(next) => updateValue(key, next)} />
                    ) : ["NUMBER", "INT", "INTEGER", "FLOAT", "SLIDER"].includes(type) ? (
                        <InputNumber className="!w-full" disabled={isRunning} value={value === undefined || value === null || value === "" ? null : Number(value)} {...workflowFieldNumberBounds(field)} onChange={(next) => updateValue(key, next)} />
                    ) : ["BOOLEAN", "BOOL"].includes(type) ? (
                        <Switch disabled={isRunning} checked={value === true || value === "true" || value === 1} onChange={(next) => updateValue(key, next)} />
                    ) : (
                        <Input.TextArea disabled={isRunning} autoSize={{ minRows: 1, maxRows: 5 }} value={String(value ?? "")} onChange={(event) => updateValue(key, event.target.value)} />
                    )}
                </div>;
            }) : <div className="py-4 text-center text-xs opacity-50">使用工作流默认参数</div>}
        </div>
        {node.metadata?.errorDetails ? <div className="max-h-16 overflow-y-auto text-xs" role="alert">{node.metadata.errorDetails}</div> : null}
        <Button type="primary" className="!h-9 !w-full !shrink-0" disabled={isRunning || refreshing} icon={isRunning ? <LoaderCircle className="size-4 animate-spin" /> : <Play className="size-4" />} onMouseDown={(event) => event.stopPropagation()} onClick={onGenerate}>{isRunning ? "正在运行" : "运行"}</Button>
    </div>;
}
