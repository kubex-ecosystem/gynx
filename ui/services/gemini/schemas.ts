// FIX: Added full content for services/gemini/schemas.ts to resolve module errors.
import { Type } from '@google/genai';

const improvementSchema = {
  type: Type.OBJECT,
  properties: {
    title: { type: Type.STRING, description: 'A concise title for the improvement area.' },
    description: { type: Type.STRING, description: 'A detailed explanation of the issue and why it needs improvement.' },
    priority: { type: Type.STRING, enum: ['Low', 'Medium', 'High'], description: 'The priority of the improvement.' },
    difficulty: { type: Type.STRING, enum: ['Low', 'Medium', 'High'], description: 'The estimated difficulty to implement the improvement.' },
    businessImpact: { type: Type.STRING, description: 'How this improvement impacts the business goals.' }
  },
  required: ['title', 'description', 'priority', 'difficulty', 'businessImpact']
};

const nextStepSchema = {
  type: Type.OBJECT,
  properties: {
    title: { type: Type.STRING, description: 'A concise title for the next step.' },
    description: { type: Type.STRING, description: 'A brief description of what this step entails.' },
    difficulty: { type: Type.STRING, enum: ['Low', 'Medium', 'High'], description: 'The estimated difficulty of this step.' },
  },
  required: ['title', 'description', 'difficulty']
};

const kanbanTaskSuggestionSchema = {
    type: Type.OBJECT,
    properties: {
        title: { type: Type.STRING, description: 'A short, actionable title for the Kanban card.' },
        description: { type: Type.STRING, description: 'A detailed description for the card, derived from an improvement or next step.' },
        priority: { type: Type.STRING, enum: ['Low', 'Medium', 'High'] },
        difficulty: { type: Type.STRING, enum: ['Low', 'Medium', 'High'] },
        tags: { type: Type.ARRAY, items: { type: Type.STRING }, description: 'Relevant tags like "security", "refactor", "documentation".' }
    },
    required: ['title', 'description', 'priority', 'difficulty', 'tags']
};

export const ProjectAnalysisSchema = {
  type: Type.OBJECT,
  properties: {
    projectName: { type: Type.STRING, description: 'The name of the project being analyzed.' },
    analysisType: { type: Type.STRING, enum: ['Architecture', 'Code Quality', 'Security Analysis', 'Scalability Analysis', 'Compliance & Best Practices', 'Documentation Review'], description: 'The type of analysis performed.' },
    summary: { type: Type.STRING, description: 'A high-level executive summary of the analysis findings.' },
    strengths: { type: Type.ARRAY, items: { type: Type.STRING }, description: 'A list of key strengths of the project.' },
    improvements: { type: Type.ARRAY, items: improvementSchema, description: 'A list of suggested improvements.' },
    nextSteps: {
      type: Type.OBJECT,
      properties: {
        shortTerm: { type: Type.ARRAY, items: nextStepSchema },
        longTerm: { type: Type.ARRAY, items: nextStepSchema }
      },
      required: ['shortTerm', 'longTerm']
    },
    viability: {
      type: Type.OBJECT,
      properties: {
        score: { type: Type.INTEGER, description: 'A project viability score from 1 to 10.' },
        assessment: { type: Type.STRING, description: 'A brief justification for the viability score.' }
      },
      required: ['score', 'assessment']
    },
    roiAnalysis: {
      type: Type.OBJECT,
      properties: {
        assessment: { type: Type.STRING, description: 'Assessment of the potential Return on Investment.' },
        potentialGains: { type: Type.ARRAY, items: { type: Type.STRING }, description: 'List of potential gains from implementing improvements.' },
        estimatedEffort: { type: Type.STRING, enum: ['Low', 'Medium', 'High'], description: 'Overall estimated effort for improvements.' }
      },
      required: ['assessment', 'potentialGains', 'estimatedEffort']
    },
    maturity: {
      type: Type.OBJECT,
      properties: {
        level: { type: Type.STRING, enum: ['Prototype', 'MVP', 'Production', 'Optimized'], description: 'The project\'s current maturity level.' },
        assessment: { type: Type.STRING, description: 'Justification for the maturity level assessment.' }
      },
      required: ['level', 'assessment']
    },
    architectureDiagram: { type: Type.STRING, description: 'A MermaidJS graph TD syntax for the project architecture. Can be empty string if not applicable.' },
    suggestedQuestions: { type: Type.ARRAY, items: { type: Type.STRING }, description: 'A list of 3-4 relevant follow-up questions a user might ask about the analysis.' },
    suggestedKanbanTasks: { type: Type.ARRAY, items: kanbanTaskSuggestionSchema, description: 'Actionable tasks derived from improvements, formatted for a Kanban board.' }
  },
  required: ['projectName', 'analysisType', 'summary', 'strengths', 'improvements', 'nextSteps', 'viability', 'roiAnalysis', 'maturity', 'architectureDiagram', 'suggestedQuestions', 'suggestedKanbanTasks']
};


export const EvolutionAnalysisSchema = {
    type: Type.OBJECT,
    properties: {
        projectName: { type: Type.STRING },
        analysisType: { type: Type.STRING },
        evolutionSummary: { type: Type.STRING, description: "A summary comparing the two analyses, highlighting progress and new issues." },
        keyMetrics: {
            type: Type.OBJECT,
            properties: {
                previousScore: { type: Type.INTEGER },
                currentScore: { type: Type.INTEGER },
                previousStrengths: { type: Type.INTEGER },
                currentStrengths: { type: Type.INTEGER },
                previousImprovements: { type: Type.INTEGER },
                currentImprovements: { type: Type.INTEGER },
            },
            required: ['previousScore', 'currentScore', 'previousStrengths', 'currentStrengths', 'previousImprovements', 'currentImprovements']
        },
        resolvedImprovements: { type: Type.ARRAY, items: improvementSchema, description: "List of improvements from the previous analysis that are no longer present in the current one." },
        newImprovements: { type: Type.ARRAY, items: improvementSchema, description: "List of improvements present in the current analysis but not in the previous one." },
        persistentImprovements: { type: Type.ARRAY, items: improvementSchema, description: "List of improvements present in both analyses." }
    },
    required: ['projectName', 'analysisType', 'evolutionSummary', 'keyMetrics', 'resolvedImprovements', 'newImprovements', 'persistentImprovements']
};

export const SelfCritiqueSchema = {
    type: Type.OBJECT,
    properties: {
        confidenceScore: { type: Type.INTEGER, description: 'A score from 1-10 on how confident the AI is in the quality and accuracy of its own previous analysis.' },
        overallAssessment: { type: Type.STRING, description: 'A brief summary of the critique.' },
        positivePoints: { type: Type.ARRAY, items: { type: Type.STRING }, description: 'What the AI did well in the previous analysis.' },
        areasForRefinement: { type: Type.ARRAY, items: { type: Type.STRING }, description: 'Where the AI could have done better or been more detailed.' }
    },
    required: ['confidenceScore', 'overallAssessment', 'positivePoints', 'areasForRefinement']
};

export const DashboardInsightSchema = {
    type: Type.OBJECT,
    properties: {
        title: { type: Type.STRING, description: "A catchy, personalized title for the insight." },
        summary: { type: Type.STRING, description: "A 1-2 sentence summary of an interesting pattern or suggestion based on the user's recent activity." }
    },
    required: ['title', 'summary']
};