.PHONY: build test tidy clean docker-domino-runner docker-domino-runner-julia \
	docker-kbl-controller docker-kbl-tsdb lab-up lab-down lab-volcano-install lab-openkruise-install cdk-synth

build:
	cd controller && go build -o bin/kbl-compute ./cmd/kbl-compute
	cd controller && go build -o bin/kbl-controller ./cmd/kbl-controller
	cd controller && go build -o bin/domino-runner ./cmd/domino-runner
	cd controller && go build -o bin/kbl-tsdb ./cmd/kbl-tsdb

test:
	cd controller && go test ./...

docker-domino-runner:
	docker build -f controller/docker/domino-runner/Dockerfile \
		-t ghcr.io/jmjava/kbl-domino-runner:latest .

docker-domino-runner-julia:
	docker build -f controller/docker/domino-runner-julia/Dockerfile \
		-t ghcr.io/jmjava/kbl-domino-runner-julia:latest .

docker-kbl-controller:
	docker build -f controller/docker/kbl-controller/Dockerfile \
		-t kbl-controller:lab .

docker-kbl-tsdb:
	docker build -f controller/docker/kbl-tsdb/Dockerfile \
		-t kbl-tsdb:lab .

lab-up:
	chmod +x lab/scripts/*.sh
	./lab/scripts/up.sh

lab-down:
	chmod +x lab/scripts/*.sh
	./lab/scripts/down.sh

lab-volcano-install:
	chmod +x lab/scripts/install-volcano.sh
	./lab/scripts/install-volcano.sh

lab-openkruise-install:
	chmod +x lab/scripts/install-openkruise.sh
	./lab/scripts/install-openkruise.sh

cdk-synth:
	cd infra/aws/cdk && npm install && npm run build && npm run synth

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
