
.PHONY: test
test:
	docker build  -t go-lcov-summary:latest .
	docker build -f Dockerfile.lcov -t lcov-test:latest .
	go test -v ./cmd/go-lcov-summary -run TestIntegrationLCOVSummary