.PHONY: test build
	

MODULES := $(shell find . -name 'go.mod' -exec dirname {} \;)

define modules
	failures=""; \
	echo "$(MODULES)" | xargs -n1 -I{} sh -c 'cd {} && go test -cover ./... || failures="$$failures {}"'; \
	if [ -n "$$failures" ]; then \
		echo "FAILED MODULES:$$failures"; \
		exit 1; \
	fi
endef

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
	go tool govulncheck -test ./...
