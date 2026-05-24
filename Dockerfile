# syntax=docker/dockerfile:1.7

# ─── Stage 1: build a static binary ─────────────────────────────────────────
# Pin a specific Go minor for reproducible builds. Bump it deliberately.
FROM golang:1.22-bookworm AS build

WORKDIR /src

# Cache module downloads as their own layer. Empty in this project today,
# but future-proof so adding deps doesn't bust the whole build cache.
COPY go.mod go.sum* ./
RUN go mod download

# Copy the rest of the source.
COPY . .

# Build a fully static binary so the final image can be distroless/static.
# - CGO_ENABLED=0  : no glibc, no dynamic linker.
# - -trimpath      : reproducible paths (no /src/... in stack traces).
# - -ldflags -s -w : strip debug + symbol tables for a smaller binary.
# - GOOS / GOARCH  : honor BuildKit's TARGETOS/TARGETARCH for multi-arch.
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags='-s -w' -o /out/golang-tut .

# ─── Stage 2: minimal runtime ───────────────────────────────────────────────
# distroless/static gives us CA certs (for HTTPS to the Go Playground),
# /etc/passwd with a `nonroot` user, and tzdata. No shell, no package
# manager, no nothing else. Final image weighs ~20MB.
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

COPY --from=build /out/golang-tut /golang-tut

# Cloud Run injects $PORT (default 8080) and ignores EXPOSE, but EXPOSE
# is still useful for local `docker run -p` and image documentation.
EXPOSE 8080

# Default to the Playground runner so the container is safe by default.
# Override with `-e RUNNER=local` if you really know what you're doing.
ENV RUNNER=playground

USER nonroot:nonroot
ENTRYPOINT ["/golang-tut"]
