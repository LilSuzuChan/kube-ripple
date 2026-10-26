package main

import (
	"fmt"
)

func resourceKinds() []string { return nil }
func matches(selector interface{}, labels interface{}) bool {
	a, ok := selector.(map[string]interface{})
	if !ok || len(a) == 0 {
		return false
	}
	b, _ := labels.(map[string]interface{})
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
func candidates(objects []Obj, service Obj) int {
	count := 0
	for _, o := range objects {
		if ns(o) != ns(service) {
			continue
		}
		switch s(o, "kind") {
		case "Pod":
			if matches(get(service, "spec", "selector"), get(o, "metadata", "labels")) {
				count++
			}
		case "Deployment":
			if matches(get(service, "spec", "selector"), get(o, "spec", "template", "metadata", "labels")) {
				replicas := 1
				if get(o, "spec", "replicas") != nil {
					replicas = num(o, "spec", "replicas")
				}
				if replicas > 0 {
					count++
				}
			}
		}
	}
	return count
}
func diagnose(r *Report, after, before []Obj, n, name, src, dst string, port int) {
	services := 0
	for _, service := range after {
		if s(service, "kind") != "Service" {
			continue
		}
		services++
		old := find(before, "Service", ns(service), s(service, "metadata", "name"))
		if lenMap(get(service, "spec", "selector")) == 0 {
			r.add("selector-unmanaged", "info", "unknown", "Service has no selector; endpoint management is outside analysis", service, "spec.selector")
			continue
		}
		next := candidates(after, service)
		if old == nil {
			r.add("service-added", "info", "confirmed", "New Service in after manifests", service, "metadata")
		}
		previous := 0
		if old != nil {
			previous = candidates(before, old)
		}
		severity := "info"
		id := "selector-candidates"
		if next == 0 {
			severity = "warning"
			id = "selector-no-candidates"
		} else if next < previous {
			severity = "warning"
			id = "selector-candidates-reduced"
		}
		r.add(id, severity, "predicted", fmt.Sprintf("Candidate objects: before=%d after=%d (Pods and Deployment templates counted separately)", previous, next), service, "spec.selector / Pod labels / Deployment template labels")
	}
	for _, old := range before {
		if s(old, "kind") == "Service" && find(after, "Service", ns(old), s(old, "metadata", "name")) == nil {
			r.add("service-removed", "warning", "confirmed", "Service absent from after manifests", old, "metadata")
		}
	}
	if services == 0 {
		r.Limitations = append(r.Limitations, "No Services in after manifests.")
	}
	r.Limitations = append(r.Limitations, "Inputs are assumed to be complete before/after inventories. Candidate counts are not replica or traffic counts. Live Pods and predicted templates are distinct; controller behavior and actual traffic are unverified.")
}
func lenMap(v interface{}) int { m, _ := v.(map[string]interface{}); return len(m) }
