.PHONY: test build proto
	

MODULES := $(shell find . -maxdepth 2 -name 'go.mod' -not -path './.git/*' -not -path './vendor/*' -exec dirname {} \; | sort | grep -v '^\.$$')

define modules
	@failures=""; \
	for dir in $(MODULES); do \
		echo "==> $$dir"; \
		if ! (cd $$dir && $(1)); then \
			failures="$$failures $$dir"; \
		fi; \
	done; \
	if [ -n "$$failures" ]; then \
		echo "FAILED MODULES:$$failures"; \
		exit 1; \
	fi
endef

#################################################################################
# BUILD COMMANDS
#################################################################################
tidy:
	$(call modules, go mod tidy)

tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go

generate: tools
	protoc --proto_path=proto --go_out=grpc/proto --go_opt=paths=source_relative options/v1/auth.proto
	protoc --proto_path=proto --proto_path=grpc/proto --go_out=grpc/proto --go_opt=paths=source_relative test/test.proto
	$(call modules, go generate ./...)

#################################################################################
# TEST COMMANDS
#################################################################################
test:
	$(call modules, go test -cover ./...)

lint:
	$(call modules, go tool golangci-lint run ./...)

cover:
	go test -coverpkg ./internal/... -coverprofile coverage.out ./... && go tool cover -html=coverage.out

vuln: 
	$(call modules, go tool govulncheck -test ./...)
