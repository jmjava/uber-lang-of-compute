# Go-only domino-runner for builtin:* domino steps (smaller than Julia image).
#
# Build from repo root:
#   docker build -f controller/docker/domino-runner/Dockerfile -t kbl-domino-runner .

FROM golang:1.23-bookworm AS go-build
WORKDIR /src
COPY controller/go.mod controller/go.sum ./controller/
WORKDIR /src/controller
RUN go mod download
COPY controller/ ./
RUN CGO_ENABLED=0 go build -o /domino-runner ./cmd/domino-runner

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=go-build /domino-runner /domino-runner
ENTRYPOINT ["/domino-runner"]
