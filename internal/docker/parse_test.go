package docker

import "testing"

func TestParsePsJSONLines(t *testing.T) {
	in := `{"ID":"123","Names":"web","Image":"nginx:alpine","Status":"Up 2 minutes","State":"running","Labels":"com.docker.compose.project=demo"}`
	out := parsePsJSONLines(in)
	if len(out) != 1 { t.Fatalf("want 1 got %d", len(out)) }
	if out[0].ID != "123" || out[0].Name != "web" { t.Fatalf("unexpected: %#v", out[0]) }
	if out[0].Labels["com.docker.compose.project"] != "demo" { t.Fatalf("labels parse failed") }
}

func TestParseImagesJSONLines(t *testing.T) {
	in := `{"ID":"sha256:1","Repository":"alpine","Tag":"3.18","Size":"5MB"}\n{"ID":"sha256:2","Repository":"<none>","Tag":"<none>","Size":"1MB"}`
	out := parseImagesJSONLines(in)
	if len(out) != 2 { t.Fatalf("want 2 got %d", len(out)) }
	if out[0].RepoTags[0] != "alpine:3.18" { t.Fatalf("unexpected tag: %v", out[0].RepoTags) }
	if out[1].RepoTags != nil { t.Fatalf("dangling should have nil tags") }
}
