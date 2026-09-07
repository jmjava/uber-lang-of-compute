# Research test environment (Cloud Agent / local).
# Do not COPY the project; Cursor mounts the workspace at agent start.
FROM golang:1.23-bookworm

RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential \
        ca-certificates \
        curl \
        git \
        make \
        python3 \
    && rm -rf /var/lib/apt/lists/*

# CGO sqlite (mattn/go-sqlite3) needs gcc from build-essential.
ENV CGO_ENABLED=1 \
    GOTOOLCHAIN=local \
    GOPROXY=https://proxy.golang.org,direct

WORKDIR /workspace
