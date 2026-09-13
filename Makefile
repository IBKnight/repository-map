.PHONY: build test vet fmt fmt-check tidy ci clean

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

fmt-check:
	@out="$$(gofmt -l .)"; \
	if [ -n "$$out" ]; then echo "not gofmt-formatted:"; echo "$$out"; exit 1; fi

tidy:
	go mod tidy

ci: fmt-check vet build test

clean:
	rm -f repomap
