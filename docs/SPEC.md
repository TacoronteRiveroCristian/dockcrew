# Especificaciones Técnicas: DockCrew

Autor: Tú (con apoyo de IA)
Fecha: 2025-09-23
Versión: 0.1 (borrador inicial)

Este documento traduce el PRD en un diseño técnico implementable. Define stack, arquitectura, contratos de datos, flujos de UI, manejo de logs, empaquetado, CI/CD, seguridad y pruebas.

## 1. Alcance técnico y decisiones clave

- Objetivo: TUI moderna para gestionar Docker y Docker Compose con instalación de 1 paso.
- Stack seleccionado: Go 1.22+
  - TUI: Charm stack (Bubble Tea + Bubbles + Lip Gloss)
  - CLI/config: Cobra + Viper
  - Docker: `github.com/docker/docker` (client) y `github.com/docker/cli` para ayudas de formato/parsing. Compose v2 vía `docker` CLI (shell-out) con salida parseada.
  - Logs/regex: `regexp` precompilado, buffers circulares.
  - Releases: GoReleaser + contenedor en GHCR/Docker Hub.
- Motivos: cross-compile fácil, ecosistema TUI robusto, binarios autoincluidos, buen soporte multi-OS.

Nota: Si se prefiere Rust, se puede adaptar el diseño manteniendo módulos y contratos; differences: `bollard` para Docker, `ratatui` para TUI y `tokio` para concurrencia.

## 2. Arquitectura y módulos

Capas y responsabilidades:

- `internal/docker` (adaptador a Docker API)
  - Cliente Docker: listas, detalles, acciones, stats streaming, logs.
  - Compose: wrappers sobre `docker compose` (ls, up -d, down, restart, ps).
- `internal/domain`
  - Modelos: Container, Image, Volume, Network, Stack, LogLine, Stats.
  - Casos de uso: listar, ordenar/filtrar, ejecutar acciones, obtener logs.
- `internal/ui`
  - Programa Bubble Tea, páginas (containers, images, volumes, networks, stacks, logs), componentes (tablas, sidebar, command palette, footer/help).
  - Enrutamiento, keymap, estado global, panel de detalles.
- `internal/logs`
  - Streamer (lectura de logs), Buffer circular (cap), Parser (niveles, JSON, HTTP), Highlighter, Index para búsqueda.
- `internal/config`
  - Carga/validación (Viper), cascade env>ruta explícita>user default>embebido, watcher opcional para recarga en vivo.
- `internal/update`
  - Chequeo de versión y `self-update` (opcional fuera de contenedor).
- `internal/telemetry` (opt-in)
  - Envío mínimo de métricas anónimas (versión, OS, flags usados).
- `cmd/dockcrew`
  - Entrypoint Cobra, flags globales y subcomandos utilitarios (`--export-default-config`, `self-update`).

Diagrama (alto nivel):

UI (Bubble Tea) <-> domain <-> docker adapter | logs subsystem | config | update | telemetry

## 3. Modelos de datos (contratos)

- Container
  - id: string
  - name: string
  - image: string
  - state: string (running, exited, restarting, paused, etc.)
  - status: string (human)
  - ports: []PortBinding {ip, private, public, type}
  - createdAt: time.Time
  - labels: map[string]string
  - stack: string (compose project)
- Image
  - id: string
  - repoTags: []string
  - sizeBytes: int64
  - createdAt: time.Time
  - dangling: bool
  - layers?: []Layer {id, sizeBytes}
- Volume
  - name: string
  - driver: string
  - mountpoint: string
  - createdAt: time.Time
  - inUseBy: []string (container names)
- Network
  - id: string
  - name: string
  - driver: string
  - scope: string
  - createdAt: time.Time
- Stack (Compose)
  - name: string
  - services: []string
  - containers: []string
- Stats (stream)
  - cpuPercent: float64
  - memBytes: uint64
  - memPercent: float64
  - netRxBytes, netTxBytes: uint64
  - blockReadBytes, blockWriteBytes: uint64
- LogLine
  - ts: time.Time | nil
  - level: enum{TRACE,DEBUG,INFO,WARN,ERROR}
  - text: string
  - meta: map[string]string (service, container, httpCode, method, etc.)

## 4. Interacciones con Docker

- API Docker (a través del cliente Go):
  - Listar: Containers (List), Images (ImageList), Volumes (VolumeList), Networks (NetworkList)
  - Acciones: ContainerStart, ContainerStop, ContainerRestart, ContainerRemove, ImageRemove, VolumeRemove, NetworkRemove
  - Stats: ContainerStats (stream)
  - Logs: ContainerLogs (stream, `since`, `tail`, `follow`)
  - Inspect: ContainerInspect, ImageInspect, VolumeInspect, NetworkInspect
- Compose v2 (CLI shell-out):
  - `docker compose ls --format json`
  - `docker compose -p <project> up -d`, `down`, `restart`, `ps --format json`
  - Detección de proyecto/stack: label `com.docker.compose.project`

Errores se exponen con códigos y mensajes amigables. Shell-out debe capturar stdout/stderr, tiempo de espera y retornar errores estructurados.

## 5. Concurrencia y bucles de eventos

- UI Bubble Tea en el hilo principal.
- Goroutines para:
  - Polling de recursos (refresh_interval): actualizar cache de containers/images/volumes/networks.
  - Streams: stats/logs por contenedor.
  - Compose commands (con cancelación via context).
- Comunicación: canales Go hacia UI mediante mensajes Bubble Tea (tea.Msg) con payloads inmutables.
- Context: `context.Context` por operación con timeouts configurables.
- Backpressure: límites de colas (buffer) y caídas controladas (drop oldest) para mantener UI fluida.

## 6. UI y navegación

Páginas principales:
- Containers: tabla sortable, filtros rápidos, acciones y panel de detalles.
- Images: tabla, dangling toggle, detalles, prune.
- Volumes: tabla con estado de uso, detalles y prune.
- Networks: tabla, detalles y eliminación.
- Stacks (Compose): vista agrupada por proyecto, acciones por stack.
- Logs: visor dedicado por contenedor/servicio.

Navegación:
- Tabs superiores o keybindings directos; paleta de comandos (Ctrl+P/Ctrl+\) para saltar.
- Sidebar de detalles persistente con acciones contextuales.
- Footer con leyenda de teclas y estado (versión, update).

Keybindings (coherentes con PRD):
- Global: j/k o ↑/↓, g/G, Tab, :, ?, q
- Ordenación: Shift+[N,I,S,C,P,...] por página
- Containers: s,t,r,k,l,e,x,P
- Images: x,D,i
- Volumes: x,P,d
- Networks: x,d
- Compose: U (up -d), D (down), R (recreate), S (restart)
- Paleta: Ctrl+P o Ctrl+\

Estados y transiciones:
- Selección en tabla -> panel detalles se actualiza.
- Acción confirmable -> modal (sí/no) -> ejecuta -> notificación resultado.
- Ver logs -> navegación a página Logs -> Esc vuelve a contenedores.

## 7. Visor de logs (internos)

- Entrada: stream Docker (`ContainerLogs`) con opciones `follow`, `since`, `tail`, `timestamps`.
- Buffer: anillo (cap configurable `logs.max_lines`, p.ej. 4000). Evitar OOM.
- Parsing:
  - Detec. timestamp (varios formatos), niveles (ERROR,WARN,INFO,DEBUG,TRACE), HTTP (método/códigos), URLs, paths, IPs, UUIDs.
  - Pretty JSON: detección de líneas JSON; expand/collapse por atajo.
- Highlighting: colores por nivel y patrones; configurables por tema.
- Búsqueda/filtrado:
  - Filtro simple (texto) y regex `/pattern/` (compilación cacheada, timeout de eval.).
  - Índices ligeros (offsets) para saltos rápidos.
- Virtual scrolling: solo render de líneas visibles + prefetch de próximas.
- Copiado: integración con clipboard del host cuando sea posible (fuera del contenedor) o alternativas (imprimir selección). En contenedor, exponer texto en portapapeles vía utilidades si están presentes.

## 8. Configuración

Ruta y cascada:
1) `DOCKCREW_CONFIG` (ruta explícita)
2) `~/.config/dockcrew/config.yaml` (Linux), `~/Library/Application Support/DockCrew/config.yaml` (macOS), `%APPDATA%/DockCrew/config.yaml` (Windows)
3) Valores embebidos por defecto.

Esquema (YAML):
```yaml
app:
  refresh_interval: 5.0   # s
  check_for_update: true
  telemetry: false        # opt-in
ui:
  theme: default          # default|dark|light|high-contrast|custom
  mouse: true
  wrap_main_panel: false
  palette_shortcut: "ctrl+p"  # ctrl+p|ctrl+\\
logs:
  max_lines: 4000
  tail: 400
  since: "30m"
  highlight_levels: true
  pretty_json: true
  regex_timeout_ms: 50
docker:
  host: "unix:///var/run/docker.sock" # o npipe para Windows
compose:
  enabled: true
  binary: "docker"  # path a docker CLI
  default_project: ""
colors:  # cuando theme=custom
  title: "#96E072"
  help: "#EE5D43"
  background: "#23262E"
  footer: "#00E8C6"
  success: "#96E072"
  error: "#EE5D43"
```

Validación: tipos estrictos y rangos aceptables; errores con mensajes claros. `--export-default-config` genera un archivo completo.

## 9. Errores y UX de fallos

- Docker no accesible: banner en footer + sugerencias (docker running, pertenencia a grupo docker).
- Acción inválida (volumen en uso, red en uso): modal con motivo y alternativas.
- Compose no encontrado: deshabilitar pestaña Stacks y mostrar hint para instalar/actualizar Docker Compose v2.
- Logs sin permiso: nota no intrusiva y fallback a modo sin logs.

## 10. Seguridad

- Respeto al principio de mínimo privilegio; no requerir root.
- Uso del socket Docker: advertencias y documentación para permisos del usuario.
- No ejecutar plugins/acciones externas sin confirmación o sandbox (fase posterior).
- Sin telemetría por defecto. Si se activa, anonimizada, sin contenido sensible.

## 11. Rendimiento (presupuestos y técnicas)

- Listas de 500+ contenedores: UI responsiva (<100ms para scroll/selección).
- Logs 50k líneas: <300MB RSS; GC saludable.
- Técnicas: buffers circulares, render incremental, throttling de updates, batch de mensajes a UI, memoización de celdas y estilos.

## 12. Empaquetado y distribución

- GoReleaser: binarios linux-amd64/arm64, darwin-universal, windows-amd64.
- Homebrew tap propio inicialmente; PR a homebrew-core cuando esté estable.
- Docker image: `ghcr.io/<owner>/dockcrew:TAG` con `docker` bin no incluido (monta el socket del host). Entrada:
  - `docker run --rm -it -v /var/run/docker.sock:/var/run/docker.sock -v ~/.config/dockcrew:/.config/dockcrew ghcr.io/<owner>/dockcrew:latest`
- Script instalador: descarga binario + checksum + instalación en `~/.local/bin`.

## 13. CI/CD

GitHub Actions workflows:
- `ci.yml`:
  - Go 1.22; `go vet`, `golangci-lint`, `go test ./...` (con docker disponible si es posible)
- `release.yml`:
  - En tags `v*`: GoReleaser build y publicación a Releases; build/push de imagen a GHCR.
- `nightly.yml` (opcional): build nightly y auto-update flag.

## 14. Telemetría (opt-in)

- Eventos: app_start, page_view, action_invoked (agregados), versión, OS.
- Transporte: HTTP POST a endpoint configurable; reintentos limitados; off por defecto.

## 15. Flags y feature gates

- `--no-compose` para desactivar soporte Compose.
- `--no-mouse` para desactivar mouse.
- `--logs-max-lines`, `--logs-since`, `--logs-tail` para overrides rápidos.

## 16. Extensibilidad (fase posterior)

- Definición de comandos personalizados vía YAML:
```yaml
commands:
  - name: "prune all"
    scope: "global"  # global|container|image|volume|network
    run: "docker system prune -f"
    confirm: true
  - name: "exec zsh"
    scope: "container"
    run: "docker exec -it {{.ContainerID}} zsh"
```
- Sandbox y env seguro en fases futuras (JS/WASM a evaluar).

## 17. Estrategia de pruebas

- Unit tests: domain, logs parser, config loader, keymap.
- Integration: docker client contra daemon local (con etiqueta `integration` para activar en CI en contenedores linux).
- Snapshot tests UI (golden files) para tablas/leyendas.
- End-to-end (opcional): script que lanza la TUI en modo headless y simula entradas.

## 18. Versionado y compatibilidad

- SemVer (MAJOR.MINOR.PATCH)
- Compatibilidad mínima Docker Engine API 1.41; degradación funcional si API menor.

## 19. Glosario

- TUI: Terminal User Interface
- Stack/Proyecto: grupo Compose identificado por `com.docker.compose.project`
- Tail/Since: parámetros de recorte de logs

---

Checklist de salida de arquitectura (para TodoList):
- [ ] Estructura de módulos creada
- [ ] Cliente Docker con listas/acciones y tests básicos
- [ ] Soporte Compose (ls, ps, up, down, restart)
- [ ] TUI base (Bubble Tea), tablas y navegación
- [ ] Visor de logs con buffer y filtros
- [ ] Config loader+validator y `--export-default-config`
- [ ] Empaquetado con GoReleaser y contenedor GHCR
- [ ] CI básico (lint, test, build)
- [ ] Doc de instalación y atajos
