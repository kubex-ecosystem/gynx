# Frontend Integration Roadmap (GNyx)

## 1. Frontend Architecture Overview

### 1.1 Stack e Shell

- Build/runtime: `Vite + React 19 + TypeScript + TailwindCSS`.
- UI shell central em `src/App.tsx`, com:
  - `AuthProvider` global (`src/context/AuthContext.tsx`)
  - layout (`src/components/layout/Layout.tsx`)
  - navegação lateral por seções
- Navegação atual é hash-based (`window.location.hash`) e **não** usa React Router.

### 1.2 Estrutura de pastas (alto nível)

- `src/pages`: páginas de operação/admin/autenticação.
- `src/components/features`: módulos funcionais principais.
- `src/services`: integração backend + fallback/offline/mock.
- `src/store`: Zustand para chaves de providers e auth contextual (RBAC).
- `src/core/llm`: wrappers e providers locais (OpenAI/Anthropic/Gemini/DeepSeek).
- `src/hooks`: hooks de API, PWA, multi-provider e utilidades de UX/estado.

### 1.3 Padrões de estado

- Estado local pesado nos componentes (muito `useState` em features).
- Contexts:
  - `AuthContext` (sessão/autenticação)
  - `LanguageContext` (i18n)
- Zustand:
  - `useProvidersStore` (BYOK + roteamento de provider)
  - `useAuthStore` (RBAC/tenant mock, pouco integrado ao fluxo real)
- Persistência local:
  - `localStorage` (histórico/configs)
  - `IndexedDB` (`storageService`, `enhancedAPI`, autosave/draft)

### 1.4 Camadas de serviço e comunicação backend

Padrão atual é híbrido, com 4 caminhos diferentes:

1. `fetch` direto em context/pages/components
2. `configService` para `/api/v1/config`
3. `unifiedAIService` para `/api/v1/unified`
4. `enhancedAPI` + `multiProviderService` (offline + fallback + providers locais)

Isso entrega flexibilidade, mas cria fragmentação de contratos, erros e autenticação.

### 1.5 Fluxo de autenticação atual

- `AuthContext` decide modo real vs demo via env (`DEMO_MODE`).
- Real:
  - valida sessão em `/api/v1/me`
  - login em `/api/v1/auth/sign-in`
  - logout em `/api/v1/sign-out`
- Demo:
  - usa cookies mock (`gnyx_access_token`, `gnyx_refresh_token`) + usuário fake.

### 1.6 Forças arquiteturais

- Base modular já existente por domínio de produto.
- Camada de offline/PWA relativamente madura (`enhancedAPI`, SW, fila offline).
- Serviços específicos por domínio já separados (`gatewayService`, `syncService`, `agentsService`, etc.).
- BYOK e multi-provider já modelados.

### 1.7 Riscos arquiteturais

- Fragmentação de integração backend (múltiplos estilos de acesso API).
- Grande volume de lógica de dados dentro de componentes.
- Uso extensivo de mocks em produção funcional.
- Módulos com UI pronta mas sem integração real.
- Código legado paralelo e parcialmente desconectado (ex.: `src/PromptCrafter.tsx`, screens antigas).

---

## 2. Module Domain Map

| Domínio | Módulo | Responsabilidade | Componentes principais | Estado usado | Serviços usados | Capacidade backend implícita |
| --- | --- | --- | --- | --- | --- | --- |
| Operation & Data | Gateway | Monitoramento operacional do gateway | `pages/GatewayDashboard.tsx` | Local (`metrics`, `logs`) | `gatewayService` | métricas em tempo real, logs paginados, health operacional |
| Operation & Data | Mail Hub | Caixa de entrada inteligente e resumo de e-mails | `pages/MailHub.tsx` | Local com mock em memória | Nenhum (mock inline) | inbox/listagem, leitura, labels IA, resumo IA, ações (reply/forward/tag) |
| Operation & Data | Sync | Integrações e cronjobs | `pages/DataSync.tsx` | Local | `syncService` | CRUD de integrações, agendamentos, status de execução |
| Intelligence & Analysis | Playground | Teste streaming de prompts | `components/features/Playground.tsx` | Local + `useProvidersStore` | `streamingService` | stream token/chunk, seleção provider efetivo, telemetria de stream |
| Intelligence & Analysis | Analyzer | Upload CSV + análise assistida + chart/table | `components/features/DataAnalyzer.tsx` | Local + `localStorage` | `useMultiProvider` (`multiProviderService`) | pipeline de análise tabular, geração de plano/código, execução segura no backend |
| Intelligence & Analysis | Prompt Crafter | Geração de prompt estruturado | `components/features/PromptCrafter.tsx` | Local + `localStorage` + IndexedDB | `configService`, `unifiedAIService` | providers ativos, modelos, geração por provider, token usage, histórico |
| Intelligence & Analysis | Agents | Geração/gestão de blueprints de agentes | `components/features/AgentsGenerator.tsx` | Local | `configService`, `agentsService` | geração automática de agentes, persistência CRUD, export markdown |
| Intelligence & Analysis | Chat | Interface de chat | `components/features/ChatInterface.tsx` | Local | Nenhum efetivo no App atual | envio/recebimento de mensagens com provider e contexto |
| Creative Lab | Summarizer | Resumo de conteúdo com tom/limite | `components/features/ContentSummarizer.tsx` | Local | Nenhum efetivo no App atual | summarization endpoint com parâmetros de estilo |
| Creative Lab | Code | Geração de código guiada por stack/constraints | `components/features/CodeGenerator.tsx` | Local | Nenhum efetivo no App atual | code generation endpoint + controles de formato |
| Creative Lab | Images | Geração de prompt para imagem | `components/features/ImageGenerator.tsx` | Local | Nenhum efetivo no App atual | endpoint de prompt crafting para imagem |
| Administration | Workspace | Configurações de workspace/tenant | `pages/WorkspaceSettings.tsx` | Local simulado | Nenhum (simulação) | leitura/atualização de tenant, plano, retenção, região |
| Administration | Providers | BYOK e roteamento de providers | `pages/ProvidersSettings.tsx` | `useProvidersStore` | `unifiedAIService.testProvider` | teste de provider, persistência segura de chaves, provider default/global |
| Access | Auth | Login + OAuth | `pages/Auth.tsx`, `context/AuthContext.tsx` | Context + cookies | `fetch` direto | sign-in, profile, sign-out, OAuth callback |
| Access | Invite | Validação/aceite de convite | `pages/AcceptInvite.tsx` | Local | `inviteService` | validação token, criação de senha, aceite de convite |

Dependências cruzadas relevantes:

- `App.tsx` centraliza orquestração de módulos por hash section.
- `configService` e `unifiedAIService` impactam Prompt/Agents/Providers.
- `multiProviderService` e `enhancedAPI` influenciam Analyzer/Playground e fallback offline.

---

## 3. Backend Integration Gap Analysis

## 3.1 Áreas com mock/temporário hoje

1. **Mail Hub 100% mockado no componente**

- `src/pages/MailHub.tsx`
- Dados hardcoded de e-mails, resumo IA e ações sem backend.

1. **Gateway com simulação condicional por env**

- `src/services/gatewayService.ts`
- `VITE_SIMULATE_AUTH=true` retorna `mockMetrics` e `mockLogs`.

1. **Sync com simulação condicional por env**

- `src/services/syncService.ts`
- `mockConnections` e `mockCronjobs` quando simulado.

1. **Invite com simulação condicional por env**

- `src/services/inviteService.ts`
- validação/aceite fake em modo simulado.

1. **Auth com modo demo**

- `src/context/AuthContext.tsx`
- sessão mockada e fallback local.

1. **Workspace Settings sem backend**

- `src/pages/WorkspaceSettings.tsx`
- save apenas com `setTimeout`; `tenant_demo_123` hardcoded.

1. **Fallback demo/offline no núcleo de IA**

- `src/services/configService.ts` (`getDemoConfig`)
- `src/services/unifiedAIService.ts` (`generateDemoPrompt`)
- `src/services/enhancedAPI.ts` (offline templates + queue)
- `src/services/geminiService.ts` (mock prompt)

## 3.2 Módulos com UI pronta, sem integração efetiva no fluxo atual

1. `ChatInterface`

- `src/components/features/ChatInterface.tsx`
- No App, `onSend` não é injetado; envio real não ocorre.

1. `CodeGenerator`

- `src/components/features/CodeGenerator.tsx`
- `onGenerate` opcional não conectado no App.

1. `ContentSummarizer`

- `src/components/features/ContentSummarizer.tsx`
- `onSummarize` opcional não conectado no App.

1. `ImageGenerator`

- `src/components/features/ImageGenerator.tsx`
- `onCraftPrompt` opcional não conectado no App.

## 3.3 Lógica de dados que deve sair da camada de componente

1. **DataAnalyzer**

- `src/components/features/DataAnalyzer.tsx`
- Constrói prompt, chama LLM, recebe código e executa com `new Function(...)` no cliente.
- Recomendado mover planejamento/execução para backend (ou sandbox dedicado).

1. **PromptCrafter**

- `src/components/features/PromptCrafter.tsx`
- mistura UI + carga de config + history + autosave + share URL + geração.
- Recomendado separar em hooks de domínio (`usePromptGeneration`, `usePromptHistory`, `usePromptDraft`).

## 3.4 Inconsistências de integração backend

1. **Padrões de endpoint divergentes**

- coexistem `/v1/*` e `/api/v1/*` em serviços distintos.
- risco de ambientes com base path único quebrarem parte dos módulos.

1. **Camadas paralelas de acesso API**

- `fetch` direto + `configService` + `unifiedAIService` + `enhancedAPI` + `api.ts`.
- alto custo de manutenção e observabilidade.

1. **Circularidade entre serviços centrais**

- `multiProviderService` importa `enhancedAPI` e `enhancedAPI` importa `multiProviderService`.
- funciona hoje, mas aumenta risco de bootstrap/estado inconsistente.

1. **Fluxo de invite público potencialmente bloqueado pelo guard**

- `App.tsx` define públicas apenas `landing` e `auth`; `accept-invite` não entra na allowlist.
- pode impedir onboarding externo por link de convite.

1. **Código legado desconectado da shell principal**

- `src/PromptCrafter.tsx`, `src/screens/*`, `src/components/settings/*` convivem com a nova shell.
- aumenta dívida e ambiguidade do fluxo oficial.

## 3.5 Locais que exigirão integração backend (lista objetiva)

- Mail Hub: `src/pages/MailHub.tsx`
- Workspace config persistente: `src/pages/WorkspaceSettings.tsx`
- Chat/Summarizer/Code/Images handlers: `src/components/features/*.tsx` (módulos citados)
- Analyzer execução segura de análise: `src/components/features/DataAnalyzer.tsx`
- Normalização de auth/session endpoints: `src/context/AuthContext.tsx`
- Normalização de API contracts/baseURL: `src/services/*` (especialmente `api.ts`, `enhancedAPI.ts`, `configService.ts`, `unifiedAIService.ts`)
- Substituição progressiva de mocks por providers reais: `gatewayService.ts`, `syncService.ts`, `inviteService.ts`

---

## 4. ROI Priority List

## HIGH ROI

1. **Camada API unificada (`core/http`)**

- Problema: integração fragmentada em múltiplas estratégias.
- Oportunidade: centralizar cliente HTTP, auth/cookies, errors, retry e base path.
- Impacto esperado: reduz refatorações futuras em quase todos os módulos; acelera integração backend.

1. **Conectar módulos criativos e chat ao backend real**

- Problema: Chat/Code/Summarizer/Images estão funcionais no UI, mas sem pipeline de dados no App.
- Oportunidade: criar handlers padrão e conectar em `App.tsx` com `unifiedAIService`/novo façade.
- Impacto esperado: libera 4 áreas de produto com baixo esforço incremental e alto valor visível.

1. **DataAnalyzer: remover execução dinâmica no cliente**

- Problema: execução de código gerado por IA via `new Function` no browser.
- Oportunidade: mover planejamento/execução para endpoint backend (ou sandbox isolado).
- Impacto esperado: melhora segurança, previsibilidade e governança; reduz risco crítico.

1. **Estratégia única de mocks/simulação por feature flag**

- Problema: simulação espalhada por envs/serviços/componentes.
- Oportunidade: concentrar em camada `mocks` com contrato por domínio.
- Impacto esperado: troca controlada mock->real sem quebrar UI; menor custo de QA.

1. **Hardening de autenticação/onboarding**

- Problema: fluxo de convite potencialmente bloqueado e endpoints de sessão inconsistentes.
- Oportunidade: alinhar política de rotas públicas + contratos auth.
- Impacto esperado: reduz risco de bloqueio de onboarding e defeitos de sessão em produção.

## MEDIUM ROI

1. **Modularização por domínio (`src/modules/*`)**

- Problema: lógica de domínio ainda muito distribuída entre `pages/components/services`.
- Oportunidade: vertical slice por domínio (ui + hooks + services + types).
- Impacto esperado: manutenção mais previsível e onboarding técnico mais rápido.

1. **Observabilidade front para integração backend**

- Problema: erros e fallbacks pouco rastreáveis entre camadas.
- Oportunidade: padronizar telemetria de request-id, status, modo (`server/byok/demo/offline`).
- Impacto esperado: reduz MTTR e acelera depuração de integração.

1. **Padronização de persistência local**

- Problema: histórico/draft/configs em vários pontos com estratégias diferentes.
- Oportunidade: unificar em hooks/repositórios de storage.
- Impacto esperado: menos bugs de estado e migração de dados mais simples.

1. **Cobertura de testes de integração frontend-backend**

- Problema: suite principal de API está praticamente comentada.
- Oportunidade: reativar testes com mocks MSW/Vitest e contratos reais.
- Impacto esperado: confiança de release e redução de regressões.

## LOW ROI

1. **Refino visual de componentes legados não utilizados**

- Problema: telas antigas coexistem e poluem manutenção.
- Oportunidade: deprecar/remover após migração completa.
- Impacto esperado: limpeza incremental; baixo impacto de negócio imediato.

1. **Otimizações de micro-UX antes da consolidação backend**

- Problema: investir em polimento antes de estabilizar integrações causa retrabalho.
- Oportunidade: postergar para fase pós-consolidação.
- Impacto esperado: ganho marginal no curto prazo.

1. **Ajustes avançados de PWA sem necessidade operacional imediata**

- Problema: já existe base funcional de offline e instalação.
- Oportunidade: manter como backlog até fechar integrações centrais.
- Impacto esperado: baixo frente aos itens críticos de integração.

---

## 5. Implementation Strategy (Top Priorities)

## 5.1 Target architecture pattern

- Padrão recomendado: **Domain-driven frontend com API façade central**.
- Estrutura alvo:
  - `src/core/http/*` (cliente HTTP, auth, errors, retries, interceptors)
  - `src/modules/<domain>/{ui,hooks,services,types,mappers}`
  - `src/mocks/<domain>` (cenários e adapters de simulação)
- Estado:
  - server-state via hooks de domínio
  - UI-state local nos componentes
  - sessão/provider settings em stores dedicadas

## 5.2 High ROI Execution Plan

### Fase 1 - API foundation (obrigatória)

***Criar/refatorar***

- `src/core/http/client.ts`
- `src/core/http/endpoints.ts`
- `src/core/http/errors.ts`
- `src/core/http/auth.ts`
- `src/core/http/types.ts`

***Ações***

- Definir base path único e política de credenciais.
- Normalizar resposta/erro com envelope comum.
- Criar helpers `get/post/put/delete` tipados.

***Resultado esperado***

- Todos os serviços passam a depender da mesma infraestrutura de rede.

### Fase 2 - Conectar módulos que já têm UI pronta

***Refatorar***

- `src/components/features/ChatInterface.tsx`
- `src/components/features/CodeGenerator.tsx`
- `src/components/features/ContentSummarizer.tsx`
- `src/components/features/ImageGenerator.tsx`
- `src/App.tsx`

***Criar***

- `src/modules/creative/services/creativeService.ts`
- `src/modules/chat/services/chatService.ts`
- `src/modules/creative/hooks/useCreativeGeneration.ts`
- `src/modules/chat/hooks/useChat.ts`

***Ações***

- Injetar handlers reais em `App.tsx`.
- Unificar modo `server/byok/demo` no retorno dessas chamadas.

***Resultado esperado***

- 4 módulos saem de estado “UI-only” para operação real.

### Fase 3 - Hardening de Analyzer e segurança de execução

***Refatorar***

- `src/components/features/DataAnalyzer.tsx`

***Criar***

- `src/modules/analyzer/services/analyzerService.ts`
- `src/modules/analyzer/types/contracts.ts`

***Ações***

- Substituir `new Function` por chamada backend para execução controlada.
- Frontend passa a exibir apenas resultados estruturados.

***Resultado esperado***

- elimina vetor crítico de execução dinâmica no browser.

### Fase 4 - Mock strategy unificada

***Criar/refatorar***

- `src/mocks/index.ts`
- `src/mocks/scenarios.ts`
- `src/mocks/domains/{gateway,sync,invite,mail,workspace}.ts`
- refactor de `gatewayService.ts`, `syncService.ts`, `inviteService.ts`, `MailHub.tsx`, `WorkspaceSettings.tsx`

***Ações***

- Centralizar decisão mock/real por feature flag.
- Remover mocks inline de componentes.

***Resultado esperado***

- rollout mock->backend previsível por domínio.

### Fase 5 - Auth/onboarding consistency

***Refatorar***

- `src/App.tsx`
- `src/context/AuthContext.tsx`
- `src/pages/AcceptInvite.tsx`

***Ações***

- Tornar `accept-invite` rota pública explícita.
- Revisar contratos de sign-in/sign-out/me.
- Padronizar sessão e refresh behavior.

***Resultado esperado***

- onboarding sem bloqueio e sessão consistente entre ambientes.

## 5.3 Sequenciamento recomendado

1. Fase 1 (API foundation)
2. Fase 5 (auth/onboarding)
3. Fase 2 (módulos UI-only)
4. Fase 3 (analyzer hardening)
5. Fase 4 (mock unification)

## 5.4 Definition of done por item HIGH ROI

- Contrato de API tipado e documentado.
- Sem endpoint duplicado (`/v1` vs `/api/v1`) dentro do mesmo fluxo.
- Módulos críticos sem fallback silencioso não rastreado.
- Mocks ativáveis por flag central e removíveis por domínio.
- Fluxos auth/invite validados com smoke tests E2E.

---

## Executive Summary

O frontend já possui base modular robusta e pronta para consolidação, mas a maior oportunidade de ROI está em **unificar a integração backend** e **transformar módulos já prontos de UI em funcionalidades realmente conectadas**. O plano acima prioriza redução de retrabalho futuro, aumento de previsibilidade de integração e aceleração de entrega por domínio.
