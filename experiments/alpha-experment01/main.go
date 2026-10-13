package main
import("context";"encoding/json";"flag";"fmt";"io";"log";"net/http";"os";"os/exec";"path/filepath";"strings";"time")
var root=flag.String("projects-root","C:\\Dev\\Projects","local project checkouts")
var port=flag.String("port","18999","local port")
var client=&http.Client{Timeout:2*time.Second}
type Project struct{ID string `json:"id"`;Name string `json:"name"`;Repo string `json:"repo"`;Path string `json:"path"`;Local bool `json:"local"`;Branch string `json:"branch"`;Commit string `json:"commit"`}
func git(dir string,a ...string)string{c,x:=context.WithTimeout(context.Background(),time.Second);defer x();args:=append([]string{"-C",dir},a...);b,e:=exec.CommandContext(c,"git",args...).Output();if e!=nil{return ""};return strings.TrimSpace(string(b))}
func project(id,name,url,dir string)Project{p:=Project{ID:id,Name:name,Repo:url,Path:dir};if s,e:=os.Stat(dir);e==nil&&s.IsDir(){p.Local=true;p.Branch=git(dir,"branch","--show-current");p.Commit=git(dir,"rev-parse","--short","HEAD")};return p}
func fetch(path string)(*http.Response,error){return client.Get("http://127.0.0.1:18777"+path)}
func alive()bool{r,e:=fetch("/api/ping");if e!=nil{return false};defer r.Body.Close();b,_:=io.ReadAll(io.LimitReader(r.Body,120));return r.StatusCode==200&&strings.Contains(string(b),"dashboard")}
func relay(w http.ResponseWriter,path,ct string){r,e:=fetch(path);if e!=nil{http.Error(w,"Existing Alpha offline",503);return};defer r.Body.Close();if r.StatusCode!=200{w.WriteHeader(r.StatusCode);return};w.Header().Set("Content-Type",ct);w.Header().Set("Cache-Control","no-store");io.Copy(w,io.LimitReader(r.Body,2<<20))}
func main(){flag.Parse();mux:=http.NewServeMux()
mux.HandleFunc("GET /",func(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/"{http.NotFound(w,r);return};w.Header().Set("Content-Type","text/html; charset=utf-8");w.Header().Set("Cache-Control","no-store");io.WriteString(w,page)})
mux.HandleFunc("GET /api/projects",func(w http.ResponseWriter,r *http.Request){dir:=filepath.Join(*root,"Dashboard-Alpha-Owner-Test");if _,e:=os.Stat(dir);os.IsNotExist(e){dir=filepath.Join(*root,"Dashboard")};p:=[]Project{project("dashboard","Dashboard","https://github.com/Arameair/Dashboard",dir),project("teams","Teams","https://github.com/Arameair/Teams",filepath.Join(*root,"Teams"))};w.Header().Set("Content-Type","application/json");w.Header().Set("Cache-Control","no-store");json.NewEncoder(w).Encode(map[string]any{"projects":p,"alphaOnline":alive()})})
mux.HandleFunc("GET /api/alpha/state",func(w http.ResponseWriter,r *http.Request){relay(w,"/api/state","application/json")})
mux.HandleFunc("GET /api/alpha/log",func(w http.ResponseWriter,r *http.Request){relay(w,"/api/log?id=owner-alpha-smoke","text/plain")})
mux.HandleFunc("GET /api/alpha/shot",func(w http.ResponseWriter,r *http.Request){relay(w,"/api/shot?id=owner-alpha-smoke","image/jpeg")})
fmt.Println("alpha-experment01 at http://127.0.0.1:"+*port+"/");fmt.Println("Read-only prototype; existing Alpha is unchanged.");log.Fatal(http.ListenAndServe("127.0.0.1:"+*port,mux))}
