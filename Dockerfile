# Imagen mínima para ejecutar dockcrew dentro de contenedor
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags "-s -w" -o /out/dockcrew ./cmd/dockcrew

FROM alpine:3.20
RUN apk add --no-cache ca-certificates bash
# Nota: se asume que el host monta el binario docker y el socket
# o que 'docker' está disponible en el contenedor (no recomendado en prod)
COPY --from=builder /out/dockcrew /usr/local/bin/dockcrew
ENTRYPOINT ["dockcrew"]
