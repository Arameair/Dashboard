package main

import (
 "errors"
 "fmt"
 "os"
 "path/filepath"
 "strings"
)

// requireOwnedBuild confines candidate binaries to a reviewed, existing test-build
// root. It is only a validation primitive: it does NOT confer launch approval.
func requireOwnedBuild(root, candidate string) error {
 if !filepath.IsAbs(root) || !filepath.IsAbs(candidate) {
  return errors.New("managed build root and binary must be absolute")
 }
 canonicalRoot, err := filepath.EvalSymlinks(filepath.Clean(root))
 if err != nil { return fmt.Errorf("managed build root unavailable: %w",err) }
 if canonicalRoot!=filepath.Clean(root){return errors.New("managed build root is an alias")}
 info,err:=os.Stat(canonicalRoot)
 if err!=nil || !info.IsDir(){return errors.New("managed build root must be an existing directory")}
 canonicalFile,err:=filepath.EvalSymlinks(filepath.Clean(candidate))
 if err!=nil{return fmt.Errorf("managed binary unavailable: %w",err)}
 if canonicalFile!=filepath.Clean(candidate){return errors.New("managed binary path is an alias")}
 rel,err:=filepath.Rel(canonicalRoot,canonicalFile)
 if err!=nil{return errors.New("managed binary and root differ in volume")}
 if rel=="." || rel==".." || filepath.IsAbs(rel) || strings.HasPrefix(rel,".."+string(filepath.Separator)){
  return errors.New("binary outside owned build root")
 }
 if err:=validExecutable(canonicalFile);err!=nil{return err}
 return nil
}
