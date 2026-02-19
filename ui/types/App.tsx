import { NotificationType, Theme } from "./Enums";
import { IntegrationSettings } from "./Integrations";

// Interfaces for App state
export interface ProjectFile {
  id: number;
  name: string;
  content: string;
  isFragment?: boolean;
}

export interface Notification {
  id: number;
  message: string;
  type: NotificationType;
  duration?: number;
}


export interface FeatureFlags {
  enableExperimentalFeatures: boolean;
  enableBetaFeatures: boolean;
}

export interface AIUsageMetadata {
  promptTokenCount: number;
  candidatesTokenCount: number;
  totalTokenCount: number;
}

export interface AppSettings {
  [x: string]: any;
  // UI settings
  theme: Theme;

  // General settings
  enableTelemetry: boolean;
  autoAnalyze: boolean;
  saveHistory: boolean;

  // Dashboard settings
  enableDashboardInsights?: boolean;

  // IntegrationsTabProps
  integrations?: IntegrationSettings;
}



export interface UserProfile {
  name: string;
  email?: string;
  avatar?: string;
  plan?: 'free' | 'pro' | 'enterprise';
  isEmailVerified?: boolean;

  // Other profile-related fields

}
