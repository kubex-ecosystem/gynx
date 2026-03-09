# Frontend Integration Roadmap (GNyx) - Snapshot 2026-03-09 Post-Shell

## 1. Executive Summary

Estado atual do frontend, após as últimas iterações:

- A fundação HTTP permanece consolidada em `src/core/http/*`.
- Chat, Summarizer, Code e Images seguem integrados via `unifiedAIService`.
- Mocks foram centralizados em `src/mocks/*` por domínio.
- `MailHub` e `Workspace` deixaram de concentrar regra de estado na página e agora possuem módulos próprios.
- A navegação hash-based foi endurecida com util central de parse/build/guard em `src/core/navigation/hashRoutes.ts`.
- `accept-invite` passou a funcionar como rota pública especial, inclusive com suporte correto a `token` em hash query.

Leitura objetiva:

- A base frontend ficou mais previsível.
- O acoplamento mais problemático agora está menos na shell e mais nos legados de integração e nos mocks ainda ativos.
- O principal risco técnico restante continua sendo o `DataAnalyzer`, mas ele está explicitamente adiado nesta trilha por decisão de escopo.

---

## 2. Updated Architecture Snapshot

### 2.1 Shell e navegação

- Shell principal continua em `src/App.tsx`.
- Navegação segue hash-based, sem React Router.
- A diferença agora é que o roteamento deixou de ficar implícito e disperso:
  - `src/core/navigation/hashRoutes.ts` centraliza:
    - seções válidas
    - parsing do hash
    - construção do hash
    - decisão de rota standalone
    - guard de acesso público/privado
    - navegação utilitária

### 2.2 Estrutura por camadas

- `src/core/http/*`: cliente HTTP, auth, endpoints, errors, types.
- `src/core/navigation/*`: governança da navegação hash.
- `src/services/*`: integração, orchestration e compatibilidade legada.
- `src/mocks/*`: cenários simulados por domínio.
- `src/modules/*`: modularização vertical em expansão.

### 2.3 Modularização por domínio já ativa

Domínios já modularizados de forma visível:

- `src/modules/chat/*`
- `src/modules/creative/*`
- `src/modules/mail/*`
- `src/modules/workspace/*`

Isso já reduz o peso de `App.tsx` e das páginas como lugar de regra de negócio.

---

## 3. Domain Map (Current State)

| Domínio | Estado atual | Backend real | Maturidade frontend |
| --- | --- | --- | --- |
| Auth | Integrado com demo fallback | Parcial | Boa |
| Invite | Híbrido real/mock | Parcial | Boa |
| Chat | Integrado | Sim | Boa |
| Summarizer | Integrado | Sim | Boa |
| Code | Integrado | Sim | Boa |
| Images | Integrado | Sim | Boa |
| Prompt | Integrado com fallback | Parcial | Boa |
| Agents | Integrado | Parcial | Boa |
| Gateway | Híbrido real/mock | Parcial | Média |
| Sync | Híbrido real/mock | Parcial | Média |
| Mail Hub | Modularizado, mock-backed | Não | Média |
| Workspace | Modularizado, mock-backed | Não | Média |
| Providers | Integrado parcial | Parcial | Média |
| Analyzer | Híbrido com risco técnico | Parcial | Frágil |

Interpretação:

- `MailHub` e `Workspace` evoluíram de "páginas mockadas" para "domínios frontend organizados, ainda sem backend".
- `Gateway` e `Sync` continuam híbridos, mas agora se beneficiam de mocks centralizados.
- O shell de navegação deixou de ser um ponto de fragilidade prioritário.

---

## 4. What Changed Since The Previous Snapshot

### 4.1 Concluído

1. Centralização de mocks por domínio

- `src/mocks/scenarios.ts`
- `src/mocks/domains/gateway.ts`
- `src/mocks/domains/sync.ts`
- `src/mocks/domains/invite.ts`
- `src/mocks/domains/mail.ts`
- `src/mocks/domains/workspace.ts`

1. Mail Hub modularizado

- `src/modules/mail/types.ts`
- `src/modules/mail/services/mailService.ts`
- `src/modules/mail/hooks/useMailHub.ts`

Ganhos:

- busca real no dataset mockado
- estados explícitos de `loading/error/empty`
- seleção/estrela/leitura fora da página

1. Workspace modularizado

- `src/modules/workspace/types.ts`
- `src/modules/workspace/services/workspaceService.ts`
- `src/modules/workspace/hooks/useWorkspaceSettings.ts`

Ganhos:

- leitura/salvamento/reset centralizados
- estado de feedback desacoplado da página
- persistência mock tratada como serviço

1. Hardening da navegação e guards

- `src/core/navigation/hashRoutes.ts`
- `src/App.tsx`
- `src/components/layout/Sidebar.tsx`
- `src/context/AuthContext.tsx`
- `src/pages/AcceptInvite.tsx`
- `src/pages/Landing.tsx`
- `src/pages/WorkspaceSettings.tsx`
- `src/components/features/Welcome.tsx`

Ganhos:

- rota `accept-invite` com query suportada corretamente
- hash routing mais determinístico
- sidebar colapsada sempre navega para destinos válidos
- CTA de simulação de convite volta a ser funcional

### 4.2 Resolvido em relação ao relatório anterior

- `MailHub` não está mais preso a estado inline da página.
- `WorkspaceSettings` não está mais preso a persistência inline da página.
- a navegação hash não está mais espalhada sem governança.
- `accept-invite` deixou de conflitar com o próprio route guard.

---

## 5. Current Gap Analysis

### 5.1 Gaps frontend ainda abertos

1. `DataAnalyzer` continua com risco estrutural

- ainda executa abordagem insegura no cliente
- permanece fora desta trilha por decisão de escopo

1. `MailHub` e `Workspace` ainda não têm backend real

- agora têm módulos bem definidos
- ainda faltam contratos reais para troca mock -> API

1. Camadas legadas ainda coexistem

- `api.ts`
- `enhancedAPI.ts`
- `multiProviderService`

Isso não quebra o fluxo atual, mas mantém custo de manutenção e ambiguidades de contrato.

1. Estratégia de erro ainda não foi unificada

- `APIError` legado
- `HttpError` moderno

### 5.2 Gaps frontend já bem preparados para a próxima fase

1. Mail Hub

- pronto para trocar `mailService` mock por contrato real sem reescrever a página

1. Workspace

- pronto para trocar `workspaceService` mock por contrato real sem reescrever a página

1. Navegação

- pronta para novos CTAs e rotas sem proliferar `window.location.hash` solto

---

## 6. ROI Reordered For Frontend-Only Continuation

### HIGH ROI

1. Unificar contratos de erro e reduzir legados de integração

- alvo: `api.ts`, `enhancedAPI.ts`, adapters e consumers mais críticos
- motivo: isso reduz inconsistência transversal em todo o frontend

1. Expandir modularização vertical para áreas ainda pesadas

- alvo principal: `Providers`, partes administrativas e fluxos que ainda vivem concentrados em página/componente
- motivo: o padrão novo já provou valor em Mail/Workspace

1. Fechar a governança de simulação/demo

- alvo: `AuthContext`, `configService`, `gatewayService`, `syncService`, `inviteService`
- motivo: há mock centralizado, mas a política demo/real ainda não está totalmente padronizada

### MEDIUM ROI

1. Refinar feedback UX dos módulos operacionais restantes

- `GatewayDashboard`
- `DataSync`
- `ProvidersSettings`

1. Limpar navegação residual e chamadas antigas diretas

- sobrou pouco, mas vale uma passada de fechamento para evitar regressão futura

1. Atualizar documentação operacional do frontend

- útil agora porque a arquitetura já mudou de fato

### DEFERRED BY SCOPE

1. Hardening do `DataAnalyzer`

- continua sendo o maior risco técnico
- foi explicitamente adiado nesta trilha para não abrir frente backend agora

---

## 7. Explicit Next Sequence (Frontend-Only)

Sequência recomendada a partir daqui, um item por vez:

### Sequência 1

**Unificar contrato de erro e reduzir ambiguidade entre `APIError` e `HttpError`**

Objetivo:

- deixar o frontend falar um idioma único de erro
- reduzir adapters improvisados nos serviços

Arquivos mais prováveis:

- `src/core/http/errors.ts`
- `src/services/api.ts`
- `src/services/enhancedAPI.ts`
- consumers que ainda dependem do shape legado

### Sequência 2

**Padronizar governança demo/mock/real**

Objetivo:

- consolidar como cada domínio decide entre mock, demo e backend
- reduzir decisões distribuídas em contexto, serviço e página

Alvos:

- `src/context/AuthContext.tsx`
- `src/services/configService.ts`
- `src/services/gatewayService.ts`
- `src/services/syncService.ts`
- `src/services/inviteService.ts`

### Sequência 3

**Expandir modularização para áreas administrativas restantes**

Objetivo:

- aplicar o mesmo padrão de `modules/*` onde ainda há concentração de regra em página

Alvos mais prováveis:

- `ProvidersSettings`
- partes de onboarding/admin

### Sequência 4

**Revisão documental final do frontend**

Objetivo:

- deixar arquitetura, convenções e limites do modo demo registrados com precisão

---

## 8. Small Summary: Done vs Pending

### Feito

- fundação HTTP consolidada
- módulos IA principais conectados
- mocks centralizados por domínio
- `MailHub` modularizado
- `Workspace` modularizado
- navegação/guard hash endurecidos
- fluxo de convite público estabilizado

### Ainda falta

- backend real para `MailHub` e `Workspace`
- unificação final de contrato de erro
- governança demo/mock/real mais rígida
- redução dos legados em `api.ts` / `enhancedAPI.ts`
- hardening do `DataAnalyzer` quando essa frente voltar ao escopo

---

## 9. Updated Definition Of Done For The Current Frontend Track

Esta trilha frontend-only pode ser considerada madura quando:

- domínios principais estiverem modularizados
- navegação e guards estiverem estáveis
- mocks estiverem governados de forma única
- contratos de erro estiverem unificados
- troca mock -> backend puder ocorrer por serviço, sem refatorar página inteira

Hoje, o projeto já avançou materialmente nessa direção.
