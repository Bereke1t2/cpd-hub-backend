# Template for Phase 1 — copy to: ./Dockerfile (repo root)
# Multi-stage build: compile a static binary, run it on distroless (non-root, ~15MB).

# ---- build stage ----
FROM golang:1.22-alpine AS build
WORKDIR /src

# Cache deps first
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

# ---- run stage ----
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/server /app/server
# migrations are applied by the migrate CLI / on boot; bundle them if you run them in-container
COPY --from=build /src/migrations /app/migrations

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
