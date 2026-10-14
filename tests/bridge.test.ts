import { test } from "node:test";
import assert from "node:assert/strict";
import { createRequire } from "node:module";
const require=createRequire(import.meta.url);
const {operations,importSchema}=require("../electron/main/operations.cjs");
test("bridge operations reject extra authority and arbitrary paths",()=>{
  assert.equal(operations.openProject.schema.safeParse({id:crypto.randomUUID(),path:"outside"}).success,false);
  assert.equal(operations.configureChannel.schema.safeParse({name:"渠道",pluginId:"provider.openai",baseUrl:"https://example.com/v1",models:[{name:"model",kind:"text"}],userId:"unexpected"}).success,false);
  assert.equal(Object.hasOwn(operations,"executeShell"),false);
  assert.equal(Object.hasOwn(operations,"fetch"),false);
  assert.equal(importSchema.safeParse({name:"素材",bytes:new ArrayBuffer(0)}).success,false);
  assert.equal(importSchema.safeParse({name:"素材",bytes:new ArrayBuffer(1),path:"outside"}).success,false);
});
