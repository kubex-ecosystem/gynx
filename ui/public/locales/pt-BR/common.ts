import { TranslationMessages } from '../types';

export const commonPtBR: TranslationMessages = {
  header: {
    title: "Analisador GemX",
    subtitle: "Transforme a documentação do seu projeto em insights acionáveis com análise orientada por IA."
  },
  navigation: {
    dashboard: "Painel",
    newAnalysis: "Nova Análise",
    currentAnalysis: "Análise Atual",
    kanban: "Kanban",
    history: "Histórico",
    chat: "Chat",
    // FIX: Added missing evolution key
    evolution: "Evolução"
  },
  actions: {
    analyzeProject: "Analisar Projeto",
    analyzing: "Analisando",
    uploadFile: "Carregar Arquivo",
    showExample: "Mostre-me um exemplo",
    exitExample: "Sair do Modo Exemplo",
    load: "Carregar",
    showMore: "Mostrar Mais",
    view: "Ver",
    createKanbanBoard: "Criar Quadro Kanban",
    viewKanbanBoard: "Ver Quadro Kanban"
  },
  common: {
    title: "Título",
    description: "Descrição",
    priorityLabel: "Prioridade",
    difficultyLabel: "Dificuldade",
    delete: "Excluir",
    save: "Salvar",
    cancel: "Cancelar",
    confirm: "Confirmar",
    connect: "Conectar",
    notConnected: "Não Conectado"
  },
  priority: {
    Low: "Baixa",
    Medium: "Média",
    High: "Alta"
  },
  difficulty: {
    Low: "Baixa",
    Medium: "Média",
    High: "Alta"
  },
  effort: {
    Low: "Baixo",
    Medium: "Médio",
    High: "Alto"
  },
  status: {
    TODO: "A Fazer",
    InProgress: "Em Progresso",
    Done: "Concluído",
    Blocked: "Bloqueado"
  },
  loader: {
    loading: "Carregando",
    uploading: "Enviando",
    analyzing: "Analisando",
    saving: "Salvando",
    // FIX: Added missing loader keys
    message: "Analisando seu projeto...",
    subMessage: "Isso pode levar alguns instantes.",
    ariaLabel: "Analisando conteúdo, por favor aguarde.",
    steps: [
      "Analisando estrutura de arquivos...",
      "Avaliando arquitetura...",
      "Verificando qualidade do código...",
      "Identificando possíveis melhorias...",
      "Compilando o relatório..."
    ]
  },
  feedback: {
    success: "Sucesso",
    error: "Erro",
    warning: "Aviso",
    info: "Informação"
  },
  network: {
    connecting: "Conectando",
    connected: "Conectado",
    disconnected: "Desconectado",
    connectionError: "Erro de Conexão",
    retry: "Tentar Novamente",
    online: "Online",
    // FIX: Added missing offline key
    offline: "Você está offline"
  },
  tokenUsage: {
    title: "Aviso de Uso de Tokens",
    inputTokens: "Tokens de Entrada",
    outputTokens: "Tokens de Saída",
    totalTokens: "Total de Tokens",
    estimatedCost: "Custo Estimado"
  },
  settings: {
    language: "Idioma",
    theme: "Tema",
    appearance: "Aparência"
  },
  showExample: "Mostre-me um exemplo",
  analysisTitle: "Título da Análise",
  save: "Salvar"
};