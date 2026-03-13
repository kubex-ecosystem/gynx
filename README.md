# GNyx

Portuguese (Brazil) version: [docs/README.pt-BR.md](./docs/README.pt-BR.md)

## Table of Contents

- [Overview](#overview)
- [Current Product State](#current-product-state)
- [Core Capabilities](#core-capabilities)
- [Ecosystem Context](#ecosystem-context)
- [Architecture Overview](#architecture-overview)
- [Repository Layout](#repository-layout)
- [Runtime Model](#runtime-model)
- [Primary Commands](#primary-commands)
- [Active HTTP Surface](#active-http-surface)
- [Authentication and Access](#authentication-and-access)
- [AI Providers and Runtime](#ai-providers-and-runtime)
- [BI and Metadata-Driven Generation](#bi-and-metadata-driven-generation)
- [Domus Boundary](#domus-boundary)
- [Configuration and Runtime Home](#configuration-and-runtime-home)
- [Development Notes](#development-notes)
- [Documentation Map](#documentation-map)
- [Screenshots](#screenshots)
- [Current Focus](#current-focus)

## Overview

`GNyx` is the product-facing application layer of the Kubex ecosystem.

It combines:

- a Go gateway/backend
- an embedded React frontend (`GNyx-UI`)
- real AI provider execution through a unified runtime path
- authentication, session, invitation, and access flows
- administrative and operational surfaces
- a typed boundary into `Domus`
- shared infrastructure from `Kbx`

At this point, `GNyx` is no longer just a shell or architectural proposal. It already exposes real product behavior and demonstrable runtime value.

## Current Product State

Currently operational or materially consolidated:

- embedded frontend delivery through the backend
- real provider runtime via the active gateway path
- working AI-assisted frontend features through backend routes
- hardened provider registry consumption between `Kbx` and `GNyx`
- functional sign-in, refresh, logout, and `me` bootstrap flows
- enriched `/me` payload with `access_scope`, memberships, and effective permissions
- tenant-aware frontend bootstrap and route guarding
- `Access Management` MVP for members, invites, and controlled role reassignment
- metadata-driven BI proof of concept backed by real Sankhya catalog ingestion
- `BI Studio` frontend page with generation, schema copy, and ZIP export

Still evolving:

- broader business-flow coverage across every screen
- deeper plan and entitlement governance
- richer multi-tenant / team-level RBAC
- broader BI domain coverage beyond the first grounded slice
- further Domus-backed expansion of operational areas still using hybrid or mock behavior

## Core Capabilities

At the current stage, `GNyx` provides or partially provides:

- unified provider execution and streaming
- provider catalog, status, config, and health exposure
- chat, summarization, code, image-prompt, and prompt-crafting flows
- authentication, sessions, invitation acceptance, and access bootstrap
- tenant-aware shell behavior and RBAC-aware navigation
- provider administration and BYOK-oriented flows
- access management for members, invites, and role reassignment
- metadata synchronization orchestration for Sankhya BI catalog
- grounded BI board generation and export

## Ecosystem Context

`GNyx` operates as the orchestration and product layer across four related projects:

- `GNyx-UI`: embedded frontend experience
- `Domus`: data-service substrate, typed stores, migrations, and PostgreSQL runtime
- `Kbx`: shared infrastructure such as config, crypto, mail utilities, and provider registry
- `GETL`: ETL/sync utility now used for Sankhya catalog ingestion

`Logz` also participates in the broader ecosystem, but it is intentionally treated as a shared logging foundation rather than a `GNyx`-specific evolution target.

## Architecture Overview

`GNyx` currently sits across these layers:

1. `cmd/` and CLI entrypoints for gateway and supporting operational commands
2. `internal/runtime/` for boot, wiring, and gateway startup
3. `internal/api/routes/` for the active HTTP surface
4. `internal/features/` for product-oriented runtime features such as BI
5. `internal/services/` for domain services such as invites
6. `internal/dsclient/` for the single integration point into `Domus`
7. `frontend/` for the embedded product UI

A critical design rule remains in place:

- `GNyx` should reach Domus through the `internal/dsclient` single integration point, not through arbitrary cross-repo shortcuts.

## Repository Layout

```text
cmd/                    Cobra entrypoints
config/                 local config, provider config, GETL sync manifests
frontend/               embedded React application (GNyx-UI)
internal/api/           HTTP routes and route wiring
internal/auth/          token and auth-related runtime logic
internal/dsclient/      Domus integration boundary
internal/features/      product features such as BI generation
internal/runtime/       gateway boot and runtime composition
internal/services/      domain services such as invite flows
.notes/                 task docs, context, and implementation analysis
```

## Runtime Model

The most used local gateway flow today is:

```bash
gnyx gateway up -e ./config/.env.local -D
```

Equivalent source-driven flow from the repo:

```bash
go run ./cmd/main.go gateway up -e ./config/.env.local -D
```

`GNyx` relies on:

- runtime config from env files and runtime home
- active provider config under `~/.kubex/gnyx/config/providers.yaml`
- frontend assets built from `frontend/`
- Domus-backed data stores through `internal/dsclient`

## Primary Commands

Run the gateway:

```bash
go run ./cmd/main.go gateway up -e ./config/.env.local -D
```

Build the backend:

```bash
go build ./...
```

Build the frontend:

```bash
cd frontend
pnpm exec vite build
```

Sync Sankhya metadata catalog into PostgreSQL through the GNyx orchestration path:

```bash
go run ./cmd/main.go metadata sankhya sync \
  --env-file ./config/.env.local \
  --pg-dsn 'postgres://kubex_adm:admin123@localhost:5432/postgres?sslmode=disable'
```

Important note:

- because `GETL` currently depends on `godror`, some build environments may require `CGO=1` when compiling code paths that import `GETL` transitively.

## Active HTTP Surface

The active gateway currently exposes, among others:

- `GET /api/v1/health`
- `GET /api/v1/healthz`
- `GET /api/v1/providers`
- `GET /api/v1/config`
- `GET /api/v1/test?provider=...`
- `POST /api/v1/unified`
- `POST /api/v1/unified/stream`
- `GET /me`
- `GET /auth/me`
- `GET /api/v1/access/members`
- `PATCH /api/v1/access/members/:user_id/role`
- `GET /api/v1/bi/catalog/status`
- `POST /api/v1/bi/boards/generate`
- `POST /api/v1/bi/boards/export`

The HTTP surface is intentionally evolving from general infrastructure endpoints into feature-oriented product routes.

## Authentication and Access

Authentication is already functional and security-sensitive paths have been hardened.

Current access posture:

- login, refresh, logout, and `me` are operational
- refresh expiration and revocation checks were hardened
- invitation acceptance was hardened for retry safety and partial-failure recovery
- `/me` now returns `access_scope`, memberships, team memberships, pending access, and effective permissions
- frontend route guards now depend on real authenticated access state rather than legacy mock assumptions
- `Access Management` provides a first tenant-scoped management surface

Current RBAC MVP boundary:

- tenant-scoped, not team-scoped
- effective permissions exposed for the active tenant
- visually enforced in the frontend for high-ROI administrative areas
- includes controlled role reassignment, but not direct permission editing

## AI Providers and Runtime

Providers are no longer theoretical in `GNyx`.

They are actively consumed through the backend runtime and exposed to the frontend.

Current practical state:

- provider registry hardening happened in `Kbx`
- `GNyx` now uses a single effective provider config path during boot
- runtime catalog and config exposure are available through the active gateway
- frontend tools can explicitly select providers per feature
- `Groq` is currently the most reliable provider for the BI generation demo path
- `Gemini` is supported, but tends to be slower and more likely to fall back for the BI contract shape

## BI and Metadata-Driven Generation

One of the most important new fronts is now materially implemented.

Current flow:

1. Sankhya BI metadata CSVs are ingested into PostgreSQL under `sankhya_catalog`
2. `Domus` registers external metadata load state in `public.external_metadata_registry`
3. `GNyx` uses grounded metadata queries to prepare board-generation context
4. provider runtime generates or partially generates a board plan
5. backend validates and compiles the result into a `DashboardSchema`
6. frontend `BI Studio` exposes generation, inspection, copy, and ZIP export

Current BI slice status:

- first grounded domain: `sales`
- generation modes: `llm`, `llm_recovered`, `fallback_template`
- ZIP export contains the generated schema, board plan, grounding context, and generation metadata
- the frontend only exposes the feature as operational when the catalog is actually available

## Domus Boundary

`GNyx` does not treat Domus as an incidental dependency.

It uses Domus as the active data-service boundary for:

- user and session-related runtime data
- invitations and pending access
- companies / tenants and memberships through mixed typed-store plus SQL composition paths
- external metadata registry reads for BI metadata readiness

A significant rule still applies:

- data expansion should keep converging toward the single `internal/dsclient` boundary instead of creating parallel access paths.

## Configuration and Runtime Home

`GNyx` uses a runtime-home model under:

```text
~/.kubex/gnyx/
```

Important operational rule:

- this runtime home should be treated as the active source of truth when materialized
- missing config should be created conservatively from passed values or safe defaults
- existing config must not be destructively overwritten in repeated or parallel runs

Typical active files include:

- `~/.kubex/gnyx/config/providers.yaml`
- runtime secrets, config fragments, and generated material

Repo-local config still matters for development and bootstrap, especially:

- `config/.env.local`
- `config/providers.yaml`
- `config/getl/sankhya_catalog/*`

## Development Notes

Recommended backend validation:

```bash
go build ./...
go test ./...
```

Recommended frontend validation:

```bash
cd frontend
pnpm exec tsc --noEmit
pnpm exec vite build
```

Recommended provider/runtime smoke:

- start the gateway from source
- hit `GET /api/v1/providers`
- hit `POST /api/v1/unified`
- validate provider availability and returned health/config state

Recommended BI smoke:

- confirm `GET /api/v1/bi/catalog/status`
- generate a board through `POST /api/v1/bi/boards/generate`
- export through `POST /api/v1/bi/boards/export`
- validate the frontend `BI Studio` path

## Documentation Map

Important local documentation:

- [`frontend/README.md`](./frontend/README.md)
- [`docs/README.pt-BR.md`](./docs/README.pt-BR.md)
- [`.notes/analyzis/global-execution-plan/`](./.notes/analyzis/global-execution-plan)
- [`.notes/analyzis/gnyx-skw-dynamic-ui/`](./.notes/analyzis/gnyx-skw-dynamic-ui)
- [`.notes/context/`](./.notes/context)

## Screenshots

Placeholder suggestions:

- `[Screenshot Placeholder: Welcome and tenant-aware shell]`
- `[Screenshot Placeholder: Access Management]`
- `[Screenshot Placeholder: Providers Settings with real runtime state]`
- `[Screenshot Placeholder: BI Studio generation result]`

## Current Focus

The current baseline is strong enough to support the next major steps:

- broader frontend-to-backend consolidation on top of real access and provider runtime
- richer BI/domain generation slices beyond the first `sales` proof of concept
- broader Domus-backed data and governance evolution without abandoning current operational flows
