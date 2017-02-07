all: petrel

FILES := $$(find . -name '*.go' | grep -vE 'vendor') 
SOURCE_PATH := handler orm validation cache server

unused:
	go get honnef.co/go/unused/cmd/unused

golint:
	go get github.com/golang/lint/golint  

lint: golint unused
	@for path in $(SOURCE_PATH); do echo "golint $$path"; golint $$path"/..."; done;
	@for path in $(SOURCE_PATH); do echo "unused $$path"; unused "./"$$path; done;
	@for path in $(SOURCE_PATH); do echo "gofmt -s -l -w $$path";  gofmt -s -l -w $$path;  done;
	go tool vet $(FILES) 2>&1
	go tool vet --shadow $(FILES) 2>&1

clean:
	@rm -rf bin

fmt: 
	@for path in $(SOURCE_PATH); do echo "gofmt -s -l -w $$path";  gofmt -s -l -w $$path;  done;


petrel:
	go build -o bin/$@ -ldflags '$(LDFLAGS)' ./main.go


test:
	@for path in $(SOURCE_PATH); do echo "go test ./$$path"; go test "./"$$path; done;


