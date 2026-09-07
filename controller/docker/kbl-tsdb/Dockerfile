# kbl-tsdb node-local store server.
# Build from repo root:
#   docker build -f controller/docker/kbl-tsdb/Dockerfile -t kbl-tsdb:lab .
#
# CGO is required: the server imports mattn/go-sqlite3.

FROM golang:1.23-bookworm AS go-build
RUN apt-get update && apt-get install -y --no-install-recommends gcc libc6-dev \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY controller/go.mod controller/go.sum ./controller/
WORKDIR /src/controller
RUN go mod download
COPY controller/ ./
RUN CGO_ENABLED=1 go build -o /kbl-tsdb ./cmd/kbl-tsdb

FROM gcr.io/distroless/cc-debian12:nonroot
COPY --from=go-build /kbl-tsdb /kbl-tsdb
USER nonroot:nonroot
EXPOSE 9090
ENTRYPOINT ["/kbl-tsdb"]
