# Makefile for github.com/clarenous/proxyot

# consts
LINT_REPORT=lint.report
TEST_REPORT=test.report

# commands

lint:
	@echo "make lint: begin"
	@golangci-lint run --no-config --issues-exit-code=0 \
	--exclude=".vscode" --exclude=".idea" --exclude="bin" --exclude="vendor" \
	--timeout=600s | tee $(LINT_REPORT)
	@echo "make lint: end"

test:
	@echo "make test: begin"
	@go test ./... 2>&1 | tee $(TEST_REPORT)
	@echo "make test: end"
