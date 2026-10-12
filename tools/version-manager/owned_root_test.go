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
 rootAlias:=filepath.Join(t.TempDir(),"aliased-root")
 if err:=os.Symlink(root,rootAlias);err==nil {
  if err:=requireOwnedBuild(rootAlias,inside);err==nil{t.Fatal("aliased managed root accepted")}
  aliasedCandidate:=filepath.Join(rootAlias,"builds","candidate.exe")
  if err:=requireOwnedBuild(rootAlias,aliasedCandidate);err==nil{t.Fatal("aliased root and candidate accepted")}
 }
}

func TestOwnedBuildWithinIsolatedAlphaDirectory(t *testing.T) {
 base:=t.TempDir()
 root:=filepath.Join(base,"version-manager-alpha","builds","synthetic")
 if err:=os.MkdirAll(root,0700);err!=nil{t.Fatal(err)}
 exe:=filepath.Join(root,"dashboard.exe")
 if err:=os.WriteFile(exe,[]byte("synthetic manager-controlled build"),0700);err!=nil{t.Fatal(err)}
 if err:=requireOwnedBuild(root,exe);err!=nil{t.Fatalf("isolated manager build rejected: %v",err)}
 if err:=validExecutable(exe);err==nil{t.Fatal("catalog registration accepted restricted managed build path")}
 if err:=os.Remove(exe);err!=nil{t.Fatal(err)}
 if err:=os.Mkdir(exe,0700);err!=nil{t.Fatal(err)}
 if err:=requireOwnedBuild(root,exe);err==nil{t.Fatal("directory pretending to be executable accepted")}
}
