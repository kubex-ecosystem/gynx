import { ProjectAnalysis, SelfCritiqueAnalysis } from "./Analysis";
import { KanbanState } from "./Kanban";

// FIX: Added import from google/genai for Content type
import { Content } from "@google/genai";

// Project & History Types
export interface HistoryItem {
  id: number;
  timestamp: string;
  analysis: ProjectAnalysis;
}

export interface Project {
  id: string;
  name: string;
  createdAt: string;
  updatedAt: string;
  history: HistoryItem[];
  kanban: KanbanState | null;
  chatHistories: Record<number, Content[]>; // key is history item ID
  critiques?: Record<number, SelfCritiqueAnalysis>; // key is history item ID
  contextFiles: string[];
}

export interface NewProject {
  name: string;
}

export interface UpdateProject {
  name?: string;
  contextFiles?: string[];
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

export interface ContextFileState {
  files: ContextFile[];
  isLoading: boolean;
  error: string | null;
}

export interface NewContextFile {
  name: string;
  content: string;
  isFragment?: boolean;
}

export interface UpdateContextFile {
  id: string;
  name?: string;
  content?: string;
  isFragment?: boolean;
}

export interface DeleteContextFile {
  id: string;
}

export interface ReorderContextFiles {
  sourceIndex: number;
  destinationIndex: number;
}

