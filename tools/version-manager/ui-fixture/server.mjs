import http from 'node:http';import {readFile} from 'node:fs/promises';import {fileURLToPath} from 'node:url';import {resolve,dirname} from 'node:path';import {fresh,logPage,remove,register} from './scenarios.mjs';
const staticDir=resolve(dirname(fileURLToPath(import.meta.url)),'../static');
export function makeServer(){const state=fresh();const assets=new Map([['/','index.html'],['/index.html','index.html'],['/app.js','app.js'],['/styles.css','styles.css']]);
return http.createServer(async(req,res)=>{res.setHeader('X-Alpha-Fixture','synthetic-only');res.setHeader('Cache-Control','no-store');res.setHeader('X-Content-Type-Options','nosniff');
const send=(status,data)=>{res.writeHead(status,{'Content-Type':'application/json; charset=utf-8'});res.end(JSON.stringify(data))};const fail=(status,error,message=error)=>send(status,{error,message});
const path=new URL(req.url||'/', 'http://127.0.0.1').pathname;
if(path==='/__fixture/reset'&&req.method==='POST'){Object.assign(state,fresh());return send(200,{ok:true,mock:true})}
if(path==='/__fixture/fault'&&req.method==='POST'){let obj;try{obj=JSON.parse(await body(req))}catch{return fail(400,'BAD_JSON')}for(const name of ['failLog','failStatus','rejectAdd','disconnected'])if(typeof obj[name]==='boolean')state[name]=obj[name];return send(200,{ok:true,mock:true})}
if(state.disconnected&&path.startsWith('/api/'))return fail(503,'DISCONNECTED');
if(path==='/api/versions'&&req.method==='GET')return send(200,{versions:state.versions,selectedId:state.selectedId,runningId:state.runningId});
if(path==='/api/status'&&req.method==='GET')return state.failStatus?fail(500,'SYNTHETIC_STATUS_FAILURE'):send(200,{state:'mock-only',selectedId:state.selectedId,runningId:state.runningId,lastKnownGoodId:state.lastKnownGoodId,pid:null,port:null,message:'MOCK ONLY — no executable is running'});
if(path==='/api/log'&&req.method==='GET'){if(state.failLog)return fail(500,'SYNTHETIC_LOG_FAILURE');const n=Number(new URL(req.url,'http://127.0.0.1').searchParams.get('cursor')||'0');return send(200,logPage(state,Number.isSafeInteger(n)&&n>=0?n:0))}
if(path==='/api/select'&&req.method==='POST'){let b;try{b=JSON.parse(await body(req))}catch{return fail(400,'BAD_JSON')}if(!state.versions.some(x=>x.id===b.id))return fail(404,'NOT_FOUND');state.selectedId=b.id;return send(200,{ok:true})}
if(path==='/api/versions'&&req.method==='POST'){let b;try{b=JSON.parse(await body(req))}catch{return fail(400,'BAD_JSON')}const result=register(state,b);return result.error?fail(result.status,result.error):send(result.status,result.body)}
if(path.startsWith('/api/versions/')&&req.method==='DELETE'){const result=remove(state,decodeURIComponent(path.slice('/api/versions/'.length)));return result.error?fail(result.status,result.error):send(result.status,result.body)}
if(path==='/api/approval-preview'&&req.method==='GET')return fail(422,'BUILD_UNVERIFIED');
if(path==='/api/approvals'&&req.method==='POST')return fail(403,'APPROVAL_REQUIRED');
if(['/api/launch','/api/stop','/api/rollback'].includes(path)&&req.method==='POST')return fail(501,'NOT_IMPLEMENTED','MOCK ONLY — no launch, stop, or rollback');
if(req.method==='GET'&&assets.has(path)){try{const file=await readFile(resolve(staticDir,assets.get(path)));const type=path.endsWith('.js')?'text/javascript':path.endsWith('.css')?'text/css':'text/html';res.writeHead(200,{'Content-Type':type+'; charset=utf-8'});return res.end(file)}catch{return fail(404,'STATIC_UNAVAILABLE')}}
return fail(404,'NOT_FOUND');
})}
async function body(req){let result='';for await(const chunk of req){result+=chunk;if(result.length>16384)throw Error('too large')}return result}
if(process.argv[1]&&resolve(process.argv[1])===fileURLToPath(import.meta.url)){const port=Number(process.env.ALPHA_FIXTURE_PORT||0);if(port&&!Number.isInteger(port))throw Error('Invalid port');makeServer().listen(port,'127.0.0.1',function(){console.log('MOCK ONLY — Dashboard Alpha UI fixture http://127.0.0.1:'+this.address().port);console.log('No real process launch, browser control, sign-in or GitHub fetch.')})}
