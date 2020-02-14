ci: build format lint test

test:
	-go test ./... -v -cover -coverprofile=cover.out
	go tool cover -html=cover.out

lint:
	golangci-lint run ./... -v
	gocyclo -over 15 .

format:
	goimports -l -w -local "listize" .
	gofmt -l -w -s .

build:
	go build -o program
