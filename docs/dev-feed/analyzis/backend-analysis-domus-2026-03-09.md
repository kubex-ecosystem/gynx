# Domus Analysis

Date: 2026-03-09
Method: source-code-first analysis. Docs/README were not used as primary truth.

## 1. Role in the Ecosystem

`Domus` is the data runtime and datastore abstraction platform for the ecosystem.

It is broader than a models package and broader than a plain database client.

Observed responsibilities:

- root database config loading/bootstrap
- database connection lifecycle
- datastore/store factory
- typed stores for product entities
- adapter factory for store/repository/ORM combinations
- provider abstraction for service startup, health and migrations

## 2. Public Surface Used by Consumers

The main public integration surface is:

- `domus/client/client.go`
- `domus/client/stores.go`

This is important because consumers such as `GNyx` should depend on this public client layer, not on Domus internals.

## 3. `client/client.go`

The public client exposes a substantial contract:

- `Init`
- `GetConn`
- `Config`
- `ConfigPath`
- `GetReference`
- `Close`
- `GetDriver`
- `GetStore`
- typed store getters
- adapter factory builders

Assessment:

- Domus exposes both generic and typed access paths
- it is intended to be the stable consumer-facing runtime API

## 4. Store Factory and Typed Stores

`client/stores.go` and `internal/datastore/factory.go` show the real shape of the datastore abstraction.

Observed capabilities:

- generic store resolution by driver/store name
- typed helpers for user, company, invite, pending access, integration
- executor-based constructors for specific store types
- adapter configuration helpers for repository/store/ORM combinations

Assessment:

- Domus is built to support both generic dynamic store resolution and typed product-aware store access

## 5. Connection and Runtime Lifecycle

## 5.1 `internal/engine/db_manager.go`

This file is central.

Observed responsibilities:

- load or bootstrap root config
- validate enabled DB configs
- construct driver by registered type
- connect with timeout
- keep active connection map
- select default connection
- health check all connections
- secure connection retrieval with reconnect behavior

Assessment:

- `DatabaseManager` is the operational core of Domus runtime
- it owns "live" database state, not just static config

## 5.2 `internal/engine/engine.go`

This file wraps the manager in a higher-level runtime facade.

Observed contract:

- bootstrap runtime
- expose config
- expose manager
- health check
- secure connection retrieval
- graceful shutdown

Assessment:

- Domus is trying to present a controlled runtime abstraction above the raw manager

## 6. Provider and Infrastructure Layer

`internal/provider/provider.go` shows that Domus abstracts infrastructure providers, not just DB access.

Observed interfaces:

- `Provider`
- `MigratableProvider`
- `RootConfigProvider`

These include:

- start
- health
- stop
- migration preparation
- migration execution
- root-config based startup

This is a meaningful architectural decision.

Domus is not only a datastore library. It also wants to orchestrate backing services.

## 7. Dockerstack Backend

`internal/backends/dockerstack/adapter.go` shows a concrete backend implementation.

Observed behavior:

- maps provider start specs into legacy DB config
- initializes Docker-backed services
- extracts endpoints
- supports migration-related capabilities
- covers PostgreSQL, MongoDB, Redis and RabbitMQ concepts

Assessment:

- Domus is carrying infra-local startup behavior for development or managed runtime scenarios
- this increases power, but also expands responsibility

## 8. Architectural Strengths

- public client surface exists
- connection lifecycle is centralized
- store factory model is explicit
- typed stores and generic stores coexist coherently
- provider abstractions are formalized instead of hidden in scripts

## 9. Architectural Risks

## 9.1 Scope Breadth

Domus spans:

- config management
- connection runtime
- store creation
- repository/ORM adapter creation
- provider orchestration
- migrations

Risk:

- the package boundary can become too broad
- infrastructure concerns and data abstraction concerns may evolve at different speeds

## 9.2 Legacy-to-New Shape Tension

In the Dockerstack adapter, there are explicit conversion bridges between new provider specs and legacy config/runtime shapes.

Risk:

- backend evolution may be slowed by translation layers
- semantics can drift if the legacy and new abstractions stop aligning

## 9.3 Consumer Assumption Risk

Consumers may think of Domus as "just stores" while the codebase behaves like a runtime platform.

Risk:

- backend changes that seem internal may affect consumer expectations unexpectedly

## 10. Guidance Before Modifying Domus

1. Distinguish carefully between public client API and internal runtime implementation.
2. Treat `client/*` as consumer contract.
3. Treat `internal/engine/*` as runtime core.
4. Treat `internal/provider/*` and `internal/backends/*` as infrastructure orchestration layer.
5. Any change to store creation should be validated against both typed helpers and generic factory paths.

## 11. Bottom Line

`Domus` is the ecosystem's data runtime platform. It should be analyzed and changed as a platform component, not as a simple persistence package.
