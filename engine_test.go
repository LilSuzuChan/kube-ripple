package main

import (
	"os"
	"strings"
	"testing"
)

func TestFixture(t *testing.T) {
	r := Report{Findings: []Finding{}}
	var before, objects []Obj
	var e error
	before, e = read("examples/before.json")
	if e != nil {
		t.Fatal(e)
	}
	objects, e = read("examples/after.json")
	if e != nil {
		t.Fatal(e)
	}
	diagnose(&r, objects, before, "default", "demo", "client", "backend", 80)
	found := false
	for _, f := range r.Findings {
		if f.ID == "selector-no-candidates" {
			found = true
		}
		if len(f.Evidence) == 0 {
			t.Fatalf("missing evidence: %+v", f)
		}
	}
	if !found {
		t.Fatalf("expected diagnostic absent: %+v", r.Findings)
	}
}
func TestMalformedInput(t *testing.T) {
	p := t.TempDir() + "/bad.yaml"
	if e := os.WriteFile(p, []byte("x: ["), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := read(p); e == nil {
		t.Fatal("malformed YAML accepted")
	}
}
func TestMissingSubject(t *testing.T) {
	r := Report{}
	missing(&r, "Pod", "default", "absent")
	if r.Findings[0].Classification != "unknown" {
		t.Fatal("missing subject is not unknown")
	}
}
func TestNamespaceIsolation(t *testing.T) {
	objects := []Obj{{"kind": "Pod", "metadata": Obj{"name": "same", "namespace": "other"}}}
	if find(objects, "Pod", "default", "same") != nil {
		t.Fatal("cross namespace match")
	}
}
func TestSecretNotEchoed(t *testing.T) {
	r := Report{}
	secret := Obj{"kind": "Secret", "metadata": Obj{"name": "secret"}, "data": Obj{"password": "sensitive-marker"}}
	diagnose(&r, []Obj{secret}, nil, "default", "demo", "client", "backend", 80)
	for _, f := range r.Findings {
		if strings.Contains(f.Summary, "sensitive-marker") {
			t.Fatal("secret exposed")
		}
	}
}
func TestTemplateChangeDetected(t *testing.T) {
	before, _ := read("examples/before.json")
	after, _ := read("examples/before.json")
	d := find(after, "Deployment", "default", "backend")
	d["spec"].(map[string]interface{})["template"].(map[string]interface{})["metadata"].(map[string]interface{})["labels"] = Obj{"app": "wrong"}
	r := Report{}
	diagnose(&r, after, before, "default", "", "", "", 80)
	for _, f := range r.Findings {
		if f.ID == "selector-no-candidates" {
			return
		}
	}
	t.Fatal("unchanged Service with changed template not detected")
}
func TestZeroReplicasNotCandidate(t *testing.T) {
	o, _ := read("examples/before.json")
	find(o, "Deployment", "default", "backend")["spec"].(map[string]interface{})["replicas"] = 0
	if candidates(o, find(o, "Service", "default", "backend")) != 0 {
		t.Fatal("scaled-to-zero template counted")
	}
}
