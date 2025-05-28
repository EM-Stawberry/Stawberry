.PHONY:golint
golint:
	golangci-lint run -c .golangci.yaml

.PHONY:gofmt
gofmt:
	gofumpt -l -w .
	goimports -w .

MOCKS_DESTINATION=tests/mocks
.PHONY: mocks
# put the files with interfaces you'd like to mock in prerequisites
# wildcards are allowed
mocks: internal/handler/user.go internal/domain/service/token/token.go
	@echo "Generating mocks..."
	@rm -rf $(MOCKS_DESTINATION)
	@for file in $^ ; do \
		out_path=$$(echo $$file | sed 's|^internal/||'); \
		out_dir=$$(dirname $(MOCKS_DESTINATION)/$$out_path); \
		mkdir -p $$out_dir; \
		mockgen -source=$$file -destination=$(MOCKS_DESTINATION)/$$out_path; \
	done

.PHONY: test
test:
	go test -v -coverprofile=cov.out ./...
	go tool cover -func=cov.out

coverage:
	go tool cover -html=cov.out


# Frontend
npm-install:
	cd frontend && npm install

npm-run:
	cd frontend && npm run dev

# Deploy
docker-build:
	docker build --platform linux/amd64 -t localhost:5000/stawberry -f deploy/Dockerfile .

docker-push:
	docker push localhost:5000/stawberry

compose-up:
	docker-compose -f deploy/docker-compose.yml up -d

compose-down:
	docker-compose -f deploy/docker-compose.yml down
