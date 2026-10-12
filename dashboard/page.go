package main

const page = `<!doctype html><html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Dashboard — Prototype</title>
<style>
:root{--bg:#f3f2ee;--panel:#fff;--ink:#1c1c1a;--mute:#6b6a65;--line:#dedcd5;--acc:#1f6f43;--warn:#b7791f;--bad:#b3261e;--log:#14171a;--logink:#d9dee3}
@media (prefers-color-scheme:dark){:root{--bg:#17181a;--panel:#212326;--ink:#ecebe7;--mute:#9b9a95;--line:#34363a;--log:#0d0f11}}
*{box-sizing:border-box}
html,body{height:100%}
body{margin:0;font:15px/1.4 system-ui,Segoe UI,sans-serif;background:var(--bg);color:var(--ink);min-height:100%;display:flex;flex-direction:column}
header{display:flex;align-items:center;gap:10px;padding:10px 16px;border-bottom:1px solid var(--line);background:var(--panel);flex-wrap:wrap}
header h1{font-size:17px;margin:0 12px 0 0}.prototype-tag{display:inline-block;font:700 10px/1.5 system-ui,Segoe UI,sans-serif;letter-spacing:.08em;border:1px solid var(--acc);border-radius:4px;color:var(--acc);padding:1px 5px;margin-left:8px;vertical-align:middle}
.workspace-status{font-size:13px;color:var(--mute)}
.dashboard-tools{position:relative}
.dashboard-tools summary{cursor:pointer;font-size:14px;padding:8px 12px;border:1px solid var(--line);border-radius:8px;list-style:none}
.settings-panel{position:absolute;right:0;top:calc(100% + 8px);z-index:2;width:300px;padding:16px;background:var(--panel);border:1px solid var(--line);border-radius:10px;box-shadow:0 8px 24px #0002;display:flex;flex-direction:column;gap:12px}
.settings-panel h2{font-size:15px;margin:0}.settings-panel p{font-size:13px;margin:0;color:var(--mute)}
.pill{font-size:13px;padding:4px 10px;border-radius:99px;border:1px solid var(--line)}
.pill.ok{color:var(--acc);border-color:var(--acc)} .pill.no{color:var(--warn);border-color:var(--warn)}
.sp{flex:1}
button{font:inherit;font-size:14px;padding:8px 14px;border-radius:8px;border:1px solid var(--line);background:var(--panel);color:var(--ink);cursor:pointer}
button:hover{border-color:var(--mute)} button:disabled{opacity:.4;cursor:default}
button.pri{background:var(--acc);border-color:var(--acc);color:#fff;font-weight:600}
button.stop{background:var(--bad);border-color:var(--bad);color:#fff}
button.fin{background:var(--acc);border-color:var(--acc);color:#fff}
button.fin.pulse{animation:p 1.2s infinite}@keyframes p{50%{box-shadow:0 0 0 6px #1f6f4340}}
.wrap{flex:1;display:flex;min-height:0;min-width:0}
.project-sidebar{width:290px;flex:none;border-right:1px solid var(--line);overflow:auto;background:var(--panel)}
.sidebar-head{padding:16px 16px 12px;border-bottom:1px solid var(--line)}
.sidebar-head h2{font-size:16px;margin:0 0 10px}.sidebar-head .bar{gap:6px}.sidebar-head button{padding:6px 10px;font-size:13px}
nav{padding:0 10px 10px}
nav h2{font-size:12px;text-transform:uppercase;letter-spacing:.06em;color:var(--mute);margin:14px 6px 6px}
.item{display:flex;gap:10px;align-items:center;padding:10px;border-radius:8px;cursor:pointer;border:1px solid transparent}
.item:hover{background:var(--bg)} .item.sel{border-color:var(--acc);background:var(--bg)}
.dot{width:11px;height:11px;border-radius:50%;background:var(--line);flex:none}
.dot.running{background:#2b7de9;animation:b 1s infinite} .dot.waiting{background:var(--acc);animation:b 1s infinite}
.dot.ok{background:var(--acc)} .dot.failed{background:var(--bad)} .dot.stopped{background:var(--warn)}
@keyframes b{50%{opacity:.35}}
.item .n{font-weight:600} .item .s{font-size:12px;color:var(--mute)}
main{flex:1;display:flex;flex-direction:column;min-width:0;max-width:100%;overflow:hidden;padding:16px;gap:10px}
main h2{margin:0;font-size:20px} .about{color:var(--mute);margin:0;max-width:900px}
.project-heading{display:flex;align-items:flex-start;gap:12px}.project-heading>div{flex:1}
.eyebrow{font-size:12px;color:var(--mute);text-transform:uppercase;letter-spacing:.06em;margin-bottom:4px}
.project-controls{padding:12px 0;border-bottom:1px solid var(--line)}
.run-output{flex:1;display:flex;flex-direction:column;min-height:260px;gap:8px}.run-output h3{font-size:14px;margin:0}.run-output .status{font-size:13px;color:var(--mute)}
.project-details{font-size:13px;color:var(--mute)}.project-details summary{cursor:pointer}.project-details dl{display:grid;grid-template-columns:auto 1fr;gap:6px 12px}.project-details dt{font-weight:600}.project-details dd{margin:0;overflow-wrap:anywhere}
.bar{display:flex;flex-wrap:wrap;gap:8px;align-items:center}
.status{font-weight:600}
pre{flex:1;min-height:200px;overflow-wrap:anywhere;margin:0;background:var(--log);color:var(--logink);padding:12px;border-radius:8px;overflow:auto;white-space:pre-wrap;font:13px/1.45 ui-monospace,Consolas,monospace}
.shot img{max-height:180px;border:1px solid var(--line);border-radius:6px}
.empty{color:var(--mute);margin:auto;text-align:center}
dialog{border:1px solid var(--line);border-radius:10px;background:var(--panel);color:var(--ink);max-width:560px;width:92vw}
dialog label{display:block;margin:10px 0 4px;font-size:13px;color:var(--mute)}
dialog input{width:100%;font:inherit;padding:8px;border:1px solid var(--line);border-radius:6px;background:var(--bg);color:var(--ink)}
.err{color:var(--bad);font-size:13px;min-height:1em}
.problems{color:var(--bad);font-size:13px;padding:0 6px}
@media (max-width:760px){body{height:auto}.wrap{display:block}main{display:flex;width:100%}main{overflow:visible}pre{min-height:260px}.project-sidebar{width:auto;max-height:38vh;border-right:0;border-bottom:1px solid var(--line)}.project-heading{flex-wrap:wrap}.settings-panel{position:fixed;left:12px;right:12px;top:70px;width:auto}}
</style></head><body>
<header>
 <h1>Dashboard <span class="prototype-tag">PROTOTYPE</span></h1>
 <span id="workspace-status" class="workspace-status">Connecting…</span>
 <span class="sp"></span>
 <details class="dashboard-tools"><summary>Dashboard settings</summary><div class="settings-panel">
  <h2>Shared browser engine</h2>
  <p>Used by browser projects in this workspace.</p>
  <span id="eng" class="pill">…</span>
  <button id="inst">Install browser engine</button>
 </div></details>
 <button class="stop" id="stopall" disabled>■ Stop all projects</button>
</header>
<div class="wrap">
 <aside class="project-sidebar">
  <div class="sidebar-head"><h2>Projects</h2><div class="bar">
   <button id="add">+ Add project</button><button id="rescan">Rescan</button>
   <button id="openall" title="Open the workspace containing project definitions">Open projects folder</button>
  </div></div>
  <nav id="nav" aria-label="Projects"></nav>
 </aside>
 <main id="main"><p class="empty">Pick a project on the left.</p></main>
</div>
<dialog id="dlg"><form method="dialog" id="f">
 <h3 style="margin:0 0 6px">Add a project you already have</h3>
 <p class="about" style="font-size:13px">Point at a folder on this PC and the program inside it. It shows up in the sidebar with a Run button.</p>
 <label>Name</label><input name="Name" placeholder="Round 2 playwright (my copy)" required>
 <label>Folder (full path)</label><input name="Folder" placeholder="C:\Users\dusti\Teams-Browser-Test\prototypes\round2\playwright" required>
 <label>Program in that folder</label><input name="Exe" placeholder="round2-demo.exe" required>
 <label>Arguments (optional)</label><input name="Args" placeholder="-smoke">
 <p class="err" id="ferr"></p>
 <div class="bar"><button class="pri" value="ok" id="fok">Add</button><button value="cancel" formnovalidate>Cancel</button></div>
</form></dialog>
<script>
let S={projects:[]},sel=null,next=0,logFor=null,logRun='',logSerial=0,logBusy=false;
try{sel=localStorage.getItem('sel')}catch(e){}
const $=id=>document.getElementById(id);
async function post(u){const r=await fetch(u,{method:'POST'});if(!r.ok){alert(await r.text());}refresh();return r}
const label={idle:'idle',running:'running',waiting:'waiting for you: click Finish',ok:'finished OK',failed:'failed',stopped:'stopped'};
function esc(s){return String(s).replace(/[&<>"]/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;'}[c]))}
function renderNav(){
 let h='',g=null;
 for(const p of S.projects){
  if(p.Group!==g){g=p.Group;h+='<h2>'+esc(g||'Projects')+'</h2>'}
  h+='<button type="button" class="item'+(p.ID===sel?' sel':'')+'" style="width:100%;text-align:left" aria-pressed="'+(p.ID===sel)+'" data-id="'+esc(p.ID)+'"><span class="dot '+p.State+'"></span><span><span class="n" style="display:block">'+esc(p.Name)+'</span><span class="s">'+esc(label[p.State])+(p.Current?' · '+esc(p.Current):'')+'</span></span></button>';
 }
 if(!S.projects.length)h='<p class="about" style="padding:8px">No projects found in '+esc(S.folder)+'.</p>';
 if(S.problems&&S.problems.length)h+='<div class="problems">'+S.problems.map(esc).join('<br>')+'</div>';
 const nav=$('nav');
 if(nav.innerHTML===h)return; // avoid rebuilding identical navigation on every poll
 nav.innerHTML=h;
 for(const el of document.querySelectorAll('.item'))el.onclick=()=>{sel=el.dataset.id;try{localStorage.setItem('sel',sel)}catch(e){};renderNav();renderMain(true)};
}
function cur(){return S.projects.find(p=>p.ID===sel)}
function renderMain(fresh){
 const p=cur();
 if(!p){$('main').innerHTML='<p class="empty">Pick a project on the left.</p>';return}
 const busy=p.State==='running'||p.State==='waiting';
 if(fresh||!$('log')||logFor!==p.ID){
  let h='<div class="project-heading"><div><div class="eyebrow">Selected project</div><h2>'+esc(p.Name)+'</h2><p class="about">'+esc(p.About||'')+'</p></div><button id="openp">Open project folder</button></div>';
  if(!p.ExeFound)h+='<p class="err">Program not found: '+esc(p.Exe)+' in '+esc(p.Folder)+'</p>';
  h+='<div class="bar project-controls" id="acts" aria-label="Project actions">';
  p.Actions.forEach((a,i)=>{h+='<button data-a="'+i+'" class="'+(a.primary?'pri':'')+'" title="'+esc(a.note||'')+'">'+esc(a.label)+'</button>'});
  h+='<span class="sp"></span><button class="fin" id="fin" title="Presses Enter in the program: a clean close">✔ Finish</button><button class="stop" id="stop">■ Stop</button></div>';
  h+='<section class="run-output" aria-label="Run output"><div class="bar"><h3>Run output</h3><span class="sp"></span><span class="status" id="st"></span></div><pre id="log" aria-label="Project log"></pre><div class="shot" id="shot"></div></section>';
  h+='<details class="project-details"><summary>Project details</summary><dl><dt>Folder</dt><dd>'+esc(p.Folder)+'</dd><dt>Program</dt><dd>'+esc(p.Exe)+'</dd><dt>Run logs</dt><dd>'+esc(S.logs||'runtime/logs')+'</dd></dl></details>';
  $('main').innerHTML=h;logFor=p.ID;next=0;logRun='';logSerial++;
  for(const b of document.querySelectorAll('#acts [data-a]'))b.onclick=()=>post('/api/run?p='+encodeURIComponent(p.ID)+'&a='+b.dataset.a);
  $('fin').onclick=()=>post('/api/finish?p='+encodeURIComponent(p.ID));
  $('stop').onclick=()=>post('/api/stop?p='+encodeURIComponent(p.ID));
  $('openp').onclick=()=>post('/api/open?p='+encodeURIComponent(p.ID));
  pollLog();
 }
 for(const b of document.querySelectorAll('#acts [data-a]'))b.disabled=busy;
 $('fin').disabled=p.State!=='waiting';$('fin').classList.toggle('pulse',p.State==='waiting');
 $('stop').disabled=!busy;
 $('st').textContent='Status: '+label[p.State]+(p.Current?' — '+p.Current:'')+(p.Since?' (since '+p.Since+')':'');
 const sh=$('shot');const want=p.HasShot?'1':'0';
 if(sh.dataset.v!==want+p.State){sh.dataset.v=want+p.State;sh.innerHTML=p.HasShot?'<div class="about" style="font-size:13px">Latest screenshot</div><img src="/api/shot?p='+encodeURIComponent(p.ID)+'&t='+Date.now()+'">':''}
}
async function pollLog(){
 if(!sel||logBusy)return;
 const id=sel,serial=logSerial,from=next,run=logRun;
 logBusy=true;
 try{
  const r=await fetch('/api/log?p='+encodeURIComponent(id)+'&from='+from+'&run='+encodeURIComponent(run));
  if(!r.ok)throw new Error('log request failed');
  const j=await r.json();
  if(id!==logFor||serial!==logSerial)return;
  const l=$('log');if(!l)return;
  if(j.reset)l.textContent='';
  if(j.lines.length){const end=l.scrollTop+l.clientHeight>=l.scrollHeight-30;l.textContent+=j.lines.join('\n')+'\n';if(end||j.reset)l.scrollTop=l.scrollHeight}
  next=j.next;logRun=j.run;
 }catch(e){}finally{logBusy=false}
}
async function refresh(){
 try{const r=await fetch('/api/state');S=await r.json()}catch(e){$('eng').textContent='Dashboard closed: reopen Dashboard.exe';$('eng').className='pill no';return}
 if(!S.projects.some(p=>p.ID===sel))sel=S.projects.length?S.projects[0].ID:null;
 const running=S.projects.filter(p=>p.State==='running'||p.State==='waiting').length;
 $('workspace-status').textContent=S.projects.length+' project'+(S.projects.length===1?'':'s')+' · '+running+' running';
 $('stopall').disabled=!running;
 $('eng').textContent=S.engine?'Browser engine: installed':'Browser engine: not installed yet';
 $('eng').className='pill '+(S.engine?'ok':'no');
 $('inst').textContent=S.engine?'Reinstall engine':'Install browser engine';
 $('inst').className=S.engine?'':'pri';
 renderNav();renderMain(false);
}
$('inst').onclick=async()=>{const r=await fetch('/api/install',{method:'POST'});if(!r.ok)return alert(await r.text());const j=await r.json();sel=j.project;renderNav();renderMain(true);refresh()};
$('stopall').onclick=()=>post('/api/stopall');
$('rescan').onclick=()=>post('/api/rescan');
$('openall').onclick=()=>post('/api/open');
$('add').onclick=()=>{$('ferr').textContent='';$('dlg').showModal()};
$('fok').onclick=async e=>{e.preventDefault();const f=$('f');if(!f.reportValidity())return;
 const d=Object.fromEntries(new FormData(f));const r=await fetch('/api/add',{method:'POST',body:JSON.stringify(d)});
 if(!r.ok){$('ferr').textContent=await r.text();return}
 const j=await r.json();sel=j.id;$('dlg').close();f.reset();await refresh();renderMain(true)};
setInterval(()=>{refresh();pollLog()},1000);refresh();
</script></body></html>`
