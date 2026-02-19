// FIX: Added full content for contexts/ProjectContext.tsx to resolve module errors.
import * as React from 'react';

import {
  createContext, ReactNode,
  useCallback, // ===== CHAT MANAGEMENT =====, [currentAnalysis, userSettings.userApiKey, addNotification]);, useContext, useEffect, useState } from 'react';
  useContext, useEffect, useState
} from 'react';
import { v4 as uuidv4 } from 'uuid';
import { exampleProject } from '../data/exampleAnalysis';
import { usePersistentState } from '../hooks/usePersistentState';
import { getAllProjects, setProject } from '../lib/idb';
import {
  analyzeProject,
  compareAnalyses,
  createChat,
  generateDashboardInsight,
  generateSelfCritique,
} from '../services/gemini/api';
import {
  AnalysisType,
  DashboardInsight,
  EvolutionAnalysis,
  HistoryItem,
  KanbanCard,
  KanbanState,
  KanbanTaskSuggestion,
  Project,
  ProjectAnalysis,
  ViewType
} from '../types';
import { useNotification } from './NotificationContext';
import { useUser } from './UserContext';
// FIX: Replaced deprecated ChatMessage with Content and Chat
import { Chat, Content } from '@google/genai';

interface ProjectContextType {
  projects: Project[];
  activeProjectId: string | null;
  setActiveProjectId: (id: string | null) => void;
  activeProject: Project | null;
  isExample: boolean;
  currentView: ViewType;
  setCurrentView: (view: ViewType) => void;
  isAnalyzing: boolean;
  isChatLoading: boolean;
  isHistoryPanelOpen: boolean;
  setIsHistoryPanelOpen: (isOpen: boolean) => void;

  // Analysis and data
  currentAnalysis: ProjectAnalysis | null;
  activeHistoryId: number | null;
  evolutionAnalysis: EvolutionAnalysis | null;
  kanbanState: KanbanState | null;
  setKanbanState: (state: KanbanState) => void;

  // Chat
  currentChatHistory: Content[];
  suggestedQuestions: string[];

  // Actions
  handleAnalyze: (projectName: string, context: string, analysisType: AnalysisType) => Promise<void>;
  handleSendMessage: (message: string) => Promise<void>;
  handleSelectHistoryItem: (id: number) => void;
  handleCompareHistoryItems: (id1: number, id2: number) => Promise<void>;
  handleDeleteHistoryItem: (id: number) => Promise<void>;
  handleCreateKanbanBoard: () => void;
  handleClearHistory: () => void;
  handleImportHistory: (data: any) => Promise<void>;
  handleExportHistory: (file: File) => Promise<void>;

  // Dashboard
  dashboardInsight: DashboardInsight | null;
  isInsightLoading: boolean;
}

const handleExportHistory = async (file: File): Promise<void> => {
  const text = await file.text();
  try {
    const data = JSON.parse(text);
    if (!data.id || !data.name) {
      throw new Error("Invalid project data.");
    }
    const project: Project = data;
    // Trigger file download
  } catch (error) {
    console.error("Failed to import history:", error);
    throw new Error("Invalid file format.");
  }
};

const handleImportHistory = async (data: any): Promise<void> => {
  try {
    const project: Project = data;
    if (!project.id || !project.name) {
      throw new Error("Invalid project data.");
    }
    // Save to IndexedDB
    await setProject(project);
  } catch (error) {
    console.error("Failed to import history:", error);
    throw new Error("Invalid file format.");
  }
};

const ProjectContext = createContext<ProjectContextType | undefined>(undefined);

export const ProjectContextProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const { addNotification } = useNotification();
  const {
    userSettings,
    usageTracking,
    incrementTokenUsage,
    canUseTokens,
    name: userName,
    email: userEmail,
    getUserTrackingMetadata
  } = useUser();

  // ===== STATE MANAGEMENT =====
  const [userProfile, setUserProfile] = useState<{ name: string }>({ name: userName || 'User' });
  const [projects, setProjects] = useState<Project[]>([]);
  const [activeProjectId, setActiveProjectId] = usePersistentState<string | null>('activeProjectId', null);

  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const [isChatLoading, setIsChatLoading] = useState(false);
  const [isHistoryPanelOpen, setIsHistoryPanelOpen] = useState(false);

  const [currentView, setCurrentView] = usePersistentState<ViewType>('currentView', ViewType.Dashboard);
  const [activeHistoryId, setActiveHistoryId] = useState<number | null>(null);
  const [chatInstance, setChatInstance] = useState<Chat | null>(null);

  const [dashboardInsight, setDashboardInsight] = useState<DashboardInsight | null>(null);
  const [isInsightLoading, setIsInsightLoading] = useState(false);

  // ===== DERIVED STATE =====
  const activeProject = projects.find(p => p.id === activeProjectId) ?? null;
  const isExample = activeProject?.id === exampleProject.id;

  const currentHistoryItem = activeProject?.history.find(h => h.id === activeHistoryId) ?? activeProject?.history[activeProject.history.length - 1] ?? null;
  const currentAnalysis = currentHistoryItem?.analysis ?? null;
  const currentChatHistory = (activeProject && activeHistoryId && activeProject.chatHistories[activeHistoryId]) || (currentAnalysis ? [{ role: 'model', parts: [{ text: "Hello! Ask me anything about this analysis." }] }] : []);
  const evolutionAnalysis = activeHistoryId === -1 ? (currentAnalysis as unknown as EvolutionAnalysis) : null;
  const suggestedQuestions = currentAnalysis?.suggestedQuestions || [];
  const kanbanState = activeProject?.kanban ?? null;

  // ===== DATA FETCHING & PERSISTENCE =====
  useEffect(() => {
    const loadProjects = async () => {
      const storedProjects = await getAllProjects();
      if (storedProjects.length === 0) {
        setProjects([exampleProject]);
      } else {
        setProjects([exampleProject, ...storedProjects]);
      }
    };
    loadProjects();
  }, []);

  const updateProject = useCallback(async (updatedProject: Project) => {
    if (isExample) return;
    const newProjects = projects.map(p => p.id === updatedProject.id ? updatedProject : p);
    setProjects(newProjects);
    await setProject(updatedProject);
  }, [projects, isExample]);

  // ===== CHAT MANAGEMENT =====
  useEffect(() => {
    if (currentAnalysis && userSettings.userApiKey) {
      try {
        const newChat = createChat(userSettings.userApiKey, currentAnalysis);
        setChatInstance(newChat);
      } catch (error: any) {
        addNotification({ message: error.message, type: 'error' });
      }
    }
  }, [currentAnalysis, userSettings.userApiKey, addNotification]);

  // ===== DASHBOARD INSIGHTS =====
  useEffect(() => {
    const fetchInsight = async () => {
      if (currentView === ViewType.Dashboard && userSettings.enableDashboardInsights && userSettings.userApiKey && projects.length > 1) {
        setIsInsightLoading(true);
        try {
          const userProjects = projects.filter((p: Project) => p.id !== exampleProject.id);
          const recentHistory = userProjects.flatMap((p: Project) => p.history).sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()).slice(0, 5);
          if (recentHistory.length > 0) {
            // Check token usage limits
            if (!canUseTokens(500)) {
              setDashboardInsight(null);
              setIsInsightLoading(false);
              return;
            }

            const metadata = getUserTrackingMetadata();

            // Criar um UserProfile a partir dos dados do usuário
            const userProfile = {
              name: userName || 'User',
              email: userEmail || '',
              preferences: userSettings,
            };

            const insight = await generateDashboardInsight(userProfile, recentHistory, userSettings.userApiKey);

            setDashboardInsight(insight);
          }
        } catch (error: any) {
          // Fail silently, it's not a critical feature
          console.warn("Could not generate dashboard insight:", error.message);
          setDashboardInsight(null);
        } finally {
          setIsInsightLoading(false);
        }
      }
    };
    fetchInsight();
  }, [currentView, projects, userSettings.enableDashboardInsights, userSettings.userApiKey]);

  // ===== ACTIONS / HANDLERS =====
  const handleAnalyze = async (projectName: string, context: string, analysisType: AnalysisType) => {
    if (!userSettings.userApiKey) {
      addNotification({ message: 'Please set your Gemini API key in the settings.', type: 'error' });
      return;
    }

    setIsAnalyzing(true);
    try {
      let projectToUpdate: Project | null = activeProject;
      if (!projectToUpdate) {
        const newProject: Project = {
          id: uuidv4(), name: projectName, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString(),
          history: [], kanban: null, chatHistories: {}, contextFiles: []
        };
        setProjects(prev => [...prev, newProject]);
        setActiveProjectId(newProject.id);
        projectToUpdate = newProject;
        await setProject(newProject);
      }

      let analysisResult: ProjectAnalysis;

      if (analysisType === AnalysisType.SelfCritique) {
        if (!currentAnalysis) throw new Error("No analysis available to critique.");
        const critiqueResult = await generateSelfCritique(currentAnalysis, userSettings.userApiKey);
        const updatedProject = {
          ...projectToUpdate,
          critiques: { ...projectToUpdate.critiques, [currentHistoryItem!.id]: critiqueResult }
        };
        await updateProject(updatedProject);
        addNotification({ message: 'Self-critique completed successfully!', type: 'success' });
        setCurrentView(ViewType.Analysis); // Stay on the analysis view to see the critique button
        return; // Exit early
      } else {
        analysisResult = await analyzeProject(context, analysisType, userSettings.userApiKey);
      }

      const newHistoryItem: HistoryItem = {
        id: Date.now(),
        timestamp: new Date().toISOString(),
        analysis: analysisResult,
      };

      const updatedProject = {
        ...projectToUpdate,
        history: userSettings.saveHistory ? [...projectToUpdate.history, newHistoryItem] : [newHistoryItem],
        chatHistories: { ...projectToUpdate.chatHistories, [newHistoryItem.id]: [] },
        updatedAt: new Date().toISOString(),
      };

      await updateProject(updatedProject);
      setActiveHistoryId(newHistoryItem.id);
      setCurrentView(ViewType.Analysis);
      addNotification({ message: 'Analysis complete!', type: 'success' });
    } catch (error: any) {
      addNotification({ message: error.message, type: 'error' });
    } finally {
      setIsAnalyzing(false);
    }
  };

  const handleSendMessage = async (message: string) => {
    if (!chatInstance || !activeProject || !activeHistoryId) return;

    const userMessage: Content = { role: 'user', parts: [{ text: message }] };
    const currentHistory = activeProject.chatHistories[activeHistoryId] || [];
    const updatedHistory = [...currentHistory, userMessage];

    // Optimistically update UI
    const optimisticallyUpdatedProject = { ...activeProject, chatHistories: { ...activeProject.chatHistories, [activeHistoryId]: updatedHistory } };
    updateProject(optimisticallyUpdatedProject);

    setIsChatLoading(true);
    try {
      const result = await chatInstance.sendMessage({ message });
      const modelMessage: Content = { role: 'model', parts: [{ text: result.text }] };
      const finalHistory = [...updatedHistory, modelMessage];

      const finalUpdatedProject = { ...activeProject, chatHistories: { ...activeProject.chatHistories, [activeHistoryId]: finalHistory } };
      updateProject(finalUpdatedProject);

    } catch (error: any) {
      addNotification({ message: `Chat error: ${error.message}`, type: 'error' });
      // Revert optimistic update on error
      updateProject(activeProject);
    } finally {
      setIsChatLoading(false);
    }
  };

  const handleSelectHistoryItem = (id: number) => {
    setActiveHistoryId(id);
    setCurrentView(ViewType.Analysis);
    setIsHistoryPanelOpen(false);
  };

  const handleCompareHistoryItems = async (id1: number, id2: number) => {
    if (!activeProject) return;
    const item1 = activeProject.history.find(h => h.id === id1);
    const item2 = activeProject.history.find(h => h.id === id2);

    if (!item1 || !item2) {
      addNotification({ message: "Could not find selected history items.", type: 'error' });
      return;
    }

    setIsAnalyzing(true);
    setIsHistoryPanelOpen(false);
    try {
      const [previous, current] = [item1, item2].sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
      const evolutionResult = await compareAnalyses(previous.analysis, current.analysis, userSettings.userApiKey || '');

      // This is a bit of a hack: we create a temporary history item to display the evolution.
      const evolutionHistoryItem: HistoryItem = {
        id: -1, // Special ID to signify comparison
        timestamp: new Date().toISOString(),
        analysis: evolutionResult as unknown as ProjectAnalysis,
      };

      const updatedProject = {
        ...activeProject,
        history: [...activeProject.history, evolutionHistoryItem]
      };
      // We don't save this temporary item to DB
      setProjects(projects.map(p => p.id === updatedProject.id ? updatedProject : p));

      setActiveHistoryId(evolutionHistoryItem.id);
      setCurrentView(ViewType.Evolution);
      addNotification({ message: 'Comparison complete!', type: 'success' });
    } catch (error: any) {
      addNotification({ message: error.message, type: 'error' });
    } finally {
      setIsAnalyzing(false);
    }
  };

  const handleDeleteHistoryItem = async (id: number) => {
    if (!activeProject || isExample) return;
    const updatedHistory = activeProject.history.filter(h => h.id !== id);
    const updatedChatHistories = { ...activeProject.chatHistories };
    delete updatedChatHistories[id];
    const updatedProject = { ...activeProject, history: updatedHistory, chatHistories: updatedChatHistories };
    await updateProject(updatedProject);
    addNotification({ message: "History item deleted.", type: 'success' });
  };

  const handleCreateKanbanBoard = () => {
    if (!activeProject || !currentAnalysis?.suggestedKanbanTasks || isExample) return;

    const newCards: Record<string, KanbanCard> = {};
    const cardIds: string[] = [];

    currentAnalysis.suggestedKanbanTasks.forEach((task: KanbanTaskSuggestion) => {
      const id = uuidv4();
      newCards[id] = { id, ...task };
      cardIds.push(id);
    });

    const newKanbanState: KanbanState = {
      cards: newCards,
      columns: {
        backlog: { id: 'backlog', title: 'Backlog', cardIds: cardIds },
        todo: { id: 'todo', title: 'To Do', cardIds: [] },
        inProgress: { id: 'inProgress', title: 'In Progress', cardIds: [] },
        done: { id: 'done', title: 'Done', cardIds: [] },
      },
      columnOrder: ['backlog', 'todo', 'inProgress', 'done'],
    };

    const updatedProject = { ...activeProject, kanban: newKanbanState };
    updateProject(updatedProject);
    setCurrentView(ViewType.Kanban);
    addNotification({ message: "Kanban board created successfully!", type: 'success' });
  };

  const setKanbanState = (state: KanbanState) => {
    if (!activeProject || isExample) return;
    const updatedProject = { ...activeProject, kanban: state };
    updateProject(updatedProject);
  }

  const handleClearHistory = () => {
    if (!activeProject || isExample) return;
    const updatedProject = { ...activeProject, history: [], chatHistories: {} };
    updateProject(updatedProject);
  };

  // Clean up comparison analysis when view changes
  useEffect(() => {
    if (currentView !== ViewType.Evolution && activeProject?.history.some(h => h.id === -1)) {
      const cleanedHistory = activeProject.history.filter(h => h.id !== -1);
      const updatedProject = { ...activeProject, history: cleanedHistory };
      setProjects(projects.map(p => p.id === updatedProject.id ? updatedProject : p));
      // Reset to latest analysis
      setActiveHistoryId(cleanedHistory[cleanedHistory.length - 1]?.id ?? null);
    }
  }, [currentView, activeProject, projects]);


  const value: ProjectContextType = {
    projects,
    activeProjectId,
    setActiveProjectId,
    activeProject,
    isExample,
    currentView,
    setCurrentView,
    isAnalyzing,
    isChatLoading,
    isHistoryPanelOpen,
    setIsHistoryPanelOpen,
    currentAnalysis,
    activeHistoryId,
    evolutionAnalysis,
    kanbanState,
    setKanbanState,
    currentChatHistory,
    suggestedQuestions,
    handleImportHistory,
    handleExportHistory,
    handleAnalyze,
    handleSendMessage,
    handleSelectHistoryItem,
    handleCompareHistoryItems,
    handleDeleteHistoryItem,
    handleCreateKanbanBoard,
    handleClearHistory,
    dashboardInsight,
    isInsightLoading
  };

  return <ProjectContext.Provider value={value}>{children}</ProjectContext.Provider>;
};

export const useProjectContext = (): ProjectContextType => {
  const context = useContext(ProjectContext);
  if (context === undefined) {
    throw new Error('useProjectContext must be used within a ProjectContextProvider');
  }
  return context;
};
