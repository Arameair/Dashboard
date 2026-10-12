// UI-only. Each sandboxed iframe is mounted exactly once; focus is a CSS class.
const grid=document.querySelector('#grid'),all=document.querySelector('#all');
const sources=new Map([['a','telemetry-alpha-a'],['b','telemetry-alpha-b']]);
const maxEntries=120;
for(const tile of grid.querySelectorAll('.tile')){
  const output=tile.querySelector('.log');let entries=[];
  const show=()=>{output.replaceChildren(...entries.map(x=>{const p=document.createElement('div');p.textContent=x;return p}));output.scrollTop=output.scrollHeight};
  tile.addEventListener('click',event=>{
    if(event.target.classList.contains('focus')){grid.classList.add('focused');for(const t of grid.querySelectorAll('.tile'))t.classList.toggle('active',t===tile);all.hidden=false;all.focus();return}
    if(event.target.classList.contains('tab')){const tab=event.target.dataset.tab;for(const b of tile.querySelectorAll('.tab')){b.classList.toggle('current',b===event.target);b.setAttribute('aria-pressed',String(b===event.target))}tile.classList.toggle('logs-only',tab==='logs');tile.querySelector('.updates').hidden=tab!=='updates';tile.classList.toggle('updates-only',tab==='updates')}
  });
  tile.appendProperty=message=>{entries=[...entries, message].slice(-maxEntries);show()};
}
window.addEventListener('message',event=>{
  // For this local synthetic demo only: accept messages only from the exact existing iframe window.
  const tile=[...grid.querySelectorAll('.tile')].find(t=>t.querySelector('iframe').contentWindow===event.source);
  if(!tile||event.data?.kind!=='synthetic-workspace-event'||event.data?.id!==sources.get(tile.id)||typeof event.data?.message!=='string')return;
  const msg=event.data.message.slice(0,180);tile.appendProperty(msg);
  tile.querySelector('.state').textContent=msg.includes('DISCONNECTED')?'SIMULATED DISCONNECTED':'SIMULATED';
});
all.addEventListener('click',()=>{grid.classList.remove('focused');for(const t of grid.querySelectorAll('.tile'))t.classList.remove('active');all.hidden=true;grid.querySelector('.tile.active .focus')?.focus();grid.querySelector('#a .focus').focus()});
for(const b of document.querySelectorAll('.tab.current'))b.setAttribute('aria-pressed','true');
