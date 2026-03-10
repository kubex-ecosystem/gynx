# GNyx

GNyx is the product-facing gateway and workspace application of the Kubex ecosystem. It combines a Go backend, an embedded React frontend, real AI provider integration, authentication and invitation flows, operational surfaces, and a progressive multi-tenant access model.

## Table of Contents

- [What GNyx Is](#what-gnyx-is)
- [Current Product State](#current-product-state)
- [Core Capabilities](#core-capabilities)
- [How the Ecosystem Fits Together](#how-the-ecosystem-fits-together)
- [Architecture Overview](#architecture-overview)
- [Repository Structure](#repository-structure)
- [Runtime Model](#runtime-model)
- [Active HTTP Surface](#active-http-surface)
- [Frontend Product Areas](#frontend-product-areas)
- [Authentication and Access](#authentication-and-access)
- [Providers and AI Runtime](#providers-and-ai-runtime)
- [Data and Domus Boundary](#data-and-domus-boundary)
- [Development Workflow](#development-workflow)
- [Running GNyx Locally](#running-gnyx-locally)
- [Configuration and Runtime Home](#configuration-and-runtime-home)
- [Documentation Map](#documentation-map)
- [Screenshots](#screenshots)
- [Current Focus](#current-focus)
- [Status Notes](#status-notes)

## What GNyx Is

GNyx is not just a backend API and not just a frontend workspace.

It is the runtime product layer that currently brings together:

- a Go gateway application
- an embedded web UI (`GNyx-UI`)
- real AI provider execution through a unified backend path
- authentication, session, invitation, and user-access flows
- operational and administrative surfaces
- a data-service boundary into `Domus`
- shared ecosystem capabilities supplied by `Kbx`

In practical terms, GNyx is the application users and operators interact with directly, while other ecosystem projects provide infrastructure, shared services, or runtime support around it.

## Current Product State

The repository already supports real, demonstrable product behavior.

Currently functional or materially consolidated:

- the frontend is embedded into the backend delivery flow
- real backend routes exist for provider/runtime usage
- real AI provider usage is working through the active gateway path
- `Chat`, `Prompt Crafter`, `Summarizer`, `Code`, and `Image Prompt` flows are wired through the backend
- authentication flows are operational
- invitation acceptance is operational and has already been hardened for retry safety
- refresh-session expiration and revocation checks were hardened recently
- provider runtime and registry consumption were consolidated across `Kbx` and `GNyx`

Still in progress or intentionally incomplete:

- full frontend-to-backend coverage across every screen and business flow
- deeper multi-tenant and RBAC consolidation
- plans and entitlements as a fully closed access-governance layer
- broader Domus expansion beyond the currently proven store/runtime surface
- some areas of the frontend still rely on mock-backed or hybrid behavior

## Core Capabilities

At the current stage, GNyx provides or partially provides:

- unified AI provider execution
- provider health, catalog, and runtime configuration exposure
- browser-based AI-assisted workflows
- authentication and session management
- invitation and onboarding flows
- workspace and administrative surfaces
- mail and sync-related product areas
- repository- and metrics-oriented runtime surfaces
- CLI-driven gateway operation and supporting utilities

## How the Ecosystem Fits Together

GNyx lives inside a broader ecosystem, but it is the product-facing center of the current effort.

### GNyx + GNyx UI

This repository contains:

- the Go backend and gateway runtime
- the embedded frontend application
- business-facing routes and runtime composition

The frontend lives in [frontend/README.md](./frontend/README.md) and is built into the product delivery flow.

### Domus

`Domus` is the data runtime and datastore platform behind the current data-service path.

Today, GNyx should reach Domus through the single integration point under `internal/dsclient`. That boundary is operational and intentionally treated as the safe path until a more evolved integration model is deliberately introduced.

### Kbx

`Kbx` provides shared capabilities used by GNyx and other Kubex projects, especially:

- AI provider registry and abstractions
- mail and IMAP helpers
- shared config/defaulting behavior
- reusable operational utilities

### Logz

`Logz` is the logging foundation. It is important to the runtime, but it is not considered the current target for ecosystem-specific product changes.

## Architecture Overview

From a product-engineering point of view, the repository currently has five major layers.

### 1. Gateway and runtime entrypoints

Main areas:

- [`cmd/`](./cmd)
- [`internal/runtime/`](./internal/runtime)
- [`internal/app/`](./internal/app)

These areas start and compose the active server runtime, gateway behavior, dependency wiring, and feature exposure.

### 2. HTTP and API surface

Main areas:

- [`internal/api/`](./internal/api)
- [`internal/features/`](./internal/features)
- [`internal/web/`](./internal/web)

This layer exposes the active HTTP routes, auth flows, invite flows, and the embedded UI.

### 3. Domain and service logic

Main areas:

- [`internal/domain/`](./internal/domain)
- [`internal/services/`](./internal/services)
- [`internal/auth/`](./internal/auth)

This layer holds business logic such as session/auth handling, invitation orchestration, provider-facing logic, and supporting domain services.

### 4. Data-service boundary

Main areas:

- [`internal/dsclient/`](./internal/dsclient)

This is the current safe integration point between GNyx and Domus-backed data access.

### 5. Embedded frontend application

Main areas:

- [`frontend/`](./frontend)
- [`ui/`](./ui)
- runtime-serving pieces under [`internal/web/`](./internal/web)

The frontend is the product UI. It is built and then served by the Go application.

## Repository Structure

```text
.
├── cmd/               # CLI entrypoints such as gateway and docs servers
├── config/            # Local project config files and examples
├── docs/              # Project documentation site sources
├── frontend/          # GNyx-UI frontend application
├── internal/
│   ├── api/           # HTTP routes, controllers, invite API
│   ├── app/           # Container/bootstrap/runtime composition
│   ├── auth/          # JWT, auth config, auth controllers/middleware
│   ├── domain/        # Domain contracts and models
│   ├── dsclient/      # Domus-facing single integration point
│   ├── features/      # Runtime features and adapters
│   ├── runtime/       # Active gateway runtime
│   ├── services/      # Business and supporting services
│   ├── web/           # Embedded UI serving
│   └── ...
├── support/           # Build/install/documentation scripts and hooks
├── templates/         # Email templates and other runtime templates
├── tests/             # Project-level test assets and support
├── ui/                # Additional UI-related assets/legacy surfaces
├── Makefile
└── README.md
```

## Runtime Model

GNyx currently operates as a combined backend and product shell.

Important runtime characteristics:

- the gateway is the main active runtime entrypoint
- the frontend is built and served by the backend runtime
- provider configuration is loaded from the active runtime home when available
- the runtime is expected to materialize safe defaults under `~/.kubex` when needed
- existing runtime state under `~/.kubex` should not be overwritten destructively

Operationally, the active home directory matters because GNyx is not only reading local project files. It is also designed to run against persistent runtime state outside the repository.

## Active HTTP Surface

The currently relevant backend HTTP surface includes at least the following product-facing areas.

### Auth and identity

- `POST /api/v1/auth/sign-up`
- `POST /api/v1/auth/sign-in`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/sign-out`
- `GET /api/v1/me`
- `GET /api/v1/auth/me`

### Runtime and providers

- `GET /api/v1/health`
- `GET /api/v1/healthz`
- `GET /api/v1/providers`
- `GET /api/v1/config`
- `GET /api/v1/test?provider=...`
- `POST /api/v1/unified`
- `POST /api/v1/unified/stream`

### Invitation and onboarding

Invitation routes are active and tied into the current onboarding flow, including the public acceptance path used by the frontend.

## Frontend Product Areas

The embedded frontend currently includes these major user-facing areas:

- landing and sign-in
- invite acceptance
- welcome workspace shell
- gateway dashboard
- prompt crafter
- chat
- summarizer
- code generator
- image prompt generator
- providers settings
- data sync
- mail hub
- workspace settings
- agents
- playground
- data analyzer

Not all of these are equally mature. Some are already backed by real backend/provider flows, while others are still hybrid or mock-backed.

For the full frontend breakdown, see [frontend/README.md](./frontend/README.md).

## Authentication and Access

Authentication is already operational and recently hardened further.

Current practical state:

- sign-in is functional
- refresh is functional
- `me` is functional
- session refresh expiration and revocation checks were explicitly hardened
- invitation acceptance is functional and more retry-safe than before

What is not yet fully closed at the same level:

- full access-governance consolidation across multi-tenant scope, RBAC, plans, and entitlements
- a completely explicit domain model for active access scope across all product areas

That is the next major domain focus for the repository.

## Providers and AI Runtime

GNyx is now operating with a real provider-backed execution path rather than only UI proposal or local abstraction.

Current state:

- provider runtime was hardened in `Kbx`
- GNyx now consumes a single effective provider path more coherently
- the gateway exposes real config, health, and unified execution routes
- the frontend can use real providers through backend routes
- provider selection by tool is now supported in the frontend

The provider stack is still evolving, but it has already crossed the line from “proposed” to “functionally demonstrable”.

## Data and Domus Boundary

The current rule for safe integration is straightforward:

- GNyx should access Domus through `internal/dsclient`

This boundary already backs meaningful parts of the runtime, including user/invite/company-related paths. The current integration is operational, but it is also understood to be transitional rather than the final long-term architecture.

Important nuance:

- Domus is not only a collection of stores; it is a data-runtime platform
- GNyx is not yet fully integrated with all desirable data/business flows
- deeper integration should proceed carefully and incrementally

## Development Workflow

Typical local work involves:

1. building and validating the Go application
2. building the frontend when needed
3. running the gateway with a local environment file
4. exercising the frontend and backend together

A real command sequence already used in local work has looked like this:

```sh
go fmt ./... && go vet ./... && go build -v ./... && go mod tidy && make build-dev && gnyx gateway up -e ./config/.env.local -D
```

During active backend iteration inside the repo, using the source tree directly is often the clearest option:

```sh
go run ./cmd gateway up -e ./config/.env.local -D
```

## Running GNyx Locally

### Prerequisites

Recommended local prerequisites include:

- Go `1.26.0`
- Node.js compatible with the frontend toolchain
- `pnpm`
- `jq`
- access to the companion projects when working on ecosystem-wide changes
- an initialized runtime home under `~/.kubex` or permission for GNyx to materialize it safely

### Frontend build

The frontend can be worked on directly from `frontend/`.

Typical commands:

```sh
cd frontend
pnpm install
pnpm exec tsc --noEmit
pnpm exec vite build
```

### Gateway run

From the repository root:

```sh
go run ./cmd gateway up -e ./config/.env.local -D
```

### Local health checks

Useful runtime checks include:

```sh
curl http://localhost:5000/api/v1/healthz
curl http://localhost:5000/api/v1/providers
curl http://localhost:5000/api/v1/config
```

## Configuration and Runtime Home

GNyx relies on both repository-local files and persistent runtime-home material.

### Repository-local examples

Important local files include:

- [`config/.env.local`](./config/.env.local)
- [`config/providers.yaml`](./config/providers.yaml)
- files under [`config/`](./config)

### Runtime-home material

Typical active runtime-home paths include:

- `~/.kubex/gnyx/config/config.json`
- `~/.kubex/gnyx/config/providers.yaml`
- `~/.kubex/gnyx/config/mail_config.json`
- `~/.kubex/gnyx/gnyx.crt`
- `~/.kubex/gnyx/gnyx.key`
- `~/.kubex/gnyx/secrets/*`

Expected runtime-home behavior:

- if required material does not exist, the runtime may create it from provided values or defaults
- if required material already exists, later runs should not overwrite it destructively
- the `domus/volumes` area under `~/.kubex` is operational state and should be treated differently from normal config files

## Documentation Map

Useful project documents include:

- [frontend/README.md](./frontend/README.md)
- [frontend/docs/README.pt-BR.md](./frontend/docs/README.pt-BR.md)
- [docs/](./docs)
- [.notes/analyzis/](./.notes/analyzis)
- [.notes/analyzis/global-execution-plan/README.md](./.notes/analyzis/global-execution-plan/README.md)

Important recent planning and analysis material includes:

- ecosystem and project backend analysis files in `.notes/analyzis/`
- the global execution plan under `.notes/analyzis/global-execution-plan/`

## Screenshots

This README intentionally leaves product screenshots as placeholders for now.

Suggested future additions:

- `[Placeholder]` product shell / welcome screen
- `[Placeholder]` providers settings showing live runtime state
- `[Placeholder]` chat or prompt crafter using a real provider
- `[Placeholder]` invite acceptance flow

## Current Focus

The current macro direction is already defined and tracked.

Near-term focus:

1. consolidate the access foundation
2. integrate frontend and backend by vertical slices instead of trying to close everything at once
3. expand Domus pragmatically only after the product and access model are more coherent

The active plan is documented in:

- [.notes/analyzis/global-execution-plan/README.md](./.notes/analyzis/global-execution-plan/README.md)

## Status Notes

This README describes the project as it exists now: partially consolidated, already functional in important product-facing areas, and still evolving in access governance, full backend coverage, and broader data-platform integration.

It should be read as a product-facing technical README, not as a claim that every proposed capability is already complete.
