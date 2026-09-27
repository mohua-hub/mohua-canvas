export const runtime = "nodejs";
export const dynamic = "force-dynamic";

export function GET() {
    const token = process.env.MOHUA_DESKTOP_TOKEN;
    return new Response(token || "Not Found", {
        status: token ? 200 : 404,
        headers: { "Cache-Control": "no-store" },
    });
}
