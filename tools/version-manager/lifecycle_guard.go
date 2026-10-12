package main

import (
 "errors"
 "fmt"
)

// childIdentity is a synthetic, in-memory observation. A port ping alone
// cannot establish OS process ownership; the production runtime MUST first
// independently bind the listener to the spawned OS process.
type childIdentity struct {
 PID int
 StartToken string
 Port int
 BuildDigest string
 ListenerPID int
 Ready bool
}

func verifyOwnedChild(expected, observed childIdentity) error {
 if expected.PID<=0 || expected.Port<=0 || expected.Port==8777 || expected.StartToken=="" || expected.BuildDigest=="" {
  return errors.New("missing valid managed-child identity")
 }
 if !observed.Ready || observed.PID!=expected.PID || observed.ListenerPID!=expected.PID ||
  observed.StartToken!=expected.StartToken || observed.Port!=expected.Port ||
  observed.BuildDigest!=expected.BuildDigest {
  return fmt.Errorf("observed listener is not the approved managed child")
 }
 return nil
}

func authorizeOwnedStop(expected, observed childIdentity) error {
 // Never signal a PID based only on an old manager status or a successful HTTP ping.
 return verifyOwnedChild(expected,observed)
}
