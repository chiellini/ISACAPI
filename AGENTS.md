# Repository Guidelines

## Project Structure & Module Organization

`backend/` is the Go service: `cmd/server` is the application entry point, `internal/` contains handlers, services, repositories, middleware, and integrations, `ent/` holds generated ORM code, and `migrations/` holds schema changes. Keep the handler → service → repository dependency direction; handlers and services must not import repositories directly.

`frontend/` is a Vue 3 + TypeScript Vite application. Put screens in `src/views/`, reusable UI in `src/components/`, state in `src/stores/`, API clients in `src/api/`, and translations in `src/i18n/`. Static and deployment materials live in `assets/`, `docs/`, `deploy/`, and `scripts/`.

## Build, Test, and Development Commands

- `pnpm --dir frontend install --frozen-lockfile` installs the locked frontend dependencies.
- `pnpm --dir frontend dev` starts Vite; it proxies API traffic to `http://localhost:8080` by default.
- `make build` builds the Go server and bundles the frontend into `backend/internal/web/dist`.
- `make test` runs backend tests plus frontend linting, type checks, and critical Vitest cases.
- `make -C backend test` runs `go test ./...` and `golangci-lint run ./...`.
- `pnpm --dir frontend run test:run` or `test:coverage` runs Vitest once, with or without coverage.

## Coding Style & Naming Conventions

Format Go with `gofmt`; comply with `backend/.golangci.yml` and handle errors explicitly. Use lowercase package names and `*_test.go` test files. For Vue/TypeScript, use two-space indentation, strict types, `PascalCase.vue` components, `useX.ts` composables, and the `@/` source alias. Keep changes focused; do not reformat unrelated files. Run `pnpm --dir frontend run lint:check` before submitting frontend work.

## Brand and Visual Identity

For product, storefront, documentation, and UI iconography, use the ISAC AI logo at `assets/公司icon.jpeg` as the visual source of truth. Do not substitute generic, third-party, or legacy brand icons unless explicitly requested. Reuse this approved ISAC AI asset in README branding where an icon is displayed.

## Testing Guidelines

Place frontend tests beside the feature in `__tests__/` and name them `*.spec.ts`; Vitest runs in jsdom. The configured coverage thresholds are 80% globally for statements, branches, functions, and lines. Add or update tests for behavior changes, especially API compatibility, billing, authentication, migrations, and visible UI states.

## Commit & Pull Request Guidelines

Follow the existing Conventional Commit style: `feat(auth): support combined admin provider role` or `fix(openai): preserve image function tools`. Use an imperative, scoped subject. PRs should explain the user impact, list validation commands, link relevant issues, and include screenshots for UI changes. Include migration and lockfile updates when applicable. Never commit secrets, local databases, generated `node_modules`, or `.pnpm-store`; review `DEV_GUIDE.md` before changing local setup or dependency tooling.

## Upstream Sync Policy

`README.md`, `README_CN.md`, and `README_JA.md` are fork-owned. Preserve the official links to `https://isacai.space` and `https://isacai.cn` during `upstream` syncs. Reject other sponsors, advertisements, affiliate/referral links, promo codes, and commercial calls to action unless explicitly requested. Reject upstream README changes and review `git diff -- README.md README_CN.md README_JA.md` before committing.
