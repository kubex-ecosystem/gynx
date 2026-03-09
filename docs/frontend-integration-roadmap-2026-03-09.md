# Frontend Integration Roadmap (GNyx) - Snapshot 2026-03-09

## 1. Frontend Architecture Overview

### 1.1 Stack e shell

- Build/runtime: `Vite + React 19 + TypeScript + TailwindCSS`.
- Shell principal centralizada em `src/App.tsx`, com `AuthProvider`, layout, header, footer e navegação lateral por domínio.
- Navegação continua hash-based (`window.location.hash`), sem React Router.

### 1.2 Estrutura de pastas

- `src/pages`: páginas operacionais, autenticação e administração.
- `src/components/features`: módulos de produto.
- `src/services`: integração backend, fallback/offline e orchestration.
- `src/core/http`: fundação HTTP unificada.
- `src/modules`: início da modularização por domínio, já ativa para chat e creative.
- `src/core/llm`, `src/hooks`, `src/store`: wrappers de IA, hooks de integração e estado compartilhado.

### 1.3 Estado e persistência

- Estado local ainda é dominante nos módulos de feature.
- Contextos: `AuthContext`, `LanguageContext`.
- Zustand: `useProvidersStore`, `useAuthStore`.
- Persistência local: `localStorage`, cookies e IndexedDB (`enhancedAPI`, drafts, histórico).

### 1.4 Comunicação backend

Situação atual:

- `src/core/http/client.ts` é a base oficial de rede.
- `httpClient` já aparece em 12 arquivos do `src`.
- `fetch` direto fora de testes caiu para 1 chamada real, a interna do próprio `httpClient`.

Camadas em operação:

1. `core/http` com `client`, `types`, `auth`, `errors`, `endpoints`.
2. `api.ts` legado adaptado sobre `httpClient`.
3. `enhancedAPI` para offline/cache/queue, também sobre `httpClient`.
4. Serviços de domínio (`unifiedAIService`, `chatService`, `creativeService`, `agentsService`, etc.).

### 1.5 Fluxo de autenticação e onboarding

- Sessão real: `/me`, `/auth/sign-in`, `/sign-out` via `httpClient`.
- Demo mode continua disponível em `AuthContext`.
- `accept-invite` agora está corretamente incluído na allowlist pública do `App`.

### 1.6 Forças arquiteturais

- Fundação HTTP consolidada e tipada.
- Fase 2 executada: chat, summarizer, code e images agora têm wiring real no `App`.
- Catálogo de endpoints centralizado em `src/core/http/endpoints.ts`.
- Início real de modularização vertical com `src/modules/chat` e `src/modules/creative`.
- Camada offline/PWA segue robusta para cenários de fallback.

### 1.7 Riscos arquiteturais atuais

- `multiProviderService` e `enhancedAPI` continuam com acoplamento circular.
- `APIError` legado e `HttpError` coexistem.
- Ainda há namespaces mistos entre fluxos antigos (`api.ts`, `enhancedAPI`) e paths novos centralizados.
- `DataAnalyzer` segue executando código gerado por IA com `new Function(...)` no browser.
- Mocks continuam espalhados em Mail Hub, Workspace, Gateway, Sync, Invite e partes de auth/config.

---

## 2. Module Domain Map

| Domínio | Módulo | Responsabilidade | Estado de integração | Serviços principais | Capacidades backend implícitas |
| --- | --- | --- | --- | --- | --- |
| Operation & Data | Gateway | Métricas e logs operacionais | Híbrido (real/simulado) | `gatewayService` | métricas, logs, health operacional |
| Operation & Data | Mail Hub | Inbox inteligente e ações de e-mail | Mockado | mock local em `MailHub.tsx` | inbox, labels IA, resumo, ações de atendimento |
| Operation & Data | Sync | Conexões e jobs de sincronização | Híbrido (real/simulado) | `syncService` | CRUD integrações, jobs e histórico |
| Intelligence & Analysis | Playground | Streaming de geração | Integrado | `streamingService` | SSE/token stream por provider |
| Intelligence & Analysis | Analyzer | Upload CSV + análise | Híbrido com risco | `useMultiProvider` | análise assistida e execução segura no backend |
| Intelligence & Analysis | Prompt Crafter | Geração estruturada de prompts | Integrado com fallback demo | `configService`, `unifiedAIService` | geração estruturada, metadados, seleção de provider |
| Intelligence & Analysis | Agents | Geração e persistência de agents | Integrado | `agentsService`, `configService` | gerar, listar, salvar, remover e exportar AGENTS |
| Intelligence & Analysis | Chat | Conversa multi-turno | Integrado no shell principal | `chatService`, `unifiedAIService` | chat contextual por provider |
| Creative Lab | Summarizer | Resumo por tom/limite | Integrado no shell principal | `creativeService`, `unifiedAIService` | sumarização parametrizada |
| Creative Lab | Code | Geração de snippet/scaffold | Integrado no shell principal | `creativeService`, `unifiedAIService` | code generation com constraints |
| Creative Lab | Images | Geração de prompt visual | Integrado no shell principal | `creativeService`, `unifiedAIService` | prompt engineering para imagem |
| Administration | Workspace | Configurações do workspace | Simulado local | sem serviço backend dedicado | leitura/atualização de tenant, região, plano |
| Administration | Providers | BYOK e saúde de providers | Integrado parcial | `unifiedAIService`, `useProvidersStore` | teste de provider, escolha default, gestão de chaves |
| Access | Auth | Login, sessão e logout | Integrado + demo fallback | `AuthContext`, `httpClient` | sign-in, sessão, sign-out, OAuth start |
| Access | Invite | Aceite de convite | Híbrido (real/simulado) | `inviteService` | validar token e concluir onboarding |

Dependências cruzadas relevantes:

- `App.tsx` continua sendo o orquestrador global dos domínios.
- `chatService` e `creativeService` passam a ser a fachada de Fase 2 para os módulos conectados.
- `configService` e `unifiedAIService` seguem como núcleo da experiência de IA.

---

## 3. Backend Integration Gap Analysis

### 3.1 Áreas com mock ou simulação ativa

1. `src/pages/MailHub.tsx` continua 100% mockado.
2. `src/pages/WorkspaceSettings.tsx` continua com persistência simulada.
3. `src/services/gatewayService.ts` ainda usa `VITE_SIMULATE_AUTH`.
4. `src/services/syncService.ts` ainda usa `VITE_SIMULATE_AUTH`.
5. `src/services/inviteService.ts` ainda usa `VITE_SIMULATE_AUTH`.
6. `src/context/AuthContext.tsx` ainda mantém modo demo.
7. `src/services/configService.ts` ainda devolve fallback demo.
8. `src/services/geminiService.ts` permanece como caminho mockado/alternativo.

### 3.2 Itens resolvidos desde o snapshot anterior

1. Chat deixou de ser `UI-only`.
2. Summarizer deixou de ser `UI-only`.
3. Code Generator deixou de ser `UI-only`.
4. Image Generator deixou de ser `UI-only`.
5. `accept-invite` deixou de estar bloqueado pelo route guard.

### 3.3 Gaps críticos ainda abertos

1. `src/components/features/DataAnalyzer.tsx`

- Continua executando `new Function(...)`.
- Este segue sendo o maior risco técnico e de segurança do frontend.

1. `src/pages/MailHub.tsx`

- UI pronta, mas sem qualquer contrato backend real.

1. `src/pages/WorkspaceSettings.tsx`

- Painel pronto, porém sem leitura/gravação persistente.

1. `src/services/enhancedAPI.ts` e `src/services/multiProviderService.ts`

- Ainda merecem desacoplamento para reduzir bootstrap frágil.

### 3.4 Inconsistências arquiteturais restantes

1. Convivência entre `APIError` e `HttpError`.
2. Mistura entre camadas novas (`modules/*`, `core/http`) e fluxos antigos fortemente centrados em componente.
3. Padrão de mocks ainda distribuído, sem governança única por domínio.

---

## 4. ROI Priority List

## HIGH ROI

1. **Hardening do Analyzer**

- Problema: execução dinâmica de código no browser.
- Oportunidade: mover planejamento e execução para backend/sandbox.
- Impacto esperado: remove o principal risco técnico do frontend atual.

1. **Substituir mocks de Mail Hub e Workspace por contratos reais**

- Problema: duas áreas visíveis continuam sem backend.
- Oportunidade: transformar UI pronta em módulos operacionais.
- Impacto esperado: alto valor funcional com baixa ambiguidade de produto.

1. **Desacoplar `enhancedAPI` de `multiProviderService`**

- Problema: dependência circular aumenta complexidade e risco de regressão.
- Oportunidade: separar responsabilidades entre offline/cache e orchestration de provider.
- Impacto esperado: bootstrap mais previsível e menor custo de manutenção.

1. **Unificar estratégia de mocks por domínio**

- Problema: simulação espalhada em páginas, serviços e contextos.
- Oportunidade: centralizar feature flags e cenários em `src/mocks/*`.
- Impacto esperado: rollout mock->real mais previsível.

1. **Consolidar contrato único de erro**

- Problema: `APIError` legado e `HttpError` convivem.
- Oportunidade: fechar a transição para um único padrão ou adapters explícitos.
- Impacto esperado: menor custo de depuração e integração.

## MEDIUM ROI

1. Expandir `src/modules/*` para Prompt, Agents, Analyzer e Admin.
2. Reativar testes de integração API/frontend.
3. Adicionar observabilidade por request-id, provider, modo e fallback.

## LOW ROI

1. Polimento visual extra antes da consolidação backend dos módulos restantes.
2. Limpeza de telas legadas desconectadas da shell principal.
3. Melhorias avançadas de PWA sem requisito imediato de produto.

---

## 5. Implementation Strategy (Top Priorities)

### 5.1 Estado do plano

- **Fase 1: concluída** no escopo planejado.
- **Fase 2: concluída** com handlers reais no `App` e novas fachadas de domínio.
- Próxima prioridade objetiva: **Fase 3 (Analyzer seguro)**.

### 5.2 Próxima sequência recomendada

### Fase 3 - Analyzer seguro

Refatorar:

- `src/components/features/DataAnalyzer.tsx`

Criar:

- `src/modules/analyzer/services/analyzerService.ts`
- `src/modules/analyzer/types/contracts.ts`

Ações:

- Substituir `new Function(...)` por chamada backend.
- Fazer o frontend renderizar apenas resultado estruturado.

### Fase 4 - Mail Hub + Workspace

Refatorar:

- `src/pages/MailHub.tsx`
- `src/pages/WorkspaceSettings.tsx`

Criar:

- `src/modules/mail/services/mailService.ts`
- `src/modules/workspace/services/workspaceService.ts`

Ações:

- Definir contratos mínimos de leitura/escrita.
- Remover mocks inline e mover fallback para camada controlada.

### Fase 5 - Governança de mocks

Criar/refatorar:

- `src/mocks/index.ts`
- `src/mocks/scenarios.ts`
- `src/mocks/domains/*`

Ações:

- Centralizar decisão mock/real.
- Tirar lógica simulada de componentes e contextos.

### Fase 6 - Modularização contínua

Expandir a convenção já iniciada:

- `src/modules/prompt/*`
- `src/modules/agents/*`
- `src/modules/analyzer/*`
- `src/modules/admin/*`

### 5.3 Definition of Done atualizada

- Sem módulos críticos ainda em `UI-only`.
- Onboarding público funcional.
- Analyzer sem execução dinâmica no cliente.
- Mail Hub e Workspace com contratos backend reais.
- Estratégia de mock centralizada e removível por domínio.

---

## Executive Summary

O frontend entrou em um estado mais maduro entre os snapshots de 2026-03-08 e 2026-03-09. A fundação HTTP foi consolidada, a Fase 2 foi executada com sucesso e os quatro módulos criativos/chat já operam com wiring real para backend. O principal gargalo agora deixou de ser “infra de integração” e passou a ser “segurança e fechamento funcional dos módulos restantes”, com foco imediato em `DataAnalyzer`, `MailHub` e `Workspace`.

FRONTEND ANALYSIS AND ROI ROADMAP COMPLETE - READY FOR ARCHITECT REVIEW
