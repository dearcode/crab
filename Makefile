all: crab

FILES := $$(find . -name '*.go' | grep -vE 'vendor') 
SOURCE_PATH := orm validation cache server log util


golint:
	go get github.com/golang/lint/golint  

megacheck:
	go get honnef.co/go/tools/cmd/megacheck

lint: golint megacheck
	@for path in $(SOURCE_PATH); do echo "golint $$path"; golint $$path"/..."; done;
	@for path in $(SOURCE_PATH); do echo "gofmt -s -l -w $$path";  gofmt -s -l -w $$path;  done;
	go tool vet $(FILES) 2>&1
	megacheck ./...

clean:
	@rm -rf bin

fmt: 
	@for path in $(SOURCE_PATH); do echo "gofmt -s -l -w $$path";  gofmt -s -l -w $$path;  done;


crab:
	go build -o bin/$@ -ldflags '$(LDFLAGS)' ./main.go


test:
	@for path in $(SOURCE_PATH); do echo "go test ./$$path"; go test "./"$$path; done;


