import * as React from 'react';
import { useState } from 'react';
import { motion } from 'framer-motion';
import { Lightbulb, Send, ThumbsUp, ThumbsDown, KanbanSquare } from 'lucide-react';
import { useProjectContext } from '../../contexts/ProjectContext';

const SuggestionsDisplay: React.FC = () => {
    const { 
        suggestedQuestions, 
        handleSendMessage, 
        isChatLoading, 
        currentAnalysis,
        kanbanState,
        handleCreateKanbanBoard 
    } = useProjectContext();
    const [feedback, setFeedback] = useState<'good' | 'bad' | null>(null);

    const hasSuggestions = suggestedQuestions.length > 0;
    const canCreateKanban = currentAnalysis?.suggestedKanbanTasks && !kanbanState;

    if (!hasSuggestions && !canCreateKanban) {
        return null;
    }

    const handleSuggestionClick = (question: string) => {
        if (!isChatLoading) {
            handleSendMessage(question);
        }
    };

    return (
        <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.5, duration: 0.5 }}
            className="p-6 bg-gray-800/50 border border-gray-700 rounded-xl"
        >
            <div className="flex items-start gap-4">
                <div className="shrink-0 w-10 h-10 rounded-full bg-purple-900/50 flex items-center justify-center border border-purple-800">
                    <Lightbulb className="w-5 h-5 text-purple-400" />
                </div>
                <div className="flex-grow">
                    <h3 className="text-lg font-semibold text-white">Next Steps & Suggestions</h3>
                    
                    {canCreateKanban && (
                        <div className="mt-4 p-4 bg-gray-900/50 border border-gray-700 rounded-lg">
                            <p className="text-sm text-gray-300 mb-3">The AI has identified actionable tasks from this analysis. Create a Kanban board to start tracking them.</p>
                            <button 
                                onClick={handleCreateKanbanBoard}
                                className="w-full sm:w-auto flex items-center justify-center gap-2 px-4 py-2 text-sm font-semibold text-white bg-purple-600 rounded-md hover:bg-purple-700"
                            >
                                <KanbanSquare className="w-4 h-4" /> Create Kanban Board
                            </button>
                        </div>
                    )}
                    
                    {hasSuggestions && (
                        <div className="mt-4">
                            <p className="text-sm text-gray-400 mb-4">Not sure what to ask? Here are some ideas to get the conversation started:</p>
                            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                                {suggestedQuestions.map((q, i) => (
                                    <button
                                        key={i}
                                        onClick={() => handleSuggestionClick(q)}
                                        disabled={isChatLoading}
                                        className="group flex items-center justify-between text-left p-3 bg-gray-900/50 border border-gray-700 rounded-lg hover:bg-gray-700/80 transition-colors disabled:opacity-60 disabled:cursor-not-allowed"
                                    >
                                        <span className="text-sm text-gray-300">{q}</span>
                                        <Send className="w-4 h-4 text-gray-500 group-hover:text-white transition-colors" />
                                    </button>
                                ))}
                            </div>
                            <div className="mt-4 flex items-center justify-end gap-2">
                                <p className="text-xs text-gray-500">Were these suggestions helpful?</p>
                                <button
                                    onClick={() => setFeedback('good')}
                                    className={`p-1.5 rounded-full transition-colors ${feedback === 'good' ? 'bg-green-500/30 text-green-400' : 'text-gray-400 hover:bg-gray-700'}`}
                                    aria-label="Like"
                                >
                                    <ThumbsUp className="w-4 h-4" />
                                </button>
                                <button
                                    onClick={() => setFeedback('bad')}
                                    className={`p-1.5 rounded-full transition-colors ${feedback === 'bad' ? 'bg-red-500/30 text-red-400' : 'text-gray-400 hover:bg-gray-700'}`}
                                    aria-label="Dislike"
                                >
                                    <ThumbsDown className="w-4 h-4" />
                                </button>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </motion.div>
    );
};

export default SuggestionsDisplay;