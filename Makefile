.PHONY: test build
	
#################################################################################
# TEST COMMANDS
#################################################################################
test:
	go test -cover ./... 

lint:
	go tool golangci-lint run ./...

cover:
	go test -coverpkg ./internal/... -coverprofile coverage.out ./... && go tool cover -html=coverage.out

vuln: 
	go tool govulncheck -test ./...
