package docker

import (
	"context"
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/TacoronteRiveroCristian/dockcrew/internal/domain"
)

// Client define la interfaz del adaptador Docker que usará la UI
// Implementación inicial: shell-out al binario `docker` para evitar dependencias.
type Client interface {
	ListContainers(ctx context.Context) ([]domain.Container, error)
	ListContainersAll(ctx context.Context, all bool) ([]domain.Container, error)
	ListImages(ctx context.Context) ([]domain.Image, error)
	RemoveImage(ctx context.Context, ref string, force bool) error
	PruneImages(ctx context.Context) error
	StartContainer(ctx context.Context, id string) error
	StopContainer(ctx context.Context, id string) error
	RestartContainer(ctx context.Context, id string) error
	RemoveContainer(ctx context.Context, id string, force bool) error
}

// ShellClient usa el binario `docker` instalado en el sistema.
type ShellClient struct {
	Binary   string        // por defecto "docker"
	Timeout  time.Duration // por defecto 10s
}

func NewShellClient() *ShellClient {
	return &ShellClient{Binary: "docker", Timeout: 10 * time.Second}
}

func (c *ShellClient) run(ctx context.Context, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, c.Binary, args...)
	return cmd.CombinedOutput()
}

func (c *ShellClient) ListContainers(ctx context.Context) ([]domain.Container, error) {
	return c.ListContainersAll(ctx, false)
}

func (c *ShellClient) ListContainersAll(ctx context.Context, all bool) ([]domain.Container, error) {
	// Usar json por línea para parseo más robusto
	args := []string{"ps"}
	if all { args = append(args, "-a") }
	args = append(args, "--format", "{{json .}}")
	out, err := c.run(ctx, args...)
	if err != nil {
		return nil, err
	}
	return parsePsJSONLines(string(out)), nil
}

func (c *ShellClient) ListImages(ctx context.Context) ([]domain.Image, error) {
	out, err := c.run(ctx, "images", "--format", "{{json .}}")
	if err != nil {
		return nil, err
	}
	return parseImagesJSONLines(string(out)), nil
}

func (c *ShellClient) RemoveImage(ctx context.Context, ref string, force bool) error {
	args := []string{"rmi"}
	if force { args = append(args, "-f") }
	args = append(args, ref)
	_, err := c.run(ctx, args...)
	return err
}

func (c *ShellClient) PruneImages(ctx context.Context) error {
	_, err := c.run(ctx, "image", "prune", "-f")
	return err
}

func (c *ShellClient) StartContainer(ctx context.Context, id string) error {
	_, err := c.run(ctx, "start", id)
	return err
}

func (c *ShellClient) StopContainer(ctx context.Context, id string) error {
	_, err := c.run(ctx, "stop", id)
	return err
}

func (c *ShellClient) RestartContainer(ctx context.Context, id string) error {
	_, err := c.run(ctx, "restart", id)
	return err
}

func (c *ShellClient) RemoveContainer(ctx context.Context, id string, force bool) error {
	args := []string{"rm"}
	if force { args = append(args, "-f") }
	args = append(args, id)
	_, err := c.run(ctx, args...)
	return err
}

// helpers simples (sin allocaciones innecesarias)
func splitLines(s string) []string {
	res := make([]string, 0, 64)
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' || s[i] == '\r' {
			res = append(res, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		res = append(res, s[start:])
	}
	return res
}

func splitTab(s string) []string {
	res := make([]string, 0, 8)
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\t' {
			res = append(res, s[start:i])
			start = i + 1
		}
	}
	res = append(res, s[start:])
	return res
}

// Estructuras auxiliares para parseo JSON de docker templates
type psLine struct {
	ID        string            `json:"ID"`
	Names     string            `json:"Names"`
	Image     string            `json:"Image"`
	Status    string            `json:"Status"`
	State     string            `json:"State"`
	Ports     string            `json:"Ports"`
	Labels    string            `json:"Labels"`
	CreatedAt string            `json:"CreatedAt"`
}

type imageLine struct {
	ID        string `json:"ID"`
	Repository string `json:"Repository"`
	Tag       string `json:"Tag"`
	Size      string `json:"Size"`
	CreatedAt string `json:"CreatedAt"`
}

func parsePsJSONLines(s string) []domain.Container {
	lines := splitLines(s)
	res := make([]domain.Container, 0, len(lines))
	for _, ln := range lines {
		if strings.TrimSpace(ln) == "" { continue }
		var p psLine
		if err := json.Unmarshal([]byte(ln), &p); err != nil { continue }
		labels := map[string]string{}
		if p.Labels != "" {
			// docker imprime labels como "k=v,foo=bar", parseo best-effort
			for _, kv := range strings.Split(p.Labels, ",") {
				parts := strings.SplitN(strings.TrimSpace(kv), "=", 2)
				if len(parts) == 2 { labels[parts[0]] = parts[1] }
			}
		}
		res = append(res, domain.Container{
			ID:     p.ID,
			Name:   p.Names,
			Image:  p.Image,
			Status: p.Status,
			State:  p.State,
			Labels: labels,
			// CreatedAt: parseable con time.Parse en futuras iteraciones
		})
	}
	return res
}

func parseImagesJSONLines(s string) []domain.Image {
	lines := splitLines(s)
	res := make([]domain.Image, 0, len(lines))
	for _, ln := range lines {
		if strings.TrimSpace(ln) == "" { continue }
		var im imageLine
		if err := json.Unmarshal([]byte(ln), &im); err != nil { continue }
		tag := im.Repository
		if im.Tag != "" && im.Tag != "<none>" {
			if tag == "<none>" || tag == "" { tag = im.Tag } else { tag = tag + ":" + im.Tag }
		}
		res = append(res, domain.Image{ ID: im.ID, RepoTags: nonEmptySlice(tag) })
	}
	return res
}

func nonEmptySlice(s string) []string {
	if strings.TrimSpace(s) == "" || s == "<none>" { return nil }
	return []string{s}
}
