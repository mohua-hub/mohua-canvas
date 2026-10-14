import { test } from "node:test";
import assert from "node:assert/strict";
import type { Project, Job } from "../contracts/desktop";
import { CanvasController } from "../renderer/features/canvas/controller";
import { executeTool, parseReply } from "../renderer/features/agent/tools";

function canvas(){
  const controller=new CanvasController();
  const project:Project={id:crypto.randomUUID(),name:"项目",revision:1,createdAt:"",updatedAt:"",document:{version:1,nodes:[],edges:[],viewport:{x:0,y:0,zoom:1}}};
  controller.load(project);return {controller,project};
}
test("acknowledging a save retains edits made during the request",()=>{
  const {controller,project}=canvas();
  controller.createNode({kind:"text",title:"第一笔",x:0,y:0});
  const saved=controller.getSnapshot();
  controller.createNode({kind:"image",title:"新想法",x:100,y:0});
  controller.acknowledge({...project,revision:2,document:saved.document},saved.sequence);
  assert.equal(controller.getSnapshot().document.nodes.length,2);
  assert.equal(controller.getSnapshot().project?.revision,2);
  assert.equal(controller.getSnapshot().dirty,true);
});
test("failed graph commands leave the live document intact",()=>{
  const {controller}=canvas();
  const a=controller.createNode({kind:"text",title:"甲",x:0,y:0});
  const b=controller.createNode({kind:"text",title:"乙",x:100,y:0});
  controller.connectNodes({source:a,target:b});
  const before=controller.getSnapshot();
  assert.throws(()=>controller.connectNodes({source:b,target:a}));
  assert.equal(controller.getSnapshot(),before);
  controller.removeNodes({ids:[a]});
  assert.equal(controller.getSnapshot().document.edges.length,0);
  controller.undo();
  assert.equal(controller.getSnapshot().document.edges.length,1);
});
test("a consumed result never overwrites subsequent edits",()=>{
  const {controller,project}=canvas();
  const id=controller.createNode({kind:"text",title:"草稿",x:0,y:0});
  const job:Job={id:crypto.randomUUID(),projectId:project.id,nodeId:id,kind:"text",status:"succeeded",text:"生成内容",assetId:"",errorCode:"",createdAt:"",updatedAt:""};
  controller.trackJob(id,job);controller.applyJob({nodeId:id,job});
  const node=controller.getSnapshot().document.nodes[0];
  assert.equal(node.content,"生成内容");assert.equal(node.jobId,"");
  controller.editNode({...node,content:"用户修订"});
  controller.applyJob({nodeId:id,job});
  assert.equal(controller.getSnapshot().document.nodes[0].content,"用户修订");
  controller.load({...project,document:controller.getSnapshot().document});
  controller.applyJob({nodeId:id,job});
  assert.equal(controller.getSnapshot().document.nodes[0].content,"用户修订");
});
test("undoing an applied generation does not reapply it on the next poll",()=>{
  const {controller,project}=canvas();
  const id=controller.createNode({kind:"text",title:"草稿",content:"原稿",x:0,y:0});
  const job:Job={id:crypto.randomUUID(),projectId:project.id,nodeId:id,kind:"text",status:"succeeded",text:"生成内容",assetId:"",errorCode:"",createdAt:"",updatedAt:""};
  controller.trackJob(id,job);controller.applyJob({nodeId:id,job});controller.undo();controller.applyJob({nodeId:id,job});
  assert.equal(controller.getSnapshot().document.nodes[0].content,"原稿");
  assert.equal(controller.getSnapshot().document.nodes[0].jobId,"");
});
test("Agent parsing cannot expand the tool permissions",()=>{
  const {controller}=canvas();
  for(const call of [
    {tool:"shell",args:{command:"anything"}},
    {tool:"create_node",args:{kind:"text",title:"节点",x:0,y:0,path:"unexpected"}},
    {tool:"delete_nodes",args:{ids:["invalid"]}},
  ])assert.throws(()=>parseReply(JSON.stringify({reply:"",commands:[call]})));
  const response=parseReply(JSON.stringify({reply:"计划",commands:[{tool:"create_node",args:{kind:"text",title:"建议",x:0,y:0}}]}));
  assert.equal(controller.getSnapshot().document.nodes.length,0);
  executeTool(controller,response.commands[0]);
  assert.equal(controller.getSnapshot().document.nodes.length,1);
});
