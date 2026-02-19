# Resumo do Projeto: GemX Analyzer

## 1. Visão Geral

O **GemX Analyzer** é uma aplicação web de página única (SPA) projetada para atuar como uma ferramenta de análise de projetos de software. Utilizando a API do Google Gemini, a aplicação recebe documentação de projeto (como `READMEs`, notas de lançamento, etc.) e gera insights estruturados e acionáveis. O objetivo é fornecer aos desenvolvedores e gerentes de projeto uma avaliação rápida e inteligente sobre a viabilidade, maturidade, pontos fortes e áreas de melhoria de um projeto.

---

## 2. Arquitetura e Stack Tecnológica

A aplicação é construída com uma arquitetura moderna de frontend, priorizando a reatividade, a persistência de dados no lado do cliente e uma experiência de usuário fluida.

### Stack Principal
- **Framework:** React 19
- **Linguagem:** TypeScript
- **Build Tool:** Vite
- **Estilização:** Tailwind CSS
- **Animações:** Framer Motion
- **Ícones:** Lucide React

### Gerenciamento de Estado
O estado da aplicação é dividido em duas categorias:
- **Dados de Projeto:** Todos os dados relacionados a um projeto (nome, histórico de análises, quadro Kanban, chats) são encapsulados em um único objeto `Project`. Esses projetos são armazenados de forma robusta no **IndexedDB** em um `objectStore` dedicado chamado `projects`. Isso garante isolamento total e escalabilidade.
- **Configurações Globais:** Configurações da aplicação e perfil do usuário são gerenciados pelo hook `usePersistentState`, que utiliza um `objectStore` genérico 'keyval' no IndexedDB com fallback para `localStorage`.

### Integração com a IA (Gemini API)
- A comunicação com a API Gemini é abstraída em uma camada de serviço (`services/gemini/`).
- A aplicação utiliza o modelo `gemini-2.5-flash` para as análises.
- Para garantir respostas estruturadas e consistentes, a aplicação define um `responseSchema` no formato JSON para as chamadas à API, o que minimiza a necessidade de parsing complexo de texto no cliente.

### Internacionalização (i18n)
- A aplicação suporta múltiplos idiomas (atualmente `en-US` e `pt-BR`).
- A tradução é gerenciada por um `LanguageContext` e um hook customizado `useTranslation`.
- Os textos são armazenados em **módulos TypeScript** (`.ts`) localizados em `public/locales/`, que são carregados dinamicamente.

### Estrutura de Diretórios
```
/
├── components/     # Componentes React reutilizáveis, organizados por feature
├── contexts/       # Provedores de contexto para estado global
├── data/           # Dados estáticos, como o modo de exemplo
├── docs/           # Documentação do projeto
├── hooks/          # Hooks customizados (usePersistentState, useTranslation, etc.)
├── lib/            # Utilitários de baixo nível (ex: idb.ts)
├── public/         # Assets públicos, incluindo os arquivos de tradução
├── services/       # Lógica de comunicação com APIs externas (Gemini)
└── types/          # Definições de tipos e interfaces TypeScript
```

---

## 3. Funcionalidades Implementadas

- **Análise de Projetos com IA:**
  - O usuário pode colar texto ou importar o contexto de um repositório GitHub.
  - **Cinco tipos de análise** estão disponíveis: Viabilidade Geral, Auditoria de Segurança, Revisão de Escalabilidade, Qualidade de Código e Revisão de Documentação.
  - A resposta da IA é exibida em um formato rico e estruturado.

- **Chat Interativo com IA:**
  - Após cada análise, o usuário pode interagir com um assistente de IA para aprofundar os insights.
  - **Sugestões Proativas:** A IA gera e sugere perguntas contextuais para guiar a conversa.

- **Dashboard de Métricas:**
  - Exibe um painel de controle com estatísticas agregadas para o projeto ativo.
  - KPIs incluem: total de análises, pontuação média, tipo mais comum e uso de tokens.
  - Apresenta um gráfico de tendência da evolução da pontuação de viabilidade.

- **Histórico e Comparação:**
  - Todas as análises (se habilitado) são salvas dentro do objeto do projeto correspondente no IndexedDB.
  - Um painel de histórico permite visualizar, carregar ou excluir análises passadas.
  - Funcionalidade de **comparação** para gerar um "relatório de evolução" via IA.

- **Quadro Kanban:**
  - A partir de uma análise, o usuário pode gerar um quadro Kanban pré-populado com tarefas baseadas nas sugestões da IA.

- **Gerenciamento de Múltiplos Projetos:**
  - A aplicação é centrada em projetos, permitindo que o usuário crie e alterne entre diferentes projetos, cada um com seu próprio histórico, kanban e chats isolados.

- **Gerenciamento de Dados do Usuário:**
  - Configurações da aplicação e perfil do usuário são persistidos.
  - Funcionalidade de **importar/exportar** todos os dados da aplicação (projetos, configurações, perfil) em um único arquivo JSON para backup e migração entre dispositivos.

- **UI/UX:**
  - Tema escuro consistente, animações fluidas, notificações de feedback e design responsivo.