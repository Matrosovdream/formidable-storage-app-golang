# syntax=docker/dockerfile:1.7

# -----------------------------------------------------------------------------
# Base layer: Go toolchain + module cache. Shared by dev and prod builds.
# -----------------------------------------------------------------------------
FROM golang:1.26-alpine AS base
RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /src
ENV CGO_ENABLED=0 GOFLAGS=-buildvcs=false
COPY go.mod go.sum ./
RUN go mod download

# -----------------------------------------------------------------------------
# Dev image: includes air for hot reload. Source is mounted at runtime.
# -----------------------------------------------------------------------------
FROM base AS dev
RUN go install github.com/air-verse/air@latest
# Note: source is bind-mounted from compose. Don't COPY anything here.
EXPOSE 8080
ENTRYPOINT ["air", "-c", ".air.toml"]

# -----------------------------------------------------------------------------
# Builder: compiles a specific binary (web or worker), selected by BIN arg.
# -----------------------------------------------------------------------------
FROM base AS builder
ARG BIN=web
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/${BIN}

# -----------------------------------------------------------------------------
# Runtime: minimal distroless image with just the binary.
# -----------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot AS runtime
COPY --from=builder /out/app /app
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app"]
