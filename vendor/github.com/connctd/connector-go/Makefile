GO_BUILD_ENV_VARS			?= CGO_ENABLED=0 GO111MODULE=on GOFLAGS=-mod=vendor
GO_BUILD_ENV_TEST_VARS			?= CGO_ENABLED=1 GO111MODULE=on GOFLAGS=-mod=vendor
GO_COVER 				?= $(GO_BUILD_ENV_TEST_VARS) go test -race -covermode=atomic -coverprofile=single.coverprofile
GO_TEST					?= $(GO_BUILD_ENV_TEST_VARS) go test -v -race $(shell $(GO_BUILD_ENV_VARS) go list ./... | grep -v /vendor/)

test:
	$(GO_TEST)

cover:
	@echo Running tests
	@$(eval PKGS := $(shell $(GO_BUILD_ENV_VARS) go list ./...))
	@echo "mode: atomic" >  merged.coverprofile
	@$(foreach PKG, $(PKGS), $(GO_COVER) $(PKG) || exit 1 ; cat single.coverprofile | grep -v "mode:" >> merged.coverprofile || true;)
	@$(GO_BUILD_ENV_VARS) go tool cover --html ./merged.coverprofile -o coverage.html