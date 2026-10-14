const { z } = require("zod");
const utf8 = (max, min = 0) => z.string().refine(s => Buffer.byteLength(s, "utf8") >= min && Buffer.byteLength(s, "utf8") <= max);
const id = z.string().regex(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/);
const reference = z.union([id, z.literal("")]);
const kind = z.enum(["text", "image", "video", "audio"]);
const position = z.number().finite().min(-1e6).max(1e6);
const node = z.object({
  id, kind, x:position, y:position, title:utf8(120,1), content:utf8(32000), prompt:utf8(32000),
  channelId:reference, model:utf8(200), assetId:reference, jobId:reference,
}).strict();
const document = z.object({
  version:z.literal(1), nodes:z.array(node).max(2000),
  edges:z.array(z.object({id,source:id,target:id}).strict()).max(4000),
  viewport:z.object({x:position,y:position,zoom:z.number().finite().min(.1).max(4)}).strict(),
}).strict();
const empty = z.object({}).strict();
const resource = z.object({id}).strict();
const route = (method, path, schema, body = input => input) => ({method,path,schema,body});
const operations = Object.freeze({
  listProjects: route("GET", () => "/projects", empty),
  createProject: route("POST", () => "/projects", z.object({name:utf8(120,1)}).strict()),
  openProject: route("GET", i => "/projects/"+i.id, resource),
  saveProject: route("POST", i => "/projects/"+i.id+"/save", z.object({id,name:utf8(120,1),expectedRevision:z.number().int().positive(),document}).strict(), ({id,...body}) => body),
  deleteProject: route("DELETE", i => "/projects/"+i.id, resource),
  listChannels: route("GET", () => "/channels", empty),
  configureChannel: route("POST", () => "/channels", z.object({
    id:id.optional(),name:utf8(120,1),pluginId:utf8(200,1),baseUrl:utf8(2000,1),
    models:z.array(z.object({name:utf8(200,1),kind}).strict()).min(1).max(50),apiKey:utf8(8192).optional(),
  }).strict()),
  deleteChannel: route("DELETE", i => "/channels/"+i.id, resource),
  listAssets: route("GET", () => "/assets", empty),
  deleteAsset: route("DELETE", i => "/assets/"+i.id, resource),
  listPlugins: route("GET", () => "/plugins", empty),
  submitGeneration: route("POST", () => "/jobs", z.object({requestId:id,projectId:id,nodeId:id,channelId:id,model:utf8(200,1),kind,prompt:utf8(32000,1)}).strict()),
  listJobs: route("GET", i => "/jobs?projectId="+i.projectId, z.object({projectId:id}).strict()),
  getJob: route("GET", i => "/jobs/"+i.id, resource),
  cancelJob: route("POST", i => "/jobs/"+i.id+"/cancel", resource, () => ({})),
  listLibrary: route("GET", () => "/library", empty),
  saveLibraryEntry: route("POST", () => "/library", z.object({id:id.optional(),kind:z.enum(["prompt","skill"]),name:utf8(120,1),content:utf8(32000,1)}).strict()),
  deleteLibraryEntry: route("DELETE", i => "/library/"+i.id, resource),
});
const importSchema = z.object({name:utf8(240,1),bytes:z.instanceof(ArrayBuffer).refine(b => b.byteLength > 0 && b.byteLength <= 64*1024*1024)}).strict();
module.exports = { operations, importSchema, id };
