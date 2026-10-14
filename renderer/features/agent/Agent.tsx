import { useEffect, useRef, useState } from "react";
import { ArrowUp, Check, Sparkles } from "lucide-react";
import type { Channel, Job, LibraryEntry } from "../../../contracts/desktop";
import type { CanvasController } from "../canvas/controller";
import { agentPrompt, executeTool, parseReply, type ToolCall } from "./tools";

export function Agent({controller,channels,library,jobs,onSubmit,onError}:{
  controller:CanvasController;channels:Channel[];library:LibraryEntry[];jobs:Job[];
  onSubmit:(nodeId:string,prompt:string)=>Promise<Job>;onError:(e:unknown)=>void;
}){
  const [prompt,setPrompt]=useState(""),[model,setModel]=useState(""),[skill,setSkill]=useState("");
  const [reply,setReply]=useState(""),[commands,setCommands]=useState<{call:ToolCall;done:boolean}[]>([]);
  const [busy,setBusy]=useState(false),[jobID,setJobID]=useState("");
  const mounted=useRef(true);
  useEffect(()=>{mounted.current=true;return()=>{mounted.current=false}},[]);
  const job=jobs.find(j=>j.id===jobID);
  useEffect(()=>{
    if(!job||job.status==="queued"||job.status==="running")return;
    setJobID("");setBusy(false);
    if(job.status!=="succeeded"){onError(new Error(job.errorCode||"Agent 任务已取消"));return}
    try{const result=parseReply(job.text);setReply(result.reply);setCommands(result.commands.map(call=>({call,done:false})))}catch(e){onError(e)}
  },[job,onError]);
  const submit=async()=>{
    if(!prompt.trim()||busy)return;
    setBusy(true);setReply("");setCommands([]);
    try{
      const [channelId,modelName]=JSON.parse(model) as [string,string];
      const context=agentPrompt(controller,prompt,library.find(e=>e.id===skill)?.content??"");
      const v=controller.getSnapshot().document.viewport;
      const id=controller.createNode({kind:"text",title:"Agent · "+prompt.slice(0,24),x:(320-v.x)/v.zoom,y:(180-v.y)/v.zoom});
      controller.editNode({id,title:"Agent · "+prompt.slice(0,24),content:"",prompt,channelId,model:modelName});
      const receipt=await onSubmit(id,context);
      if(mounted.current)setJobID(receipt.id);
    }catch(e){if(mounted.current){setBusy(false);onError(e)}}
  };
  return <div className="agent-panel"><div className="panel-heading"><h3><Sparkles size={16}/>创作助手</h3><span className="tag">Agent</span></div>
    <p className="muted small">一起整理思路、扩写想法，或规划画布。每个画布操作由你确认。</p>
    <label>文本模型<select value={model} onChange={e=>setModel(e.target.value)}><option value="">选择渠道与模型</option>{channels.flatMap(c=>c.models.filter(m=>m.kind==="text").map(m=><option key={c.id+m.name} value={JSON.stringify([c.id,m.name])}>{c.name} / {m.name}</option>))}</select></label>
    <label>使用 Skill<select value={skill} onChange={e=>setSkill(e.target.value)}><option value="">不使用</option>{library.filter(e=>e.kind==="skill").map(e=><option key={e.id} value={e.id}>{e.name}</option>)}</select></label>
    <form onSubmit={e=>{e.preventDefault();void submit()}}><label>你的想法<textarea rows={6} required value={prompt} onChange={e=>setPrompt(e.target.value)} placeholder="例如：帮我把这个故事拆成三个场景…"/></label>
      <button className="button primary full" disabled={busy||!model||!prompt.trim()}><ArrowUp size={15}/>{busy?"正在构思…":"发送给助手"}</button></form>
    {reply?<div className="agent-reply">{reply}</div>:null}
    {commands.map((entry,i)=><article className="tool-card" key={i}><strong>{entry.call.tool==="create_node"?"添加节点":entry.call.tool==="connect_nodes"?"连接节点":"删除节点"}</strong><pre>{JSON.stringify(entry.call.args,null,2)}</pre>
      <button className="button secondary full" disabled={entry.done} onClick={()=>{try{executeTool(controller,entry.call);setCommands(cs=>cs.map((c,n)=>n===i?{...c,done:true}:c))}catch(e){onError(e)}}}><Check size={14}/>{entry.done?"已执行":"确认执行"}</button></article>)}
  </div>;
}
