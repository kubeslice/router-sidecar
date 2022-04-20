.PHONY: compile
compile: ## Compile the proto file.
	protoc -I pkg/proto pkg/proto/router_sidecar.proto --go_out=paths=source_relative:pkg/proto --go-grpc_out=pkg/proto --go-grpc_opt=paths=source_relative

.PHONY: kubeslice-router-sidecar
kubeslice-router-sidecar: ## Build and run avesha sidecar.
	go build -race -ldflags "-s -w" -o bin/kubeslice-router-sidecar main.go
	bin/kubeslice-router-sidecar
