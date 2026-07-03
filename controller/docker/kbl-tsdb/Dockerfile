# kbl-tsdb node-local store server.
# Build from repo root:
#   docker build -f controller/docker/kbl-tsdb/Dockerfile -t kbl-tsdb:lab .

FROM golang:1.23-bookworm AS go-build
WORKDIR /src
COPY controller/go.mod controller/go.sum ./controller/
WORKDIR /src/controller
RUN go mod download
COPY controller/ ./
RUN CGO_ENABLED=0 go build -o /kbl-tsdb ./cmd/kbl-tsdb

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=go-build /kbl-tsdb /kbl-tsdb
USER nonroot:nonroot
EXPOSE 9090
ENTRYPOINT ["/kbl-tsdb"]
