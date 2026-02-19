// FIX: Added full content for services/gemini/prompts.ts to resolve module errors.
import { ProjectAnalysis, AnalysisType, HistoryItem, UserProfile } from '../../types';

const analysisPromptSystemInstruction = `You are a world-class senior software architect and project manager. Your task is to analyze project documentation provided by the user and generate a comprehensive, structured JSON response. Adhere strictly to the provided JSON schema. Be insightful, critical, and provide actionable advice. For the architecture diagram, you MUST use MermaidJS 'graph TD' syntax.`;

export const getAnalysisPrompt = (projectContext: string, analysisType: AnalysisType): string => {
  return `
    Project Context:
    ---
    ${projectContext}
    ---
    Analysis Request:
    Please perform a deep analysis of the provided project context. 
    Focus specifically on **${analysisType}**. 
    
    Based on your analysis, provide a detailed response in JSON format.
    - For the architecture diagram, generate valid MermaidJS graph TD syntax. If no architecture is described, return an empty string.
    - The 'suggestedQuestions' should be insightful follow-ups a user might ask.
    - The 'suggestedKanbanTasks' should be directly actionable items derived from the 'improvements' section.
  `;
};

export const getEvolutionAnalysisPrompt = (
    previousAnalysis: ProjectAnalysis,
    currentAnalysis: ProjectAnalysis
): string => {
    return `
    Here are two analyses of the same project, "${currentAnalysis.projectName}", taken at different times.
    
    PREVIOUS ANALYSIS:
    ---
    ${JSON.stringify(previousAnalysis, null, 2)}
    ---
    
    CURRENT ANALYSIS:
    ---
    ${JSON.stringify(currentAnalysis, null, 2)}
    ---
    
    Please provide an evolution analysis comparing these two snapshots. Identify which improvements were resolved, which are new, and which persist. Summarize the overall evolution of the project.
    `;
};

export const getSelfCritiquePrompt = (analysis: ProjectAnalysis): string => `
    Here is a project analysis you previously generated:
    ---
    ${JSON.stringify(analysis, null, 2)}
    ---
    Please perform a self-critique of this analysis. Evaluate its quality, depth, and helpfulness. 
    - How confident are you in its accuracy?
    - What did you do well?
    - What could have been improved or made more specific?
    Provide your critique in the specified JSON format.
`;

export const getDashboardInsightPrompt = (
    userProfile: UserProfile,
    recentHistory: HistoryItem[]
): string => {
    const historySummary = recentHistory.map(h => ({
        type: h.analysis.analysisType,
        score: h.analysis.viability.score,
        date: h.timestamp,
        strengths: h.analysis.strengths.length,
        improvements: h.analysis.improvements.length
    }));

    return `
    You are an AI assistant for a software analysis tool. Your goal is to provide a brief, personalized, and encouraging insight for the user on their dashboard.

    User Profile:
    Name: ${userProfile.name}

    Recent Activity (last 5 analyses):
    ${JSON.stringify(historySummary, null, 2)}

    Based on this data, generate a single, concise insight. It could be a trend you notice, a suggestion for a different analysis type, or a comment on their progress.
    Keep it short and engaging. Address the user by their name.
    
    Example: "Hi ${userProfile.name}, great job on improving the viability score on your last project! Maybe try a Security Analysis next to cover all your bases."
    
    Generate the insight in the specified JSON format.
    `;
};

export const getChatPrompt = (
    analysisContext: ProjectAnalysis,
): string => {
    return `You are a helpful AI assistant specialized in analyzing the provided software project analysis. Your knowledge is strictly limited to the following JSON data. Do not invent information. Answer the user's questions based only on this context. Be concise and helpful.

    Analysis Context:
    ---
    ${JSON.stringify(analysisContext, null, 2)}
    ---
    `;
};

export { analysisPromptSystemInstruction };