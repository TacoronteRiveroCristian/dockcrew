# DockCrew

TUI moderna para gestionar Docker y Docker Compose desde la terminal, con foco en instalación sencilla y alto rendimiento.

- Documento de Producto (PRD): consulte `docs/PRD.md` para la visión completa, alcance del MVP, requisitos y roadmap.
- Especificaciones Técnicas (SPEC): consulte `docs/SPEC.md` para diseño técnico, arquitectura, contratos y CI/CD.

## Estado

Borrador inicial del PRD y SPEC disponibles en `docs/`. Implementación en preparación.

## Estado del Repositorio

Este repositorio ha sido inicializado localmente. Asegúrate de configurar el remoto y realizar el primer push.

## Inspiraciones

- Ducker (Rust)
- Lazydocker (Go)
- DockTUI (Python)

## Construcción y ejecución (Go 1.22+)

- Requisitos: Go 1.22+
- Build local: `make build`
- Ejecutar versión: `make run`

Si prefieres sin Makefile:
- `go build -ldflags "-X main.version=0.0.0-dev" -o bin/dockcrew ./cmd/dockcrew`
- `./bin/dockcrew --version`

## Roadmap corto (MVP)

Ver `TODO.md` y `docs/SPEC.md` para el checklist de implementación.

## Comandos nuevos

- `dockcrew containers [-a|--all]`
- `dockcrew start <id|name>`
- `dockcrew stop <id|name>`
- `dockcrew restart <id|name>`
- `dockcrew rm [-f] <id|name>`

## Ejecutarlo en contenedor (opcional)

Construir imagen local y ejecutar montando el socket de Docker del host:

- Build imagen: `docker build -t dockcrew:dev .`
- Run: `docker run --rm -it -v /var/run/docker.sock:/var/run/docker.sock dockcrew:dev containers`
