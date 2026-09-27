import { readFileSync } from "node:fs";

const tag = process.argv[2];
const changelog = readFileSync(new URL("../CHANGELOG.md", import.meta.url), "utf8");
const section = changelog.split(/^## /m).find((entry) => entry.split(/\r?\n/, 1)[0].split(" ")[0] === tag);
if (!section) throw new Error(`CHANGELOG.md 缺少 ${tag}`);
console.log(section.slice(section.indexOf("\n") + 1).trim());
console.log("\n安装包包含 Node 运行时和 Next.js 服务。Windows：exe / msi；macOS：按 x64 或 aarch64 选择 dmg；Linux：AppImage / deb。当前安装包未配置开发者证书签名，macOS 使用临时签名。账号和云端功能需要另行配置后端。\n");
