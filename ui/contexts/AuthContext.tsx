import * as React from 'react';
import { createContext, ReactNode, useContext, useEffect, useState } from 'react';
import { usePersistentState } from '../hooks/usePersistentState';

// Define the shape of our user - expandable for future auth features
interface User {
  id?: string;
  name: string;
  email?: string;
  avatar?: string;
  token?: string;
  refreshToken?: string;
  expiresAt?: number;
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (userData: User) => void;
  logout: () => void;
  updateUser: (userData: Partial<User>) => void;
  // Future auth methods ready for implementation
  refreshAuth?: () => Promise<void>;
  checkAuthStatus?: () => Promise<boolean>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

// A provider component with persistent auth state
export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [user, setUser] = usePersistentState<User | null>('authUser', null);
  const [isLoading, setIsLoading] = useState(false);

  // Derived state
  const isAuthenticated = user !== null;

  // Check if token is expired (if token system is used)
  const isTokenValid = (user: User | null): boolean => {
    if (!user || !user.expiresAt) return true; // No expiration set
    return Date.now() < user.expiresAt;
  };

  // Initialize auth state on app load
  useEffect(() => {
    const initializeAuth = async () => {
      setIsLoading(true);

      // Check if user session is still valid
      if (user && !isTokenValid(user)) {
        // Token expired, logout user
        setUser(null);
      }

      setIsLoading(false);
    };

    initializeAuth();
  }, []); // Only run once on mount

  // Login function - ready for real auth integration
  const login = (userData: User) => {
    setUser(userData);
  };

  // Enhanced logout function
  const logout = () => {
    setUser(null);
    // Future: Clear other auth-related storage, revoke tokens, etc.
  };

  // Update user data
  const updateUser = (userData: Partial<User>) => {
    if (user) {
      setUser({ ...user, ...userData });
    }
  };

  // Mock login function for current phase (backward compatibility)
  const mockLogin = () => {
    login({
      id: 'mock-user-id',
      name: 'Mock User',
      email: 'mock@example.com'
    });
  };

  const value = {
    user,
    isAuthenticated,
    isLoading,
    login: mockLogin, // Keep current behavior for this phase
    logout,
    updateUser
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

// Custom hook to easily access auth context
export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
