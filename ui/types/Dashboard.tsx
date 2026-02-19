import { UsageMetadata } from "./Analysis";

// Dashboard Types
export interface DashboardInsight {
  title: string;
  summary: string;
  usageMetadata?: UsageMetadata;
}

export interface DashboardSettings {
  enableProjectSummaries: boolean;
  enableImprovementTrends: boolean;
  enableViabilityScores: boolean;
  enableRoiAnalyses: boolean;
  enableMaturityLevels: boolean;
  enableStrengthsAndImprovements: boolean;
  enableDashboardInsights: boolean;
}

export interface DashboardData {
  projectSummaries: string[];
  improvementTrends: { date: string; improvementCount: number }[];
  viabilityScores: { projectName: string; score: number }[];
  roiAnalyses: { projectName: string; assessment: string }[];
  maturityLevels: { projectName: string; level: string }[];
  strengthsAndImprovements: { projectName: string; strengths: string[]; improvements: string[] }[];
  insights: DashboardInsight[];
  improvementCount: number;
  averageViabilityScore: number;
  averageMaturityLevel: string;
}

