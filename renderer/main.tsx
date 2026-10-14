import { Component, type ReactNode } from "react";
import { createRoot } from "react-dom/client";
import "@xyflow/react/dist/style.css";
import "./styles.css";
import { App } from "./app/App";
import { APIClient } from "./services/api";
import { CanvasController } from "./features/canvas/controller";
import { loadNodePlugins } from "./features/canvas/node-plugins";

class ErrorBoundary extends Component<{children:ReactNode},{failed:boolean}> {
  state={failed:false};
  static getDerivedStateFromError(){return {failed:true}}
  render(){return this.state.failed?<div className="startup-error"><h1>工作台暂时无法显示</h1><p>请重新启动墨华画布。本地已保存的项目仍在工作空间内。</p></div>:this.props.children}
}
const root=createRoot(document.getElementById("root")!);
try{
  const api=new APIClient(window.mohua),controller=new CanvasController(),nodePlugins=loadNodePlugins();
  root.render(<ErrorBoundary><App api={api} controller={controller} nodePlugins={nodePlugins}/></ErrorBoundary>);
}catch{
  root.render(<div className="startup-error"><h1>请通过 Electron 打开墨华画布</h1><p>开发时运行 npm run dev。</p></div>);
}
