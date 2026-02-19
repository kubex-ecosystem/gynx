import * as React from 'react';

import { Info, Loader2, Send, Sparkles, User } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { useProjectContext } from '../../contexts/ProjectContext';

// FIX: Replaced deprecated ChatMessage with Content
import { Content } from '@google/genai';

const ChatMessageBubble: React.FC<{ message: Content }> = ({ message }) => {
  const isUser = message.role === 'user';
  return (
    <div className={`flex items-start gap-3 ${isUser ? 'justify-end' : ''}`}>
      {!isUser && (
        <div className="w-8 h-8 rounded-full bg-purple-900/50 flex items-center justify-center border border-purple-800 shrink-0">
          <Sparkles className="w-5 h-5 text-purple-400" />
        </div>
      )}
      <div
        className={`max-w-xl p-3 rounded-xl text-white ${isUser ? 'bg-blue-600 rounded-br-none' : 'bg-gray-700 rounded-bl-none'
          }`}
      >
        <p className="whitespace-pre-wrap">{(message.parts || [])[0].text}</p>
      </div>
      {isUser && (
        <div className="w-8 h-8 rounded-full bg-gray-600 flex items-center justify-center shrink-0">
          <User className="w-5 h-5 text-gray-300" />
        </div>
      )}
    </div>
  );
};

const ChatPanel: React.FC = () => {
  const {
    currentChatHistory,
    isChatLoading,
    handleSendMessage,
    suggestedQuestions,
    activeProject,
    isExample
  } = useProjectContext();
  const [input, setInput] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [currentChatHistory, isChatLoading]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (input.trim() && !isChatLoading) {
      handleSendMessage(input);
      setInput('');
    }
  };

  const handleSuggestionClick = (question: string) => {
    handleSendMessage(question);
  };

  if (!activeProject) {
    return (
      <div className="h-full flex flex-col items-center justify-center bg-gray-900/30 border border-gray-800 rounded-xl text-center">
        <h3 className="text-xl font-bold">No Project Loaded</h3>
        <p className="text-gray-400 mt-2">Create or select a project to start chatting.</p>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col bg-gray-900/30 border border-gray-800 rounded-xl">
      {/* Messages */}
      <div className="flex-grow p-4 overflow-y-auto space-y-4">
        {isExample && (
          <div className="p-3 bg-purple-900/50 border border-purple-700 text-purple-300 rounded-lg flex items-center gap-3 text-sm">
            <Info className="w-5 h-5 shrink-0" />
            <p>You are chatting in example mode. The conversation will not be saved.</p>
          </div>
        )}
        {currentChatHistory.map((msg, index) => (
          <ChatMessageBubble key={index} message={msg} />
        ))}
        {isChatLoading && currentChatHistory.length > 0 && currentChatHistory[currentChatHistory.length - 1].role === 'user' && (
          <div className="flex items-start gap-3">
            <div className="w-8 h-8 rounded-full bg-purple-900/50 flex items-center justify-center border border-purple-800 shrink-0">
              <Sparkles className="w-5 h-5 text-purple-400" />
            </div>
            <div className="max-w-xl p-3 rounded-xl text-white bg-gray-700 rounded-bl-none">
              <Loader2 className="w-5 h-5 animate-spin text-gray-400" />
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input Area */}
      <div className="p-4 border-t border-gray-700">
        {currentChatHistory.length <= (isExample ? 1 : 0) && suggestedQuestions.length > 0 && (
          <div className="mb-3 grid grid-cols-1 md:grid-cols-2 gap-2">
            {suggestedQuestions.map((q, i) => (
              <button key={i} onClick={() => handleSuggestionClick(q)}
                className="p-2 text-sm text-left bg-gray-800/60 border border-gray-700 rounded-lg hover:bg-gray-700/80 transition-colors"
              >
                {q}
              </button>
            ))}
          </div>
        )}

        <form onSubmit={handleSubmit} className="flex items-center gap-2">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Ask something about the analysis..."
            className="flex-grow p-2 bg-gray-900 border border-gray-600 rounded-md text-sm"
            disabled={isChatLoading}
          />
          <button title='Send Message' type="submit" disabled={isChatLoading || !input.trim()} className="p-2 bg-purple-600 text-white rounded-md hover:bg-purple-700 disabled:bg-gray-600">
            <Send className="w-5 h-5" />
          </button>
        </form>
      </div>
    </div>
  );
};

export default ChatPanel;
