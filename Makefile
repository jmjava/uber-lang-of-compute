.PHONY: build test tidy clean

build:
	cd controller && go build -o bin/kbl-compute ./cmd/kbl-compute
	cd controller && go build -o bin/kbl-controller ./cmd/kbl-controller

test:
	cd controller && go test ./...

tidy:
	cd controller && go mod tidy

clean:
	rm -rf controller/bin controller/kbl-compute

install-crds:
	kubectl apply -f crds/

run-controller-local:
	cd controller && go run ./cmd/kbl-controller --store-root /tmp/kbl-store

run-finance-example:
	cd controller && go build -o bin/kbl-compute ./cmd/kbl-compute
	./controller/bin/kbl-compute --workflow examples/finance-curve-snapshot/workflow.yaml \
		--store /tmp/kbl-finance/store.db --replay-log /tmp/kbl-finance/replay.json
