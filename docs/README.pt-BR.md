# GNyx

English version: [../README.md](../README.md)

## Sumário

- [Visão Geral](#visão-geral)
- [Estado Atual do Produto](#estado-atual-do-produto)
- [Capacidades Principais](#capacidades-principais)
- [Contexto do Ecossistema](#contexto-do-ecossistema)
- [Visão Geral da Arquitetura](#visão-geral-da-arquitetura)
- [Estrutura do Repositório](#estrutura-do-repositório)
- [Modelo de Runtime](#modelo-de-runtime)
- [Comandos Principais](#comandos-principais)
- [Superfície HTTP Ativa](#superfície-http-ativa)
- [Autenticação e Acesso](#autenticação-e-acesso)
- [Providers de IA e Runtime](#providers-de-ia-e-runtime)
- [BI e Geração Guiada por Metadados](#bi-e-geração-guiada-por-metadados)
- [Boundary com o Domus](#boundary-com-o-domus)
- [Configuração e Runtime Home](#configuração-e-runtime-home)
- [Notas de Desenvolvimento](#notas-de-desenvolvimento)
- [Mapa de Documentação](#mapa-de-documentação)
- [Screenshots](#screenshots)
- [Foco Atual](#foco-atual)

## Visão Geral

`GNyx` é a camada de aplicação voltada ao produto dentro do ecossistema Kubex.

Ele combina:

- um gateway/backend em Go
- um frontend React embarcado (`GNyx-UI`)
- execução real de providers de IA por um caminho unificado de runtime
- fluxos de autenticação, sessão, convite e acesso
- superfícies administrativas e operacionais
- um boundary tipado com o `Domus`
- infraestrutura compartilhada vinda do `Kbx`

Neste ponto, `GNyx` já não é apenas uma casca ou proposta arquitetural. Ele já expõe comportamento real de produto e valor demonstrável em runtime.

## Estado Atual do Produto

Atualmente operacionais ou materialmente consolidados:

- entrega do frontend embarcado pelo backend
- runtime real de providers pelo caminho ativo do gateway
- features assistidas por IA funcionando no frontend via rotas do backend
- consolidação do consumo do registry de providers entre `Kbx` e `GNyx`
- fluxos funcionais de sign-in, refresh, logout e bootstrap de `me`
- payload enriquecido de `/me` com `access_scope`, memberships e permissões efetivas
- bootstrap tenant-aware no frontend e guards de rota baseados em acesso real
- `Access Management` MVP para membros, convites e reatribuição controlada de role
- prova de conceito de BI guiada por metadados reais do catálogo Sankhya
- tela `BI Studio` com geração, cópia de schema e exportação em ZIP

Ainda em evolução:

- cobertura mais ampla de fluxos de negócio em todas as telas
- governança mais profunda de planos e entitlements
- RBAC multi-tenant / team-level mais rico
- expansão da frente de BI para além da primeira slice grounded
- avanço de áreas ainda híbridas ou mockadas para consumo real via Domus/backend

## Capacidades Principais

No estágio atual, o `GNyx` fornece ou parcialmente fornece:

- execução unificada de providers e streaming
- exposição de catálogo, status, config e health de providers
- fluxos de chat, sumarização, código, image-prompt e prompt crafting
- autenticação, sessões, aceite de convite e bootstrap de acesso
- shell tenant-aware e navegação sensível a RBAC
- administração de providers e fluxos BYOK
- gestão de acesso para membros, convites e troca de role
- orquestração de sincronização de metadados Sankhya
- geração grounded de boards BI com exportação

## Contexto do Ecossistema

`GNyx` opera como a camada de orquestração e produto sobre quatro projetos relacionados:

- `GNyx-UI`: experiência frontend embarcada
- `Domus`: substrate de dados, stores tipados, migrações e runtime PostgreSQL
- `Kbx`: infraestrutura compartilhada como config, crypto, mail utilities e provider registry
- `GETL`: utilitário de ETL/sync agora usado na ingestão do catálogo Sankhya

`Logz` também participa do ecossistema mais amplo, mas é tratado intencionalmente como fundação compartilhada de logging e não como alvo específico de evolução do `GNyx`.

## Visão Geral da Arquitetura

Hoje o `GNyx` se distribui por estas camadas:

1. `cmd/` e entrypoints Cobra para gateway e comandos operacionais
2. `internal/runtime/` para boot, wiring e startup do gateway
3. `internal/api/routes/` para a superfície HTTP ativa
4. `internal/features/` para features orientadas a produto, como BI
5. `internal/services/` para serviços de domínio, como invites
6. `internal/dsclient/` para o single integration point com o `Domus`
7. `frontend/` para a UI embarcada do produto

Uma regra crítica de design continua valendo:

- `GNyx` deve alcançar o `Domus` pelo single integration point `internal/dsclient`, e não por atalhos arbitrários entre repositórios.

## Estrutura do Repositório

```text
cmd/                    entrypoints Cobra
config/                 config local, providers e manifests GETL
frontend/               aplicação React embarcada (GNyx-UI)
internal/api/           rotas HTTP e wiring de rotas
internal/auth/          lógica de tokens e autenticação
internal/dsclient/      boundary de integração com Domus
internal/features/      features de produto como BI
internal/runtime/       boot do gateway e composição de runtime
internal/services/      serviços de domínio como invite flows
.notes/                 tarefas, contexto e análises de implementação
```

## Modelo de Runtime

O fluxo local mais usado hoje para o gateway é:

```bash
gnyx gateway up -e ./config/.env.local -D
```

Fluxo equivalente rodando a partir do código-fonte:

```bash
go run ./cmd/main.go gateway up -e ./config/.env.local -D
```

O `GNyx` depende de:

- config de runtime vinda de env files e runtime home
- config ativa de providers em `~/.kubex/gnyx/config/providers.yaml`
- assets do frontend gerados a partir de `frontend/`
- stores do `Domus` acessados por `internal/dsclient`

## Comandos Principais

Subir o gateway:

```bash
go run ./cmd/main.go gateway up -e ./config/.env.local -D
```

Compilar o backend:

```bash
go build ./...
```

Compilar o frontend:

```bash
cd frontend
pnpm exec vite build
```

Sincronizar o catálogo Sankhya para o PostgreSQL pelo caminho orquestrado do `GNyx`:

```bash
go run ./cmd/main.go metadata sankhya sync \
  --env-file ./config/.env.local \
  --pg-dsn 'postgres://kubex_adm:admin123@localhost:5432/postgres?sslmode=disable'
```

Nota importante:

- como o `GETL` depende atualmente de `godror`, alguns ambientes de build podem exigir `CGO=1` ao compilar caminhos que importam `GETL` de forma transitiva.

## Superfície HTTP Ativa

O gateway ativo hoje expõe, entre outras:

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

A superfície HTTP está evoluindo intencionalmente de endpoints mais infraestruturais para rotas de produto orientadas a feature.

## Autenticação e Acesso

A autenticação já é funcional e os caminhos sensíveis de segurança passaram por hardening.

Postura atual de acesso:

- login, refresh, logout e `me` estão operacionais
- expiração e revogação de refresh foram endurecidas
- aceite de convite foi endurecido para retry seguro e recuperação de falha parcial
- `/me` agora retorna `access_scope`, memberships, team memberships, pending access e permissões efetivas
- guards do frontend agora dependem do estado real de acesso autenticado, e não de suposições mockadas legadas
- `Access Management` fornece a primeira superfície de gestão tenant-scoped

Boundary atual do RBAC MVP:

- escopo por tenant, não por team
- permissões efetivas expostas para o tenant ativo
- enforcement visual no frontend para áreas administrativas de maior ROI
- inclui troca controlada de role, mas não edição direta de permissões

## Providers de IA e Runtime

Os providers já não são teóricos no `GNyx`.

Eles são consumidos de forma real pelo runtime backend e expostos ao frontend.

Estado prático atual:

- o hardening do registry ocorreu no `Kbx`
- o `GNyx` agora usa um único caminho efetivo de config de providers durante o boot
- catálogo e config de runtime ficam disponíveis pelo gateway ativo
- as ferramentas do frontend podem selecionar provider explicitamente por feature
- `Groq` é hoje o provider mais confiável para a demo de geração BI
- `Gemini` é suportado, mas tende a ser mais lento e a cair com mais frequência em fallback na forma atual do contrato BI

## BI e Geração Guiada por Metadados

Uma das frentes mais importantes novas já está materialmente implementada.

Fluxo atual:

1. CSVs de metadados BI do Sankhya são ingeridos no PostgreSQL sob `sankhya_catalog`
2. o `Domus` registra o estado de carga em `public.external_metadata_registry`
3. o `GNyx` usa queries grounded nesses metadados para preparar contexto de geração
4. o runtime de provider gera ou parcialmente gera um plano de board
5. o backend valida e compila o resultado para um `DashboardSchema`
6. o frontend `BI Studio` expõe geração, inspeção, cópia e exportação em ZIP

Estado atual da slice de BI:

- primeiro domínio grounded: `sales`
- modos de geração: `llm`, `llm_recovered`, `fallback_template`
- o ZIP exportado contém schema gerado, board plan, grounding context e generation metadata
- o frontend só expõe a feature como operacional quando o catálogo realmente está disponível

## Boundary com o Domus

`GNyx` não trata o `Domus` como dependência incidental.

Ele usa o `Domus` como boundary ativo de dados para:

- dados de usuário e sessão
- convites e pending access
- companies / tenants e memberships por uma composição de stores tipados e SQL local
- leitura do external metadata registry para readiness da frente de BI

Uma regra importante continua valendo:

- a expansão de dados deve continuar convergindo para o boundary único de `internal/dsclient`, em vez de criar caminhos paralelos.

## Configuração e Runtime Home

O `GNyx` usa um modelo de runtime home em:

```text
~/.kubex/gnyx/
```

Regra operacional importante:

- esse runtime home deve ser tratado como fonte ativa de verdade quando materializado
- config ausente deve ser criada de forma conservadora a partir dos valores passados ou defaults seguros
- config já existente não deve sofrer sobrescrita destrutiva em execuções repetidas ou paralelas

Arquivos tipicamente ativos incluem:

- `~/.kubex/gnyx/config/providers.yaml`
- secrets, fragments de config e material gerado em runtime

A config local do repositório continua relevante para desenvolvimento e bootstrap, especialmente:

- `config/.env.local`
- `config/providers.yaml`
- `config/getl/sankhya_catalog/*`

## Notas de Desenvolvimento

Validação recomendada do backend:

```bash
go build ./...
go test ./...
```

Validação recomendada do frontend:

```bash
cd frontend
pnpm exec tsc --noEmit
pnpm exec vite build
```

Smoke recomendado de provider/runtime:

- subir o gateway a partir do código-fonte
- chamar `GET /api/v1/providers`
- chamar `POST /api/v1/unified`
- validar disponibilidade dos providers e o estado retornado de health/config

Smoke recomendado da frente BI:

- confirmar `GET /api/v1/bi/catalog/status`
- gerar um board via `POST /api/v1/bi/boards/generate`
- exportar via `POST /api/v1/bi/boards/export`
- validar o caminho `BI Studio` no frontend

## Mapa de Documentação

Documentação local importante:

- [`frontend/README.md`](../frontend/README.md)
- [`docs/README.pt-BR.md`](./README.pt-BR.md)
- [`.notes/analyzis/global-execution-plan/`](../.notes/analyzis/global-execution-plan)
- [`.notes/analyzis/gnyx-skw-dynamic-ui/`](../.notes/analyzis/gnyx-skw-dynamic-ui)
- [`.notes/context/`](../.notes/context)

## Screenshots

Sugestões de placeholders:

- `[Screenshot Placeholder: Welcome e shell tenant-aware]`
- `[Screenshot Placeholder: Access Management]`
- `[Screenshot Placeholder: Providers Settings com estado real do runtime]`
- `[Screenshot Placeholder: BI Studio com resultado de geração]`

## Foco Atual

A baseline atual já é forte o bastante para sustentar os próximos passos:

- consolidação mais ampla de frontend-backend sobre acesso real e runtime de providers
- slices de geração BI mais ricas além da primeira prova de conceito de `sales`
- evolução mais ampla do `Domus` sem abandonar os fluxos operacionais que já funcionam hoje
