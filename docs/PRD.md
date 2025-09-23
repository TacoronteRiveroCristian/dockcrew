# PRD: DockCrew — TUI moderna para Docker y Compose

Autor: Tú (con apoyo de IA)
Fecha: 2025-09-23
Versión: 0.1 (borrador inicial)

## 1. Contexto y objetivo

DockCrew será una aplicación TUI (Terminal User Interface) para administrar y observar entornos Docker y Docker Compose de forma ágil, moderna y extremadamente fácil de instalar. Se inspira en Ducker (Rust/ratatui), Lazydocker (Go/gocui) y DockTUI (Python/Textual), pero con un foco claro en:

- Experiencia más actual (UI pulida, accesible, con atajos coherentes y búsqueda/filtrado en vivo)
- Instalación de un paso, sin toolchains (binarios precompilados y contenedor oficial listo)
- Rendimiento robusto en hosts con muchos contenedores/logs
- Extensibilidad segura (comandos y paneles personalizados sin reiniciar)

## 2. Público objetivo

- Desarrolladores que trabajan con múltiples contenedores/compose stacks locales
- DevOps/SRE que requieren observabilidad ligera desde terminal
- Equipos que desean una alternativa a Docker Desktop para flujos diarios
- Usuarios en Linux/macOS/WSL2 que prefieren TUI con mouse opcional

## 3. Problema a resolver

- Gestión fragmentada entre varias terminales y comandos
- Curva de aprendizaje de Docker/Docker Compose y memorizar flags
- Herramientas existentes: o no tienen instalación trivial, o no escalan bien con logs/estados masivos, o carecen de ciertas comodidades modernas (paleta de comandos, búsqueda enriquecida, theming accesible)

## 4. Principios de diseño

- Fast-first: UI fluida, baja latencia, consumo contenido de CPU/RAM
- Zero-friction install: binarios para Linux/macOS/Windows, imagen Docker oficial y script único
- Composable UX: paneles con layout flexible, paleta de comandos, fuzzy-search
- Seguro por defecto: acciones destructivas con confirmaciones claras, modo lectura
- Extensible y portable: config YAML + plugins externos aislados

## 5. Alcance del MVP

### 5.1 Funcionalidades núcleo

- Dashboard general: visión de contenedores, imágenes, volúmenes, redes, y stacks Compose
- Gestión de contenedores: start/stop/restart/remove, ver stats (CPU, RAM, IO), abrir shell (exec) con selección de intérprete
- Logs en tiempo real: follow, filtros, regex, resaltado de niveles, límites de memoria, desde `since` configurable
- Compose awareness: agrupar por proyecto, acciones up/down/recreate por stack
- Gestión de imágenes: listar, eliminar, ver capas y tamaño, prune
- Gestión de volúmenes y redes: listar, eliminar (cuando no estén en uso), detalles
- Búsqueda/filtrado: en tablas/paneles con fuzzy match y columnas ordenables con toggle asc/desc
- Paleta de comandos: Ctrl+P (o Ctrl+\) con búsqueda rápida de acciones
- Atajos coherentes y ayuda contextual: leyenda inferior + panel de ayuda rápido con `?`
- Theming y accesibilidad: tema claro/oscuro, alto contraste, tamaño de fuente relativo, soporte parcial de mouse

### 5.2 Plataformas y compatibilidad

- Docker Engine API >= 1.41 (Docker 20.10+)
- Docker Compose v2 (comando `docker compose`), opcional pero soportado
- Sistemas: Linux, macOS, Windows (vía WSL2 para TUI nativa) y Docker container

### 5.3 Instalación (MVP)

- Binarios precompilados: tarballs en Releases + Homebrew tap + AUR/winget más adelante
- Script único: `curl -fsSL https://get.dockcrew.io | bash` (instala en `~/.local/bin` por defecto)
- Imagen oficial: `docker run --rm -it -v /var/run/docker.sock:/var/run/docker.sock -v ~/.config/dockcrew:/.config/dockcrew ghcr.io/<org>/dockcrew:latest`

### 5.4 Configuración

- Archivo YAML: `~/.config/dockcrew/config.yaml` (Linux), `~/Library/Application Support/DockCrew/config.yaml` (macOS), `%APPDATA%\DockCrew\config.yaml` (Windows)
- Variables clave: refresh_interval, logs.tail, logs.since, logs.max_lines, colors/theme, docker.host, compose.enabled, ui.mouse, telemetry (opt-in)
- Comando para exportar defaults: `dockcrew --export-default-config`

## 6. Diferenciadores clave

- Instalación 1-click real: binarios + contenedor oficial listos desde el día 1
- Log Viewer de alto rendimiento con:
  - Virtual scrolling, lazy parsing, highlighting de niveles/HTTP/JSON
  - Búsqueda en vivo con regex y filtros rápidos por nivel/servicio
  - Marcadores temporales y copiar al portapapeles (X11/Wayland/macOS/WSL)
- UX moderna: paleta de comandos, fuzzy-search y layouts memorables por usuario
- Extensibilidad segura: comandos y vistas personalizadas via YAML/JS (sandbox) sin reiniciar

## 7. Requisitos funcionales

1. Contenedores
   - Listar, ordenar (nombre, imagen, estado, puertos, creado)
   - Acciones: start/stop/restart/kill/remove/prune
   - Exec con selección interactiva de shell: bash, sh, zsh, powershell (si existe)
   - Stats en tiempo real (CPU%, Mem, Net I/O, Block I/O)
2. Imágenes
   - Listar con tags/tamaño/fecha
   - Eliminar individual/en lote (dangling toggle), prune
   - Ver layers (si disponible)
3. Volúmenes
   - Listar, ver en uso por contenedores, eliminar si libre, prune
4. Redes
   - Listar, ver driver/scope, eliminar si libre
5. Compose
   - Detectar stacks por `docker compose ls` y agrupar contenedores por proyecto
   - Acciones por stack: up -d, down, restart, recreate service
6. Logs
   - Follow por contenedor/servicio, tail configurable, since configurable
   - Filtros: texto, regex `/pattern/`, por nivel (INFO/WARN/ERROR)
   - Coloreado contextual y pretty-print JSON con expand/collapse
7. Paleta de comandos
   - Acceso con Ctrl+P o Ctrl+\; busca por nombre/descripción/atajos
8. Configuración
   - Carga en cascada: env > ruta explícita > default user > embebido
   - Validación con mensajes claros; fallback seguro a defaults
9. Actualización
   - Chequeo opcional de nueva versión; `dockcrew self-update` (si no es contenedor)
10. Telemetría (opt-in)
   - Eventos mínimos: versión, OS, features usadas; 100% desactivable

## 8. Requisitos no funcionales

- Rendimiento: listas 500+ contenedores fluida; logs 50k líneas con <300MB RAM
- UX: 60 FPS en terminales modernas; operaciones críticas con confirmación
- Seguridad: no enviar datos sensibles; respeto a permisos del socket Docker
- Portabilidad: binarios estáticos cuando sea posible
- Observabilidad: logs internos en `~/.local/share/dockcrew/dockcrew.log`, nivel configurable

## 9. Arquitectura propuesta

- Lenguaje: Rust (ratatui/tuirealm + reqwest + tokio) O Go (Bubble Tea/Charm ecosystem). Decisión:
  - Opción A (Rust) prioriza rendimiento y binarios pequeños
  - Opción B (Go) prioriza ecosistema TUI maduro (Bubble Tea) y facilidad de cross-compile
- Capa Docker:
  - Cliente nativo de Docker Engine (Rust: bollard; Go: moby/moby, docker/cli)
  - Para compose: shell out a `docker compose` con parsing robusto o usar Docker SDK si disponible
- Módulos:
  - core/docker_client
  - domain/{containers,images,volumes,networks,compose}
  - ui/{pages,components,keymap,command_palette}
  - logs/{streamer,parser,highlighter,buffer}
  - config/{loader,validator,watcher}
  - telemetry/{metrics,consent}
  - updates/self_update
- Concurrencia: tokio (Rust) o goroutines (Go)
- Persistencia: config YAML + layout state en `~/.local/share/dockcrew/`

## 10. Experiencia de usuario (UX)

- Layout por pestañas: Contenedores, Imágenes, Volúmenes, Redes, Stacks
- Panel lateral de detalles contextual con acciones rápidas
- Barra inferior con leyenda de atajos + estado (versión, update disponible)
- Paleta de comandos modal; `?` abre ayuda compacta con keymap
- Soporte de mouse opcional; selección de texto sin interferencia (toggle)

## 11. Keybindings (propuesta inicial)

- Global: j/k o ↑/↓ navegar; g/G top/bottom; Tab cambia panel; : prompt; ? ayuda; q salir
- Ordenación por columnas (Shift+letra), similar a Ducker; repetición invierte orden
- Contenedores: s start; t stop; r restart; k kill; l logs; e exec; x remove; P prune
- Imágenes: x remove; D prune dangling; i inspect/layers
- Volúmenes: x remove; P prune; d details
- Redes: x remove; d details
- Compose: U up -d; D down; R recreate; S stack restart
- Paleta: Ctrl+P o Ctrl+\

## 12. Instalación y distribución

- Releases: GitHub Releases con binarios para linux-amd64, linux-arm64, macOS (universal), Windows (amd64)
- Homebrew tap oficial y PR a homebrew-core cuando madure
- Imagen en GHCR y Docker Hub
- Script instalador firma/verificación SHASUM y no requiere root

## 13. Métricas de éxito (MVP)

- Tiempo de primera instalación < 30s (binario) / < 10s (docker run ya cacheado)
- CPU idle ~0–2%; pico < 20% durante polling/logs intensivos
- Abrir logs 10k líneas en < 1s y navegación fluida
- Sesiones/usuarios activos semanales y NPS internos del equipo

## 14. Riesgos y mitigaciones

- Compatibilidad Docker/Compose: test matrix CI contra múltiples versiones
- Permisos del socket: guías para grupo docker y mensajes claros
- Consumo en logs masivos: virtual scrolling + límites configurables
- Cruce de teclas con terminal/tmux: doc y toggles para mouse/keys
- Self-update multiplataforma: fallback a mensajes/links si no soportado

## 15. Roadmap sugerido

- 0.1 (este PRD): Esqueleto del proyecto, cliente Docker, listas básicas, logs tail, acciones start/stop/remove, config loader, empaquetado binario y contenedor
- 0.2: Compose stacks, paleta de comandos, ordenación y filtros, theming
- 0.3: Log viewer avanzado (regex, pretty JSON, marcadores), exec con selector de shell, prune/inspect
- 0.4: Telemetría opt-in, self-update, portapapeles, AUR/Homebrew Core
- 0.5: Extensiones/Plugins (YAML/JS sandbox), perfiles remotos (a evaluar)

## 16. Aceptación (criterios de done MVP)

- Instalable con binario y con docker run
- Navegación fluida por contenedores/imágenes/volúmenes/redes
- Logs follow con filtros básicos y límites de memoria
- Acciones start/stop/restart/remove confiables con confirmaciones
- Tests básicos de integración contra daemon docker local en CI

## 17. Anexos: Comparativa rápida

- Ducker (Rust): gran rendimiento, hotkeys bien pensadas, instalación vía cargo/brew/pacman, sin binarios en releases para todas las plataformas en todo momento; sorting por columnas, exec limitado a bash inicialmente
- Lazydocker (Go): súper popular, soporte amplio (brew, scoop, choco, docker), UI con gocui, mouse support, vista de métricas/graphs y muchas integraciones
- DockTUI (Python+Textual): dockerizado por defecto (cero deps), log viewer rico (regex, highlight, expand JSON), postura moderna de UX

DockCrew combinará instalación simple (como DockTUI), rendimiento (como Ducker) y cobertura funcional + ecosistema (como Lazydocker).
