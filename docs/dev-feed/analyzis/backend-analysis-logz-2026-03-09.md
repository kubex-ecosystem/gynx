# Logz Analysis

Date: 2026-03-09
Method: source-code-first analysis. Docs/README were not used as primary truth.

## 1. Role in the Ecosystem

`Logz` is the logging substrate used by the rest of the ecosystem.

It is not a thin wrapper. It contains:

- logger core
- formatter abstraction and registry
- writer abstraction
- staged processing manager

This makes it a platform dependency, not only a utility package.

## 1.1 Scope Position For The Current Initiative

User context makes the intended handling of `Logz` explicit:

- `Logz` should not be treated as a specific change target of the current `GNyx + Domus + Kbx` backend initiative
- it is shared by other projects beyond this ecosystem slice

So its relevance here is architectural dependency mapping, not planned intervention.

## 2. Core Logger Layer

Observed in `internal/core/logger.go`:

- `Logger`
- generic `LoggerZ[T]`
- logger setup through formatter/output/metadata/options

Assessment:

- logging semantics are rich and configurable
- callers rely on more than basic print methods

## 3. Processing Pipeline

Observed in `internal/manager/manager.go`:

- validate
- pre-hooks
- format
- post-hooks
- write

This is implemented as an explicit staged manager with control/state semantics.

Assessment:

- Logz defines a real logging pipeline
- hooks and formatting are part of its behavioral contract

## 4. Formatter and Writer Layers

Observed capabilities:

- formatter registry supporting multiple output styles
- writer abstraction over stdout, stderr and files

Assessment:

- Logz is designed to normalize output behavior across environments and projects

## 5. Architectural Strengths

- central logging behavior is formalized
- pipeline stages are explicit
- formatter and writer concerns are separated
- ecosystem projects can share operational logging conventions

## 6. Architectural Risks

## 6.1 Deep Dependency

Because `GNyx`, `Domus` and `Kbx` all use Logz deeply:

Risk:

- any incompatible change has broad blast radius

## 6.2 Behavioral Contract Risk

Projects may rely on:

- logger initialization behavior
- formatter names and outputs
- hook execution ordering
- writer semantics

Risk:

- changing internals without treating them as contract can create subtle system-wide regressions

## 7. Guidance Before Modifying Logz

1. Treat Logz as a platform library.
2. Review all upstream consumers before changing logger setup or manager semantics.
3. Preserve stage ordering unless there is a deliberate ecosystem-wide migration plan.
4. Document behavioral changes because they may alter runtime observability for every other project.
5. For the current backend scope, prefer not modifying `Logz` at all unless a blocking issue leaves no other option.

## 8. Bottom Line

`Logz` is the lowest shared operational layer in this ecosystem. It is foundational enough that changes should be handled with the same care as changes to a shared runtime library, not as a local refactor.
