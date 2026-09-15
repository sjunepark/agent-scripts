"use strict";
const assert = require("node:assert/strict"), fs = require("node:fs"), os = require("node:os"), path = require("node:path");
const { spawnSync } = require("node:child_process");
const { test, before, after } = require("node:test");
const { check } = require("../plugins/sjskills-maintenance/scripts/session-start.cjs");
const { classifyStatus, report, DAY } = require("../plugins/sjskills-maintenance/scripts/status.cjs");
const { observePlugin, provenance, mainRef, fetchJSON, RETRY } = require("../plugins/sjskills-maintenance/scripts/observation.cjs");
const { nativeExecutable, runProcess } = require("../plugins/sjskills-maintenance/scripts/process.cjs");
const repo = path.resolve(__dirname,".."), plugin = path.join(repo,"plugins/sjskills-maintenance");
const version = JSON.parse(fs.readFileSync(path.join(plugin,".codex-plugin/plugin.json"))).version;
const now = Date.now(), commit = "a".repeat(40);
const config = '[marketplaces.personal]\nsource_type = "git"\nsource = "https://github.com/sjunepark/agent-scripts.git"\nref = "main"\n';
const event = { hook_event_name:"SessionStart",source:"startup",cwd:repo };
const healthyPlugin = {plugin:true,findings:[],incomplete:[]};
function healthy(configured=true) {
  return { operation:"status",result:"success",error:null,status:{projectConfiguration:configured?"configured":"not-configured",...(configured?{projectRoot:repo}:{})},
    cliAdvisory:{comparison:"equal",runningVersion:"1.3.0",availableVersion:"1.3.0",freshness:"fresh",cached:false,observedAt:new Date(now).toISOString()},
    advisories:(configured?["global","project"]:["global"]).map(scope=>({scope,freshness:"fresh",cached:false,observedAt:new Date(now).toISOString(),findings:[]})) };
}
function temp(t) {
  const base=fs.realpathSync(os.tmpdir()), root=fs.mkdtempSync(path.join(base,"sjskills-check-"));
  t.after(()=>{const relative=path.relative(base,fs.realpathSync(root));assert.ok(relative&&!relative.startsWith("..")&&!path.isAbsolute(relative));fs.rmSync(root,{recursive:true,force:true});});
  return root;
}
function copy(from,to) {fs.mkdirSync(to,{recursive:true});for(const e of fs.readdirSync(from,{withFileTypes:true})){const a=path.join(from,e.name),b=path.join(to,e.name);if(e.isDirectory())copy(a,b);else fs.writeFileSync(b,fs.readFileSync(a));}}
function observation(root, overrides={}) {return {dataRoot:path.join(root,"data"),now,provenance:async()=>version,get:async url=>url.includes("/commits/")?{sha:commit}:{name:"sjskills-maintenance",version},...overrides};}
function cached(root, published=version) {fs.mkdirSync(root,{recursive:true});fs.writeFileSync(path.join(root,"plugin-observation.json"),JSON.stringify({key:`https://github.com/sjunepark/agent-scripts.git@main:${version}`,checkedAt:now,outcome:"success",version:published,commit}));}

test("fresh complete status is silent, with an inapplicable project and ahead CLI allowed",()=>{
  for(const configured of [true,false]) for(const comparison of ["equal","ahead"]){const v=healthy(configured);v.cliAdvisory.comparison=comparison;assert.equal(report([classifyStatus(v,now)]),null);}
});
test("stale, absent, unknown, disabled and malformed evidence cannot be healthy",()=>{
  const mutations=[v=>delete v.cliAdvisory,v=>v.cliAdvisory.comparison="uncomparable",v=>v.cliAdvisory.comparison="future",v=>v.cliAdvisory.runningVersion="dev",v=>v.cliAdvisory.observedAt=new Date(now+1).toISOString(),v=>v.advisories.pop(),v=>v.advisories.push(v.advisories[0]),v=>v.status.projectConfiguration="skipped",v=>v.status.projectConfiguration="unknown",v=>v.advisories[0].freshness="stale",v=>v.advisories[0].findings=null,v=>v.advisories[0].observedAt=new Date(now-DAY).toISOString(),v=>v.advisories[0].error="private auth secret",v=>v.result="unavailable"];
  for(const mutate of mutations){const value=healthy();mutate(value);const r=report([classifyStatus(value,now)]);assert.match(r.systemMessage,/incomplete/);assert.doesNotMatch(r.systemMessage,/secret/);assert.equal(r.hookSpecificOutput,undefined);}
});
test("known findings survive partial failure, additive fields and repeated checks",()=>{
  const value=healthy();value.newField=true;value.cliAdvisory.comparison="update";
  value.advisories[0].findings=[{category:"missing",skill:"sensitive name",reason:"missing"}];value.advisories[1].freshness="unavailable";
  const one=report([classifyStatus(value,now),{plugin:true,findings:["plugin update available"],incomplete:[]}]);
  assert.match(one.systemMessage,/CLI update available; global skills differ; plugin update available; check incomplete/);
  assert.doesNotMatch(one.systemMessage,/sensitive/);assert.deepEqual(one,report([classifyStatus(value,now),{plugin:true,findings:["plugin update available"],incomplete:[]}]));
});
test("adapter runs exactly one status in event cwd, retains nonzero output, ignores other events",async()=>{
  const calls=[];const options={platform:"darwin",arch:"arm64",now,resolve:()=>process.execPath,observe:async()=>healthyPlugin,run:async(exe,args,opts)=>{calls.push({exe,args,cwd:opts.cwd});return {code:0,stdout:JSON.stringify(healthy())};}};
  assert.equal(await check(event,options),null);assert.deepEqual(calls,[{exe:process.execPath,args:["--json","status"],cwd:repo}]);
  assert.equal(await check({...event,source:"resume"},options),null);
  for(const source of ["compact","clear","startup-extra"])assert.equal(await check({...event,source},options),null);
  assert.equal(await check({...event,hook_event_name:"SubagentStart"},options),null);assert.equal(calls.length,2);
  const value=healthy();value.cliAdvisory.comparison="update";
  const result=await check(event,{...options,run:async()=>({code:1,stdout:JSON.stringify(value)})});assert.match(result.systemMessage,/CLI update available.*status process/);
  assert.match((await check(event,{...options,platform:"linux"})).systemMessage,/unsupported/);
});
test("plugin observation handles expiry, rollback, failure cooldown and installed version changes",async t=>{
  const root=temp(t);let calls=0;const opts=observation(root,{get:async url=>{calls++;return url.includes("/commits/")?{sha:commit}:{name:"sjskills-maintenance",version};}});
  assert.deepEqual(await observePlugin(opts),healthyPlugin);assert.equal(calls,2);
  await observePlugin({...opts,now:now+DAY-1});assert.equal(calls,2);
  await observePlugin({...opts,now:now+DAY});assert.equal(calls,4);
  await observePlugin({...opts,now:now-1});assert.equal(calls,6);
  await observePlugin({...opts,provenance:async()=>"0.1.0+codex.changed"});assert.equal(calls,8);
  const fail={...opts,now:now+DAY*2,get:async()=>{calls++;throw Error("secret");}};
  assert.ok((await observePlugin(fail)).incomplete.length);const count=calls;
  await observePlugin({...fail,now:fail.now+RETRY-1});assert.equal(calls,count);
  await observePlugin({...fail,now:fail.now+RETRY});assert.equal(calls,count+1);
});
test("plugin immutable observation rejects bad source, remote content and unsafe cache; does not steal locks",async t=>{
  const root=temp(t);assert.ok(mainRef(config));assert.equal(mainRef(config.replace('"main"','"dev"')),false);assert.equal(mainRef(config+config),false);
  const opts=observation(root);fs.mkdirSync(opts.dataRoot);fs.mkdirSync(path.join(opts.dataRoot,"plugin-observation.lock"));
  let requests=0;assert.ok((await observePlugin({...opts,get:async()=>{requests++;}})).incomplete.length);assert.equal(requests,0);
  assert.ok(fs.existsSync(path.join(opts.dataRoot,"plugin-observation.lock")));
  const outside=path.join(root,"outside");fs.mkdirSync(outside);const linked=path.join(root,"linked");fs.symlinkSync(outside,linked,process.platform==="win32"?"junction":"dir");
  assert.ok((await observePlugin({...opts,dataRoot:linked})).incomplete.length);assert.deepEqual(fs.readdirSync(outside),[]);
  assert.ok((await observePlugin({...opts,dataRoot:path.join(root,"remote"),get:async()=>({sha:"../../bad"})})).incomplete.length);
  assert.ok((await observePlugin({...opts,provenance:async()=>{throw Error("disabled");}})).incomplete.length);
});
test("plugin cache contention preserves independent fresh observations",async t=>{
  const root=temp(t);let release;const gate=new Promise(r=>release=r);let entered;const started=new Promise(r=>entered=r);
  const opts=observation(root,{get:async url=>{entered();await gate;return url.includes("/commits/")?{sha:commit}:{name:"sjskills-maintenance",version};}});
  const first=observePlugin(opts);await started;assert.ok((await observePlugin(opts)).incomplete.length);release();assert.deepEqual(await first,healthyPlugin);assert.deepEqual(await observePlugin(opts),healthyPlugin);
});

let nativeRoot, executable;
before(()=>{nativeRoot=fs.mkdtempSync(path.join(fs.realpathSync(os.tmpdir()),"sjskills-native-"));executable=path.join(nativeRoot,process.platform==="win32"?"sjskills.exe":"sjskills");const r=spawnSync("go",["build","-o",executable,path.join(repo,"scripts/fixtures/startup-cli.go")],{encoding:"utf8",windowsHide:true});assert.equal(r.status,0,r.stderr);});
after(()=>{if(nativeRoot){assert.ok(path.basename(nativeRoot).startsWith("sjskills-native-"));fs.rmSync(nativeRoot,{recursive:true,force:true});}});
test("registered command executes controlled native checks in both native shells without protected writes",t=>{
  const root=temp(t), installed=path.join(root,"user's 한글 $workspace & tools","plugin"), data=path.join(root,"data"), home=path.join(root,"home"), bin=path.join(root,"bin");
  copy(plugin,installed);fs.mkdirSync(home);fs.mkdirSync(bin);cached(data);fs.writeFileSync(path.join(home,"config.toml"),config);
  for(const name of ["sjskills","codex"])fs.copyFileSync(executable,path.join(bin,name+(process.platform==="win32"?".exe":"")));
  const log=path.join(root,"commands.jsonl"), beforeConfig=fs.readFileSync(path.join(home,"config.toml"));
  const env={...process.env,PSExecutionPolicyPreference:"Restricted",PATH:bin+path.delimiter+process.env.PATH,CODEX_HOME:home,HOME:home,USERPROFILE:home,PLUGIN_ROOT:installed,PLUGIN_DATA:data,HOOK_FIXTURE_LOG:log,HOOK_FIXTURE_STATUS:JSON.stringify(healthy(false)),HOOK_FIXTURE_PLUGINS:JSON.stringify({installed:[{pluginId:"sjskills-maintenance@personal",enabled:true,installed:true,version,marketplaceSource:{sourceType:"git",source:"https://github.com/sjunepark/agent-scripts.git"}}]})};
  const hook=JSON.parse(fs.readFileSync(path.join(installed,"hooks/hooks.json"))).hooks.SessionStart[0].hooks[0];assert.equal(hook.timeout,40);assert.equal(hook.additionalContextLimit,undefined);
  const shells=process.platform==="win32"?[["powershell.exe",["-NoProfile","-NonInteractive","-Command",hook.commandWindows]],[process.env.COMSPEC||"cmd.exe",["/d","/s","/c",hook.commandWindows]]]:[["/bin/sh",["-c",hook.command]]];
  for(const [shell,args] of shells)for(const fallback of [false,true]){const shellEnv={...env};if(fallback){delete shellEnv.PLUGIN_ROOT;delete shellEnv.PLUGIN_DATA;shellEnv.CLAUDE_PLUGIN_ROOT=installed;shellEnv.CLAUDE_PLUGIN_DATA=data;}const result=spawnSync(shell,args,{env:shellEnv,cwd:root,input:JSON.stringify({...event,cwd:root}),encoding:"utf8",windowsHide:true,timeout:10000});assert.equal(result.status,0,result.stderr);if(process.platform==="linux")assert.match(JSON.parse(result.stdout).systemMessage,/unsupported/);else assert.equal(result.stdout,"");}
  if(process.platform!=="linux"){const commands=fs.readFileSync(log,"utf8").trim().split("\n").map(JSON.parse);assert.equal(commands.filter(c=>c.args[0]==="--json").length,shells.length*2);for(const c of commands)assert.ok(JSON.stringify(c.args)==='["--json","status"]'||JSON.stringify(c.args)==='["plugin","list","--marketplace","personal","--available","--json"]');}
  assert.deepEqual(fs.readFileSync(path.join(home,"config.toml")),beforeConfig);for(const protectedName of [".agents",".claude","plugins","auth.json"])assert.equal(fs.existsSync(path.join(home,protectedName)),false);assert.deepEqual(fs.readdirSync(data),["plugin-observation.json"]);
});
test("native resolution rejects a building wrapper without executing it",t=>{
  const root=temp(t);const name=process.platform==="win32"?"sjskills.exe":"sjskills";fs.writeFileSync(path.join(root,name),"#!/bin/sh\necho bad\n",{mode:0o755});assert.throws(()=>nativeExecutable("sjskills",{PATH:root}));assert.equal(nativeExecutable("sjskills",{PATH:nativeRoot}),fs.realpathSync(executable));
});
test("bounded process output and cancellation clean up owned descendants",async t=>{
  await assert.rejects(runProcess(process.execPath,["-e","process.stdout.write('x'.repeat(20000));setInterval(()=>{},1000)"],{cwd:repo,limit:1000,signal:AbortSignal.timeout(3000)}));
  const root=temp(t),pidFile=path.join(root,"pid");
  const code="const c=require('child_process').spawn(process.execPath,['-e','setInterval(()=>{},1000)'],{stdio:'ignore',windowsHide:true});require('fs').writeFileSync(process.argv[1],String(c.pid));setInterval(()=>{},1000)";
  const start=Date.now();await assert.rejects(runProcess(process.execPath,["-e",code,pidFile],{cwd:root,signal:AbortSignal.timeout(1500)}));assert.ok(Date.now()-start<5000);
  const pid=Number(fs.readFileSync(pidFile,"utf8"));await assertStopped(pid);
});
test("malformed hook input is bounded and never echoes untrusted content",()=>{
  for(const input of ["secret", "null", "[]", "x".repeat(1024*1024+1)]){const r=spawnSync(process.execPath,[path.join(plugin,"scripts/session-start.cjs")],{input,encoding:"utf8",windowsHide:true,timeout:5000});assert.equal(r.status,0,r.stderr);const value=JSON.parse(r.stdout);assert.match(value.systemMessage,/incomplete/);assert.doesNotMatch(value.systemMessage,/secret/);assert.equal(value.hookSpecificOutput,undefined);}
});

test("read-only provenance rejects disabled, mismatched and custom-source installations",async t=>{
  const root=temp(t);fs.writeFileSync(path.join(root,"config.toml"),config);
  const entry={pluginId:"sjskills-maintenance@personal",enabled:true,installed:true,version,marketplaceSource:{sourceType:"git",source:"https://github.com/sjunepark/agent-scripts.git"}};
  const opts={root:plugin,env:{CODEX_HOME:root},resolve:()=>process.execPath,run:async(exe,args)=>{assert.deepEqual(args,["plugin","list","--marketplace","personal","--available","--json"]);return {code:0,stdout:JSON.stringify({installed:[entry]})};}};
  assert.equal(await provenance(opts),version);
  for(const mutate of [e=>e.enabled=false,e=>e.version="0.1.0+codex.other",e=>e.marketplaceSource.sourceType="local",e=>e.marketplaceSource.source="https://example.com/repo"]){const changed=structuredClone(entry);mutate(changed);await assert.rejects(provenance({...opts,run:async()=>({code:0,stdout:JSON.stringify({installed:[changed]})})}));}
  fs.writeFileSync(path.join(root,"config.toml"),config.replace('"main"','"dev"'));await assert.rejects(provenance(opts));
});
test("metadata transport bounds responses, rejects redirects, and cancels network work",async t=>{
  const server=require("node:http").createServer((req,res)=>{
    if(req.url==="/large"){res.end("x".repeat(129*1024));return;}
    if(req.url==="/redirect"){res.writeHead(302,{Location:"/ok"});res.end();return;}
    if(req.url==="/hang")return;
    res.end(JSON.stringify({sha:commit}));
  });
  await new Promise(resolve=>server.listen(0,"127.0.0.1",resolve));t.after(()=>{server.closeAllConnections();server.close();});
  const url=`http://127.0.0.1:${server.address().port}`;
  assert.deepEqual(await fetchJSON(url+"/ok",AbortSignal.timeout(2000)),{sha:commit});
  await assert.rejects(fetchJSON(url+"/large",AbortSignal.timeout(2000)));
  await assert.rejects(fetchJSON(url+"/redirect",AbortSignal.timeout(2000)));
  await assert.rejects(fetchJSON(url+"/hang",AbortSignal.timeout(100)));
});
test("warm plugin evidence never suppresses fresh local skill findings",async t=>{
  const root=temp(t);const opts=observation(root);await observePlugin(opts);
  const value=healthy();const options={platform:"darwin",arch:"x64",now,resolve:()=>process.execPath,run:async()=>({code:0,stdout:JSON.stringify(value)}),observe:()=>observePlugin({...opts,get:async()=>{throw Error("must use cache");}})};
  assert.equal(await check(event,options),null);
  value.advisories[1].findings=[{category:"update",skill:"changed",reason:"local-modification"}];
  assert.match((await check(event,options)).systemMessage,/project skills differ/);
});
test("read-only cache errors preserve fresh results without modifying foreign state",async t=>{
  const root=temp(t),data=path.join(root,"data");fs.mkdirSync(data);const state=path.join(data,"plugin-observation.json");fs.mkdirSync(state);
  assert.ok((await observePlugin(observation(root))).incomplete.length);assert.ok(fs.statSync(state).isDirectory());
  const elsewhere=path.join(root,"elsewhere");fs.mkdirSync(elsewhere);const foreign=path.join(elsewhere,"foreign");fs.writeFileSync(foreign,"unchanged");
  const other=path.join(root,"other");fs.mkdirSync(other);fs.linkSync(foreign,path.join(other,"plugin-observation.json"));
  assert.ok((await observePlugin({...observation(root),dataRoot:other})).incomplete.length);assert.equal(fs.readFileSync(foreign,"utf8"),"unchanged");
});

test("mixed valid and invalid findings preserve known drift",()=>{
  const value=healthy();value.advisories[0].findings=[{category:"missing",skill:"known",reason:"missing"},{category:"unknown"}];
  const message=report([classifyStatus(value,now)]).systemMessage;
  assert.match(message,/global skills differ/);assert.match(message,/global verification/);
});
test("native parent exit cannot leave a pipe-holding descendant running",async t=>{
  const root=temp(t),file=path.join(root,"orphan-pid"),begin=Date.now();
  await runProcess(executable,["--orphan",process.execPath,file],{cwd:root,signal:AbortSignal.timeout(3000)}).catch(()=>{});
  assert.ok(Date.now()-begin<5500);const pid=Number(fs.readFileSync(file,"utf8"));
  await assertStopped(pid);
});

async function assertStopped(pid) {
  let running=true;
  for(let i=0;i<50;i++){
    try{process.kill(pid,0);if(process.platform==="linux"&&fs.readFileSync(`/proc/${pid}/stat`,"utf8").includes(") Z ")){running=false;break;}}
    catch(error){if(error.code==="ESRCH"||error.code==="ENOENT"){running=false;break;}throw error;}
    await new Promise(resolve=>setTimeout(resolve,20));
  }
  if(running)process.kill(pid);
  assert.equal(running,false,"owned descendant survived completion/cancellation");
}
