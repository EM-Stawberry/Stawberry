golint:
	golangci-lint run -c .golangci.yaml
.PHONY:golint

gofmt:
	gofumpt -l -w .
	goimports -w .
.PHONY:gofmt

MOCKS_DESTINATION=tests/mocks
.PHONY: mocks
# put the files with interfaces you'd like to mock in prerequisites
# wildcards are allowed
mocks: internal/handler/user.go
	@echo "Generating mocks..."
	@rm -rf $(MOCKS_DESTINATION)
	@for file in $^ ; do \
		out_path=$$(echo $$file | sed 's|^internal/||'); \
		out_dir=$$(dirname $(MOCKS_DESTINATION)/$$out_path); \
		mkdir -p $$out_dir; \
		mockgen -source=$$file -destination=$(MOCKS_DESTINATION)/$$out_path; \
	done

# Frontend
npm-install:
	cd frontend && npm install

npm-run:
	cd frontend && npm run dev
