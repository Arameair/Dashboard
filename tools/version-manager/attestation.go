package main

import (
 "crypto/sha256"
 "encoding/hex"
 "errors"
 "fmt"
 "io"
 "os"
 "path/filepath"
)

// binaryEvidence is an ephemeral observation, not executable approval.
// Approval must separately bind the observed digest to an owner-reviewed build.
type binaryEvidence struct {
 Path string
 Digest string
 Size int64
}

func inspectBinary(path string) (binaryEvidence,error) {
 if err:=validExecutable(path);err!=nil{return binaryEvidence{},err}
 clean:=filepath.Clean(path)
 f,err:=os.Open(clean);if err!=nil{return binaryEvidence{},err}
 defer f.Close()
 before,err:=f.Stat();if err!=nil{return binaryEvidence{},err}
 if !before.Mode().IsRegular(){return binaryEvidence{},errors.New("executable is not regular file")}
 h:=sha256.New()
 n,err:=io.Copy(h,f);if err!=nil{return binaryEvidence{},err}
 after,err:=f.Stat();if err!=nil{return binaryEvidence{},err}
 if !os.SameFile(before,after) || n!=before.Size() || before.Size()!=after.Size() || !before.ModTime().Equal(after.ModTime()) {
  return binaryEvidence{},errors.New("executable changed during inspection")
 }
 return binaryEvidence{Path:clean,Digest:hex.EncodeToString(h.Sum(nil)),Size:n},nil
}

func verifyBinary(expected binaryEvidence) error {
 actual,err:=inspectBinary(expected.Path)
 if err!=nil{return err}
 if expected.Path!=actual.Path||expected.Digest==""||expected.Digest!=actual.Digest||expected.Size!=actual.Size {
  return fmt.Errorf("executable content/path changed since approval")
 }
 return nil
}
