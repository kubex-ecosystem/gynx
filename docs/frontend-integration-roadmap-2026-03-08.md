# Frontend Integration Roadmap (GNyx) - Snapshot 2026-03-08

## 1. Frontend Architecture Overview

### 1.1 Stack e Shell

- Build/runtime: `Vite + React 19 + TypeScript + TailwindCSS`.
- Shell principal em `src/App.tsx` com `AuthProvider`, layout e sidebar por domínio.
- Navegação hash-based (`window.location.hash`), sem React Router.

### 1.2 Estrutura de pastas (alto nível)

- `src/pages`: páginas operacionais e administrativas.
- `src/components/features`: módulos de produto.
- `src/services`: integração backend, fallback/offline e orquestração de providers.
- `src/core/http`: cliente HTTP unificado.
- `src/core/llm`: wrappers/providers locais.
- `src/store`: Zustand para auth contextual e providers.
- `src/hooks`: integração API, PWA, providers e produtividade.

### 1.3 Estado e persistência

- Estado local intenso nos módulos de feature (`useState`/`useEffect`).
- Contexto global: `AuthContext`, `LanguageContext`.
- Zustand: `useProvidersStore`, `useAuthStore`.
- Persistência local: `localStorage`, cookies e IndexedDB (`enhancedAPI`, drafts/autosave).

### 1.4 Comunicação backend (estado atual)

O projeto avançou para uma base unificada:

- `src/core/http/client.ts` está ativo e já é usado por serviços críticos.
- `httpClient` já aparece em 12 arquivos do `src`.
- `fetch` direto fora de testes caiu para 1 chamada real (a interna do próprio `httpClient`).

Camadas ainda coexistentes (com sobreposição funcional):

1. `httpClient` (fundação nova)
2. `api.ts` (agora encapsulando `httpClient` + `APIError` legada)
3. `enhancedAPI` (offline/caching/queue, agora também sobre `httpClient`)
4. `unifiedAIService` / `configService` / serviços de domínio

### 1.5 Fluxo de autenticação atual

- `AuthContext` usa modo real vs demo por env (`DEMO_MODE`/`VITE_DEMO_MODE`).
- Real: `/me`, `/auth/sign-in`, `/sign-out` via `httpClient`.
- Demo: usuário/sessão mock com cookies locais.

### 1.6 Forças arquiteturais

- Modularidade de produto já consolidada no frontend.
- Fundação HTTP centralizada implantada e validada por tipagem.
- Camada offline/PWA madura (`enhancedAPI`, fila offline, cache local).
- Serviços de domínio já separados (`agents`, `gateway`, `sync`, `invite`, etc.).

### 1.7 Riscos arquiteturais atuais

- Guard de rota pública inconsistente para convite: `accept-invite` existe na shell, mas não está em `publicSections` em `App.tsx`.
- Contratos/nomespaces de endpoint ainda misturados (`/v1`, `/api/v1`, paths relativos ao `DEFAULT_BASE_URL`).
- Dualidade de erros (`HttpError` e `APIError`) coexistindo.
- Circularidade entre `multiProviderService` e `enhancedAPI` ainda presente.
- Módulos de UI continuam sem pipeline de backend no App (chat/summarizer/code/images).

---

## 2. Module Domain Map

| Domínio | Módulo | Responsabilidade | Estado de integração | Serviços principais | Capacidades backend implícitas |
| --- | --- | --- | --- | --- | --- |
| Operation & Data | Gateway | Métricas e logs operacionais | Híbrido (real/simulado) | `gatewayService` | métricas do gateway, logs com paginação/filtro |
| Operation & Data | Mail Hub | Inbox inteligente e ações de e-mail | Mockado | mock local em `MailHub.tsx` | inbox, labels IA, resumo, ações de atendimento |
| Operation & Data | Sync | Conexões e jobs de sincronização | Híbrido (real/simulado) | `syncService` | CRUD integrações, jobs/cron, status e histórico |
| Intelligence & Analysis | Playground | Streaming de geração | Integrado | `streamingService` | SSE/token stream por provider |
| Intelligence & Analysis | Analyzer | Upload CSV + análise | Híbrido (com risco) | `useMultiProvider` | execução de análises em ambiente controlado |
| Intelligence & Analysis | Prompt Crafter | Prompt estruturado por provider | Integrado com fallback demo | `configService`, `unifiedAIService` | geração estruturada, metadados de uso |
| Intelligence & Analysis | Agents | Geração e persistência de agentes | Integrado | `agentsService`, `configService` | gerar/listar/salvar/remover/exportar AGENTS |
| Intelligence & Analysis | Chat | Conversa multi-turno | UI-only no shell principal | opcional (`onSend`) | sessão de chat com contexto + provider |
| Creative Lab | Summarizer | Resumos por tom/limite | UI-only no shell principal | opcional (`onSummarize`) | sumarização parametrizada |
| Creative Lab | Code | Geração de snippet/scaffold | UI-only no shell principal | opcional (`onGenerate`) | code generation com constraints |
| Creative Lab | Images | Geração de prompt visual | UI-only no shell principal | opcional (`onCraftPrompt`) | prompt engineering para imagem |
| Administration | Workspace | Configurações do workspace | Simulado local | sem serviço backend dedicado | leitura/atualização de tenant/região/plano |
| Administration | Providers | BYOK e saúde de provider | Integrado parcial | `unifiedAIService`, `useProvidersStore` | health check/teste provider, gestão de chaves |
| Access | Auth | Login/sessão/logout | Integrado + demo fallback | `AuthContext` + `httpClient` | sign-in, sessão, sign-out, OAuth start |
| Access | Invite | Aceite de convite | Híbrido (real/simulado) | `inviteService` | validar token e concluir onboarding |

Dependências cruzadas relevantes:

- `App.tsx` é o orquestrador de domínio, mas sem handlers reais para 4 módulos criativos/chat.
- `configService` e `unifiedAIService` são críticos para Prompt/Providers/Agents.
- `enhancedAPI` e `multiProviderService` continuam centrais para fallback e providers locais.

---

## 3. Backend Integration Gap Analysis

### 3.1 Áreas com mock/simulação ativa

1. `src/pages/MailHub.tsx` (mock completo de e-mails e ações).
2. `src/pages/WorkspaceSettings.tsx` (persistência simulada com `setTimeout` e `tenant_demo_123`).
3. `src/services/gatewayService.ts` (`VITE_SIMULATE_AUTH=true` => mock).
4. `src/services/syncService.ts` (`VITE_SIMULATE_AUTH=true` => mock).
5. `src/services/inviteService.ts` (`VITE_SIMULATE_AUTH=true` => mock).
6. `src/context/AuthContext.tsx` (modo demo).
7. `src/services/configService.ts` (fallback demo de configuração).
8. `src/services/geminiService.ts` (mock quando API key ausente).

### 3.2 UI pronta sem integração efetiva no fluxo principal

No `App.tsx`, os módulos abaixo são renderizados sem handlers de backend:

1. `ChatInterface` (`onSend` opcional não injetado).
2. `ContentSummarizer` (`onSummarize` opcional não injetado).
3. `CodeGenerator` (`onGenerate` opcional não injetado).
4. `ImageGenerator` (`onCraftPrompt` opcional não injetado).

### 3.3 Lógica que ainda deveria sair da camada de componente

1. `src/components/features/DataAnalyzer.tsx`

- Continua executando código gerado por IA via `new Function(...)` no browser.
- Precisa migrar execução para backend/sandbox controlada.

1. `src/components/features/PromptCrafter.tsx`

- Mantém múltiplas responsabilidades no componente (draft, histórico, share, geração, config).
- Alto custo de manutenção e teste.

### 3.4 Inconsistências de integração

1. Namespaces mistos de endpoint (`/v1`, `/api/v1`, rotas relativas) em serviços diferentes.
2. Contratos de erro duplicados (`APIError` legado e `HttpError` novo).
3. Circularidade de dependência: `multiProviderService <-> enhancedAPI`.
4. Rota de convite potencialmente bloqueada por guard: `accept-invite` fora da allowlist pública em `App.tsx`.

### 3.5 Locais que exigirão integração backend imediata

- `src/pages/MailHub.tsx`
- `src/pages/WorkspaceSettings.tsx`
- `src/components/features/ChatInterface.tsx` + wiring no `App.tsx`
- `src/components/features/ContentSummarizer.tsx` + wiring no `App.tsx`
- `src/components/features/CodeGenerator.tsx` + wiring no `App.tsx`
- `src/components/features/ImageGenerator.tsx` + wiring no `App.tsx`
- `src/components/features/DataAnalyzer.tsx` (execução segura no backend)

---

## 4. ROI Priority List

## HIGH ROI

1. **Completar a consolidação da camada HTTP (fase 1 em andamento)**

- Problema: base já existe, mas ainda há sobreposição de contratos e namespaces.
- Oportunidade: fechar `core/http` com `endpoints`, `errors`, `auth` e migração final de contratos.
- Impacto esperado: redução de regressão cross-módulo e aceleração de integração backend.

1. **Conectar chat + creative modules no App com handlers reais**

- Problema: 4 módulos seguem UI-only no fluxo principal.
- Oportunidade: ligar handlers padronizados com `unifiedAIService`/façade de domínio.
- Impacto esperado: entrega funcional imediata em áreas de alto valor visível.

1. **Hardening de segurança no Analyzer**

- Problema: execução dinâmica no cliente (`new Function`).
- Oportunidade: mover execução para backend/sandbox.
- Impacto esperado: redução de risco crítico de segurança/compliance.

1. **Corrigir consistência de auth/onboarding (invite route guard)**

- Problema: fluxo de convite pode ser interceptado por proteção de rota atual.
- Oportunidade: alinhar allowlist pública e contratos de sessão.
- Impacto esperado: onboarding previsível e menor atrito de aquisição.

1. **Unificar estratégia de mock por domínio**

- Problema: simulação distribuída em componentes e serviços.
- Oportunidade: centralizar feature flags/cenários em `src/mocks/*`.
- Impacto esperado: troca mock->real com menor custo de QA.

## MEDIUM ROI

1. Modularização por domínio (`src/modules/*`) para reduzir acoplamento horizontal.
2. Observabilidade frontend por request-id/modo (`server`, `byok`, `demo`, `offline`).
3. Reativar testes de integração API (suite comentada em `src/tests/api.test.ts`).

## LOW ROI

1. Polimento visual de telas legadas antes de consolidar backend.
2. Otimizações finas de PWA além do necessário para integração.
3. Limpeza de componentes antigos sem impacto direto de negócio imediato.

---

## 5. Implementation Strategy (Top Priorities)

### 5.1 Situação do plano anterior

- **Fase 1 (API foundation): parcialmente concluída**.
- Fundação `httpClient` está em produção no frontend.
- Migração de serviços críticos já executada (`auth`, `agents`, `gateway`, `sync`, `invite`, `config`, `unified`, `streaming`, `enhanced`, `api`).

### 5.2 Próxima sequência recomendada

### Fase 1B - Fechamento da fundação HTTP

Criar/refatorar:

- `src/core/http/endpoints.ts`
- `src/core/http/errors.ts`
- `src/core/http/auth.ts`
- `src/core/http/types.ts`
- ajustes em `src/services/api.ts` e `src/services/enhancedAPI.ts`

Ações:

- Padronizar catálogo de endpoints.
- Definir contrato único de erro (ou adapter explícito de legado).
- Remover ambiguidade de base path entre `/v1` e `/api/v1`.

### Fase 2 - Backend wiring dos módulos UI-only

Refatorar:

- `src/App.tsx`
- `src/components/features/ChatInterface.tsx`
- `src/components/features/ContentSummarizer.tsx`
- `src/components/features/CodeGenerator.tsx`
- `src/components/features/ImageGenerator.tsx`

Criar:

- `src/modules/chat/services/chatService.ts`
- `src/modules/creative/services/creativeService.ts`
- hooks de domínio para orchestration de estado assíncrono.

### Fase 3 - Analyzer seguro

Refatorar:

- `src/components/features/DataAnalyzer.tsx`

Criar:

- `src/modules/analyzer/services/analyzerService.ts`
- `src/modules/analyzer/types/contracts.ts`

Ações:

- Substituir execução dinâmica local por execução backend/sandbox.
- Retornar somente resultado estruturado para renderização.

### Fase 4 - Auth/onboarding consistency

Refatorar:

- `src/App.tsx`
- `src/context/AuthContext.tsx`
- `src/pages/AcceptInvite.tsx`

Ações:

- Incluir `accept-invite` explicitamente na allowlist pública.
- Validar fluxo convite sem sessão ativa.

### Fase 5 - Governança de mocks

Criar/refatorar:

- `src/mocks/index.ts`
- `src/mocks/scenarios.ts`
- `src/mocks/domains/*`
- mover mocks de `MailHub`, `WorkspaceSettings`, `gatewayService`, `syncService`, `inviteService`.

### 5.3 Definition of Done (atualizada)

- Nenhum módulo crítico com integração “silenciosamente simulada”.
- Rotas de onboarding funcionais sem autenticação prévia.
- Contrato de erro único para chamadas HTTP.
- Endpoints centralizados e auditáveis.
- Analyzer sem execução dinâmica de código no browser.

---

## Executive Summary

A arquitetura frontend evoluiu bem desde a última análise: a fundação HTTP central já está implantada e a maior parte das integrações críticas deixou `fetch` direto. O maior ganho de ROI agora está em fechar a padronização de contratos/endpoints, ligar os módulos ainda UI-only ao backend real e eliminar o vetor de risco do Analyzer. Com isso, o projeto entra em uma fase de integração previsível, com menor retrabalho e melhor velocidade de entrega.

FRONTEND ANALYSIS AND ROI ROADMAP COMPLETE - READY FOR ARCHITECT REVIEW
