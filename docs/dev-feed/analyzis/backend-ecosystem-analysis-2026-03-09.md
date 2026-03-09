# Backend Ecosystem Analysis

Date: 2026-03-09
Scope: `GNyx`, `Domus`, `Kbx`, `Logz`
Method: code-first analysis based on source structure and runtime entrypoints. Existing docs and READMEs were intentionally not treated as source of truth.

## 1. Executive Summary

The backend ecosystem is not a set of isolated services. It is a layered composition with clear directional dependencies:

`GNyx -> Domus -> Kbx -> Logz`

and also:

`GNyx -> Kbx -> Logz`

`GNyx -> Logz`

`Domus -> Kbx -> Logz`

`Domus -> Logz`

From the code that is actually in use today, the roles are:

- `GNyx`: application runtime, HTTP API, auth, UI embedding, provider-facing product logic, orchestration.
- `Domus`: data service runtime, connection and store factory, typed datastore facade, backend/provider orchestration for infrastructure-aware data access.
- `Kbx`: ecosystem utility layer, especially LLM provider registry, mail/IMAP helpers, config/env/default loaders, shared abstractions.
- `Logz`: logging substrate used by the other three projects.

The most important architectural rule that is already reflected in the code is:

- `GNyx` should access `Domus` through the SIP in `internal/dsclient`, not by coupling directly to `Domus` internals.

That rule is mostly respected.

## 2. Real Dependency Graph

## 2.1 Module-Level Dependencies

Observed from `go.mod` and imports:

- `gnyx/go.mod`
  - requires `domus`
  - requires `kbx`
  - requires `logz`
  - uses local replace for `domus` and `kbx`
- `domus/go.mod`
  - requires `kbx`
  - requires `logz`
- `kbx/go.mod`
  - requires `logz`
- `logz/go.mod`
  - no Kubex internal dependency below it

This makes `Logz` the lowest-level shared substrate and `GNyx` the highest-level product runtime.

## 2.2 Runtime Dependency Intent

From actual code paths:

- `GNyx` does not own LLM provider implementation. It delegates to `kbx/tools/providers`.
- `GNyx` does not own generic mail/IMAP primitives. It delegates to `kbx/mailing` and related helpers.
- `GNyx` does not own base logging mechanics. It delegates to `logz`.
- `GNyx` does not directly instantiate `Domus` internals in feature code. It wraps `domus/client` behind `internal/dsclient`.
- `Domus` is both a data facade and an infrastructure-aware runtime. It is not only a passive models package.

## 3. High-Level Architecture by Responsibility

## 3.1 GNyx

Primary concerns found in code:

- application bootstrap
- dependency container
- HTTP wiring through Gin
- auth and route registration
- invite flow
- email integration surface
- UI embedding into the Go binary
- provider registry loading
- service composition over datastore and external integrations

Core flow observed:

`cmd/main.go -> internal/module -> internal/app/server.go -> internal/app/container.go -> internal/runtime/wire/http_wire.go -> internal/api/routes`

## 3.2 Domus

Primary concerns found in code:

- root database configuration lifecycle
- connection bootstrap and health
- typed and generic store factory
- adapter factory for store/repository/ORM combinations
- provider abstraction for infrastructure startup and migrations

Domus is not just "models". It is a backend platform layer for data runtime.

## 3.3 Kbx

Primary concerns found in code:

- provider registry for LLMs
- provider abstractions and model/chat contracts
- mailer and IMAP access
- config/env/default helpers
- generic utility packages reused by product and infrastructure layers

Kbx behaves like an ecosystem commons package that holds cross-product abstractions.

## 3.4 Logz

Primary concerns found in code:

- logger core
- formatter registry
- writer abstraction
- staged log processing manager

Logz is a true shared foundation, not a thin wrapper around `log`.

## 4. Cross-Project Integration Map

## 4.1 GNyx <-> Domus

This is the most sensitive integration boundary.

Observed code facts:

- `GNyx` imports `github.com/kubex-ecosystem/domus/client` only inside `internal/dsclient`.
- `internal/dsclient/client.go` exposes aliases and wrapper constructors around `domus/client`.
- `internal/dsclient/datastore/datastore.go` centralizes singleton-style init and typed store access.
- `internal/app/container.go` creates the DS client and uses it during bootstrap.
- feature/runtime code consumes the DS abstraction from inside `GNyx`, not raw `Domus` internals.

Assessment:

- the intended SIP boundary exists and is meaningful
- the rule is valuable and should be preserved
- future backend changes should prefer evolving `internal/dsclient` rather than importing new `Domus` internals across `GNyx`

## 4.2 GNyx <-> Kbx

Observed code facts:

- provider registry in `GNyx` is a thin alias over `kbx/tools/providers`
- mail-related logic depends on `kbx` mail abstractions
- multiple runtime/config helpers come from `kbx/get`, `kbx/is`, and other utility packages

Assessment:

- `GNyx` is operationally dependent on `Kbx`, especially for AI/provider and utility behavior
- provider behavior changes in `Kbx` can directly affect `GNyx` runtime semantics

## 4.3 GNyx <-> Logz

Observed code facts:

- `logz` is imported pervasively across runtime/bootstrap paths
- `cmd/main.go`, `server.go`, container bootstrap and route layers use `logz` directly

Assessment:

- logging is not an optional cross-cutting concern here; it is deeply structural
- major logging contract changes would ripple through all projects

## 4.4 Domus <-> Kbx

Observed code facts:

- `Domus` uses internal `module/kbx` types for config and root configuration lifecycle
- provider and engine layers rely on `kbx`-shaped config structs and helpers

Assessment:

- `Domus` is not independent from `Kbx`
- `Kbx` acts partly as shared config vocabulary for `Domus`

## 4.5 Domus <-> Logz / Kbx <-> Logz

Observed code facts:

- both `Domus` and `Kbx` use `Logz` as foundational logging substrate

Assessment:

- `Logz` is a strict foundation library in this ecosystem

## 5. Architectural Strengths

- There is a visible attempt to preserve boundaries through wrapper layers.
- `GNyx` has a central container/bootstrap path instead of ad hoc wiring everywhere.
- `Domus` exposes a public client surface rather than forcing consumers into internal packages.
- `Kbx` centralizes abstractions that otherwise would likely be duplicated.
- `Logz` is structured enough to support consistent behavior across projects.
- UI embedding is treated as a first-class feature of `GNyx`, not a deployment afterthought.

## 6. Architectural Tensions and Risks

## 6.1 GNyx Has Hybrid Data Access Semantics

In `internal/app/container.go`, `GNyx` bootstraps both:

- the `Domus` DS client
- a local GORM path and adapter factory fallback

This means the data layer is not purely one thing. It is a hybrid:

- public `Domus` client/store usage
- local ORM fallback path
- generic controller construction on top

Risk:

- future changes can accidentally split behavior between DS-backed and ORM-backed flows
- debugging data issues may become harder because there is more than one effective access model

## 6.2 Domus Has Broad Scope

Domus currently spans:

- config loading
- connection lifecycle
- store factory
- typed repositories/stores
- backend/provider orchestration
- migration-capable provider abstractions

Risk:

- Domus can become both the domain data service and the infrastructure orchestration layer at once
- changes in infrastructure behavior may affect consumers expecting a narrower datastore contract

## 6.3 Kbx Is Powerful but Sensitive

The LLM registry in `kbx/tools/providers/registry.go` contains:

- config loading
- env fallback logic
- provider construction
- model info handling
- partial provider support with explicit TODO paths

Risk:

- behavior can drift by provider
- adding or changing providers is likely to have ecosystem-level impact
- `GNyx` inherits those inconsistencies because it delegates rather than wraps deeply

## 6.4 Logging Is Deeply Coupled

`Logz` is used at startup, runtime, and infrastructure layers.

Risk:

- any incompatible change to logger initialization, formatter expectations, or manager pipeline semantics would have wide blast radius

## 6.5 Integration Rule Must Stay Explicit

The `internal/dsclient` boundary in `GNyx` is strategically important.

Risk:

- if future feature work bypasses it and imports `Domus` internals directly, the boundary collapses and backend evolution becomes much more expensive

## 7. Recommended Rules Before Backend Changes

These should be treated as operating constraints for the next phase.

1. Preserve `GNyx -> internal/dsclient -> domus/client` as the only sanctioned data-service path from `GNyx`.
2. Avoid introducing new direct imports from `GNyx` into `Domus` internals.
3. Treat `Kbx` provider behavior as shared ecosystem infrastructure, not as an app-local implementation detail.
4. Treat `Logz` changes as platform changes with cross-repo impact.
5. When modifying bootstrap flows, review `cmd`, `module`, `container`, and `runtime/wire` together. They form one startup chain.
6. When modifying data behavior, review both DS client usage and GORM fallback paths together.

## 8. Priority Questions for the Next Backend Analysis Iteration

These are the next questions worth answering before code changes:

1. Which GNyx features still depend on fallback ORM behavior instead of pure DS store behavior?
2. How complete is the Domus store surface relative to what GNyx product features already need?
3. Where exactly does invite/auth/session state live across `GNyx` and `Domus`?
4. Which provider features in `Kbx` are production-ready versus partially stubbed?
5. Which Logz behaviors are relied on as API contract versus just convenience?

## 9. Bottom Line

The ecosystem already has a meaningful architecture. It is not chaotic, but it is sensitive.

The safest reading is:

- `GNyx` is the product runtime
- `Domus` is the data runtime and store abstraction platform
- `Kbx` is the shared integration/utilities/platform library
- `Logz` is the logging substrate

The main point of care before touching backend code is not lack of structure. It is the number of deep, real dependencies already composed into runtime bootstrap.
