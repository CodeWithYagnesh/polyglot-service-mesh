DOCKER_USERNAME ?= yagneshjariwala
VERSION ?= latest
SERVICES := auth by_who category entry frontend
.PHONY: all build build-all push deploy clean help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build-all    Build all Docker images"
	@echo "  build-auth   Build only the Auth service"
	@echo "  push-all     Push all images to Docker Hub"
	@echo "  deploy       Apply all K8s manifests (Apps + Istio)"
	@echo "  delete       Delete all K8s resources"
	@echo "  clean        Remove local docker images"

build-all: $(addprefix build-,$(SERVICES))

build-%:
	@echo "🏗️  Building $* service..."
	@if [ "$*" = "frontend" ]; then \
		docker build -t $(DOCKER_USERNAME)/$*-service:$(VERSION) ./frontend; \
	else \
		docker build -t $(DOCKER_USERNAME)/$*-service:$(VERSION) ./backend/$*; \
	fi

push-all: $(addprefix push-,$(SERVICES))

push-%:
	@echo "🚀 Pushing $* service..."
	docker push $(DOCKER_USERNAME)/$*-service:$(VERSION)

deploy:
	@echo "☸️  Deploying to Kubernetes..."
	kubectl apply -f k8s/ns.yaml
	kubectl apply -R -f k8s/apps/
	kubectl apply -f k8s/istio/
	@echo "✅ Deployment complete!"

delete:
	@echo "🔥 Deleting resources..."
	kubectl delete -R -f k8s/apps/
	kubectl delete -f k8s/istio/

minikube-env:
	@echo "Run this command in your terminal:"
	@echo "eval \$$(minikube docker-env)"
