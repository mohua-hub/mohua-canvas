"use client";

import { AuditOutlined, FileTextOutlined, HomeOutlined, PictureOutlined, SettingOutlined, ToolOutlined } from "@ant-design/icons";
import { Button, Flex, Layout, Menu, Typography, theme } from "antd";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import type { ReactNode } from "react";

import { AppActions } from "@/components/layout/app-actions";
import { adminLayoutStyle } from "@/lib/app-theme";

const adminMenus = [
    { key: "/settings/ai-logs", icon: <AuditOutlined />, label: "AI 日志" },
    { key: "/settings/prompts", icon: <FileTextOutlined />, label: "提示词管理" },
    { key: "/settings/skills", icon: <ToolOutlined />, label: "Skill 管理" },
    { key: "/settings/assets", icon: <PictureOutlined />, label: "素材库" },
    { key: "/settings", icon: <SettingOutlined />, label: "系统设置" },
];

export default function SettingsLayout({ children }: { children: ReactNode }) {
    const { token: antToken } = theme.useToken();
    const pathname = usePathname();
    const router = useRouter();
    const activeKey = pathname.startsWith("/settings/assets") ? "/settings/assets" : pathname.startsWith("/settings/skills") ? "/settings/skills" : pathname.startsWith("/settings/prompts") ? "/settings/prompts" : pathname.startsWith("/settings/ai-logs") ? "/settings/ai-logs" : "/settings";
    const pageTitle = activeKey === "/settings/assets" ? "素材库" : activeKey === "/settings/skills" ? "Skill 管理" : activeKey === "/settings/prompts" ? "提示词管理" : activeKey === "/settings/ai-logs" ? "AI 日志" : "设置";

    return (
        <Layout hasSider style={{ height: "100vh", overflow: "hidden", background: antToken.colorBgLayout }}>
            <Layout.Sider width={adminLayoutStyle.siderWidth} style={{ height: "100vh", overflow: "hidden", background: antToken.colorBgContainer, borderRight: `1px solid ${antToken.colorBorder}` }}>
                <Flex align="center" gap={12} style={{ height: adminLayoutStyle.brandHeight, padding: "0 20px", borderBottom: `1px solid ${antToken.colorBorderSecondary}` }}>
                    <span aria-hidden style={{ display: "inline-block", width: 30, height: 30, background: antToken.colorText, WebkitMask: "url(/logo.svg) center / contain no-repeat", mask: "url(/logo.svg) center / contain no-repeat" }} />
                    <Typography.Text strong style={{ fontSize: 18, letterSpacing: 0 }}>
                        画布设置
                    </Typography.Text>
                </Flex>
                <Menu
                    mode="inline"
                    selectedKeys={[activeKey]}
                    style={adminLayoutStyle.menu}
                    items={adminMenus.map((item) => ({
                        ...item,
                        label: (
                            <Link href={item.key} style={{ color: "inherit" }}>
                                {item.label}
                            </Link>
                        ),
                        style: adminLayoutStyle.menuItem,
                    }))}
                />
                <Flex vertical gap={8} style={{ position: "absolute", bottom: 0, insetInline: 0, padding: 12, borderTop: `1px solid ${antToken.colorBorder}`, background: antToken.colorBgContainer }}>
                    <Button block icon={<HomeOutlined />} onClick={() => router.push("/canvas")}>
                        前往画布
                    </Button>
                </Flex>
            </Layout.Sider>
            <Layout style={{ background: antToken.colorBgLayout }}>
                <Layout.Header
                    style={{ display: "flex", alignItems: "center", justifyContent: "space-between", height: adminLayoutStyle.headerHeight, padding: "0 24px", background: antToken.colorBgContainer, borderBottom: `1px solid ${antToken.colorBorder}` }}
                >
                    <Typography.Title level={5} style={{ margin: 0 }}>
                        {pageTitle}
                    </Typography.Title>
                    <Flex align="center" gap={4}>
                        <AppActions showConfig={false} />
                    </Flex>
                </Layout.Header>
                <Layout.Content style={{ minHeight: 0, overflow: "auto" }}>{children}</Layout.Content>
            </Layout>
        </Layout>
    );
}
