.PHONY: compile
compile: ## Compile the proto file.
	protoc -I pkg/sidecar/sidecarpb pkg/sidecar/sidecarpb/router_sidecar.proto --go_out=paths=source_relative:pkg/sidecar/sidecarpb --go-grpc_out=pkg/sidecar/sidecarpb --go-grpc_opt=paths=source_relative

.PHONY: router-sidecar
router-sidecar: ## Build and run router sidecar.
	go build -race -ldflags "-s -w" -o bin/router-sidecar main.go

.PHONY: docker-build
docker-build: router-sidecar
	docker build -t router-sidecar:latest-release --build-arg PLATFORM=amd64 . && docker tag router-sidecar:latest-release docker.io/aveshasystems/router-sidecar:latest-stable

.PHONY: docker-push
docker-push:
	docker push docker.io/aveshasystems/router-sidecar:latest-stable
