.PHONY: lint
lint:
	cd app; golangci-lint run -v --out-format tab --path-prefix app/

.PHONY: lint-ci
lint-ci:
	cd app; golangci-lint run -v --timeout=15m

.PHONY: docker-local-env-down
docker-local-env-down:
	@docker-compose -f docker-compose.local.yml stop

.PHONY: docker-local-env-up
docker-local-env-up: docker-local-env-down
	@docker-compose -f docker-compose.local.yml up -d

.PHONY: run
run:
	@cd app; go build -o app cmd/software/main.go && ./app -config ../configs/config.local.yml

.PHONY: test
test:
	cd app; CGO_ENABLED=1 go test -v -race -count=1 ./...

.PHONY: test-integration
test-integration: up-test-env
	cd app; CGO_ENABLED=1 go test --tags="integration" -v -race -count=1 -p=1 ./... -run ^TestIntegration*

.PHONY: migration-add
migration-add:
	@read -p "Enter migration name (ex. add_id_column): " name; \
	goose -dir app/internal/dal/postgres/migrations create $$name sql


.PHONY: build-docker
build-docker:
	@CGO_ENABLED=1 docker build -t software-test:local  \
		--build-arg goprivate="github.com/*" \
		--build-arg machine="github.com" \
		--build-arg login=${GITLAB_LOGIN} \
		--build-arg password=${GITLAB_TOKEN} \
		-f Dockerfile .

.PHONY: build-docker-linux
build-docker-linux:
	@GOOS=linux GOARCH=amd64 docker buildx build --platform linux/amd64 --load -t software-test:local  \
		--build-arg goprivate="github.com/*" \
		--build-arg machine="github.com" \
		--build-arg login=${GITLAB_LOGIN} \
		--build-arg password=${GITLAB_TOKEN} \
		-f Dockerfile .

.PHONY: build-docker-arm64
build-docker-arm64:
	@GOARCH=arm64 CGO_ENABLED=1 docker build -t software-test:local  \
		--build-arg goprivate="github.com/*" \
		--build-arg machine="github.com" \
		--build-arg login=${GITLAB_LOGIN} \
		--build-arg password=${GITLAB_TOKEN} \
		-f Dockerfile.arm64.dockerfile .

.PHONY: mockgen
mockgen:
	@cd app; go mod tidy && \
        mockery --dir=internal/domain/game/service \
				--name=gameStorage \
				--structname=gameStorageMock \
				--filename=service_mock.go \
				--output=internal/domain/game/service/ \
				--outpkg=service