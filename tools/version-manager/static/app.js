'use strict';
(() => {
 const $ = id => document.getElementById(id);
 const ui = {versions:[],selectedId:null,runningId:null,status:{state:'disconnected'},connected:false,busy:false,logCursor:0,events:[]};
 const text = (id,value) => {$(id).textContent=String(value??'—')};
 const label = v => v ? (v.name||v.id||'Unnamed') : 'None';
 const current = () => ui.versions.find(v=>v.id===ui.selectedId);
 const safeError = err => String(err?.message || err || 'Unknown error').slice(0,350);
 async function api(path,method='GET',body) {
   const controller=new AbortController(), timeout=setTimeout(()=>controller.abort(),8000);
   try{
     const res=await fetch(path,{method,headers:body?{'Content-Type':'application/json'}:{},body:body?JSON.stringify(body):undefined,signal:controller.signal,cache:'no-store'});
     const raw=await res.text(); let data;
     try{data=raw?JSON.parse(raw):{}}catch{data={message:raw.slice(0,300)}}
     if(!res.ok)throw Error(data.error||data.message||('HTTP '+res.status));
     return data;
   }finally{clearTimeout(timeout)}
 }
 function failure(message){$('error').hidden=false;text('error',message)}
 function clearFailure(){$('error').hidden=true;text('error','')}
 function setConnected(value){ui.connected=value;text('connection',value?'Connected to manager':'Manager disconnected')}
 function render(){
   const catalog=$('catalog'), filter=$('search').value.trim().toLowerCase(), matched=ui.versions.filter(v=>[v.name,v.id,v.channel,v.branch,v.sha].some(x=>String(x||'').toLowerCase().includes(filter)));
   catalog.replaceChildren();
   for(const v of matched){
     const b=document.createElement('button');b.type='button';b.className='version';b.setAttribute('role','option');b.setAttribute('aria-selected',String(ui.selectedId===v.id));b.dataset.id=v.id;
     const name=document.createElement('strong');name.textContent=label(v);
     const detail=document.createElement('span');detail.textContent=[v.channel||'Unknown channel',v.branch||'no branch',v.sha||'no SHA'].join(' · ');
     const flags=document.createElement('span');flags.textContent=[v.installed?'Installed':'Source only',v.runnable?'Runnable':'NOT RUNNABLE',v.protected?'Protected':'',v.id===ui.runningId?'RUNNING':'',v.managed?'Managed':''].filter(Boolean).join(' · ');
     b.append(name,detail,flags);
     b.addEventListener('click',()=>runAction('/api/select','POST',{id:v.id}));
     catalog.append(b)
   }
   $('empty').hidden=matched.length>0;
   const selected=current(), st=String(ui.status.state||'unknown').toLowerCase();
   text('selected',selected?label(selected)+' · '+(selected.sha||'no SHA'):'None');
   text('running',label(ui.versions.find(v=>v.id===ui.runningId))+(ui.runningId?' · '+ui.runningId:''));
   text('state',ui.status.state||'Unknown');
   text('process',[ui.status.pid?'PID '+ui.status.pid:null,ui.status.port?'Port '+ui.status.port:null].filter(Boolean).join(' / ')||'—');
   text('status-message',ui.status.message||'');
   const enabled=ui.connected&&!ui.busy;
   $('launch').disabled=!enabled||!selected||!selected.runnable||!selected.installed||!!ui.runningId||['running','starting','stopping'].includes(st);
   $('remove').disabled=!enabled||!selected||!!selected.protected||selected.id===ui.runningId;
   $('stop').disabled=!enabled||!ui.runningId||!['running','starting','stopping'].includes(st);
   $('rollback').disabled=!enabled||!ui.status.lastKnownGoodId||ui.status.lastKnownGoodId===ui.runningId||['starting','stopping'].includes(st);
   $('add-form').querySelector('button[type=submit]').disabled=!enabled;
 }
 function renderEvents(){
   const list=$('events');list.replaceChildren();
   for(const e of ui.events.slice(-150)){
     const li=document.createElement('li');
     li.textContent=[e.at||'',e.level||'',e.kind||'',e.versionId||'',e.message||''].map(x=>String(x).slice(0,300)).join(' | ').slice(0,600);
     list.append(li)
   }
 }
 async function refresh(){
   if(ui.busy)return;
   try{
     const [catalog,status]=await Promise.all([api('/api/versions'),api('/api/status')]);
     if(!Array.isArray(catalog.versions))throw Error('Invalid versions response');
     ui.versions=catalog.versions;ui.selectedId=catalog.selectedId??null;ui.runningId=status.runningId??catalog.runningId??null;ui.status=status;
     setConnected(true);clearFailure();render();await refreshLog();
   }catch(err){setConnected(false);failure('Status unavailable: '+safeError(err));render()}
 }
 async function refreshLog(){
   try{
     const res=await api('/api/log?cursor='+encodeURIComponent(ui.logCursor));
     if(!Array.isArray(res.events)||!Number.isSafeInteger(res.next)||res.next<0)throw Error('Invalid log response');
     if(res.next<ui.logCursor){ui.events=[]}
     ui.events.push(...res.events);ui.events=ui.events.slice(-150);ui.logCursor=res.next;renderEvents()
   }catch(err){setConnected(false);failure('Log polling failed; controls disabled: '+safeError(err));render()}
 }
 async function runAction(path,method='POST',body,confirmMessage){
   if(!ui.connected||ui.busy)return;
   if(confirmMessage&&!window.confirm(confirmMessage))return;
   ui.busy=true;clearFailure();render();
   try{await api(path,method,body);ui.busy=false;await refresh()}
   catch(err){ui.busy=false;failure('Operation failed: '+safeError(err));await refresh();failure('Operation failed: '+safeError(err))}
   finally{render()}
 }
 $('search').addEventListener('input',render);
 $('add-form').addEventListener('submit',e=>{
   e.preventDefault();
   const form=e.currentTarget;if(!form.reportValidity())return;
   const d=Object.fromEntries(new FormData(form));
   runAction('/api/versions','POST',d).then(()=>{if($('error').hidden)form.reset()})
 });
 $('launch').addEventListener('click',()=>{const s=current();if(s)runAction('/api/launch','POST',{id:s.id})});
 $('remove').addEventListener('click',()=>{const s=current();if(s&&!s.protected&&s.id!==ui.runningId)runAction('/api/versions/'+encodeURIComponent(s.id),'DELETE',undefined,'Unregister '+label(s)+'? This removes the catalog entry ONLY. No build files, profiles, or data are deleted.')});
 $('stop').addEventListener('click',()=>runAction('/api/stop','POST',undefined,'Stop the managed running version?'));
 $('rollback').addEventListener('click',()=>runAction('/api/rollback','POST',undefined,'Rollback to the last known good build?'));
 $('reload').addEventListener('click',refresh);
 $('clear-log').addEventListener('click',()=>{ui.events=[];renderEvents()});
 $('catalog').addEventListener('keydown',e=>{if(e.key!=='ArrowDown'&&e.key!=='ArrowUp')return;const options=[...$('catalog').querySelectorAll('.version')];const pos=options.indexOf(document.activeElement);const next=options[Math.min(options.length-1,Math.max(0,pos+(e.key==='ArrowDown'?1:-1)))];if(next){e.preventDefault();next.focus()}});
 refresh();setInterval(refresh,3000)
})();
