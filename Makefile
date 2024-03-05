VERSION := $(shell git describe --tags --always --dirty="-dev" --match "v*.*.*" || echo "development" )
VERSION := $(VERSION:v%=%)

.PHONY: build-backend
build-backend:
	@cd ./backend && CGO_ENABLED=0 go build \
			-ldflags "-X main.version=${VERSION}" \
			-o ./bin/eth-faucet \
		github.com/flashbots/eth-faucet/cmd

.PHONY: release-backend
release-backend:
	@cd ./backend && goreleaser release --snapshot --clean

.PHONY: build-frontend
build-frontend:
	@cd ./frontend && yarn install && yarn build

.PHONY: docker-compose
docker-compose:
	@docker compose down --remove-orphans
	@docker compose up --build
