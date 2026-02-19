import * as React from 'react';
import { motion } from 'framer-motion';
import { FileText, Star, ListChecks, Zap } from 'lucide-react';

interface DashboardEmptyStateProps {
  onNavigate: () => void;
}

const KpiPlaceholder: React.FC<{ icon: React.ReactNode; description: string; delay: number }> = ({ icon, description, delay }) => (
    <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: delay }}
        className="bg-gray-800/50 border border-gray-700/80 p-4 rounded-lg flex items-start gap-4"
    >
        <div className="bg-gray-900/50 p-3 rounded-full">{icon}</div>
        <div>
            <div className="h-8 w-16 bg-gray-700 rounded-md mb-2"></div>
            <p className="text-xs text-gray-500">{description}</p>
        </div>
    </motion.div>
);

const DashboardEmptyState: React.FC<DashboardEmptyStateProps> = ({ onNavigate }) => {

  return (
    <div className="h-full flex flex-col items-center justify-center text-center p-4">
      <motion.div
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.5 }}
      >
        <h2 className="text-3xl font-bold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-teal-400">No projects yet</h2>
        <p className="mt-4 max-w-xl mx-auto text-gray-400">Get started by creating your first project analysis. Your dashboard will light up with insights once you do!</p>
        <button
          onClick={onNavigate}
          className="mt-8 px-6 py-3 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-lg font-semibold hover:shadow-2xl hover:shadow-blue-500/30 hover:scale-105 transition-all"
        >
          Start First Analysis
        </button>
      </motion.div>
      <div className="mt-16 w-full max-w-4xl grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <KpiPlaceholder icon={<FileText className="w-6 h-6 text-purple-400" />} description="Total analyses you've performed across all projects." delay={0.3} />
          <KpiPlaceholder icon={<Star className="w-6 h-6 text-yellow-400" />} description="The average viability score from your analyses." delay={0.4} />
          <KpiPlaceholder icon={<ListChecks className="w-6 h-6 text-teal-400" />} description="The analysis type you use most frequently." delay={0.5} />
          <KpiPlaceholder icon={<Zap className="w-6 h-6 text-blue-400" />} description="Tokens consumed by your analyses this month." delay={0.6} />
      </div>
    </div>
  );
};

export default DashboardEmptyState;