import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig(({command})=>({
  root: "renderer",
  base: "./",
  plugins: [react(),{
    name:"desktop-production-csp",
    transformIndexHtml(html){
      if(command!=="build")return html;
      return html.replace("script-src 'self' 'unsafe-inline'","script-src 'self'").replace("connect-src 'self' ws://127.0.0.1:5173","connect-src 'none'");
    },
  }],
  server: { host: "127.0.0.1", port: 5173, strictPort: true },
  build: { outDir: "../dist/renderer", emptyOutDir: true },
}));
