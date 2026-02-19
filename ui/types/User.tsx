import { Theme } from "./Enums";
import { IntegrationSettings } from "./Integrations";

// Configurações específicas do usuário - reutilizando tipos existentes
export interface UserSettings {
  // UI settings
  theme: Theme;

  // General settings
  enableTelemetry: boolean;
  autoAnalyze: boolean;
  saveHistory: boolean;

  // API Configuration
  userApiKey?: string;
  apiProvider?: 'openai' | 'claude' | 'gemini' | 'ollama' | 'groq' | 'custom';
  customApiEndpoint?: string;

  // Token limits and usage
  tokenLimit: number;
  dailyTokenLimit: number;
  monthlyTokenLimit: number;

  // Dashboard settings
  enableDashboardInsights: boolean;

  // Feature flags
  enableExperimentalFeatures: boolean;
  enableBetaFeatures: boolean;
}

export interface UsageTracking {
  totalTokens: number;
  dailyTokens: number;
  monthlyTokens: number;
  lastResetDate: string;
  analysisCount?: number;
  projectCount?: number;
  kanbanBoardCount?: number;
  chatSessionCount?: number;
  dashboardViewCount?: number;
}

// Metadados para rastreabilidade sem dados sensíveis
export interface UserTrackingMetadata {
  userId: string;
  userName: string; // Para display apenas
  createdAt: string;
  // Não incluir email, API keys, ou outros dados sensíveis
}

export interface User {
  id: string;
  avatarUrl?: string;
  email: string;
  name: string;
  role: 'user' | 'admin';
  isEmailVerified: boolean;

  plan: 'free' | 'pro' | 'enterprise';

  // Configurações específicas do usuário
  settings: UserSettings;
  integrations: IntegrationSettings | null;
  usageTracking: UsageTracking;

  createdAt: string;
  updatedAt: string;
}

export interface UpdateUserProfile {
  name?: string;
  avatarUrl?: string;
  email?: string;
}

export interface UpdateUserPlan {
  plan: 'free' | 'pro' | 'enterprise';
}
