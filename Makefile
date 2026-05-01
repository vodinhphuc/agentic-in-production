.PHONY: up down dev-backend dev-frontend install hooks gen \
        test test-backend test-frontend test-e2e \
        lint lint-backend lint-frontend fmt verify logs admin-password

up:
	docker compose up -d postgres

down:
	docker compose down

dev-backend:
	cd backend && air

dev-frontend:
	cd frontend && pnpm dev

install:
	cd backend && go mod download
	cd frontend && pnpm install

hooks:
	cp scripts/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit

gen:
	cd protocols && bash gen-types.sh

test: test-backend test-frontend
test-backend:
	cd backend && go test ./...
test-frontend:
	cd frontend && pnpm test
test-e2e:
	cd frontend && pnpm playwright test

lint: lint-backend lint-frontend
lint-backend:
	cd backend && go vet ./... && golangci-lint run
lint-frontend:
	cd frontend && pnpm lint && pnpm typecheck

fmt:
	cd backend && gofmt -w . && goimports -w .
	cd frontend && pnpm fmt

# the merge gate: lint + tests + codegen sync check
verify: lint test
	cd protocols && bash gen-types.sh
	@git diff --exit-code -- '*.gen.ts' '*.gen.go' 2>/dev/null || \
	  (echo "Generated files out of sync. Run 'make gen' and commit."; exit 1)

logs:
	docker compose logs -f $(SVC)

admin-password:
	@cd backend && go run ./cmd/admin-password
