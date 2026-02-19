// Types for translation structure
export interface TranslationMessages {
  header: {
    title: string;
    subtitle: string;
  };
  navigation: {
    dashboard: string;
    newAnalysis: string;
    currentAnalysis: string;
    kanban: string;
    history: string;
    chat: string;
    evolution: string;
  };
  actions: {
    analyzeProject: string;
    analyzing: string;
    uploadFile: string;
    showExample: string;
    exitExample: string;
    load: string;
    showMore: string;
    view: string;
    createKanbanBoard: string;
    viewKanbanBoard: string;
  };
  common: {
    title: string;
    description: string;
    priorityLabel: string;
    difficultyLabel: string;
    delete: string;
    save: string;
    cancel: string;
    confirm: string;
    connect: string;
    notConnected: string;
  };
  priority: {
    Low: string;
    Medium: string;
    High: string;
  };
  difficulty: {
    Low: string;
    Medium: string;
    High: string;
  };
  effort: {
    Low: string;
    Medium: string;
    High: string;
  };
  status: {
    TODO: string;
    InProgress: string;
    Done: string;
    Blocked: string;
  };
  loader: {
    loading: string;
    uploading: string;
    analyzing: string;
    saving: string;
    message: string;
    subMessage: string;
    ariaLabel: string;
    steps: string[];
  };
  feedback: {
    success: string;
    error: string;
    warning: string;
    info: string;
  };
  network: {
    connecting: string;
    connected: string;
    disconnected: string;
    connectionError: string;
    retry: string;
    online: string;
    offline: string;
  };
  tokenUsage: {
    title: string;
    inputTokens: string;
    outputTokens: string;
    totalTokens: string;
    estimatedCost: string;
  };
  settings: {
    language: string;
    theme: string;
    appearance: string;
  };
  showExample: string;
  analysisTitle: string;
  save: string;
}

export interface AnalysisTranslations {
  results: {
    title: string;
    summary: {
      title: string;
    };
    viability: {
      title: string;
      scoreLabel: string;
      assessmentLabel: string;
      scoreEvolution: string;
    };
    roi: {
      title: string;
      assessmentLabel: string;
      effortLabel: string;
      gainsLabel: string;
    };
    strengths: {
      title: string;
    };
    improvements: {
      title: string;
      impact: string;
      businessImpact: string;
    };
    nextSteps: {
      title: string;
      shortTerm: string;
      longTerm: string;
    };
    timeline: {
      title: string;
      phases: string;
      estimatedDuration: string;
    };
    risks: {
      title: string;
      technical: string;
      business: string;
      mitigation: string;
    };
    metrics: {
      title: string;
      current: string;
      target: string;
      kpi: string;
    };
    resources: {
      title: string;
      teamSize: string;
      budget: string;
      technology: string;
    };
    conclusion: {
      title: string;
      recommendation: string;
      confidence: string;
    };
  };
  comparison: {
    title: string;
    analyzing: string;
    differences: string;
    similarities: string;
    evolution: string;
    summary: string;
  };
}

export interface ChatTranslations {
  title: string;
  placeholder: string;
  send: string;
  typing: string;
  clear: string;
  history: string;
  export: string;
  import: string;
  messages: {
    welcome: string;
    error: string;
    thinking: string;
    noMessages: string;
  };
}

export interface DashboardTranslations {
  welcome: string;
  recentAnalyses: string;
  quickActions: string;
  statistics: string;
  noAnalyses: string;
  performanceMetrics: string;
  scoreEvolution: string;
  usage: {
    title: string;
    totalAnalyses: string;
    averageScore: string;
    successRate: string;
  };
  emptyState: {
    title: string;
    subtitle: string;
    cta: string;
    kpi_total_description: string;
    kpi_score_description: string;
    kpi_type_description: string;
    kpi_tokens_description: string;
  };
  kpi: {
    totalAnalyses: string;
    totalAnalyses_description: string;
    averageScore: string;
    averageScore_description: string;
    commonType: string;
    commonType_description: string;
    tokensThisMonth: string;
    tokensThisMonth_description: string;
  };
  projects: {
    title: string;
    allProjects: string;
    recentAnalyses: string;
    select: string;
    createNew: string;
    selectPrompt: {
        title: string;
        description: string;
    }
  };
  // FIX: Added missing properties for dashboard translations
  scoreTrend: {
    title: string;
  };
  recentActivity: {
    title: string;
  };
}

export interface ExampleTranslations {
  mode: {
    title: string;
    description: string;
    notice: string;
  };
  project: {
    name: string;
    description: string;
    type: string;
    domain: string;
  };
}

export interface InputTranslations {
  title: string;
  projectName: string;
  projectNamePlaceholder: string;
  importFromGithub: string;
  description: string;
  placeholder: string;
  useExample: string;
  analysisTypeTitle: string;
  analysisTypes: {
    GENERAL: {
      label: string;
      description: string;
    };
    SECURITY: {
      label: string;
      description: string;
    };
    SCALABILITY: {
      label: string;
      description: string;
    };
    CODEQUALITY: {
      label: string;
      description: string;
    };
    DOCUMENTATIONREVIEW: {
      label: string;
      description: string;
    };
  };
}

export interface KanbanTranslations {
  title: string;
  addCard: string;
  editCard: string;
  exampleModeNotice: string;
  notes: string;
  notesPlaceholder: string;
  deleteConfirm: {
      title: string;
      message: string;
      confirm: string;
  };
  columns: {
    todo: string;
    inProgress: string;
    done: string;
    blocked: string;
  };
}

export interface LandingTranslations {
  cta: string;
  featuresTitle: string;
  featuresSubtitle: string;
  dynamicPhrases: string[];
  hero: {
    title: {
      static: string;
    };
    subtitle: string;
    cta: string;
  };
  features: {
    title: string;
    aiDriven: {
      title: string;
      description: string;
    };
    comprehensive: {
      title: string;
      description: string;
    };
    actionable: {
      title: string;
      description: string;
    };
  };
  howItWorks: {
    title: string;
    step1: {
      title: string;
      description: string;
    };
    step2: {
      title: string;
      description: string;
    };
    step3: {
      title: string;
      description: string;
    };
  };
  featureDetails: {
    general: string;
    security: string;
    scalability: string;
    codeQuality: string;
    documentation: string;
  };
}

export interface SettingsTranslations {
  title: string;
  general: {
    title: string;
    language: string;
    theme: string;
  };
  notifications: {
    title: string;
    email: string;
    push: string;
    desktop: string;
  };
  privacy: {
    title: string;
    analytics: string;
    cookies: string;
  };
  account: {
    title: string;
    profile: string;
    security: string;
    billing: string;
  };
}

export interface AuthTranslations {
  logout: string;
}

export interface HistoryTranslations {
  title: string;
}

export interface ProfileTranslations {
  title: string;
  avatar: {
    change: string;
  };
  nameLabel: string;
  namePlaceholder: string;
  emailLabel: string;
  emailPlaceholder: string;
  save: string;
}

export interface TabsTranslations {
  profile: string;
  preferences: string;
  integrations: string;
  data: string;
}

// Files namespace
export interface FilesTranslations {
  title: string;
  addFromUpload: string;
  addFile: string;
  emptyState: string;
}

// Data Sources namespace
export interface DataSourcesTranslations {
  github: {
    placeholder: string;
  };
}

// GitHub Search namespace
export interface GithubSearchTranslations {
  button: string;
}

// Token Usage namespace
export interface TokenUsageTranslations {
  monthlyUsage: string;
}

// Import Export namespace
export interface ImportExportTranslations {
  title: string;
  description: string;
  warning: string;
  importLabel: string;
  exportLabel: string;
  confirm: {
    title: string;
    message: string;
  };
}

// Notifications namespace
export interface NotificationsTranslations {
  importSuccess: string;
}

// Main locale type combining all translations
export interface LocaleTranslations {
  common: TranslationMessages;
  analysis: AnalysisTranslations;
  auth: AuthTranslations;
  chat: ChatTranslations;
  dashboard: DashboardTranslations;
  dataSources: DataSourcesTranslations;
  example: ExampleTranslations;
  files: FilesTranslations;
  githubSearch: GithubSearchTranslations;
  history: HistoryTranslations;
  importExport: ImportExportTranslations;
  input: InputTranslations;
  kanban: KanbanTranslations;
  landing: LandingTranslations;
  notifications: NotificationsTranslations;
  profile: ProfileTranslations;
  settings: SettingsTranslations;
  tabs: TabsTranslations;
  tokenUsage: TokenUsageTranslations;
}

export type SupportedLocale = 'en-US' | 'pt-BR';
export type TranslationNamespace = keyof LocaleTranslations;