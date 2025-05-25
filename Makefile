.PHONY:golint
golint:
	golangci-lint run -c .golangci.yaml

.PHONY:gofmt
gofmt:
	gofumpt -l -w .
	goimports -w .

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
