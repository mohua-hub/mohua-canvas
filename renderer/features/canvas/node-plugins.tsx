import { memo, type ComponentType } from "react";
import { Handle, Position, type Node, type NodeProps } from "@xyflow/react";
import { FileText, Image, Film, AudioLines, LoaderCircle } from "lucide-react";
import type { CanvasNode, Job, Kind } from "../../../contracts/desktop";
import { assetURL } from "../../services/api";

export type FlowNode = Node<{node:CanvasNode;job?:Job}, "text"|"image"|"video"|"audio">;
export type NodePlugin = {id:string;version:string;apiVersion:1;kind:Kind;label:string;component:ComponentType<NodeProps<FlowNode>>};
const labels:Record<Kind,string>={text:"文本",image:"图片",video:"视频",audio:"音频"};
const icons={text:FileText,image:Image,video:Film,audio:AudioLines};
const MediaNode=memo(function MediaNode({data,selected}:NodeProps<FlowNode>){
  const {node,job}=data,Icon=icons[node.kind];
  const pending=job?.status==="queued"||job?.status==="running";
  return <article className={"canvas-node kind-"+node.kind+(selected?" is-selected":"")}>
    <Handle type="target" position={Position.Left}/><Handle type="source" position={Position.Right}/>
    <header><span className="node-kind"><Icon size={13}/>{labels[node.kind]}</span><span className="node-mark">·</span></header>
    <h3>{node.title}</h3>
    <div className="node-content">
      {node.kind==="text"?<p>{node.content||"写下想法，或让 AI 延展它。"}</p>:
       node.assetId?<>{node.kind==="image"?<img src={assetURL(node.assetId)} alt={node.title} draggable={false}/>:
       node.kind==="video"?<video className="nodrag nowheel" src={assetURL(node.assetId)} controls preload="metadata"/>:
       <audio className="nodrag nowheel" src={assetURL(node.assetId)} controls preload="metadata"/>}</>:
       <div className="media-placeholder"><Icon size={30} strokeWidth={1.1}/><span>{node.prompt||"选择节点，导入素材或开始生成"}</span></div>}
    </div>
    {job?<footer className={"job-state "+job.status}>{pending?<LoaderCircle size={12} className="spin"/>:null}
      {{queued:"等待生成",running:"正在生成",succeeded:"已完成",failed:"生成失败",cancelled:"已取消"}[job.status]}
      {job.errorCode?<span>{job.errorCode}</span>:null}</footer>:null}
  </article>;
});
export function loadNodePlugins(): {plugins:readonly NodePlugin[];nodeTypes:Record<string,ComponentType<NodeProps<FlowNode>>>} {
  const plugins:NodePlugin[]=(["text","image","video","audio"] as const).map(kind=>({id:"node."+kind,version:"1.0.0",apiVersion:1,kind,label:labels[kind],component:MediaNode}));
  const ids=new Set<string>(),types=new Set<string>();const nodeTypes:Record<string,ComponentType<NodeProps<FlowNode>>>={};
  for(const p of plugins){if(p.apiVersion!==1||ids.has(p.id)||types.has(p.kind))throw new Error("节点插件清单无效");ids.add(p.id);types.add(p.kind);nodeTypes[p.kind]=p.component}
  return {plugins:Object.freeze(plugins),nodeTypes:Object.freeze(nodeTypes)};
}
