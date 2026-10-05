.PHONY: build swagger
build: ## Build the mojo binary.
	go build -o mojo -buildvcs=false .
	@echo "mojo-strategy-liquidation build successful"

swagger: ## Regenerate docs/swagger.* (docs/docs.go, swagger.json, swagger.yaml)
	swag init -g main.go -o docs --parseInternal --exclude docs,application,infrastructure,domain/persistent
