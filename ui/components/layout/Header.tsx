import { Github, History, PlusCircle, Settings } from 'lucide-react';
import * as React from 'react';
import { useAuth } from '../../contexts/AuthContext';
import { useProjectContext } from '../../contexts/ProjectContext';
import { useUser } from '../../contexts/UserContext';
import { ViewType } from '../../types';

const Header: React.FC = () => {
  const { user } = useAuth();
  const { setIsUserSettingsModalOpen } = useUser();
  const {
    setIsHistoryPanelOpen,
    activeProject,
    setCurrentView,
    setActiveProjectId
  } = useProjectContext();

  const handleNewProject = () => {
    setActiveProjectId(null);
    setCurrentView(ViewType.Input);
  };

  return (
    <header className="py-4 border-b border-gray-800/50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <img src="/logo.svg" alt="GemX Logo" className="w-9 h-9 transition-transform hover:scale-110" />
          <div>
            <h1 className="text-xl font-bold text-white">GemX Analyzer</h1>
            {activeProject && (
              <p className="text-xs text-purple-400 font-mono -mt-1 truncate max-w-xs" title={activeProject.name}>
                {activeProject.name}
              </p>
            )}
          </div>
        </div>
        <div className="flex items-center gap-2 sm:gap-4">
          {/* New Project */}
          {user && (
            <>
              {/* New Project */}
              <button
                title='Create new project'
                onClick={handleNewProject}
                className="flex items-center gap-2 px-3 py-1.5 text-sm bg-gray-700/80 border border-gray-600 rounded-lg hover:bg-gray-700 text-gray-300 hover:text-white transition-colors"
                aria-label="Create new project"
              >
                <PlusCircle className="w-4 h-4" />
                <span className="hidden sm:inline">New Project</span>
              </button>

              {/* History */}
              <button
                title='View history'
                onClick={() => setIsHistoryPanelOpen(true)}
                // disabled={!activeProject}
                className="p-2 text-gray-400 rounded-full hover:bg-gray-700 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed"
                aria-label="View history"
              >
                <History className="w-5 h-5" />
              </button>

              {/* Settings */}
              <button
                title='Open settings'
                onClick={() => setIsUserSettingsModalOpen(true)}
                className="p-2 text-gray-400 rounded-full hover:bg-gray-700 hover:text-white"
                aria-label="Open settings"
              >
                <Settings className="w-5 h-5" />
              </button>
            </>
          )}
          <a
            href="https://github.com/kubex-ecosystem/analyzer"
            target="_blank"
            rel="noopener noreferrer"
            className="text-gray-400 hover:text-white transition-transform hover:scale-110"
            aria-label="View on GitHub"
          >
            <Github className="w-5 h-5" />
          </a>
        </div>
      </div>
    </header>
  );
};

export default Header;
