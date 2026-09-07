# kbl-controller operator image.
# Build from repo root:
#   docker build -f controller/docker/kbl-controller/Dockerfile -t kbl-controller:lab .
#
# CGO is required: the operator imports mattn/go-sqlite3.

FROM golang:1.23-bookworm AS go-build
RUN apt-get update && apt-get install -y --no-install-recommends gcc libc6-dev \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY controller/go.mod controller/go.sum ./controller/
WORKDIR /src/controller
RUN go mod download
COPY controller/ ./
RUN CGO_ENABLED=1 go build -o /kbl-controller ./cmd/kbl-controller

FROM gcr.io/distroless/cc-debian12:nonroot
COPY --from=go-build /kbl-controller /kbl-controller
USER nonroot:nonroot
ENTRYPOINT ["/kbl-controller"]
