import { DashboardTranslations } from '../types';

export const dashboardPtBR: DashboardTranslations = {
  welcome: "Bem-vindo ao Analisador GemX",
  recentAnalyses: "Análises Recentes",
  quickActions: "Ações Rápidas",
  statistics: "Estatísticas",
  noAnalyses: "Nenhuma análise encontrada",
  performanceMetrics: "Métricas de Performance",
  scoreEvolution: "Evolução da Pontuação",
  usage: {
    title: "Uso",
    totalAnalyses: "Total de Análises",
    averageScore: "Pontuação Média",
    successRate: "Taxa de Sucesso"
  },
  emptyState: {
    title: "Nenhum projeto ainda",
    subtitle: "Comece criando a análise do seu primeiro projeto. Seu painel ganhará vida com insights assim que você o fizer!",
    cta: "Iniciar Primeira Análise",
    kpi_total_description: "Total de análises que você realizou em todos os projetos.",
    kpi_score_description: "A pontuação média de viabilidade de suas análises.",
    kpi_type_description: "O tipo de análise que você usa com mais frequência.",
    kpi_tokens_description: "Tokens consumidos por suas análises este mês."
  },
  kpi: {
    totalAnalyses: "Total de Análises",
    totalAnalyses_description: "Análises realizadas para este projeto.",
    averageScore: "Score Médio",
    averageScore_description: "Pontuação média de viabilidade para este projeto.",
    commonType: "Tipo Mais Comum",
    commonType_description: "Tipo de análise mais frequente para este projeto.",
    tokensThisMonth: "Tokens Usados",
    tokensThisMonth_description: "Total de tokens consumidos por este projeto."
  },
  projects: {
    title: "Selecionar Projeto",
    allProjects: "Todos os Projetos",
    recentAnalyses: "Análises Recentes",
    select: "Selecione um projeto...",
    createNew: "Novo Projeto",
    selectPrompt: {
        title: "Por favor, selecione um projeto",
        description: "Escolha um projeto no menu acima para ver seu painel, ou crie um novo."
    }
  },
  scoreTrend: {
      title: "Tendência do Score de Viabilidade"
  },
  recentActivity: {
      title: "Atividade Recente"
  }
};