import { DashboardTranslations } from '../types';

export const dashboardEnUS: DashboardTranslations = {
  welcome: "Welcome to GemX Analyzer",
  recentAnalyses: "Recent Analyses",
  quickActions: "Quick Actions",
  statistics: "Statistics",
  noAnalyses: "No analyses found",
  performanceMetrics: "Performance Metrics",
  scoreEvolution: "Score Evolution",
  usage: {
    title: "Usage",
    totalAnalyses: "Total Analyses",
    averageScore: "Average Score",
    successRate: "Success Rate"
  },
  emptyState: {
    title: "No projects yet",
    subtitle: "Get started by creating your first project analysis. Your dashboard will light up with insights once you do!",
    cta: "Start First Analysis",
    kpi_total_description: "Total analyses you've performed across all projects.",
    kpi_score_description: "The average viability score from your analyses.",
    kpi_type_description: "The analysis type you use most frequently.",
    kpi_tokens_description: "Tokens consumed by your analyses this month."
  },
  kpi: {
    totalAnalyses: "Total Analyses",
    totalAnalyses_description: "Analyses performed for this project.",
    averageScore: "Average Score",
    averageScore_description: "Average viability score for this project.",
    commonType: "Most Common Type",
    commonType_description: "Most frequent analysis type for this project.",
    tokensThisMonth: "Tokens Used",
    tokensThisMonth_description: "Total tokens consumed by this project."
  },
  projects: {
    title: "Select Project",
    allProjects: "All Projects",
    recentAnalyses: "Recent Analyses",
    select: "Select a project...",
    createNew: "New Project",
    selectPrompt: {
        title: "Please select a project",
        description: "Choose a project from the dropdown above to see its dashboard, or create a new one."
    }
  },
  scoreTrend: {
      title: "Viability Score Trend"
  },
  recentActivity: {
      title: "Recent Activity"
  }
};