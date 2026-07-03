# Multi-stage image: domino-runner + pre-instantiated Julia project (ADR 0023 Model B).
#
# Build from repo root:
#   docker build -f controller/docker/domino-runner-julia/Dockerfile -t kbl-domino-runner-julia .
#
# Run locally:
#   docker run --rm -e KBL_COMMAND=julia:identity \
#     -e KBL_INPUT=/in.json -e KBL_OUTPUT=/out.json \
#     -v "$PWD/tmp-in.json:/in.json" -v "$PWD/tmp-out.json:/out.json" \
#     kbl-domino-runner-julia

FROM golang:1.23-bookworm AS go-build
WORKDIR /src
COPY controller/go.mod controller/go.sum ./controller/
WORKDIR /src/controller
RUN go mod download
COPY controller/ ./
RUN CGO_ENABLED=0 go build -o /domino-runner ./cmd/domino-runner

FROM julia:1.12-bookworm AS julia-deps
WORKDIR /opt/kbl/julia
COPY controller/julia/ ./
RUN julia --project=. -e 'using Pkg; Pkg.instantiate(); using JSON; using FinanceModels; using SpecialFunctions'

FROM julia:1.12-bookworm
WORKDIR /opt/kbl/julia
COPY --from=julia-deps /opt/kbl/julia /opt/kbl/julia
COPY --from=go-build /domino-runner /usr/local/bin/domino-runner
COPY controller/julia/scripts ./scripts

ENV KBL_JULIA_PROJECT=/opt/kbl/julia
ENV KBL_JULIA_BIN=julia
ENV PATH="/usr/local/bin:${PATH}"

ENTRYPOINT ["/usr/local/bin/domino-runner"]
