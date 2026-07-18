SERVER_DIR = server
BUILD_FLAGS = -tags netgo,osusergo -trimpath -ldflags="-s -w -buildid="

.PHONY: build test test-integration lint cover clean setup

setup:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

build:
	cd $(SERVER_DIR) && CGO_ENABLED=0 go build $(BUILD_FLAGS) -o kindling .

test:
	cd $(SERVER_DIR) && go test -cover -coverprofile=coverage.out ./...

test-integration:
	cd $(SERVER_DIR) && firebase emulators:exec --only firestore 'go test -tags=integration ./...'

lint:
	cd $(SERVER_DIR) && golangci-lint run && gosec ./...

cover:
	cd $(SERVER_DIR) && go test -cover ./...

clean:
	rm -f $(SERVER_DIR)/kindling $(SERVER_DIR)/kindling-*
