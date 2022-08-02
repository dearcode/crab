all: crab

SOURCE_PATH := orm validation cache log util http

golint:
	go install golang.org/x/lint/golint@latest

staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest

lint: golint staticcheck
	@for path in $(SOURCE_PATH); do echo "golint $$path"; golint $$path"/..."; done;
	@for path in $(SOURCE_PATH); do echo "gofmt -s -l -w $$path";  gofmt -s -l -w $$path;  done;
	go vet ./...
	staticcheck ./...

clean:
	@rm -rf bin

fmt: 
	@for path in $(SOURCE_PATH); do echo "gofmt -s -l -w $$path";  gofmt -s -l -w $$path;  done;

crab: lint
	@for path in $(SOURCE_PATH); do echo "go test ./$$path"; go test "./"$$path/...; done;


