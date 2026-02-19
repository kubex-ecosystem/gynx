// services/gemini/api.ts
// FIX: Added full content for services/gemini/api.ts to resolve module errors.

import { GoogleGenAI } from "@google/genai";
import {
  AnalysisType,
  DashboardInsight,
  EvolutionAnalysis,
  HistoryItem,
  ProjectAnalysis,
  SelfCritiqueAnalysis,
  UserProfile,
} from '../../types';

import {
  analysisPromptSystemInstruction,
  getAnalysisPrompt,
  getChatPrompt,
  getDashboardInsightPrompt,
  getEvolutionAnalysisPrompt,
  getSelfCritiquePrompt,
} from './prompts';

import {
  DashboardInsightSchema,
  EvolutionAnalysisSchema,
  ProjectAnalysisSchema,
  SelfCritiqueSchema,
} from './schemas';
import { handleGeminiError } from './utils';

// This function should be called ONLY ONCE to initialize the API.
const getGenAI = (apiKey?: string) => {
  const key = apiKey || process.env.API_KEY;
  if (!key) {
    const error = new Error("API_KEY_EMPTY");
    handleGeminiError(error);
  }
  return new GoogleGenAI({ apiKey: key });
};

type GenerateContentRequest = {
  model: string;
  // Removida a opção de string no array para alinhar com o tipo esperado pelo SDK.
  contents: Array<{ role?: string; parts?: Array<{ text: string }> }>;
  config?: {
    responseMimeType?: string;
    responseSchema?: any;
    systemInstruction?: string;
  };
};

const callGemini = async (
  apiKey: string,
  prompt: string,
  schema: object,
  systemInstruction?: string
): Promise<ProjectAnalysis | EvolutionAnalysis | SelfCritiqueAnalysis | DashboardInsight> => {
  try {
    const ai = getGenAI(apiKey);
    const request: GenerateContentRequest = {
      model: 'gemini-2.5-flash',
      contents: [{ role: 'user', parts: [{ text: prompt }] }],
      config: {
        responseMimeType: 'application/json',
        responseSchema: schema,
      }
    };

    if (systemInstruction) {
      request.config!.systemInstruction = systemInstruction;
    }

    // Preserve the original returned value (it can be a string or an object).
    const result = await ai.models.generateContent(request);

    // Safely extract the response text. The SDK can return either a string or an object with `.text`.
    let rawText = "";
    if (typeof result === "string") {
      rawText = result;
    } else if (result && typeof (result as any).text === "string") {
      rawText = (result as any).text;
    }

    // Ensure JSON.parse always receives a string; fallback to empty object if nothing found.
    const parsed = JSON.parse(rawText || "{}") as any; // The schema ensures the type.

    // Add usage metadata to the response object when available and when result is an object.
    if (result && typeof result !== "string" && (result as any).usageMetadata) {
      const um = (result as any).usageMetadata;
      parsed.usageMetadata = {
        promptTokenCount: um.promptTokenCount,
        candidatesTokenCount: um.candidatesTokenCount,
        totalTokenCount: um.totalTokenCount,
      };
    }

    return parsed;
  } catch (error) {
    handleGeminiError(error);
    throw error; // Re-throw after handling
  }
};

export const analyzeProject = async (
  projectContext: string,
  analysisType: AnalysisType,
  apiKey: string
): Promise<ProjectAnalysis> => {
  const prompt = getAnalysisPrompt(projectContext, analysisType);
  return await (callGemini(apiKey, prompt, ProjectAnalysisSchema, analysisPromptSystemInstruction) as Promise<ProjectAnalysis> || {}) as ProjectAnalysis;
};

export const compareAnalyses = async (
  previous: ProjectAnalysis,
  current: ProjectAnalysis,
  apiKey: string
): Promise<EvolutionAnalysis> => {
  const prompt = getEvolutionAnalysisPrompt(previous, current);
  return await (callGemini(apiKey, prompt, EvolutionAnalysisSchema, analysisPromptSystemInstruction) as Promise<EvolutionAnalysis> || {}) as EvolutionAnalysis;
};

export const generateSelfCritique = async (
  analysis: ProjectAnalysis,
  apiKey: string
): Promise<SelfCritiqueAnalysis> => {
  const prompt = getSelfCritiquePrompt(analysis);
  return await (callGemini(apiKey, prompt, SelfCritiqueSchema, analysisPromptSystemInstruction) as Promise<SelfCritiqueAnalysis> || {}) as SelfCritiqueAnalysis;
};

export const generateDashboardInsight = async (
  userProfile: UserProfile,
  recentHistory: HistoryItem[],
  apiKey: string
): Promise<DashboardInsight> => {
  const prompt = getDashboardInsightPrompt(userProfile, recentHistory);
  return await (callGemini(apiKey, prompt, DashboardInsightSchema) as Promise<DashboardInsight> || {}) as DashboardInsight;
};


export const createChat = (apiKey: string, analysisContext: ProjectAnalysis) => {
  const ai = getGenAI(apiKey);
  const systemInstruction = getChatPrompt(analysisContext);
  const chat = ai.chats.create({
    model: 'gemini-2.5-flash',
    config: {
      systemInstruction
    }
  });
  return chat;
};


export const testApiKey = async (apiKey: string): Promise<boolean> => {
  try {
    const ai = getGenAI(apiKey);
    // Usar formato consistente de 'contents' semelhante ao restante do arquivo
    const request: GenerateContentRequest = {
      model: 'gemini-2.5-flash',
      contents: [{ role: 'user', parts: [{ text: 'test' }] }],
      config: {
        responseMimeType: 'application/json'
      }
    };
    await ai.models.generateContent(request);
    return true;
  } catch (error) {
    handleGeminiError(error);
    return false;
  }
};
