import * as React from 'react';

// Contexts & Hooks
import { AppProvider } from './contexts/AppContext';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { ConfirmationProvider } from './contexts/ConfirmationContext';
import { NotificationProvider } from './contexts/NotificationContext';
import { UserProvider } from './contexts/UserContext';
// FIX: The original import path for ProjectContext was incorrect. Corrected the path to point to the correct file location to resolve module loading errors.
import { ProjectContextProvider, useProjectContext } from './contexts/ProjectContext';

// Components
// FIX: Added .tsx extensions to imports to be explicit, though the root issue was likely empty files.
import AnalysisResults from './components/analysis/AnalysisResults';
import EvolutionDisplay from './components/analysis/EvolutionDisplay';
import ProjectInput from './components/input/ProjectInput';
import KanbanBoard from './components/kanban/KanbanBoard';
import LandingPage from './components/landing/LandingPage';
import Header from './components/layout/Header';
import NavigationBar from './components/layout/NavigationBar';
// FIX: Corrected import path for Dashboard
import 'framer-motion';
import ChatPanel from './components/chat/ChatPanel';
import ConfirmationModal from './components/common/ConfirmationModal';
import Loader from './components/common/Loader';
import NetworkStatusIndicator from './components/common/NetworkStatusIndicator';
import NotificationContainer from './components/common/NotificationContainer';
import Dashboard from './components/dashboard/Dashboard';
import HistoryPanel from './components/history/HistoryPanel';
import UserSettingsModal from './components/settings/UserSettingsModal';

// Types
// FIX: Corrected import path for types
import { ViewType } from './types';

function DashboardWrapper() {
  const {
    currentView,
    setCurrentView,
    isAnalyzing,
    activeProject,
    currentAnalysis,
    evolutionAnalysis,
  } = useProjectContext();

  const renderContent = () => {
    switch (currentView) {
      case ViewType.Dashboard:
        return <Dashboard />;
      case ViewType.Input:
        return <ProjectInput />;
      case ViewType.Analysis:
        return activeProject && currentAnalysis ? <AnalysisResults /> : <Dashboard />;
      case ViewType.Evolution:
        return evolutionAnalysis ? <EvolutionDisplay /> : <Dashboard />;
      case ViewType.Kanban:
        return activeProject?.kanban ? <KanbanBoard /> : <Dashboard />;
      case ViewType.Chat:
        return activeProject ? <ChatPanel /> : <Dashboard />;
      default:
        return <Dashboard />;
    }
  };

  return (
    <div className="text-white min-h-screen font-sans selection:bg-purple-500/30">
      {isAnalyzing && <Loader />}

      <Header />
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <NavigationBar
          currentView={currentView}
          onNavigate={(v) => setCurrentView(v)}
          hasAnalysis={!!currentAnalysis}
          isAnalysisOpen={!!activeProject}
        />
        <div className="mt-8">
          {renderContent()}
        </div>
      </main>

      <HistoryPanel />
      <UserSettingsModal />
      <ConfirmationModal />
      <NotificationContainer />
      <NetworkStatusIndicator />
    </div>
  );
}

const App: React.FC = () => (
  <NotificationProvider>
    <AuthProvider>
      <ConfirmationProvider>
        <UserProvider>
          <AppProvider>
            <ProjectContextProvider>
              <MainApp />
            </ProjectContextProvider>
          </AppProvider>
        </UserProvider>
      </ConfirmationProvider>
    </AuthProvider>
  </NotificationProvider>
);

const MainApp: React.FC = () => {
  const { user } = useAuth();

  if (!user) {
    return <LandingPage />;
  }
  return <DashboardWrapper />;
};

export default App;
