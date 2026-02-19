import { useAuth } from '@/contexts/AuthContext.tsx';
import { useCallback, useState } from 'react';

interface LoginCredentials {
  email: string;
  password: string;
}

interface SignupData extends LoginCredentials {
  name: string;
  confirmPassword: string;
}

interface AuthError {
  message: string;
  code?: string;
}

/**
 * Enhanced auth hook with login/signup methods ready for next phase implementation
 * Provides methods for different auth providers and error handling
 */
export const useAuthActions = () => {
  const { login, logout, updateUser } = useAuth();
  const [authError, setAuthError] = useState<AuthError | null>(null);
  const [isAuthLoading, setIsAuthLoading] = useState(false);

  // Clear auth errors
  const clearError = useCallback(() => {
    setAuthError(null);
  }, []);

  // Login with email/password (ready for API integration)
  const loginWithCredentials = useCallback(async (credentials: LoginCredentials) => {
    setIsAuthLoading(true);
    setAuthError(null);

    try {
      // TODO: Replace with actual API call in next phase
      // const response = await authAPI.login(credentials);

      // Mock implementation for current phase
      await new Promise(resolve => setTimeout(resolve, 1000)); // Simulate API delay

      login({
        id: 'user-' + Date.now(),
        name: credentials.email.split('@')[0], // Use email prefix as name
        email: credentials.email,
        token: 'mock-jwt-token',
        expiresAt: Date.now() + (24 * 60 * 60 * 1000) // 24 hours
      });

    } catch (error) {
      setAuthError({
        message: error instanceof Error ? error.message : 'Login failed',
        code: 'LOGIN_ERROR'
      });
      throw error;
    } finally {
      setIsAuthLoading(false);
    }
  }, [login]);

  // Signup (ready for API integration)
  const signupWithCredentials = useCallback(async (signupData: SignupData) => {
    setIsAuthLoading(true);
    setAuthError(null);

    try {
      if (signupData.password !== signupData.confirmPassword) {
        throw new Error('Passwords do not match');
      }

      // TODO: Replace with actual API call in next phase
      // const response = await authAPI.signup(signupData);

      // Mock implementation
      await new Promise(resolve => setTimeout(resolve, 1000));

      login({
        id: 'user-' + Date.now(),
        name: signupData.name,
        email: signupData.email,
        token: 'mock-jwt-token',
        expiresAt: Date.now() + (24 * 60 * 60 * 1000)
      });

    } catch (error) {
      setAuthError({
        message: error instanceof Error ? error.message : 'Signup failed',
        code: 'SIGNUP_ERROR'
      });
      throw error;
    } finally {
      setIsAuthLoading(false);
    }
  }, [login]);

  // OAuth login (Google, GitHub, etc.) - ready for implementation
  const loginWithProvider = useCallback(async (provider: 'google' | 'github' | 'microsoft') => {
    setIsAuthLoading(true);
    setAuthError(null);

    try {
      // TODO: Implement OAuth flow in next phase
      // const response = await authAPI.loginWithProvider(provider);

      // Mock implementation
      await new Promise(resolve => setTimeout(resolve, 1500));

      login({
        id: `${provider}-user-` + Date.now(),
        name: `${provider.charAt(0).toUpperCase() + provider.slice(1)} User`,
        email: `user@${provider}.com`,
        avatar: `https://api.dicebear.com/7.x/avataaars/svg?seed=${provider}`,
        token: `mock-${provider}-token`,
        expiresAt: Date.now() + (24 * 60 * 60 * 1000)
      });

    } catch (error) {
      setAuthError({
        message: error instanceof Error ? error.message : `${provider} login failed`,
        code: `${provider.toUpperCase()}_LOGIN_ERROR`
      });
      throw error;
    } finally {
      setIsAuthLoading(false);
    }
  }, [login]);

  // Logout with cleanup
  const performLogout = useCallback(async () => {
    setIsAuthLoading(true);

    try {
      // TODO: Add API call to revoke tokens in next phase
      // await authAPI.logout();

      logout();
    } catch (error) {
      console.error('Logout error:', error);
      // Force logout even if API call fails
      logout();
    } finally {
      setIsAuthLoading(false);
    }
  }, [logout]);

  // Password reset (ready for implementation)
  const resetPassword = useCallback(async (email: string) => {
    setIsAuthLoading(true);
    setAuthError(null);

    try {
      // TODO: Implement password reset in next phase
      // await authAPI.resetPassword(email);

      // Mock implementation
      await new Promise(resolve => setTimeout(resolve, 1000));

      return { success: true, message: 'Password reset email sent' };

    } catch (error) {
      setAuthError({
        message: error instanceof Error ? error.message : 'Password reset failed',
        code: 'RESET_PASSWORD_ERROR'
      });
      throw error;
    } finally {
      setIsAuthLoading(false);
    }
  }, []);

  return {
    // Auth actions
    loginWithCredentials,
    signupWithCredentials,
    loginWithProvider,
    performLogout,
    resetPassword,

    // State
    authError,
    isAuthLoading,
    clearError,

    // Utils
    updateUser
  };
};

// Types export for use in components
export type { AuthError, LoginCredentials, SignupData };
