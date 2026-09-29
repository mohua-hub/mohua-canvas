import { readFileSync } from "node:fs";

const tag = process.argv[2];
const changelog = readFileSync(new URL("../CHANGELOG.md", import.meta.url), "utf8");
const section = changelog.split(/^## /m).find((entry) => entry.split(/\r?\n/, 1)[0].split(" ")[0] === tag);
if (!section) throw new Error(`CHANGELOG.md 缺少 ${tag}`);
console.log(section.slice(section.indexOf("\n") + 1).trim());
console.log("\n安装包包含 Node 运行时、Next.js 服务和 Go 后端。Windows：NSIS exe / msi；macOS：按 x64 或 aarch64 选择 dmg；Linux：AppImage / deb。仅 Windows NSIS 支持 Tauri 签名自动更新；MSI、macOS 和 Linux 安装包需手动更新。Windows 未配置开发者证书，macOS 使用临时签名。\n");
