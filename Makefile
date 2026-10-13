.DEFAULT_GOAL := check
.PHONY: check secrets fmt vet test vuln hooks run deploy

# What CI runs.
check: secrets fmt vet test

# Fails if any secret-looking string is in the working tree or history.
secrets:
	gitleaks detect --source . --redact --no-banner

fmt:
	@test -z "$$(gofmt -l .)" || { gofmt -l .; echo "run: gofmt -w ."; exit 1; }

vet:
	go vet ./...

test:
	go test -race -count=1 ./...

# Known vulnerabilities in the code paths we call.
vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Installs the pre-commit hook (gitleaks + gofmt).
hooks:
	git config core.hooksPath .githooks

# Local demo on http://localhost:8080 (set GEMINI_API_KEY for live calls).
run:
	go run ./cmd/server

# Cloud Run: make deploy PROJECT=<gcp-project> [REGION=asia-southeast2]
REGION ?= asia-southeast2
deploy:
	infra/deploy.sh $(PROJECT) $(REGION)
