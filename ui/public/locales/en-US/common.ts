import { TranslationMessages } from '../types';

export const commonEnUS: TranslationMessages = {
  header: {
    title: "GemX Analyzer",
    subtitle: "Transform your project documentation into actionable insights with AI-driven analysis."
  },
  navigation: {
    dashboard: "Dashboard",
    newAnalysis: "New Analysis",
    currentAnalysis: "Current Analysis",
    kanban: "Kanban",
    history: "History",
    chat: "Chat",
    // FIX: Added missing evolution key
    evolution: "Evolution"
  },
  actions: {
    analyzeProject: "Analyze Project",
    analyzing: "Analyzing",
    uploadFile: "Upload File",
    showExample: "Show me an example",
    exitExample: "Exit Example Mode",
    load: "Load",
    showMore: "Show More",
    view: "View",
    createKanbanBoard: "Create Kanban Board",
    viewKanbanBoard: "View Kanban Board"
  },
  common: {
    title: "Title",
    description: "Description",
    priorityLabel: "Priority",
    difficultyLabel: "Difficulty",
    delete: "Delete",
    save: "Save",
    cancel: "Cancel",
    confirm: "Confirm",
    connect: "Connect",
    notConnected: "Not Connected"
  },
  priority: {
    Low: "Low",
    Medium: "Medium",
    High: "High"
  },
  difficulty: {
    Low: "Low",
    Medium: "Medium",
    High: "High"
  },
  effort: {
    Low: "Low",
    Medium: "Medium",
    High: "High"
  },
  status: {
    TODO: "TODO",
    InProgress: "In Progress",
    Done: "Done",
    Blocked: "Blocked"
  },
  loader: {
    loading: "Loading",
    uploading: "Uploading",
    analyzing: "Analyzing",
    saving: "Saving",
    // FIX: Added missing loader keys
    message: "Analyzing your project...",
    subMessage: "This may take a few moments.",
    ariaLabel: "Analyzing content, please wait.",
    steps: [
      "Parsing file structure...",
      "Evaluating architecture...",
      "Checking code quality...",
      "Identifying potential improvements...",
      "Compiling the report..."
    ]
  },
  feedback: {
    success: "Success",
    error: "Error",
    warning: "Warning",
    info: "Information"
  },
  network: {
    connecting: "Connecting",
    connected: "Connected",
    disconnected: "Disconnected",
    connectionError: "Connection Error",
    retry: "Retry",
    online: "Online",
    // FIX: Added missing offline key
    offline: "You are offline"
  },
  tokenUsage: {
    title: "Token Usage Warning",
    inputTokens: "Input Tokens",
    outputTokens: "Output Tokens",
    totalTokens: "Total Tokens",
    estimatedCost: "Estimated Cost"
  },
  settings: {
    language: "Language",
    theme: "Theme",
    appearance: "Appearance"
  },
  showExample: "Show me an example",
  analysisTitle: "Analysis Title",
  save: "Save"
};