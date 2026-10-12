package main

import (
 "errors"
 "fmt"
 "path/filepath"
)

// stagedApproval is an internal, non-persistent approval snapshot. It cannot
// be created by Add Version or by a user-supplied commit SHA. No HTTP approval
// endpoint exists yet; launch remains disabled until 02 scopes a consent flow.
type stagedApproval struct {
 Root string
 Evidence binaryEvidence
 ExplicitConsent bool
}

// checkApproval must be called immediately before any future launch. Approval
// cannot be inferred from manifest booleans or source-ref identity.
func checkApproval(a stagedApproval, candidate string) error {
 if !a.ExplicitConsent {return errors.New("explicit operator approval required")}
 if a.Root=="" || a.Evidence.Path=="" || a.Evidence.Digest=="" {
  return errors.New("missing managed-root or executable fingerprint")
 }
 if filepath.Clean(candidate)!=a.Evidence.Path {
  return errors.New("requested binary differs from approved binary")
 }
 if err:=requireOwnedBuild(a.Root,candidate);err!=nil {
  return fmt.Errorf("build containment check failed: %w",err)
 }
 if err:=verifyBinary(a.Evidence);err!=nil {
  return fmt.Errorf("executable attestation expired: %w",err)
 }
 return nil
}
