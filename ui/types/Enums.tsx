// Enums

export enum Priority {
  High = 'High',
  Medium = 'Medium',
  Low = 'Low',
}

export enum Difficulty {
  High = 'High',
  Medium = 'Medium',
  Low = 'Low',
}

export enum Effort {
  High = 'High',
  Medium = 'Medium',
  Low = 'Low',
}

export enum MaturityLevel {
  Prototype = 'Prototype',
  MVP = 'MVP',
  Production = 'Production',
  Optimized = 'Optimized',
}

export type NotificationType = 'success' | 'error' | 'info';

export enum DataSourceType {
  Manual = 'MANUAL',
  GitHub = 'GITHUB',
  Jira = 'JIRA',
}

export enum ViewType {
  Dashboard = 'DASHBOARD',
  Input = 'INPUT',
  Analysis = 'ANALYSIS',
  Kanban = 'KANBAN',
  Evolution = 'EVOLUTION',
  Chat = 'CHAT',
}

export enum AnalysisType {
  Architecture = 'Architecture',
  CodeQuality = 'Code Quality',
  Security = 'Security Analysis',
  Scalability = 'Scalability Analysis',
  Compliance = 'Compliance & Best Practices',
  DocumentationReview = 'Documentation Review',
  SelfCritique = 'Self-Critique',
}

export type Theme = 'light' | 'dark' | 'system';


