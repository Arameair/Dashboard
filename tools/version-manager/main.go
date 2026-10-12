package main

import (
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "log"
 "net"
 "net/http"
 "os"
 "path/filepath"
 "strings"
 "sync"
 "time"
)

type Version struct { ID string `json:"id"`; Name string `json:"name"`; Channel string `json:"channel"`; Branch string `json:"branch"`; SHA string `json:"sha"`; Executable string `json:"executable"`; Installed bool `json:"installed"`; Runnable bool `json:"runnable"`; Protected bool `json:"protected"`; Managed bool `json:"managed"` }
type Catalog struct { Versions []Version `json:"versions"`; SelectedID string `json:"selectedId"`; RunningID string `json:"runningId"` }
type Event struct { At string `json:"at"`; Level string `json:"level"`; Kind string `json:"kind"`; VersionID string `json:"versionId"`; Message string `json:"message"` }
type Status struct { State string `json:"state"`; SelectedID string `json:"selectedId"`; RunningID string `json:"runningId"`; LastKnownGoodID string `json:"lastKnownGoodId"`; PID int `json:"pid"`; Port int `json:"port"`; Message string `json:"message"` }
type manager struct { mu sync.Mutex; path string; catalog Catalog; status Status; events []Event; eventStart int }

func newManager(path string) (*manager,error) {
 m:= &manager{path:path,status:Status{State:"IDLE"}}
 raw,err:=os.ReadFile(path)
 if err==nil { if e:=json.Unmarshal(raw,&m.catalog);e!=nil{return nil,fmt.Errorf("catalog corrupt; kept unchanged: %w",e)} } else if !errors.Is(err,os.ErrNotExist){return nil,err}
 if len(m.catalog.Versions)==0 && errors.Is(err,os.ErrNotExist) {
  m.catalog.Versions=[]Version{
   {ID:"main-source",Name:"Main (source)",Channel:"Main",Protected:true},
   {ID:"alpha-source",Name:"Alpha (source)",Channel:"Alpha",Protected:true},
   {ID:"experimental-source",Name:"Experimental (source)",Channel:"Experimental",Protected:true},
   {ID:"experimental-next",Name:"Experimental next (source)",Channel:"Experimental",Protected:true},
  }
 }
 // A persisted catalog is untrusted after restart: registration is not runtime approval.
 for i:=range m.catalog.Versions {m.catalog.Versions[i].Runnable=false}
 m.catalog.RunningID=""
 m.status.SelectedID=m.catalog.SelectedID
 return m,nil
}
func (m *manager) record(level,kind,id,msg string){ m.events=append(m.events,Event{At:time.Now().UTC().Format(time.RFC3339Nano),Level:level,Kind:kind,VersionID:id,Message:msg}); if len(m.events)>500 {drop:=len(m.events)-500;m.events=m.events[drop:];m.eventStart+=drop} }
func (m *manager) persist() error {
 if err:=os.MkdirAll(filepath.Dir(m.path),0700);err!=nil{return err}
 b,err:=json.MarshalIndent(m.catalog,"","  ");if err!=nil{return err}
 f,err:=os.CreateTemp(filepath.Dir(m.path),".catalog-*");if err!=nil{return err}
 defer os.Remove(f.Name())
 if err=f.Chmod(0600);err!=nil {f.Close();return err}
 if _,err=f.Write(b);err!=nil {f.Close();return err}
 if err=f.Sync();err!=nil {f.Close();return err}
 if err=f.Close();err!=nil{return err}
 return os.Rename(f.Name(),m.path)
}
func (m *manager) find(id string) *Version {for i:=range m.catalog.Versions {if m.catalog.Versions[i].ID==id{return &m.catalog.Versions[i]}};return nil}
func validCommitSHA(v string) bool { if len(v)<7 || len(v)>64{return false};for _,r:=range v{if !strings.ContainsRune("0123456789abcdefABCDEF",r){return false}};return true }
func validChannel(v string) bool {return v=="Main" || v=="Alpha" || v=="Experimental"}
func validExecutable(p string) error {
 if p=="" || !filepath.IsAbs(p){return errors.New("absolute executable path required")}
 clean:=filepath.Clean(p)
 canonical,err:=filepath.EvalSymlinks(clean);if err!=nil{return fmt.Errorf("executable unavailable: %w",err)}
 if canonical!=clean{return errors.New("symlinked executable path requires separate explicit review")}
 s,err:=os.Stat(canonical);if err!=nil{return fmt.Errorf("executable unavailable: %w",err)}
 if !s.Mode().IsRegular()||s.Mode()&0111==0 && filepath.Ext(strings.ToLower(clean))!=".exe" {return errors.New("not an executable file")}
 low:=strings.ToLower(clean)
 for _,part:=range []string{"teamsdashboard\\v5","\\windows\\","/etc/","/proc/","/sys/","/dev/","version-manager-alpha"} {if strings.Contains(low,part){return errors.New("restricted executable location")}}
 return nil
}
func writeJSON(w http.ResponseWriter,status int,v any){w.Header().Set("Content-Type","application/json");w.Header().Set("Cache-Control","no-store");w.WriteHeader(status);_ =json.NewEncoder(w).Encode(v)}
func bad(w http.ResponseWriter,code int,e string){writeJSON(w,code,map[string]string{"error":e})}
func decode(r *http.Request,v any) error {r.Body=http.MaxBytesReader(nil,r.Body,8192);return json.NewDecoder(r.Body).Decode(v)}
func (m *manager) handler(static string) http.Handler {
 mux:=http.NewServeMux()
 mux.HandleFunc("GET /api/versions",func(w http.ResponseWriter,r *http.Request){m.mu.Lock();defer m.mu.Unlock();writeJSON(w,200,m.catalog)})
 mux.HandleFunc("POST /api/versions",func(w http.ResponseWriter,r *http.Request){
  var v Version;if e:=decode(r,&v);e!=nil{bad(w,400,"invalid JSON");return}
  if v.Name==""||v.Channel==""||!validCommitSHA(v.SHA)||!validChannel(v.Channel)||v.Branch=="" {bad(w,400,"name, channel, branch and commit SHA required");return}
  if e:=validExecutable(v.Executable);e!=nil{bad(w,400,e.Error());return}
  h:=sha256.Sum256([]byte(v.Executable+"|"+v.SHA));v.ID=hex.EncodeToString(h[:8]);v.Installed=true;v.Runnable=false;v.Protected=false;v.Managed=false
  m.mu.Lock();defer m.mu.Unlock();if m.find(v.ID)!=nil{bad(w,409,"already registered");return}
  m.catalog.Versions=append(m.catalog.Versions,v);if e:=m.persist();e!=nil{m.catalog.Versions=m.catalog.Versions[:len(m.catalog.Versions)-1];bad(w,500,"catalog save failed");return}
  m.record("INFO","ADD",v.ID,"registered local build without executing");writeJSON(w,201,v)
 })
 mux.HandleFunc("DELETE /api/versions/{id}",func(w http.ResponseWriter,r *http.Request){
  id:=r.PathValue("id");m.mu.Lock();defer m.mu.Unlock();v:=m.find(id);if v==nil{bad(w,404,"version not found");return}
  if v.Protected||m.catalog.RunningID==id {bad(w,409,"protected or running version cannot be removed");return}
  old:=append([]Version(nil),m.catalog.Versions...);oldSelected:=m.catalog.SelectedID
  for i:=range m.catalog.Versions {if m.catalog.Versions[i].ID==id{m.catalog.Versions=append(m.catalog.Versions[:i],m.catalog.Versions[i+1:]...);break}}
  if m.catalog.SelectedID==id{m.catalog.SelectedID="";m.status.SelectedID=""}
  if e:=m.persist();e!=nil{m.catalog.Versions=old;m.catalog.SelectedID=oldSelected;m.status.SelectedID=oldSelected;bad(w,500,"catalog save failed");return}
  m.record("INFO","UNREGISTER",id,"catalog entry removed; executable and user files untouched");writeJSON(w,200,map[string]string{"result":"unregistered"})
 })
 mux.HandleFunc("POST /api/select",func(w http.ResponseWriter,r *http.Request){
  var req struct{ID string `json:"id"`};if decode(r,&req)!=nil{bad(w,400,"invalid JSON");return}
  m.mu.Lock();defer m.mu.Unlock();if m.find(req.ID)==nil{bad(w,404,"version not found");return}
  prev:=m.catalog.SelectedID;m.catalog.SelectedID=req.ID;if e:=m.persist();e!=nil{m.catalog.SelectedID=prev;bad(w,500,"catalog save failed");return}
  m.status.SelectedID=req.ID;m.record("INFO","SELECT",req.ID,"version selected; not launched");writeJSON(w,200,m.status)
 })
 // Approval endpoints fail closed until a manager-built, locally attested stage
 // and distinct operator-consent UI session exist. Catalog entries are NOT stages.
 mux.HandleFunc("GET /api/approval-preview",func(w http.ResponseWriter,r *http.Request){
  id:=r.URL.Query().Get("id")
  if id=="" {bad(w,400,"version id required");return}
  m.mu.Lock();exists:=m.find(id)!=nil;m.mu.Unlock()
  if !exists {bad(w,404,"version not found");return}
  writeJSON(w,422,map[string]string{"error":"BUILD_UNVERIFIED","message":"no manager-verified staged build is available"})
 })
 mux.HandleFunc("POST /api/approvals",func(w http.ResponseWriter,r *http.Request){
  writeJSON(w,403,map[string]string{"error":"APPROVAL_REQUIRED","message":"operator approval is unavailable until secure local staging and consent UI are verified"})
 })
 mux.HandleFunc("GET /api/status",func(w http.ResponseWriter,r *http.Request){m.mu.Lock();defer m.mu.Unlock();writeJSON(w,200,m.status)})
 mux.HandleFunc("GET /api/log",func(w http.ResponseWriter,r *http.Request){
  m.mu.Lock();defer m.mu.Unlock();cursor:=0;fmt.Sscanf(r.URL.Query().Get("cursor"),"%d",&cursor);if cursor<m.eventStart{cursor=m.eventStart};if cursor>m.eventStart+len(m.events){cursor=m.eventStart+len(m.events)}
  writeJSON(w,200,map[string]any{"events":m.events[cursor-m.eventStart:],"next":m.eventStart+len(m.events)})
 })
 for _,action:=range []string{"launch","stop","rollback"} {a:=action;mux.HandleFunc("POST /api/"+a,func(w http.ResponseWriter,r *http.Request){bad(w,501,a+" disabled until isolated process-lifecycle verification is implemented")})}
 mux.HandleFunc("GET /",func(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/"{http.FileServer(http.Dir(static)).ServeHTTP(w,r);return};http.ServeFile(w,r,filepath.Join(static,"index.html"))})
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){host:=r.Host;if h,_,e:=net.SplitHostPort(host);e==nil{host=h};if host!="127.0.0.1"&&host!="localhost"{bad(w,403,"non-loopback host rejected");return};if r.Method=="POST"||r.Method=="DELETE"{if origin:=r.Header.Get("Origin");origin!=""{if origin!="http://127.0.0.1:18776"&&origin!="http://localhost:18776"{bad(w,403,"cross-origin mutation rejected");return}}};mux.ServeHTTP(w,r)})
}
func main(){
 data:=os.Getenv("ALPHA_VM_DATA");if data==""{data=filepath.Join("runtime","version-manager-alpha")}
 m,err:=newManager(filepath.Join(data,"catalog.json"));if err!=nil{log.Fatal(err)}
 static:=os.Getenv("ALPHA_VM_STATIC");if static==""{static="static"}
 addr:="127.0.0.1:18776";ln,err:=net.Listen("tcp",addr);if err!=nil{log.Fatal(err)}
 log.Printf("alpha version manager (catalog only; launch disabled) on http://%s",addr)
 log.Fatal(http.Serve(ln,m.handler(static)))
}
