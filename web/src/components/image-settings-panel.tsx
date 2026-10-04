"use client";

import { type ReactNode } from "react";
import { ConfigProvider } from "antd";

import { type CanvasTheme } from "@/lib/canvas-theme";
import { imageResolutionOptions, normalizeImageResolution } from "@/lib/image-resolution";
import type { AiConfig } from "@/stores/use-config-store";

const qualityOptions = [
    { value: "low", label: "低" },
    { value: "medium", label: "中" },
    { value: "high", label: "高" },
];

const aspectOptions = [
    { value: "1:1", label: "1:1", width: 1, height: 1 },
    { value: "2:3", label: "2:3", width: 2, height: 3 },
    { value: "3:2", label: "3:2", width: 3, height: 2 },
    { value: "4:5", label: "4:5", width: 4, height: 5 },
    { value: "5:4", label: "5:4", width: 5, height: 4 },
    { value: "16:9", label: "16:9", width: 16, height: 9 },
    { value: "9:16", label: "9:16", width: 9, height: 16 },
    { value: "21:9", label: "21:9", width: 21, height: 9 },
    { value: "9:21", label: "9:21", width: 9, height: 21 },
    { value: "3:4", label: "3:4", width: 3, height: 4 },
    { value: "4:3", label: "4:3", width: 4, height: 3 },
    { value: "1:2", label: "1:2", width: 1, height: 2 },
    { value: "2:1", label: "2:1", width: 2, height: 1 },
    { value: "1:3", label: "1:3", width: 1, height: 3 },
    { value: "3:1", label: "3:1", width: 3, height: 1 },
];

type ImageSettingsPanelProps = {
    config: AiConfig;
    onConfigChange: (key: "quality" | "imageResolution" | "size" | "count", value: string) => void;
    theme: CanvasTheme;
    showTitle?: boolean;
    showSize?: boolean;
    showCount?: boolean;
    className?: string;
    maxCount?: number;
    quickCount?: number;
};

export function ImageSettingsPanel({ config, onConfigChange, theme, showTitle = true, showSize = true, showCount = true, className = "w-[320px] space-y-4 rounded-2xl px-1 py-0.5", maxCount = 15, quickCount = 10 }: ImageSettingsPanelProps) {
    const quality = config.quality || "medium";
    const count = Math.max(1, Math.min(maxCount, Math.floor(Math.abs(Number(config.count)) || 1)));
    const activeSize = config.size || "auto";
    const selectedAspect = aspectForSize(activeSize);
    const resolution = normalizeImageResolution(config.imageResolution || "1k");

    return (
        <ImageSettingsTheme theme={theme}>
            <div
                className={className}
                style={{ color: theme.node.text }}
                onMouseDown={(event) => {
                    event.stopPropagation();
                    if (event.target instanceof HTMLInputElement) return;
                    if (document.activeElement instanceof HTMLInputElement && event.currentTarget.contains(document.activeElement)) document.activeElement.blur();
                }}
            >
                {showTitle ? <div className="text-lg font-semibold">图像设置</div> : null}
                <div className="space-y-2.5">
                    <SettingTitle color={theme.node.muted}>图像质量</SettingTitle>
                    <div className="grid grid-cols-3 gap-1 rounded-xl p-1" style={{ background: theme.node.fill }}>
                        {qualityOptions.map((item) => (
                            <SegmentOption key={item.value} selected={quality === item.value} theme={theme} onClick={() => onConfigChange("quality", item.value)}>
                                {item.label}
                            </SegmentOption>
                        ))}
                    </div>
                </div>
                {showSize ? (
                    <>
                        <div className="space-y-2.5">
                            <SettingTitle color={theme.node.muted}>比例</SettingTitle>
                            <div className="grid grid-cols-[58px_minmax(0,1fr)] gap-2 rounded-xl p-2" style={{ background: theme.node.fill }}>
                                <AspectOption selected={selectedAspect === "auto"} theme={theme} width={1} height={1} label="自适应" onClick={() => onConfigChange("size", "auto")} />
                                <div className="grid grid-cols-4 gap-1">
                                    {aspectOptions.map((item) => (
                                        <button
                                            key={item.value}
                                            type="button"
                                            className="flex min-h-[50px] cursor-pointer flex-col items-center justify-center gap-1 rounded-lg text-xs transition hover:opacity-75"
                                            style={{ background: selectedAspect === item.value ? theme.toolbar.activeBg : "transparent", color: selectedAspect === item.value ? theme.toolbar.activeText : theme.node.muted }}
                                            onMouseDown={(event) => event.stopPropagation()}
                                            onClick={() => onConfigChange("size", item.value)}
                                        >
                                            <AspectIcon width={item.width} height={item.height} color={selectedAspect === item.value ? theme.toolbar.activeText : theme.node.muted} />
                                            <span>{item.label}</span>
                                        </button>
                                    ))}
                                </div>
                            </div>
                        </div>
                        <div className="space-y-2.5">
                            <SettingTitle color={theme.node.muted}>分辨率</SettingTitle>
                            <div className="grid grid-cols-3 gap-1 rounded-xl p-1" style={{ background: theme.node.fill }}>
                                {imageResolutionOptions.map((item) => (
                                    <SegmentOption key={item.value} selected={resolution === item.value} theme={theme} onClick={() => onConfigChange("imageResolution", item.value)}>
                                        {item.label}
                                    </SegmentOption>
                                ))}
                            </div>
                        </div>
                    </>
                ) : null}
                {showCount ? (
                    <div className="space-y-2.5">
                        <SettingTitle color={theme.node.muted}>生成张数</SettingTitle>
                        <div className="grid grid-cols-4 gap-2.5">
                            {Array.from({ length: quickCount }, (_, index) => index + 1).map((value) => (
                                <OptionPill key={value} selected={count === value} theme={theme} onClick={() => onConfigChange("count", String(value))}>
                                    {value} 张
                                </OptionPill>
                            ))}
                            <CountInput value={count} max={maxCount} theme={theme} onChange={(value) => onConfigChange("count", String(value || 1))} />
                        </div>
                    </div>
                ) : null}
            </div>
        </ImageSettingsTheme>
    );
}

export function ImageSettingsTheme({ theme, children }: { theme: CanvasTheme; children: ReactNode }) {
    return (
        <ConfigProvider
            theme={{
                token: { colorBgContainer: theme.toolbar.panel, colorBgElevated: theme.toolbar.panel, colorBorder: theme.node.stroke, colorPrimary: theme.node.activeStroke, colorText: theme.node.text, colorTextLightSolid: theme.node.panel },
                components: { Button: { defaultBg: theme.toolbar.panel, defaultBorderColor: theme.node.stroke, defaultColor: theme.node.text } },
            }}
        >
            {children}
        </ConfigProvider>
    );
}

export function imageQualityLabel(value: string) {
    return ({ auto: "自动", high: "高", medium: "中", low: "低" } as Record<string, string>)[value] || value;
}

export function imageResolutionLabel(value: string) {
    return normalizeImageResolution(value);
}

export function imageSizeLabel(size: string) {
    const aspect = aspectForSize(size);
    return aspectOptions.find((item) => item.value === aspect)?.label || (aspect === "auto" ? "自适应" : size);
}

function OptionPill({ selected, theme, onClick, children }: { selected: boolean; theme: CanvasTheme; onClick: () => void; children: ReactNode }) {
    return (
        <button
            type="button"
            className="h-9 cursor-pointer rounded-full border px-2 text-sm transition hover:opacity-80"
            style={{ background: "transparent", borderColor: selected ? theme.node.text : theme.node.stroke, color: theme.node.text }}
            onMouseDown={(event) => event.stopPropagation()}
            onClick={onClick}
        >
            {children}
        </button>
    );
}

function CountInput({ value, max, theme, onChange }: { value: number; max: number; theme: CanvasTheme; onChange: (value: number | null) => void }) {
    return (
        <label className="col-span-2 flex h-9 overflow-hidden rounded-full border text-sm" style={{ borderColor: theme.node.stroke, color: theme.node.text }}>
            <input
                type="number"
                min={1}
                max={max}
                className="min-w-0 flex-1 bg-transparent px-3 text-center outline-none [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
                style={{ color: theme.node.text, WebkitTextFillColor: theme.node.text }}
                value={value || ""}
                onChange={(event) => onChange(Number(event.target.value) || null)}
                onMouseDown={(event) => event.stopPropagation()}
            />
        </label>
    );
}

function AspectIcon({ width, height, color }: { width: number; height: number; color: string }) {
    const ratio = width / Math.max(1, height);
    const boxWidth = ratio >= 1 ? 16 : Math.max(6, 16 * ratio);
    const boxHeight = ratio >= 1 ? Math.max(6, 16 / ratio) : 16;
    return <span className="rounded-[2px] border" style={{ width: boxWidth, height: boxHeight, borderWidth: 1.5, borderColor: color }} />;
}

function SettingTitle({ children, color }: { children: string; color: string }) {
    return (
        <div className="text-xs font-medium" style={{ color }}>
            {children}
        </div>
    );
}

function aspectForSize(size: string) {
    const value = size.split("-")[0];
    if (value === "auto") return "auto";
    const exact = aspectOptions.find((item) => item.value === value);
    if (exact) return exact.value;
    const match = value.match(/^(\d+)[x:](\d+)$/);
    if (!match) return "auto";
    const ratio = Number(match[1]) / Number(match[2]);
    return aspectOptions.reduce((best, item) => Math.abs(item.width / item.height - ratio) < Math.abs(best.width / best.height - ratio) ? item : best).value;
}

function SegmentOption({ selected, theme, onClick, children }: { selected: boolean; theme: CanvasTheme; onClick: () => void; children: ReactNode }) {
    return (
        <button type="button" className="h-9 rounded-lg px-2 text-sm transition hover:opacity-75" style={{ background: selected ? theme.toolbar.activeBg : "transparent", color: selected ? theme.toolbar.activeText : theme.node.muted }} onMouseDown={(event) => event.stopPropagation()} onClick={onClick}>
            {children}
        </button>
    );
}

function AspectOption({ selected, theme, width, height, label, onClick }: { selected: boolean; theme: CanvasTheme; width: number; height: number; label: string; onClick: () => void }) {
    return (
        <button type="button" className="flex min-h-[202px] cursor-pointer flex-col items-center justify-center gap-2 rounded-lg px-1 text-xs transition hover:opacity-75" style={{ background: selected ? theme.toolbar.activeBg : "transparent", color: selected ? theme.toolbar.activeText : theme.node.muted }} onMouseDown={(event) => event.stopPropagation()} onClick={onClick}>
            <AspectIcon width={width} height={height} color={selected ? theme.toolbar.activeText : theme.node.muted} />
            <span>{label}</span>
        </button>
    );
}

export function imageFormatLabel(format: string) {
    const map: Record<string, string> = { png: "PNG", jpeg: "JPEG", webp: "WebP" };
    return map[format] || format;
}
