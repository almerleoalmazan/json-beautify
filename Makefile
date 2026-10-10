.PHONY: build test run-no-tui

build:
	go build -o json-beautify .

test:
	go vet ./... && go test ./...

run-no-tui:
	go run . --no-tui testdata/promos.json
