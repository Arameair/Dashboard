package main
import (
 "context"
 "encoding/json"
 "flag"
 "fmt"
 "io"
 "log"
 "net"
 "net/http"
 "net/url"
 "os"
 "os/exec"
 "path/filepath"
 "runtime"
 "strings"
 "sync"
 "time"
)
var root=flag.String("projects-root",`C:\Dev\Projects`,"folder containing independent Alpha checkouts")
var port=flag.String("port","18999","Project Hub local port")
var httpClient=&http.Client{Timeout:2*time.Second}
type Project struct{ID string `json:"id"`;Name string `json:"name"`;Repo string `json:"repo"`;Path string `json:"path"`;Branch string `json:"branch"`;Commit string `json:"commit"`;Ready bool `json:"ready"`;Error string `json:"error"`;URL string `json:"url"`;Status string `json:"status"`;PID int `json:"pid"`;Logs string `json:"logs"`}
type Runner struct{mu sync.Mutex; cmd *exec.Cmd; status string; logs string; err string; running bool; pid int; started time.Time}
var runners=map[string]*Runner{"dashboard":{status:"stopped"},"teams":{status:"stopped"}}
type Spec struct{id,name,repo,dir,branch,url string;port int}
func specs()[]Spec{return []Spec{
 {"dashboard","Dashboard","https://github.com/Arameair/Dashboard",filepath.Join(*root,"Dashboard-Alpha-Owner-Test"),"experiment/alpha-v5-owner-smoke-20261013","http://127.0.0.1:18877/",18877},
 {"teams","Teams","https://github.com/Arameair/Teams",filepath.Join(*root,"Teams-Alpha"),"alpha","http://127.0.0.1:18998/",18998},
}}
func spec(id string)(Spec,bool){for _,s:=range specs(){if s.id==id{return s,true}};return Spec{},false}
func git(dir string,args ...string)string{ctx,cancel:=context.WithTimeout(context.Background(),2*time.Second);defer cancel();a:=append([]string{"-C",dir},args...);b,err:=exec.CommandContext(ctx,"git",a...).Output();if err!=nil{return ""};return strings.TrimSpace(string(b))}
func verify(s Spec)(string,string){
 if st,err:=os.Stat(s.dir);err!=nil||!st.IsDir(){return "", "Local Alpha checkout missing: "+s.dir}
 branch:=git(s.dir,"branch","--show-current")
 if branch!=s.branch{return "",fmt.Sprintf("Expected branch %q, found %q. Use a separate clean clone.",s.branch,branch)}
 remote:=git(s.dir,"remote","get-url","origin")
 if !strings.Contains(strings.ToLower(remote),"arameair/"+strings.ToLower(s.name)){return "", "Repository origin mismatch; refusing launch"}
 sha:=git(s.dir,"rev-parse","--short","HEAD")
 if sha==""{return "", "No Git commit found"}
 return sha,""
}
func accessible(u string)bool{resp,e:=httpClient.Get(u);if e!=nil{return false};defer resp.Body.Close();return resp.StatusCode<500}
func status(s Spec)Project{
 sha,errText:=verify(s)
 r:=runners[s.id];r.mu.Lock();defer r.mu.Unlock()
 state:=r.status
 if errText!="" && state!="running"&&state!="building"{state="not ready"}
 p:=Project{ID:s.id,Name:s.name,Repo:s.repo,Path:s.dir,Branch:s.branch,Commit:sha,Ready:errText=="",Error:errText,URL:s.url,Status:state,PID:r.pid,Logs:r.logs}
 if r.err!=""{p.Error=r.err}
 return p
}
func logline(r *Runner,v string){r.logs+=time.Now().Format("15:04:05")+" "+v+"\n";if len(r.logs)>18000{r.logs=r.logs[len(r.logs)-18000:]}}
func apiError(w http.ResponseWriter,code int,msg string){w.Header().Set("Content-Type","application/json");w.WriteHeader(code);_ =json.NewEncoder(w).Encode(map[string]string{"error":msg})}
func allowed(w http.ResponseWriter,req *http.Request)bool{
 if req.Header.Get("Content-Type")!="application/json"{apiError(w,415,"application/json required");return false}
 if req.Header.Get("Sec-Fetch-Site")=="cross-site"{apiError(w,403,"cross-site request blocked");return false}
 if o:=req.Header.Get("Origin");o!="" {
  u,e:=url.Parse(o);if e!=nil||u.Host!="127.0.0.1:"+*port||u.Scheme!="http"{apiError(w,403,"origin mismatch");return false}
 }
 return true
}
func launch(id string){
 s,_:=spec(id);r:=runners[id]
 func(){r.mu.Lock();defer r.mu.Unlock();r.status="building";r.err="";logline(r,"Building "+s.name+" from local Alpha "+s.branch)}()
 // Never mutate the tracked checkout, only create ignored builds and isolated test data in the ProjectHub checkout.
 ownDir,err:=os.Getwd();if err!=nil {fail(r,err.Error());return}
 output:=filepath.Join(ownDir,"runtime","alpha-experment01","bin")
 data:=filepath.Join(ownDir,"runtime","alpha-experment01","data",id)
 if err=os.MkdirAll(output,0700);err!=nil{fail(r,err.Error());return}
 if err=os.MkdirAll(data,0700);err!=nil{fail(r,err.Error());return}
 binary:=filepath.Join(output,id+"-alpha.exe")
 var args []string
 if id=="dashboard"{args=[]string{"build","-mod=vendor","-o",binary,"."}}else{args=[]string{"build","-mod=readonly","-o",binary,"."}}
 cmd:=exec.Command("go",args...);if id=="dashboard"{cmd.Dir=filepath.Join(s.dir,"dashboard")}else{cmd.Dir=s.dir};cmd.Env=append(os.Environ(),"GOTOOLCHAIN=local","GOPROXY=off")
 out,e:=cmd.CombinedOutput()
 if e!=nil{fail(r,fmt.Sprintf("Build failed (no dependency downloads): %v\n%s",e,string(out)));return}
 // Explicit collision protection: never adopt a pre-existing instance.
 listener,e:=net.Listen("tcp",fmt.Sprintf("127.0.0.1:%d",s.port));if e!=nil{fail(r,"Test port occupied: "+e.Error());return};listener.Close()
 var launchArgs []string
 if id=="dashboard"{empty:=filepath.Join(data,"definitions");if e=os.MkdirAll(empty,0700);e!=nil{fail(r,e.Error());return};launchArgs=[]string{"-port",fmt.Sprint(s.port),"-data",data,"-projects",empty,"-no-open"}}else{launchArgs=[]string{"-addr",fmt.Sprintf("127.0.0.1:%d",s.port),"-data",data,"-root",s.dir,"-no-browser"}}
 child:=exec.Command(binary,launchArgs...);child.Dir=s.dir;child.Env=append(os.Environ(),"GOTOOLCHAIN=local")
 if e=child.Start();e!=nil{fail(r,"Start failed: "+e.Error());return}
 r.mu.Lock();r.cmd=child;r.running=true;r.pid=child.Process.Pid;r.status="starting";logline(r,"Launched PID "+fmt.Sprint(r.pid)+" on "+s.url);r.mu.Unlock()
 // Wait in a separate goroutine; never overwrite another launch's state.
 go func(){e:=child.Wait();r.mu.Lock();defer r.mu.Unlock();if r.cmd==child{r.cmd=nil;r.pid=0;r.running=false;if r.status!="stopping"{r.err=fmt.Sprint("Process exited: ",e);r.status="exited"}else{r.status="stopped"};logline(r,r.status)}}()
 for i:=0;i<40;i++{time.Sleep(250*time.Millisecond);r.mu.Lock();active:=r.cmd==child&&r.running;r.mu.Unlock();if !active{return};if accessible(s.url){r.mu.Lock();if r.cmd==child {r.status="running";logline(r,"HTTP service responding")};r.mu.Unlock();return}}
 fail(r,"Process started, but readiness timed out; use Stop before retry")
}
func fail(r *Runner,message string){r.mu.Lock();defer r.mu.Unlock();r.err=message;if !r.running{r.status="failed"};logline(r,message)}
func stop(id string)error{
 r:=runners[id];r.mu.Lock();if r.cmd==nil||r.cmd.Process==nil{r.mu.Unlock();return fmt.Errorf("No process owned by this prototype is running")};pid:=r.pid;r.status="stopping";logline(r,"Stopping owned process "+fmt.Sprint(pid));r.mu.Unlock()
 var err error
 if runtime.GOOS=="windows"{b,e:=exec.Command("taskkill","/PID",fmt.Sprint(pid),"/T","/F").CombinedOutput();if e!=nil{err=fmt.Errorf("taskkill: %v %s",e,string(b))}}else{r.mu.Lock();if r.cmd!=nil{err=r.cmd.Process.Kill()};r.mu.Unlock()}
 if err!=nil{fail(r,"Stop error: "+err.Error())};return err
}
func main(){
 flag.Parse();mux:=http.NewServeMux()
 mux.HandleFunc("GET /",func(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/"{http.NotFound(w,r);return};w.Header().Set("Content-Type","text/html; charset=utf-8");w.Header().Set("Cache-Control","no-store");io.WriteString(w,page)})
 mux.HandleFunc("GET /api/projects",func(w http.ResponseWriter,r *http.Request){w.Header().Set("Content-Type","application/json");w.Header().Set("Cache-Control","no-store");list:=[]Project{};for _,s:=range specs(){list=append(list,status(s))};json.NewEncoder(w).Encode(list)})
 mux.HandleFunc("POST /api/launch",func(w http.ResponseWriter,r *http.Request){if !allowed(w,r){return};id:=r.URL.Query().Get("id");s,ok:=spec(id);if !ok{apiError(w,404,"Unknown project");return};_,vErr:=verify(s);if vErr!=""{apiError(w,409,vErr);return}
  run:=runners[id];run.mu.Lock();if run.running||run.status=="building"{run.mu.Unlock();apiError(w,409,"Project already starting/running");return};run.status="building";run.mu.Unlock();go launch(id);json.NewEncoder(w).Encode(map[string]bool{"ok":true})})
 mux.HandleFunc("POST /api/stop",func(w http.ResponseWriter,r *http.Request){if !allowed(w,r){return};id:=r.URL.Query().Get("id");if _,ok:=spec(id);!ok{apiError(w,404,"Unknown project");return};if e:=stop(id);e!=nil{apiError(w,409,e.Error());return};json.NewEncoder(w).Encode(map[string]bool{"ok":true})})
 fmt.Println("Project Hub control center: http://127.0.0.1:"+*port+"/")
 fmt.Println("No automatic Git checkout, branch change, download, or profile takeover.")
 log.Fatal(http.ListenAndServe("127.0.0.1:"+*port,mux))
}
