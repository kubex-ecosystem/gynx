# GNyx Backend Analysis

Date: 2026-03-09
Method: source-code-first analysis. Docs/README were not used as primary truth.

## 1. Role in the Ecosystem

`GNyx` is the application runtime and product host. It owns:

- process bootstrap
- HTTP server lifecycle
- auth and API exposure
- invite and email entrypoints
- UI embedding into the final Go binary
- provider registry loading for AI-related capabilities
- orchestration of datastore access through its SIP

It is the top-level product runtime among the four projects analyzed.

## 2. Startup Chain

Observed startup path:

`cmd/main.go -> internal/module -> internal/app/server.go -> internal/app/container.go -> internal/runtime/wire/http_wire.go -> internal/api/routes`

This matters because backend modifications that look local often are not local. Most meaningful changes affect this chain.

## 2.1 Real Local Entrypoint In Current Use

The code-level startup path is reinforced by the real local command currently used in practice:

`go fmt ./... && go vet ./... && go build -v ./... && go mod tidy && make build-dev && gnyx gateway up -e ./config/.env.local -D`

Operational conclusions:

- the actively used runtime entrypoint is `gnyx gateway up`
- frontend embedding is part of the real startup path because `make build-dev` runs the pre-build script that builds the frontend and moves assets into `internal/features/ui/web`
- local execution depends on externalized runtime material under `~/.kubex/gnyx`, including config, JWT certs, secrets, mail config and provider config

This operational path should be treated as more important than alternative server stacks that exist in the repository.

## 3. Bootstrap Structure

## 3.1 `cmd/main.go`

`cmd/main.go` is intentionally thin. It delegates execution to the module registry command and fatal-logs on failure.

Implication:

- the runtime entrypoint is centralized
- CLI/module mechanics are part of the architecture, not incidental

## 3.2 `internal/app/server.go`

`server.go` is the first real orchestration layer.

Observed responsibilities:

- create `Container`
- bootstrap dependencies
- load provider registry from config
- initialize production middleware
- register providers with middleware
- create HTTP wire
- run Gin engine
- handle graceful shutdown

Important reading:

- `server.go` is where runtime composition becomes concrete
- changes to providers, middleware or route exposure should be reviewed here first

Operational nuance:

- although `internal/app/server.go` is architecturally important, the real example provided by the user indicates that the actively exercised server path in local usage is centered on `gnyx gateway up`
- this lowers the priority of alternate server stacks for immediate intervention unless they are proven to be part of the real boot path

## 3.3 `internal/app/container.go`

This is the core dependency assembly point.

Observed responsibilities:

- initialize datastore bootstrap path
- create DS client
- initialize GORM fallback
- create adapter factory
- initialize mailer and IMAP service
- initialize invite service
- initialize UI service
- create stores map
- create generic CRUD-style controllers

This file is one of the highest-value files for future backend work.

## 4. Real Data Access Boundary

## 4.1 Intended Rule

The user-defined architectural rule matches the code:

- access from `GNyx` to `Domus` should go through `internal/dsclient`

## 4.2 Actual Implementation

Observed facts:

- direct `Domus` imports inside `GNyx` are effectively concentrated in `internal/dsclient`
- `internal/dsclient/client.go` exports aliases and wrapper constructors over `domus/client`
- `internal/dsclient/datastore/datastore.go` centralizes DS initialization and typed store retrieval
- feature code and container bootstrap consume the SIP instead of spreading Domus-specific logic

Assessment:

- this is a valid anti-corruption layer
- it should remain the only sanctioned cross-project data boundary from GNyx

## 4.3 Why This Matters

Without this SIP, `GNyx` would start depending on:

- Domus typed stores directly
- Domus config/runtime semantics directly
- Domus internal evolution directly

That would significantly raise coupling and migration cost.

## 5. HTTP and Route Wiring

## 5.1 `internal/runtime/wire/http_wire.go`

This file is the real HTTP runtime entrypoint.

Observed responsibilities:

- set Gin mode
- create Gin engine
- load provider registry
- initialize production middleware
- register provider names in middleware
- install recovery/logger/CORS/security behavior
- register route groups under `/api/v1`
- handle auth callback redirects
- expose permissive fallback CORS when security is disabled

Assessment:

- route exposure is not only in the API package; wiring and middleware policy are defined above it
- any security or API-surface change should review this file together with route registration

Operational correction:

- based on the real command currently used, the highest-confidence active runtime path is still the gateway path, not every alternate HTTP stack present in code
- this file remains architecturally relevant, but changes should be validated first against the actually used `gateway up` path

## 5.2 `internal/api/routes/router.go`

Observed route families:

- auth
- user
- contact
- email
- invite

Observed pattern:

- route registration is container-driven
- invite routes are partially public and partially protected
- email routes only register when IMAP service exists

Assessment:

- route exposure is capability-sensitive
- service availability and config state alter the API surface at runtime

## 6. Embedded UI

`internal/features/ui/ui_routes/router.go` confirms that the UI is embedded and served by `GNyx` itself.

Observed behavior:

- use embedded FS from `UIService`
- serve static assets
- fallback to SPA `index.html`
- mount UI under `/app`

Implication:

- `GNyx` is not only an API backend
- frontend deployment behavior is directly part of backend runtime semantics

## 7. Provider Integration

`internal/features/providers/registry/registry.go` is only a thin alias over `kbx/tools/providers`.

Meaning:

- provider implementation does not live in GNyx
- GNyx's AI/provider capability is delegated almost entirely to Kbx

Implication:

- provider problems may look like GNyx problems while actually originating in Kbx

Real startup evidence adds one more important point:

- there are signs of configuration divergence in the provider layer during actual startup
- the real log shows providers being loaded successfully from `~/.kubex/gnyx/config/providers.yaml`, then later another provider configuration is loaded from the repository path `config/providers_config.yaml`, which yields an empty registry

That makes provider configuration precedence and duplication a real operational hotspot, not only a theoretical code smell.

## 8. Architectural Strengths

- bootstrap is centralized
- DS boundary exists and is real
- route registration uses container capabilities instead of global state
- UI embedding is integrated cleanly into runtime
- providers are delegated instead of reimplemented

## 9. Architectural Risks

## 9.1 Hybrid Persistence Model

`container.go` combines:

- DS client bootstrap
- GORM initialization
- generic adapters/controllers

Risk:

- the effective persistence model is not singular
- future behavior can diverge between DS-backed and ORM-backed paths

## 9.2 Overloaded Container

`container.go` is doing a lot:

- infra bootstrap
- datastore composition
- service composition
- UI setup
- controller creation

Risk:

- this file is likely to become the bottleneck for backend evolution

## 9.3 Runtime-Sensitive API Surface

Because services are conditionally registered, API surface depends on runtime readiness.

Risk:

- environments can differ subtly in behavior
- debugging missing routes may require checking bootstrap state, not only route code

## 9.4 Config Path Divergence In Real Usage

The real local startup example shows multiple config origins involved at boot:

- `./config/.env.local`
- `~/.kubex/gnyx/config/providers.yaml`
- repository-local `config/providers_config.yaml`
- `~/.kubex/gnyx/secrets`
- `~/.kubex/gnyx/*.crt|*.key`

Risk:

- configuration precedence may be inconsistent across subsystems
- a component may appear healthy because one config source is valid while another later source overrides or empties state
- debugging provider issues may require tracing both repo-local and home-directory runtime material

## 10. Direct Guidance Before Modifying GNyx Backend

1. Treat `internal/app/container.go` as the primary change-control point.
2. Preserve `internal/dsclient` as the only Domus-facing SIP.
3. When changing auth, invites, or email, inspect route registration and bootstrap together.
4. When changing providers, inspect Kbx rather than assuming logic lives in GNyx.
5. When changing deployment/runtime assumptions, inspect embedded UI routing too.
6. Treat `gnyx gateway up -e ./config/.env.local -D` as the reference local runtime until a different operational path is explicitly adopted.
7. Audit provider config precedence before changing provider-related behavior.

## 11. Bottom Line

`GNyx` is a composed application host with real boundaries already in place. The key risk is not absence of architecture. The key risk is changing one of its deep bootstrap boundaries without reviewing the entire startup chain and its delegated dependencies.
