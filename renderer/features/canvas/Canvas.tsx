import { useCallback, useMemo, useSyncExternalStore } from "react";
import { Background, BackgroundVariant, Controls, MiniMap, ReactFlow, type NodeChange } from "@xyflow/react";
import type { Job } from "../../../contracts/desktop";
import type { CanvasController } from "./controller";
import type { FlowNode, NodePlugin } from "./node-plugins";
import type { ComponentType } from "react";
import type { NodeProps } from "@xyflow/react";
const nodeColors:Record<string,string>={text:"#ddd4be",image:"#bdd0c3",video:"#d1c5da",audio:"#c2d0db"};

export function Canvas({controller,jobs,nodeTypes,onError}:{
  controller:CanvasController;jobs:Job[];nodeTypes:Record<string,ComponentType<NodeProps<FlowNode>>>;plugins:readonly NodePlugin[];onError:(error:unknown)=>void;
}){
  const state=useSyncExternalStore(controller.subscribe,controller.getSnapshot);
  const nodes=useMemo<FlowNode[]>(()=>state.document.nodes.map(node=>({
    id:node.id,type:node.kind,position:{x:node.x,y:node.y},selected:state.selected.includes(node.id),
    data:{node,job:jobs.find(j=>j.id===node.jobId)},
  })),[state.document.nodes,state.selected,jobs]);
  const edges=useMemo(()=>state.document.edges.map(e=>({...e,type:"smoothstep"})),[state.document.edges]);
  const onNodesChange=useCallback((changes:NodeChange<FlowNode>[])=>{
    const selected=new Set(controller.getSnapshot().selected);
    for(const change of changes){
      if(change.type==="position"&&change.position)controller.moveNode(change.id,change.position.x,change.position.y);
      if(change.type==="remove"){try{controller.removeNodes({ids:[change.id]})}catch(error){onError(error)}}
      if(change.type==="select"){if(change.selected)selected.add(change.id);else selected.delete(change.id)}
    }
    if(changes.some(c=>c.type==="select"))controller.select([...selected]);
  },[controller,onError]);
  return <div className="canvas-surface">
    <ReactFlow<FlowNode> key={state.project?.id} nodes={nodes} edges={edges} nodeTypes={nodeTypes}
      onNodesChange={onNodesChange} onNodeDragStart={()=>controller.beginGesture()} onNodeDragStop={()=>controller.endGesture()}
      onSelectionChange={({nodes})=>controller.select(nodes.map(n=>n.id))}
      onConnect={connection=>{if(connection.source&&connection.target){try{controller.connectNodes({source:connection.source,target:connection.target})}catch(error){onError(error)}}}}
      onEdgesDelete={edges=>controller.removeEdges(edges.map(e=>e.id))}
      onEdgeDoubleClick={(_event,edge)=>controller.removeEdges([edge.id])}
      onMoveEnd={(_event,viewport)=>controller.viewport(viewport)}
      defaultViewport={state.document.viewport} minZoom={.1} maxZoom={4}
      deleteKeyCode={null} selectionOnDrag panOnDrag={[1,2]} panOnScroll zoomOnScroll={false}
      proOptions={{hideAttribution:true}}>
      <Background variant={BackgroundVariant.Dots} gap={24} size={1} color="#c8c5bc"/>
      <Controls showInteractive={false}/><MiniMap pannable zoomable nodeColor={node=>nodeColors[node.type??""]??"#ddd"} maskColor="rgba(244,242,237,.7)"/>
    </ReactFlow>
    {nodes.length===0?<div className="canvas-welcome"><span className="eyebrow">A SPACE FOR YOUR NEXT IDEA</span><h2>从一个想法开始。</h2><p>添加文本，连接素材，让灵感在画布上生长。</p></div>:null}
    <div className="canvas-hint">滚轮平移 · 双指缩放 · 中键拖动 · 拖动端点连接 · 双击连线删除</div>
  </div>;
}
