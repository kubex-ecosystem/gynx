import { motion } from 'framer-motion';
import { FileText, ListChecks, Star, Zap } from 'lucide-react';
import * as React from 'react';
import { useMemo } from 'react';
import { useProjectContext } from '../../contexts/ProjectContext';
import { useUser } from '../../contexts/UserContext';
import { AnalysisType, ViewType } from '../../types';
import DashboardEmptyState from './DashboardEmptyState';
import DashboardInsightCard from './DashboardInsightCard';
import TrendChart from './TrendChart';

const KpiCard: React.FC<{ icon: React.ReactNode; title: string; value: string | number; description: string; delay: number }> = ({ icon, title, value, description, delay }) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    animate={{ opacity: 1, y: 0 }}
    transition={{ duration: 0.5, delay }}
    className="bg-gray-800/50 border border-gray-700/80 p-4 rounded-lg flex items-start gap-4"
  >
    <div className="bg-gray-900/50 p-3 rounded-full">{icon}</div>
    <div>
      <p className="text-3xl font-bold text-white">{value}</p>
      <p className="text-sm font-semibold text-gray-300 -mt-1">{title}</p>
      <p className="text-xs text-gray-500 mt-1">{description}</p>
    </div>
  </motion.div>
);

const Dashboard: React.FC = () => {
  const {
    projects,
    activeProjectId,
    setActiveProjectId,
    setCurrentView,
    activeProject
  } = useProjectContext();

  const { name: userName } = useUser();

  const projectOptions = useMemo(() => projects.filter(p => p.id !== 'example-project-id'), [projects]);

  const dashboardStats = useMemo(() => {
    if (!activeProject || activeProject.history.length === 0) {
      return {
        totalAnalyses: 0,
        averageScore: 0,
        commonType: 'N/A',
        scoreHistory: [],
      };
    }

    const history = activeProject.history;
    const totalAnalyses = history.length;
    // FIX: Explicitly typed the 'sum' accumulator to resolve TS error.
    const averageScore = history.reduce((sum: number, item) => sum + item.analysis.viability.score, 0) / totalAnalyses;

    const typeCounts = history.reduce((acc, item) => {
      acc[item.analysis.analysisType] = (acc[item.analysis.analysisType] || 0) + 1;
      return acc;
    }, {} as Record<AnalysisType, number>);

    const commonType = Object.keys(typeCounts).length > 0
      ? Object.entries(typeCounts).sort((a, b) => b[1] - a[1])[0][0]
      : 'N/A';

    const scoreHistory = history.map(item => item.analysis.viability.score);

    return {
      totalAnalyses,
      averageScore: parseFloat(averageScore.toFixed(1)),
      commonType,
      scoreHistory,
    };
  }, [activeProject]);

  if (projectOptions.length === 0 && !activeProject) {
    return <DashboardEmptyState onNavigate={() => setCurrentView(ViewType.Input)} />;
  }

  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row justify-between items-start gap-4">
        <div>
          <h1 className="text-3xl font-bold text-white">Welcome, {userName || 'User'}!</h1>
          <p className="text-gray-400">Here's a summary of your projects.</p>
        </div>
        <div className="w-full sm:w-auto">
          <select
            title='Select Project'
            value={activeProjectId || ''}
            onChange={(e) => setActiveProjectId(e.target.value)}
            className="w-full p-2 bg-gray-800 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-purple-500"
          >
            <option value="" disabled>Select a project...</option>
            {projectOptions.map(p => (
              <option key={p.id} value={p.id}>{p.name}</option>
            ))}
          </select>
        </div>
      </div>

      <DashboardInsightCard />

      {activeProject ? (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <KpiCard icon={<FileText className="w-6 h-6 text-purple-400" />} title="Total Analyses" value={dashboardStats.totalAnalyses} description="Analyses performed for this project." delay={0.1} />
            <KpiCard icon={<Star className="w-6 h-6 text-yellow-400" />} title="Average Score" value={dashboardStats.averageScore} description="Average viability score for this project." delay={0.2} />
            <KpiCard icon={<ListChecks className="w-6 h-6 text-teal-400" />} title="Common Type" value={dashboardStats.commonType} description="Most frequent analysis type for this project." delay={0.3} />
            <KpiCard icon={<Zap className="w-6 h-6 text-blue-400" />} title="Tokens Used" value="N/A" description="Token consumption tracking coming soon." delay={0.4} />
          </div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.5 }}
            className="bg-gray-800/50 border border-gray-700/80 p-6 rounded-lg h-64"
          >
            <h3 className="text-lg font-semibold text-white">Viability Score Trend</h3>
            <TrendChart data={dashboardStats.scoreHistory} />
          </motion.div>
        </>
      ) : (
        <div className="text-center py-16 bg-gray-800/50 border border-gray-700/80 rounded-lg">
          <h2 className="text-2xl font-bold text-white">Please select a project</h2>
          <p className="text-gray-400 mt-2">Choose a project from the dropdown to see its dashboard.</p>
        </div>
      )}
    </div>
  );
};

export default Dashboard;
