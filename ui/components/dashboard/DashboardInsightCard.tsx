import { motion } from 'framer-motion';
import { Sparkles } from 'lucide-react';
import * as React from 'react';
// FIX: Corrected import path for ProjectContext
import { useProjectContext } from '../../contexts/ProjectContext';
import { useUser } from '../../contexts/UserContext';

const DashboardInsightCard: React.FC = () => {
  const { dashboardInsight, isInsightLoading } = useProjectContext();
  const { name: userName } = useUser();

  if (isInsightLoading) {
    return (
      <div className="p-6 bg-gray-800/50 border border-gray-700 rounded-xl flex items-start gap-4 animate-pulse">
        <div className="w-12 h-12 rounded-full bg-gray-700 shrink-0"></div>
        <div className="flex-grow space-y-2">
          <div className="h-5 w-3/4 bg-gray-700 rounded"></div>
          <div className="h-4 w-full bg-gray-700 rounded"></div>
          <div className="h-4 w-5/6 bg-gray-700 rounded"></div>
        </div>
      </div>
    );
  }

  if (!dashboardInsight) {
    return null; // Don't render if there's no insight
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: -20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, ease: 'easeOut' }}
      className="group relative p-6 bg-gradient-to-br from-gray-800 to-gray-900/70 border border-gray-700 rounded-xl flex items-start gap-4 overflow-hidden"
    >
      <div
        className="absolute top-0 left-0 w-full h-full bg-gradient-to-r from-purple-500/10 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500"
        aria-hidden="true"
      />
      <div className="shrink-0 w-12 h-12 rounded-full bg-purple-900/50 flex items-center justify-center border border-purple-800 z-10">
        <Sparkles className="w-6 h-6 text-purple-400" />
      </div>
      <div className="flex-grow z-10">
        <h2 className="text-xl font-bold text-white">
          {dashboardInsight.title}
        </h2>
        <p className="mt-2 text-gray-300">{dashboardInsight.summary}</p>
      </div>
    </motion.div>
  );
};

export default DashboardInsightCard;
