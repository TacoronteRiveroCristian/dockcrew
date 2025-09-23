# CLAUDE.md - DockCrew Development Guide

## Stack & Architecture
- **Go 1.22+** con Charm stack (Bubble Tea + Bubbles + Lip Gloss)
- **Arquitectura**: UI ↔ Domain ↔ Docker (ver `docs/SPEC.md`)
- **Comandos**: Ver `Makefile` para build/test/lint

## Code Conventions
- Seguir patrones Go estándar + `gofmt`
- Usar interfaces para testabilidad (`internal/docker`)
- Separación clara entre capas
- `context.Context` para timeouts/cancelación
- Errores explícitos, no logs sensibles

## Dependencies (Core)
```go
// TUI
github.com/charmbracelet/bubbletea
github.com/charmbracelet/bubbles
github.com/charmbracelet/lipgloss

// CLI/Config
github.com/spf13/cobra
github.com/spf13/viper

// Docker
github.com/docker/docker
github.com/docker/cli
```

## Module Structure
```
cmd/dockcrew/      # Entrypoint Cobra
internal/
├── docker/        # Docker API client
├── domain/        # Modelos (Container, Image, etc.)
├── ui/           # Bubble Tea UI
├── logs/         # Log streaming/parsing
├── config/       # YAML config loader
├── update/       # Self-update
└── telemetry/    # Opt-in metrics
```

## Key Patterns

### Domain Models
```go
type Container struct {
    ID     string    `json:"id"`
    Name   string    `json:"name"`
    State  string    `json:"state"`
    // ... ver SPEC.md para estructura completa
}
```

### Docker Client Interface
```go
type Client interface {
    ListContainers(ctx context.Context) ([]domain.Container, error)
    StartContainer(ctx context.Context, id string) error
    // ... acciones CRUD
}
```

### Bubble Tea UI
```go
type Model struct {
    currentPage string
    containers  []domain.Container
    selected    int
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
```

## Configuration
- Path: `~/.config/dockcrew/config.yaml` (Linux/macOS/Windows)
- Cascada: env → explícita → user default → embebido
- Validación estricta + `--export-default-config`

## Testing
- **Unit**: `go test ./...`
- **Integration**: `go test -tags=integration ./...` (requiere Docker)
- **Coverage**: `go test -coverprofile=coverage.out ./...`

## Performance Targets
- 500+ contenedores: <100ms UI response
- 50k logs: <300MB RAM
- CPU idle ~0-2%

### Técnicas
- Buffers circulares (logs)
- Virtual scrolling (UI)
- Goroutines + channels para polling/streams
- Throttling de updates

## Security
- No root required
- Validar user input
- Socket Docker permissions: `sudo usermod -aG docker $USER`
- Telemetría opt-in únicamente

## CI/CD
- **GitHub Actions**: Go 1.22+, lint, test, multi-OS matrix
- **Release**: GoReleaser en tags `v*`
- **Container**: GHCR push automático

## References
- `docs/PRD.md`: Visión producto + roadmap
- `docs/SPEC.md`: Arquitectura técnica detallada
- `TODO.md`: Checklist implementación MVP
- `Makefile`: Comandos build/test/clean