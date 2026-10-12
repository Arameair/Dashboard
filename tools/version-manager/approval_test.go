package main

import (
 "os"
 "path/filepath"
 "testing"
)

func TestApprovalFailsClosedWithoutConsentOrAttestation(t *testing.T) {
 root:=t.TempDir()
 exe:=filepath.Join(root,"approved.exe")
 if err:=os.WriteFile(exe,[]byte("synthetic executable"),0700);err!=nil{t.Fatal(err)}
 proof,err:=inspectBinary(exe);if err!=nil{t.Fatal(err)}
 a:=stagedApproval{Root:root,Evidence:proof}
 if checkApproval(a,exe)==nil{t.Fatal("implicit catalog approval accepted")}
 a.ExplicitConsent=true
 if err:=checkApproval(a,exe);err!=nil{t.Fatalf("valid synthetic approval rejected: %v",err)}
 other:=filepath.Join(t.TempDir(),"other.exe")
 if err:=os.WriteFile(other,[]byte("different"),0700);err!=nil{t.Fatal(err)}
 if checkApproval(a,other)==nil{t.Fatal("foreign executable accepted")}
 if err:=os.WriteFile(exe,[]byte("replaced executable"),0700);err!=nil{t.Fatal(err)}
 if checkApproval(a,exe)==nil{t.Fatal("replaced executable accepted")}
 a.Evidence.Digest=""
 if checkApproval(a,exe)==nil{t.Fatal("missing digest accepted")}
}

func TestApprovedIsolatedManagedBuildCanBeAttested(t *testing.T) {
 root:=filepath.Join(t.TempDir(),"version-manager-alpha","builds","synthetic")
 if err:=os.MkdirAll(root,0700);err!=nil{t.Fatal(err)}
 exe:=filepath.Join(root,"synthetic.exe")
 if err:=os.WriteFile(exe,[]byte("synthetic test bytes"),0700);err!=nil{t.Fatal(err)}
 if err:=validExecutable(exe);err==nil{t.Fatal("restricted manager build permitted through untrusted catalog")}
 evidence,err:=inspectBinary(exe);if err!=nil{t.Fatalf("managed build fingerprint failed: %v",err)}
 approval:=stagedApproval{Root:root,Evidence:evidence,ExplicitConsent:true}
 if err:=checkApproval(approval,exe);err!=nil{t.Fatalf("trusted synthetic managed stage rejected: %v",err)}
 if err:=os.WriteFile(exe,[]byte("replaced test bytes"),0700);err!=nil{t.Fatal(err)}
 if err:=checkApproval(approval,exe);err==nil{t.Fatal("modified managed build accepted")}
}
