# kbl-controller operator image.
# Build from repo root:
#   docker build -f controller/docker/kbl-controller/Dockerfile -t kbl-controller:lab .

FROM golang:1.23-bookworm AS go-build
WORKDIR /src
COPY controller/go.mod controller/go.sum ./controller/
WORKDIR /src/controller
RUN go mod download
COPY controller/ ./
RUN CGO_ENABLED=0 go build -o /kbl-controller ./cmd/kbl-controller

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=go-build /kbl-controller /kbl-controller
USER nonroot:nonroot
ENTRYPOINT ["/kbl-controller"]
