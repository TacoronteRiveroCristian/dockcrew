package docker

import "testing"

func TestParseImagesJSONLines(t *testing.T) {
	in := "{\"ID\":\"sha256:1\",\"Repository\":\"alpine\",\"Tag\":\"3.18\",\"Size\":\"5MB\"}\n{\"ID\":\"sha256:2\",\"Repository\":\"<none>\",\"Tag\":\"<none>\",\"Size\":\"1MB\"}"
	out := parseImagesJSONLines(in)
	if len(out) != 2 {
		t.Fatalf("want 2 got %d", len(out))
	}
	if out[0].RepoTags[0] != "alpine:3.18" {
		t.Fatalf("unexpected tag: %v", out[0].RepoTags)
	}
	if out[1].RepoTags != nil {
		t.Fatalf("dangling should have nil tags")
	}
}

func TestParsePsJSONLines(t *testing.T) {
	in := `{"ID":"123","Names":"web","Image":"nginx:alpine","Status":"Up 2 minutes","State":"running","Ports":"0.0.0.0:8080->80/tcp, :::443->443/tcp","Labels":"com.docker.compose.project=demo","CreatedAt":"2025-01-10 12:30:45 -0700 MST"}
{"ID":"456","Names":"standalone","Image":"redis:7","Status":"Exited","State":"exited","Ports":"6379/tcp","Labels":"<none>","CreatedAt":"2025-01-08 08:00:00"}
{"ID":"789","Names":"svc.1","Image":"busybox","Status":"Up","State":"running","Ports":"*:3000-3001->3000-3001/tcp","Labels":"com.docker.stack.namespace=webapp"}`
	out := parsePsJSONLines(in)
	if len(out) != 3 {
		t.Fatalf("want 3 got %d", len(out))
	}
	if out[0].Stack != "demo" {
		t.Fatalf("want stack demo, got %q", out[0].Stack)
	}
	if len(out[0].Ports) != 2 || out[0].Ports[0].Public != 8080 || out[0].Ports[1].Type != "tcp" {
		t.Fatalf("unexpected ports: %#v", out[0].Ports)
	}
	if out[0].CreatedAt.IsZero() {
		t.Fatalf("expected CreatedAt parsed")
	}
	if out[1].Stack != "" {
		t.Fatalf("standalone container should not have stack: %+v", out[1])
	}
	if len(out[1].Ports) != 1 || out[1].Ports[0].Private != 6379 {
		t.Fatalf("unexpected standalone ports: %+v", out[1].Ports)
	}
	if out[2].Stack != "webapp" {
		t.Fatalf("swarm stack label not parsed: %+v", out[2])
	}
	if out[2].Ports[0].Public != 3000 || out[2].Ports[0].Private != 3000 {
		t.Fatalf("expected first port range collapse to 3000: %+v", out[2].Ports)
	}
}

func TestParsePortBindingsEdgeCases(t *testing.T) {
	b := parsePortBindings("::1:2222->22/tcp, 0.0.0.0:8080->8080/tcp, 5432/tcp")
	if len(b) != 3 {
		t.Fatalf("expected 3 bindings, got %d", len(b))
	}
	if b[0].IP != "::1" || b[0].Public != 2222 || b[0].Private != 22 {
		t.Fatalf("unexpected ipv6 parse: %+v", b[0])
	}
	if b[2].Public != 5432 || b[2].Private != 5432 {
		t.Fatalf("local-only port mismatch: %+v", b[2])
	}
}
