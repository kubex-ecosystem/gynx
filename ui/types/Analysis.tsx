import { AnalysisType, Difficulty, Effort, MaturityLevel, Priority } from "./Enums";
import { KanbanTaskSuggestion } from "./Kanban";
import { UserTrackingMetadata } from "./User";

// Interfaces for Analysis
export interface Improvement {
  title: string;
  description: string;
  priority: Priority;
  difficulty: Difficulty;
  businessImpact: string;
}

export interface NextStep {
  title: string;
  description: string;
  difficulty: Difficulty;
}

export interface Viability {
  score: number;
  assessment: string;
}

export interface RoiAnalysis {
  assessment: string;
  potentialGains: string[];
  estimatedEffort: Effort;
}

export interface ProjectMaturity {
  level: MaturityLevel;
  assessment: string;
}

export interface UsageMetadata {
  promptTokenCount: number;
  candidatesTokenCount: number;
  totalTokenCount: number;
}

export interface ProjectAnalysis {
  projectName: string;
  analysisType: AnalysisType;
  summary: string;
  strengths: string[];
  improvements: Improvement[];
  nextSteps: {
    shortTerm: NextStep[];
    longTerm: NextStep[];
  };
  viability: Viability;
  roiAnalysis: RoiAnalysis;
  maturity: ProjectMaturity;
  architectureDiagram?: string;
  suggestedQuestions?: string[];
  suggestedKanbanTasks?: KanbanTaskSuggestion[];
  usageMetadata?: UsageMetadata;
  // Rastreabilidade segura - sem dados sensíveis
  userTracking?: UserTrackingMetadata;
}

export interface KeyMetrics {
  previousScore: number;
  currentScore: number;
  scoreChange: number;
  previousStrengths: number;
  currentStrengths: number;
  previousImprovements: number;
  currentImprovements: number;
}

export interface EvolutionAnalysis {
  projectName: string;
  analysisType: AnalysisType;
  evolutionSummary: string;
  keyMetrics: KeyMetrics;
  resolvedImprovements: Improvement[];
  newImprovements: Improvement[];
  persistentImprovements: Improvement[];
  usageMetadata?: UsageMetadata;
}

export interface SelfCritiqueAnalysis {
  confidenceScore: number;
  overallAssessment: string;
  positivePoints: string[];
  areasForRefinement: string[];
  usageMetadata?: UsageMetadata;
}
