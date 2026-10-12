package main

import "testing"

func TestOwnedChildIdentityRejectsForeignOrFallback(t *testing.T) {
 expected:=childIdentity{PID:42001,StartToken:"synthetic-start-1",Port:18777,BuildDigest:"digest-A",ListenerPID:42001,Ready:true}
 if err:=verifyOwnedChild(expected,expected);err!=nil{t.Fatal(err)}
 cases:=map[string]childIdentity{
  "old listener":{PID:42001,StartToken:"synthetic-start-1",Port:18777,BuildDigest:"digest-A",ListenerPID:40001,Ready:true},
  "fallback port":{PID:42001,StartToken:"synthetic-start-1",Port:18778,BuildDigest:"digest-A",ListenerPID:42001,Ready:true},
  "wrong start identity":{PID:42001,StartToken:"stale-start",Port:18777,BuildDigest:"digest-A",ListenerPID:42001,Ready:true},
  "wrong digest":{PID:42001,StartToken:"synthetic-start-1",Port:18777,BuildDigest:"digest-B",ListenerPID:42001,Ready:true},
  "not ready":{PID:42001,StartToken:"synthetic-start-1",Port:18777,BuildDigest:"digest-A",ListenerPID:42001},
  "foreign pid":{PID:42002,StartToken:"synthetic-start-1",Port:18777,BuildDigest:"digest-A",ListenerPID:42002,Ready:true},
 }
 for name,obs:=range cases {t.Run(name,func(t *testing.T){if err:=authorizeOwnedStop(expected,obs);err==nil{t.Fatal("foreign/invalid child allowed")}})}
 expected.Port=8777
 if err:=verifyOwnedChild(expected,expected);err==nil{t.Fatal("protected default dashboard port accepted")}
}
