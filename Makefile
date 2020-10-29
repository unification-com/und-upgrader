.PHONY: test cover

TEST_RESULTS ?= coverage

build: clean
	go build -mod=readonly -o build/und_upgrader ./

test:
	go test -mod=readonly .

cover:
	mkdir -p $(TEST_RESULTS)
	go test -mod=readonly -timeout 1m -coverprofile=$(TEST_RESULTS)/cover.out -covermode=atomic .
	go tool cover -html=$(TEST_RESULTS)/cover.out -o $(TEST_RESULTS)/coverage.html

clean:
	rm -rf build/
