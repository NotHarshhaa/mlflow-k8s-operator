# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: ## Generate manifests e.g. CRD, RBAC etc.
	$(GOBIN)/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: ## Generate code
	$(GOBIN)/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt vet ## Run tests.
	go test ./... -coverprofile cover.out

##@ Build

.PHONY: build
build: fmt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: run
run: fmt vet ## Run a controller from your host.
	go run ./main.go

.PHONY: docker-build
docker-build: ## Build the docker image.
	docker build -t ${IMG} .

.PHONY: docker-build-multi
docker-build-multi: ## Build the docker image for multiple platforms.
	docker buildx build --platform linux/amd64,linux/arm64 -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push the docker image.
	docker push ${IMG}

.PHONY: docker-build-push-ghcr
docker-build-push-ghcr: ## Build and push to GHCR.
	$(MAKE) docker-build IMG=ghcr.io/NotHarshhaa/mlflow-k8s-operator:0.4.0
	docker push ghcr.io/NotHarshhaa/mlflow-k8s-operator:0.4.0

.PHONY: docker-build-push-dockerhub
docker-build-push-dockerhub: ## Build and push to Docker Hub.
	$(MAKE) docker-build IMG=NotHarshhaa/mlflow-k8s-operator:0.4.0
	docker push NotHarshhaa/mlflow-k8s-operator:0.4.0

.PHONY: docker-build-push-all
docker-build-push-all: ## Build and push to both registries.
	$(MAKE) docker-build-push-ghcr
	$(MAKE) docker-build-push-dockerhub

##@ Deployment

.PHONY: install
install: ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

.PHONY: deploy
deploy: ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: helm-deploy
helm-deploy: ## Deploy using Helm chart
	helm install mlflow-operator ./charts/mlflow-k8s-operator

.PHONY: helm-uninstall
helm-uninstall: ## Uninstall Helm chart
	helm uninstall mlflow-operator

.PHONY: helm-lint
helm-lint: ## Lint Helm chart
	helm lint ./charts/mlflow-k8s-operator

.PHONY: helm-package
helm-package: ## Package Helm chart
	helm package ./charts/mlflow-k8s-operator

##@ Dependencies

.PHONY: deps
deps: ## Download dependencies
	go mod download
	go mod tidy

.PHONY: tools
tools: ## Install controller-gen and kustomize
	@echo "Installing controller-gen..."
	@go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.11.1
	@echo "Installing kustomize..."
	@go install sigs.k8s.io/kustomize/kustomize/v4@v4.5.7

##@ Cleanup

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f cover.out
