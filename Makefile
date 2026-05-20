.PHONY: build-all push-all

VERSION ?= 1.0.3
REPO ?= adamad7

SERVICES = user-service catalog-service showtime-service booking-service frontend

build-all:
	@for service in $(SERVICES); do \
		echo "========================"; \
		echo "Building $$service..."; \
		echo "========================"; \
		docker buildx build --platform linux/amd64,linux/arm64 \
			--sbom=true \
			-t $(REPO)/$$service:$(VERSION) \
			./$$service; \
	done

push-all:
	@for service in $(SERVICES); do \
		echo "========================"; \
		echo "Pushing $$service..."; \
		echo "========================"; \
		docker buildx build --platform linux/amd64,linux/arm64 \
			--sbom=true \
			-t $(REPO)/$$service:$(VERSION) \
			--push \
			./$$service; \
	done
