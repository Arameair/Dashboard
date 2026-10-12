package main

import (
 "os"
 "path/filepath"
 "strings"
 "testing"
)

func TestBinaryEvidenceDetectsReplacement(t *testing.T) {
 dir:=t.TempDir()
 path:=filepath.Join(dir,"synthetic.exe")
 if err:=os.WriteFile(path,[]byte("first synthetic build"),0700);err!=nil{t.Fatal(err)}
 ev,err:=inspectBinary(path);if err!=nil{t.Fatal(err)}
 if len(ev.Digest)!=64||ev.Size==0{t.Fatalf("invalid fingerprint: %+v",ev)}
 if err:=verifyBinary(ev);err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(path,[]byte("different synthetic build"),0700);err!=nil{t.Fatal(err)}
 if err:=verifyBinary(ev);err==nil{t.Fatal("replaced binary accepted")}
 ev.Digest=strings.Repeat("0",64)
 if err:=verifyBinary(ev);err==nil{t.Fatal("forged digest accepted")}
}
func TestBinaryEvidenceRejectsAlias(t *testing.T){
 dir:=t.TempDir();target:=filepath.Join(dir,"one.exe");link:=filepath.Join(dir,"two.exe")
 if err:=os.WriteFile(target,[]byte("synthetic"),0700);err!=nil{t.Fatal(err)}
 if err:=os.Symlink(target,link);err!=nil{t.Skipf("symlink not permitted: %v",err)}
 if _,err:=inspectBinary(link);err==nil{t.Fatal("symlink alias accepted")}
}
