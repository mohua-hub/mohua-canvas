const http = require("node:http");
const net = require("node:net");
const { randomUUID } = require("node:crypto");

const port = 39217;
process.env.NODE_ENV = "production";
process.env.HOSTNAME = "127.0.0.1";
process.env.PORT = String(port);
process.env.MOHUA_DESKTOP_TOKEN = randomUUID();

// Tauri 持有 stdin 写端；正常退出或意外结束时关闭本地服务。
process.stdin.resume();
process.stdin.on("end", () => process.exit(0));

const probe = net.createServer();
probe.once("error", (error) => {
    console.error(`桌面端口 ${port} 不可用：${error.message}`);
    process.exit(1);
});
probe.listen(port, "127.0.0.1", () => probe.close(() => {
    require("./server.js");
    const deadline = setTimeout(() => {
        console.error("Next 服务未能在 60 秒内就绪");
        process.exit(1);
    }, 60_000);
    const poll = () => {
        const request = http.get(`http://127.0.0.1:${port}/api/desktop-health`, { timeout: 1000 }, (response) => {
            let body = "";
            response.on("data", (chunk) => { body += chunk; });
            response.on("end", () => {
                if (response.statusCode === 200 && body === process.env.MOHUA_DESKTOP_TOKEN) {
                    clearTimeout(deadline);
                    console.log("MOHUA_DESKTOP_READY");
                } else {
                    setTimeout(poll, 250);
                }
            });
            response.on("error", () => setTimeout(poll, 250));
        });
        request.on("timeout", () => request.destroy());
        request.on("error", () => setTimeout(poll, 250));
    };
    poll();
}));
