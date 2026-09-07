.PHONY: build test theory-prove review research-up research-down research-test tidy clean docker-domino-runner docker-domino-runner-julia \
	docker-kbl-controller docker-kbl-tsdb lab-up lab-down lab-volcano-install lab-openkruise-install \
	lab-verify-volcano lab-setup-wsl-home lab-volcano-up cdk-synth run-finance-example run-desk-day

build:
	cd controller && go build -o bin/kbl-compute ./cmd/kbl-compute
	cd controller && go build -o bin/kbl-controller ./cmd/kbl-controller
	cd controller && go build -o bin/domino-runner ./cmd/domino-runner
	cd controller && go build -o bin/kbl-tsdb ./cmd/kbl-tsdb

test:
	cd controller && go test ./...

theory-prove:
	cd controller && go test ./pkg/theory/ ./pkg/engine/ ./pkg/wheel/ ./pkg/routing/ ./pkg/replica/ ./pkg/cdc/ ./pkg/hash/ ./pkg/convert/ ./pkg/executor/ ./pkg/builtin/ ./pkg/store/ ./pkg/events/ ./pkg/review/ ./internal/controller/ -count=1

review: build
	cd controller && go build -o bin/kbl-review ./cmd/kbl-review
	cd controller && go test ./pkg/review/ ./pkg/engine/ ./pkg/builtin/ ./pkg/store/ ./pkg/events/ ./pkg/cdc/ ./pkg/replica/ ./internal/controller/ -count=1
	./controller/bin/kbl-review --workflow examples/finance-curve-snapshot/workflow.yaml

research-up:
	chmod +x lab/scripts/research-*.sh
	./lab/scripts/research-up.sh

research-down:
	chmod +x lab/scripts/research-down.sh
	./lab/scripts/research-down.sh

research-test:
	chmod +x lab/scripts/research-*.sh
	./lab/scripts/research-smoke.sh

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

lab-volcano-up:
	chmod +x lab/scripts/*.sh
	./lab/scripts/kind-volcano-up.sh

lab-verify-volcano:
	chmod +x lab/scripts/verify-volcano.sh
	./lab/scripts/verify-volcano.sh

lab-setup-wsl-home:
	chmod +x lab/scripts/setup-wsl-home.sh lab/scripts/*.sh
	./lab/scripts/setup-wsl-home.sh

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

run-desk-day: build
	mkdir -p /tmp/kbl-rates-desk
	./controller/bin/kbl-compute --workflow examples/rates-desk-day/workflow-risk.yaml \
		--store /tmp/kbl-rates-desk/ny.db --replay-log /tmp/kbl-rates-desk/ny-risk.json
	./controller/bin/kbl-compute --workflow examples/rates-desk-day/workflow-mid.yaml \
		--store /tmp/kbl-rates-desk/ny.db --replay-log /tmp/kbl-rates-desk/ln-mid.json
	./controller/bin/kbl-compute --workflow examples/rates-desk-day/workflow-risk-next.yaml \
		--store /tmp/kbl-rates-desk/ny.db --replay-log /tmp/kbl-rates-desk/ny-tplus1.json
