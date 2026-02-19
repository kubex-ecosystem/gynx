import { secureUserStorage, UserDataStorage } from "@/services/secureStorage";
import { IntegrationSettings } from "@/types/Integrations";
import { UsageTracking, User, UserSettings, UserTrackingMetadata } from "@/types/User";
import * as React from 'react';
import { createContext, ReactNode, useCallback, useContext, useEffect, useState } from 'react';

// Define the shape of our user context
interface UserContextType {
  // User profile data
  user: User | null;
  name: string | null;
  email: string | null;
  setName: (name: string | null) => void;
  setEmail: (email: string | null) => void;
  isEmailVerified: boolean;
  setIsEmailVerified: (isEmailVerified: boolean) => void;
  avatarUrl: string | null;
  setAvatarUrl: (avatarUrl: string | null) => void;
  plan: 'free' | 'pro' | 'enterprise' | null;
  setPlan: (plan: 'free' | 'pro' | 'enterprise' | null) => void;

  // User settings and configurations
  userSettings: UserSettings;
  setUserSettings: (settings: UserSettings | ((prev: UserSettings) => UserSettings)) => void;

  // Integrations
  integrations: IntegrationSettings | null;
  setIntegrations: (integrations: IntegrationSettings | null) => void;

  // Usage tracking
  usageTracking: UsageTracking;
  setUsageTracking: (tracking: UsageTracking | ((prev: UsageTracking) => UsageTracking)) => void;

  // Modal states
  isUserSettingsModalOpen: boolean;
  setIsUserSettingsModalOpen: (isOpen: boolean) => void;

  // Utility functions
  updateUserSetting: <K extends keyof UserSettings>(key: K, value: UserSettings[K]) => void;
  incrementTokenUsage: (tokens: number) => void;
  resetDailyUsage: () => void;
  resetMonthlyUsage: () => void;
  canUseTokens: (requestedTokens: number) => boolean;

  // Rastreabilidade segura
  getUserTrackingMetadata: () => UserTrackingMetadata | null;

  // Storage management
  saveUserData: () => Promise<void>;
  loadUserData: () => Promise<void>;
  clearUserData: () => Promise<void>;
}

// Default user settings
const defaultUserSettings: UserSettings = {
  theme: 'system',
  enableTelemetry: true,
  autoAnalyze: false,
  saveHistory: true,
  tokenLimit: 100000,
  dailyTokenLimit: 10000,
  monthlyTokenLimit: 500000,
  enableDashboardInsights: true,
  enableExperimentalFeatures: false,
  enableBetaFeatures: false,
  apiProvider: 'gemini',
};

// Default usage tracking
const defaultUsageTracking: UsageTracking = {
  totalTokens: 0,
  dailyTokens: 0,
  monthlyTokens: 0,
  lastResetDate: new Date().toISOString().split('T')[0],
};

const UserContext = createContext<UserContextType | undefined>(undefined);

export const UserProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  // Basic user profile state
  const [user, setUser] = useState<User | null>(null);
  const [name, setName] = useState<string | null>(null);
  const [email, setEmail] = useState<string | null>(null);
  const [isEmailVerified, setIsEmailVerified] = useState<boolean>(false);
  const [avatarUrl, setAvatarUrl] = useState<string | null>(null);
  const [plan, setPlan] = useState<'free' | 'pro' | 'enterprise' | null>('free');

  // User configurations state
  const [userSettings, setUserSettings] = useState<UserSettings>(defaultUserSettings);
  const [integrations, setIntegrations] = useState<IntegrationSettings | null>(null);
  const [usageTracking, setUsageTracking] = useState<UsageTracking>(defaultUsageTracking);

  // Modal state
  const [isUserSettingsModalOpen, setIsUserSettingsModalOpen] = useState(false);

  // Utility function to update individual setting
  const updateUserSetting = useCallback(<K extends keyof UserSettings>(key: K, value: UserSettings[K]) => {
    setUserSettings(prev => ({ ...prev, [key]: value }));
  }, []);

  // Token usage functions
  const incrementTokenUsage = useCallback((tokens: number) => {
    setUsageTracking(prev => ({
      ...prev,
      totalTokens: prev.totalTokens + tokens,
      dailyTokens: prev.dailyTokens + tokens,
      monthlyTokens: prev.monthlyTokens + tokens,
    }));
  }, []);

  const resetDailyUsage = useCallback(() => {
    setUsageTracking(prev => ({
      ...prev,
      dailyTokens: 0,
      lastResetDate: new Date().toISOString().split('T')[0],
    }));
  }, []);

  const resetMonthlyUsage = useCallback(() => {
    setUsageTracking(prev => ({
      ...prev,
      monthlyTokens: 0,
    }));
  }, []);

  const canUseTokens = useCallback((requestedTokens: number) => {
    if (plan === 'enterprise') return true;

    const dailyRemaining = userSettings.dailyTokenLimit - usageTracking.dailyTokens;
    const monthlyRemaining = userSettings.monthlyTokenLimit - usageTracking.monthlyTokens;

    return requestedTokens <= dailyRemaining && requestedTokens <= monthlyRemaining;
  }, [plan, userSettings, usageTracking]);

  // Rastreabilidade segura - sem dados sensíveis
  const getUserTrackingMetadata = useCallback((): UserTrackingMetadata | null => {
    if (!user || !name) return null;

    return {
      userId: user.id,
      userName: name,
      createdAt: new Date().toISOString(),
    };
  }, [user, name]);

  // Storage management - usando IndexedDB de forma segura
  const saveUserData = useCallback(async () => {
    try {
      if (!user?.id) return; // Só salvar se houver um usuário identificado

      const userData: UserDataStorage = {
        profile: { name, email, avatarUrl, plan, isEmailVerified },
        settings: userSettings,
        integrations,
        usageTracking,
        user: user ? { ...user, settings: userSettings, integrations, usageTracking } : null,
      };

      await secureUserStorage.saveUserData(user.id, userData);
    } catch (error) {
      console.error('Failed to save user data:', error);
      // Fallback para localStorage em caso de erro
      try {
        const fallbackData = {
          profile: { name, email, avatarUrl, plan, isEmailVerified },
          settings: userSettings,
          integrations,
          usageTracking,
        };
        localStorage.setItem('userDataFallback', JSON.stringify(fallbackData));
      } catch (fallbackError) {
        console.error('Fallback storage also failed:', fallbackError);
      }
    }
  }, [name, email, avatarUrl, plan, isEmailVerified, userSettings, integrations, usageTracking, user]);

  const loadUserData = useCallback(async () => {
    try {
      // Tentar carregar de diferentes fontes
      let userData: UserDataStorage | null = null;

      // Primeiro, tentar o localStorage para verificar se há um usuário anterior
      const fallbackData = localStorage.getItem('userDataFallback');
      if (fallbackData) {
        const parsed = JSON.parse(fallbackData);
        userData = {
          profile: parsed.profile,
          settings: parsed.settings,
          integrations: parsed.integrations,
          usageTracking: parsed.usageTracking,
          user: parsed.user,
        };
      }

      // Se temos um user ID, tentar carregar do IndexedDB
      if (user?.id) {
        const secureData = await secureUserStorage.loadUserData(user.id);
        if (secureData) userData = secureData;
      }

      if (userData) {
        // Carregar profile
        if (userData.profile) {
          setName(userData.profile.name);
          setEmail(userData.profile.email);
          setAvatarUrl(userData.profile.avatarUrl);
          setPlan(userData.profile.plan || 'free');
          setIsEmailVerified(userData.profile.isEmailVerified || false);
        }

        // Carregar settings
        if (userData.settings) {
          setUserSettings({ ...defaultUserSettings, ...userData.settings });
        }

        // Carregar integrations
        if (userData.integrations) {
          setIntegrations(userData.integrations);
        }

        // Carregar usage tracking
        if (userData.usageTracking) {
          setUsageTracking({ ...defaultUsageTracking, ...userData.usageTracking });
        }

        // Carregar user completo
        if (userData.user) {
          setUser(userData.user);
        }
      }
    } catch (error) {
      console.warn('Failed to load user data:', error);
    }
  }, [user?.id]);

  const clearUserData = useCallback(async () => {
    try {
      if (user?.id) {
        await secureUserStorage.clearUserData(user.id);
      }
      localStorage.removeItem('userDataFallback');

      setUser(null);
      setName(null);
      setEmail(null);
      setAvatarUrl(null);
      setPlan('free');
      setIsEmailVerified(false);
      setUserSettings(defaultUserSettings);
      setIntegrations(null);
      setUsageTracking(defaultUsageTracking);
    } catch (error) {
      console.error('Failed to clear user data:', error);
    }
  }, [user?.id]);

  // Check for daily reset
  useEffect(() => {
    const today = new Date().toISOString().split('T')[0];
    if (usageTracking.lastResetDate !== today) {
      resetDailyUsage();
    }
  }, [usageTracking.lastResetDate, resetDailyUsage]);

  // Load user data on mount
  useEffect(() => {
    loadUserData();
  }, [loadUserData]);

  // Auto-save user data when changes occur
  useEffect(() => {
    if (name || email || avatarUrl) { // Only save if there's actual user data
      saveUserData();
    }
  }, [name, email, avatarUrl, plan, isEmailVerified, userSettings, integrations, usageTracking, saveUserData]);

  const value: UserContextType = {
    // User profile
    user,
    name,
    email,
    setName,
    setEmail,
    isEmailVerified,
    setIsEmailVerified,
    avatarUrl,
    setAvatarUrl,
    plan,
    setPlan,

    // User settings
    userSettings,
    setUserSettings,

    // Integrations
    integrations,
    setIntegrations,

    // Usage tracking
    usageTracking,
    setUsageTracking,

    // Modal states
    isUserSettingsModalOpen,
    setIsUserSettingsModalOpen,

    // Utility functions
    updateUserSetting,
    incrementTokenUsage,
    resetDailyUsage,
    resetMonthlyUsage,
    canUseTokens,

    // Rastreabilidade
    getUserTrackingMetadata,

    // Storage
    saveUserData,
    loadUserData,
    clearUserData,
  };

  return <UserContext.Provider value={value}>{children}</UserContext.Provider>;
};

export const useUser = (): UserContextType => {
  const context = useContext(UserContext);
  if (context === undefined) {
    throw new Error('useUser must be used within a UserProvider');
  }
  return context;
};

