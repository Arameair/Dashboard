package main

import (
 "encoding/json"
 "net/http/httptest"
 "os"
 "path/filepath"
 "strings"
 "testing"
)

func TestCatalogOfflineAndProtected(t *testing.T) {
 m,err:=newManager(filepath.Join(t.TempDir(),"catalog.json"));if err!=nil{t.Fatal(err)}
 if len(m.catalog.Versions)<4{t.Fatalf("want four fixture sources, got %d",len(m.catalog.Versions))}
 for _,v:=range m.catalog.Versions {if !v.Protected||v.Runnable||v.Installed{t.Fatalf("source must be protected and not runnable: %+v",v)}}
 w:=httptest.NewRecorder();m.handler(t.TempDir()).ServeHTTP(w,httptest.NewRequest("DELETE","/api/versions/main-source",nil))
 if w.Code!=409{t.Fatalf("protected delete returned %d",w.Code)}
}
func TestRegisterUnregisterLeavesExecutableIntact(t *testing.T){
 dir:=t.TempDir();exe:=filepath.Join(dir,"fixture.exe");if err:=os.WriteFile(exe,[]byte("fixture only"),0700);err!=nil{t.Fatal(err)}
 path:=filepath.Join(dir,"catalog.json");m,_:=newManager(path)
 body:=`{"name":"fixture","channel":"Experimental","branch":"test","sha":"abcdef0123","executable":"`+filepath.ToSlash(exe)+`"}`
 w:=httptest.NewRecorder();m.handler(t.TempDir()).ServeHTTP(w,httptest.NewRequest("POST","/api/versions",strings.NewReader(body)))
 if w.Code!=201{t.Fatalf("register status=%d body=%s",w.Code,w.Body.String())}
 var v Version;if err:=json.Unmarshal(w.Body.Bytes(),&v);err!=nil{t.Fatal(err)}
 if v.ID==""||!v.Runnable{t.Fatalf("bad registered record: %+v",v)}
 loaded,err:=newManager(path);if err!=nil||loaded.find(v.ID)==nil{t.Fatalf("catalog not durable: %v",err)}
 w=httptest.NewRecorder();m.handler(t.TempDir()).ServeHTTP(w,httptest.NewRequest("DELETE","/api/versions/"+v.ID,nil))
 if w.Code!=200{t.Fatalf("unregister status=%d",w.Code)}
 if _,err:=os.Stat(exe);err!=nil{t.Fatalf("unregister must not delete executable: %v",err)}
}
func TestCorruptManifestNeverOverwritten(t *testing.T) {
 dir:=t.TempDir();path:=filepath.Join(dir,"catalog.json");before:=[]byte("{broken")
 if err:=os.WriteFile(path,before,0600);err!=nil{t.Fatal(err)}
 if _,err:=newManager(path);err==nil{t.Fatal("corrupt manifest must fail closed")}
 after,_:=os.ReadFile(path);if string(after)!=string(before){t.Fatal("corrupt manifest modified")}
}
func TestUnsafePathAndLaunchFailClosed(t *testing.T){
 if validExecutable("relative.exe")==nil{t.Fatal("relative executable accepted")}
 m,_:=newManager(filepath.Join(t.TempDir(),"catalog.json"))
 for _,action:=range []string{"launch","stop","rollback"}{
  w:=httptest.NewRecorder();m.handler(t.TempDir()).ServeHTTP(w,httptest.NewRequest("POST","/api/"+action,nil))
  if w.Code!=501{t.Fatalf("%s unexpectedly actionable: %d",action,w.Code)}
 }
}

func TestAbsoluteLogCursorAfterRollover(t *testing.T) {
 m,_:=newManager(filepath.Join(t.TempDir(),"catalog.json"))
 for i:=0;i<510;i++ {m.record("INFO","TEST","","synthetic event")}
 if m.eventStart!=10 {t.Fatalf("start=%d want=10",m.eventStart)}
 w:=httptest.NewRecorder();m.handler(t.TempDir()).ServeHTTP(w,httptest.NewRequest("GET","/api/log?cursor=0",nil))
 var resp struct{Events []Event `json:"events"`;Next int `json:"next"`}
 if err:=json.Unmarshal(w.Body.Bytes(),&resp);err!=nil{t.Fatal(err)}
 if len(resp.Events)!=500||resp.Next!=510{t.Fatalf("events=%d next=%d",len(resp.Events),resp.Next)}
 m.record("INFO","TEST","","one more")
 w=httptest.NewRecorder();m.handler(t.TempDir()).ServeHTTP(w,httptest.NewRequest("GET","/api/log?cursor=510",nil))
 if err:=json.Unmarshal(w.Body.Bytes(),&resp);err!=nil{t.Fatal(err)}
 if len(resp.Events)!=1||resp.Next!=511{t.Fatalf("events=%d next=%d",len(resp.Events),resp.Next)}
}
func TestRejectUntrustedHostAndOrigin(t *testing.T) {
 m,_:=newManager(filepath.Join(t.TempDir(),"catalog.json"))
 for _,tc:=range []struct{method,path,host,origin string}{
  {"GET","/api/versions","untrusted.invalid:18776",""},
  {"POST","/api/select","127.0.0.1:18776","https://untrusted.invalid"},
 }{
  r:=httptest.NewRequest(tc.method,tc.path,nil);r.Host=tc.host
  if tc.origin!=""{r.Header.Set("Origin",tc.origin)}
  w:=httptest.NewRecorder();m.handler(t.TempDir()).ServeHTTP(w,r)
  if w.Code!=403{t.Errorf("%s %s got=%d",tc.method,tc.host,w.Code)}
 }
}
