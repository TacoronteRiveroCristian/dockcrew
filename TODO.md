# TODO DockCrew (MVP)

Estado: 2025-09-23

## Infra/Repo
- [ ] Configurar remote y primer push a GitHub (evitar cuelgues VSCode; ejecutar manualmente)
- [ ] Activar protección ramas main/develop y PRs requeridos
- [ ] Añadir CODEOWNERS y plantilla de issues/PRs

## Esqueleto Go
- [x] go.mod y `cmd/dockcrew/main.go` con `--version`
- [x] Makefile con build/test y ldflags de versión
- [x] CI básico (build+test) en GitHub Actions
- [ ] README: sección de build/run y roadmap

## Módulos (según SPEC)
- [ ] `internal/domain` modelos base (Container, Image, Volume, Network, Stack, Stats, LogLine)
- [ ] `internal/docker` cliente docker (interfaces + stub)
- [ ] `internal/ui` bootstrap Bubble Tea (tabla simple)
- [ ] `internal/logs` buffer circular + tipos
- [ ] `internal/config` loader+defaults+`--export-default-config`
- [ ] `internal/update` checker (stub)
- [ ] `internal/telemetry` opt-in (stub)

## Funcionalidades MVP
- [ ] Listar contenedores y estados (stub hasta integrar Docker)
- [ ] Acciones start/stop/restart/remove (stub)
- [ ] Logs tail por contenedor (stub)
- [ ] Vista imágenes/volúmenes/redes (stub)

## Calidad
- [ ] go fmt/vet y linter (añadir golangci-lint en otra iteración)
- [ ] Tests unitarios mínimos (domain, config)
- [ ] Tests integración condicionados por Docker (tag `integration`)

## Empaquetado/Distribución
- [ ] Config inicial GoReleaser
- [ ] Dockerfile minimal para ejecutar en contenedor
- [ ] Script instalador (bash)

## Documentación
- [ ] Guía de instalación (binario y contenedor)
- [ ] Keybindings (extraído del PRD)
- [ ] Contribución (CONTRIBUTING.md)

## Progreso (2025-09-23)
- [x] CLI: `containers --all`, `start/stop/restart/rm`
- [x] Logs: ring buffer con tests
- [x] Config: defaults + validate + export YAML
- [x] Dockerfile base
- [x] GoReleaser: plantilla inicial

Siguiente hito:
- [ ] Carga YAML/env en `internal/config.Load()`
- [ ] Añadir `images rm` y `prune` (confirmación)
- [ ] Stub UI (Bubble Tea) con tabla de contenedores usando mock

---

Siguiente hito: implementar `internal/domain` con modelos y tests + stub de `internal/docker` con interfaz y mock para la UI.
