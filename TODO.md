# TODO DockCrew (MVP)

Estado: 2025-09-29

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
- [ ] `internal/domain` completar modelos faltantes (Volume, Network, Stack, Stats)
- [ ] `internal/docker` ampliar cliente (inspect, stats, compose, volumes, networks)
- [x] `internal/ui` bootstrap Bubble Tea con tabla de contenedores y refresco automático
- [ ] `internal/logs` buffer circular + tipos
- [ ] `internal/config` loader+defaults+`--export-default-config`
- [ ] `internal/update` checker (stub)
- [ ] `internal/telemetry` opt-in (stub)

## Funcionalidades MVP
- [x] Listar contenedores y estados (CLI + TUI)
- [x] Acciones start/stop/restart/remove desde UI (confirmaciones + feedback)
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

## TUI Roadmap inmediato
- [ ] Panel de detalles enriquecido (env vars, mounts, healthcheck)
- [ ] Agrupar/filtrar por `Stack` y etiquetas clave (Compose, Swarm)
- [ ] Barra de búsqueda y filtros rápidos (texto + estado)
- [x] Acciones en caliente: start/stop/restart/remove con modales de confirmación
- [x] Snackbar/toasts para errores y éxitos
- [ ] Atajos configurables + ayuda contextual (`?`)
- [ ] Integrar métricas live (CPU/Mem/Net) en panel derecho
- [ ] Vista logs en panel secundario con streaming y resaltado
- [ ] Pestañas para Images / Volumes / Networks / Stacks
- [ ] Modo compacto para terminales pequeñas (≤ 90 columnas)

## Progreso (2025-09-29)
- [x] CLI: `containers --all`, `start/stop/restart/rm`
- [x] Logs: ring buffer con tests
- [x] Config: defaults + validate + export YAML
- [x] Dockerfile base
- [x] GoReleaser: plantilla inicial
- [x] TUI base: tabla de contenedores, refresco y toggle `a` para mostrar todos

Siguiente hito:
- [ ] Añadir panel lateral con detalles extendidos (mounts, puertos, compose service)
- [ ] Barra de búsqueda y filtros rápidos en TUI
- [ ] Carga YAML/env en `internal/config.Load()`
- [ ] Añadir `images rm` y `prune` (confirmación en CLI/TUI)

---
