// Interfaces for Chat
export interface ChatMessage {
  role: 'user' | 'model';
  parts: { text: string }[];
}

export type AllChatHistories = Record<number, ChatMessage[]>;

export interface ChatState {
  histories: AllChatHistories;
  currentProjectId: number | null;
  isLoading: boolean;
  error: string | null;
}

export interface ChatRequestPayload {
  projectId: number;
  message: string;
  context: string;
  analysisType: string;
}

export interface ChatResponse {
  reply: string;
  updatedAnalysis?: any; // Replace 'any' with the actual type if available
}

export interface ChatError {
  message: string;
}

export interface ChatSettings {
  maxHistoryMessages: number;
  autoClearHistory: boolean;
}
