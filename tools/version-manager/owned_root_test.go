package main

import (
 "os"
 "path/filepath"
 "testing"
)

func TestOwnedBuildRootContainment(t *testing.T) {
 root:=t.TempDir()
 inside:=filepath.Join(root,"builds","candidate.exe")
 if err:=os.MkdirAll(filepath.Dir(inside),0700);err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(inside,[]byte("synthetic fixture"),0700);err!=nil{t.Fatal(err)}
 if err:=requireOwnedBuild(root,inside);err!=nil{t.Fatalf("owned fixture rejected: %v",err)}
 outside:=filepath.Join(t.TempDir(),"external.exe")
 if err:=os.WriteFile(outside,[]byte("synthetic"),0700);err!=nil{t.Fatal(err)}
 if err:=requireOwnedBuild(root,outside);err==nil{t.Fatal("outside binary accepted")}
 if err:=requireOwnedBuild(root,filepath.Join(root,"..",filepath.Base(filepath.Dir(root)),"x.exe"));err==nil{t.Fatal("path escape accepted")}
 alias:=filepath.Join(root,"builds","alias.exe")
 if err:=os.Symlink(outside,alias);err==nil {
  if err:=requireOwnedBuild(root,alias);err==nil{t.Fatal("symlink escape accepted")}
 }
 if err:=requireOwnedBuild("relative",inside);err==nil{t.Fatal("relative managed root accepted")}
}
