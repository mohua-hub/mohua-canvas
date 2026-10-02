"use client";

import type { CSSProperties } from "react";
import { Keyboard, Settings, Settings2 } from "lucide-react";
import Link from "next/link";
import { AnimatedThemeToggler } from "@/components/ui/animated-theme-toggler";
import { GitHubLink } from "@/components/layout/github-link";
import { VersionReleaseModal } from "@/components/layout/version-release-modal";
import { canvasThemes } from "@/lib/canvas-theme";
import { useConfigStore } from "@/stores/use-config-store";
import { useThemeStore } from "@/stores/use-theme-store";

export function AppActions({ showConfig = true, variant = "default", onOpenShortcuts }: { showConfig?: boolean; variant?: "default" | "canvas"; onOpenShortcuts?: () => void }) {
    const theme = useThemeStore((state) => state.theme);
    const setTheme = useThemeStore((state) => state.setTheme);
    const openConfigDialog = useConfigStore((state) => state.openConfigDialog);
    const iconStyle: CSSProperties | undefined = variant === "canvas" ? { color: canvasThemes[theme].node.text } : undefined;
    const iconClass = "inline-flex size-7 shrink-0 items-center justify-center text-muted-foreground transition hover:text-foreground [&_svg]:size-4";
    return <div className="inline-flex shrink-0 items-center gap-1">
        {showConfig ? <button type="button" className={iconClass} style={iconStyle} onClick={() => openConfigDialog(false)} aria-label="生成配置" title="生成配置"><Settings2 /></button> : null}
        <Link href="/settings" className={iconClass} style={iconStyle} aria-label="设置" title="设置"><Settings /></Link>
        <AnimatedThemeToggler theme={theme} onThemeChange={setTheme} className={iconClass} style={iconStyle} aria-label="切换主题" title="切换主题" />
        <VersionReleaseModal style={iconStyle} />
        <GitHubLink className="size-7 bg-transparent text-base" style={iconStyle} />
        {onOpenShortcuts ? <button type="button" className={iconClass} style={iconStyle} onClick={onOpenShortcuts} aria-label="快捷键" title="快捷键"><Keyboard /></button> : null}
    </div>;
}
