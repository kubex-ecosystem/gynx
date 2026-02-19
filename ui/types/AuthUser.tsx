// Auth Types (Just a sample structure, the real structure is another file in contexts/AuthContext.tsx)

export interface AuthUser {
  id: string;
  email: string;
  name: string;
  // Adicione outros campos conforme necessário
  avatar?: string;
  token?: string;
  refreshToken?: string;
  expiresAt?: number;
}

export interface AuthState {
  isAuthenticated: boolean;
  user: AuthUser | null;
  token: string | null;
  loading: boolean;
  error: string | null;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface SignupData extends LoginCredentials {
  name: string;
  confirmPassword: string;
}

export interface AuthError {
  message: string;
  code?: string;
}

export interface AuthContextType {
  state: AuthState;
  login: (credentials: LoginCredentials) => Promise<void>;
  signup: (data: SignupData) => Promise<void>;
  logout: () => void;
  clearError: () => void;
  // Future auth methods ready for implementation
  refreshAuth?: () => Promise<void>;
  checkAuthStatus?: () => Promise<boolean>;
}

export interface UpdateUserData {
  name?: string;
  email?: string;
  avatar?: string;
}

export interface OAuthProvider {
  name: string;
  authorizationUrl: string;
  clientId: string;
  redirectUri: string;
  scope: string;
  // Add other OAuth parameters as needed
}

export interface OAuthResponse {
  accessToken: string;
  refreshToken?: string;
  expiresIn?: number;
  tokenType?: string;
  scope?: string;
}

const isTokenValid = (user: AuthUser | null): boolean => {
  if (!user || !user.expiresAt) return true; // No expiration set
  return Date.now() < user.expiresAt;
};

const setUser = (
  user: AuthUser | null | ((prev: AuthUser | null) => AuthUser | null)
) => {
  // Aceita valor direto ou função updater (compatível com React setState)
  let nextUser: AuthUser | null;
  try {
    const prevRaw = localStorage.getItem('authUser');
    const prev: AuthUser | null = prevRaw ? (JSON.parse(prevRaw) as AuthUser) : null;
    if (typeof user === 'function') {
      nextUser = (user as (p: AuthUser | null) => AuthUser | null)(prev);
    } else {
      nextUser = user;
    }
    if (nextUser) {
      localStorage.setItem('authUser', JSON.stringify(nextUser));
    } else {
      localStorage.removeItem('authUser');
    }
  } catch {
    // Silenciar falhas de storage em ambientes restritos
  }
};

// Logout function
const logout = () => {
  setUser(null);
  // Optionally clear other auth-related state or tokens
  setToken(null);
  setLoading(false);
  setError(null);
};

// Update user details function
const updateUser = (userData: Partial<AuthUser>) => {
  setUser((prev: AuthUser | null): AuthUser | null => {
    if (!prev) return null;
    return { ...prev, ...userData };
  });
  // Optionally update other auth-related states
  setLoading(false);
  setError(null);
  // manter placeholders mínimos — não chamar setToken com updater here
  setIsLoading(false);
  setAuthError(null);

};

// Future methods for refreshing auth and checking status can be implemented here
const refreshAuth = async () => {
  // Placeholder for future token refresh logic
};

const checkAuthStatus = async () => {
  // Lê o usuário armazenado de forma segura
  try {
    const raw = localStorage.getItem('authUser');
    const stored: AuthUser | null = raw ? (JSON.parse(raw) as AuthUser) : null;
    const isAuthenticated = stored !== null && isTokenValid(stored);
    return isAuthenticated;
  } catch {
    return false;
  }
};

// Optionally set other auth-related states
const setIsLoading = (loading: boolean) => {
  // Placeholder for setting loading state
  // e.g., setState(prev => ({ ...prev, loading }));
};

const setAuthError = (error: AuthError | null) => {
  // Placeholder for setting error state
  // e.g., setState(prev => ({ ...prev, error: error ? error.message : null }));
};

const setToken = (token: string | null | ((prev: string | null) => string | null)) => {
  // Aceita valor direto ou updater; persiste em localStorage de forma segura
  try {
    const prev = localStorage.getItem('authToken');
    let next: string | null;
    if (typeof token === 'function') {
      next = (token as (p: string | null) => string | null)(prev);
    } else {
      next = token;
    }
    if (next) localStorage.setItem('authToken', next);
    else localStorage.removeItem('authToken');
  } catch {
    // Silenciar falhas de storage
  }
};

const setLoading = (loading: boolean) => {
  // Placeholder for setting loading state
  // e.g., setState(prev => ({ ...prev, loading }));
};

const setError = (error: string | null) => {
  // Placeholder for setting error state
  // e.g., setState(prev => ({ ...prev, error }));
};

// Project Types
export interface Project {
  id: number;
  name: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
  settings: ProjectSettings;
  order: number; // For custom ordering
}

export interface NewProject {
  name: string;
  description?: string;
  settings?: Partial<ProjectSettings>;
}

export interface UpdateProject {
  id: number;
  name?: string;
  description?: string;
  settings?: Partial<ProjectSettings>;
}

export interface DeleteProject {
  id: number;
}

export interface ReorderProjects {
  sourceIndex: number;
  destinationIndex: number;
}

export interface ProjectState {
  projects: Project[];
  currentProjectId: number | null;
  isLoading: boolean;
  error: string | null;
}

export interface ProjectSettings {
  autoAnalyze: boolean;
  saveHistory: boolean;
  tokenLimit: number;
  userApiKey?: string;
  githubPat?: string;
  jiraInstanceUrl?: string;
  jiraUserEmail?: string;
  jiraApiToken?: string;
}

// Context File Types
export interface ContextFile {
  id: string;
  name: string;
  content: string;
  isFragment?: boolean;
}
