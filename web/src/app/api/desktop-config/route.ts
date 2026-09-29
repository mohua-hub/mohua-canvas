export const dynamic = "force-dynamic";
export const runtime = "nodejs";

export async function GET() {
    const bridgeServerURL = process.env.MOHUA_DESKTOP === "1" ? sanitizeServerURL(process.env.MOHUA_BRIDGE_SERVER_URL || "") : "";
    return Response.json({ bridgeServerURL }, { headers: { "Cache-Control": "no-store" } });
}

function sanitizeServerURL(value: string) {
    try {
        const url = new URL(value);
        if (url.protocol !== "http:" && url.protocol !== "https:") return "";
        url.username = "";
        url.password = "";
        url.search = "";
        url.hash = "";
        return url.toString().replace(/\/+$/, "");
    } catch {
        return "";
    }
}
