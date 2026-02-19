import * as React from 'react';

import { motion, Variants } from 'framer-motion';
import { Award, Check, FileText, Info, MessageSquareQuote, Network, Target, Zap } from 'lucide-react';
import { useState } from 'react';
import { useProjectContext } from '../../contexts/ProjectContext';
import { Improvement, NextStep, Priority } from '../../types';
import DifficultyMeter from '../common/DifficultyMeter';
import MaturityKpiCard from '../common/MaturityKpiCard';
import ViabilityScore from '../common/ViabilityScore';
import MermaidDiagram from './MermaidDiagram';
import SelfCritiqueModal from './SelfCritiqueModal';
import SuggestionsDisplay from './SuggestionsDisplay';


const cardVariants: Variants = {
  hidden: { opacity: 0, y: 20 },
  visible: (i: number) => ({
    opacity: 1,
    y: 0,
    transition: { delay: i * 0.1, duration: 0.5, ease: "easeOut" },
  }),
};

const ImprovementCard: React.FC<{ item: Improvement, custom: number }> = ({ item, custom }) => {
  return (
    <motion.div variants={cardVariants} custom={custom} className="p-4 rounded-lg border bg-gray-800/40 border-gray-700/50 hover:border-gray-600 transition-all duration-300 hover:scale-[1.02]">
      <h4 className="font-semibold text-white">{item.title}</h4>
      <p className="mt-1 text-sm text-gray-400">{item.description}</p>
      <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-2 text-xs">
        <DifficultyMeter difficulty={item.difficulty} />
        <span className={`px-2 py-0.5 rounded-full font-mono text-xs ${item.priority === Priority.High ? 'bg-red-900/80 text-red-300' :
          item.priority === Priority.Medium ? 'bg-yellow-900/80 text-yellow-300' :
            'bg-blue-900/80 text-blue-300'
          }`}>{item.priority}</span>
      </div>
      <p className="mt-2 text-xs text-gray-500 italic"><strong>Business Impact:</strong> {item.businessImpact}</p>
    </motion.div>
  );
};

const NextStepCard: React.FC<{ item: NextStep, custom: number }> = ({ item, custom }) => {
  return (
    <motion.div variants={cardVariants} custom={custom} className="p-4 rounded-lg bg-gray-900/50">
      <h4 className="font-semibold text-white">{item.title}</h4>
      <p className="mt-1 text-sm text-gray-400">{item.description}</p>
      <div className="mt-3">
        <DifficultyMeter difficulty={item.difficulty} />
      </div>
    </motion.div>
  );
}

const AnalysisResults: React.FC = () => {
  const { currentAnalysis: analysis, isExample, activeProject, activeHistoryId } = useProjectContext();
  const [isCritiqueModalOpen, setIsCritiqueModalOpen] = useState(false);

  if (!analysis) return null;

  const critique = activeHistoryId ? activeProject?.critiques?.[activeHistoryId] : null;

  return (
    <>
      <div className="space-y-12">
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
        >
          <div className="text-center">
            <h1 className="text-3xl md:text-4xl font-bold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-teal-400">{`Analysis for: ${analysis.projectName}`}</h1>
            {isExample && (
              <div className="mt-4 inline-flex items-center gap-2 p-2 px-3 text-sm bg-purple-900/50 border border-purple-700 text-purple-300 rounded-full">
                <Info className="w-4 h-4" />
                This is an example analysis to demonstrate the tool's capabilities.
              </div>
            )}
          </div>
          {critique && (
            <motion.div
              className="mt-4 text-center"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 0.2 }}
            >
              <button
                onClick={() => setIsCritiqueModalOpen(true)}
                className="inline-flex items-center gap-2 px-4 py-2 text-sm font-semibold text-pink-300 bg-pink-900/50 border border-pink-700 rounded-full hover:bg-pink-900/80 transition-colors"
              >
                <MessageSquareQuote className="w-4 h-4" /> View AI Self-Critique
              </button>
            </motion.div>
          )}
        </motion.div>

        <motion.div
          className="grid grid-cols-1 lg:grid-cols-3 gap-8"
          initial="hidden"
          animate="visible"
          variants={{ visible: { transition: { staggerChildren: 0.1 } } }}
        >
          <motion.div variants={cardVariants} custom={0} className="lg:col-span-2 bg-gray-800/50 border border-gray-700 rounded-xl p-6">
            <div className="flex items-center gap-3 mb-4">
              <FileText className="w-6 h-6 text-blue-400" />
              <h3 className="text-xl font-semibold text-white">Executive Summary</h3>
            </div>
            <p className="text-gray-300 whitespace-pre-line">{analysis.summary}</p>
          </motion.div>
          <motion.div variants={cardVariants} custom={1} className="bg-gray-800/50 border border-gray-700 rounded-xl p-6 flex flex-col items-center justify-center">
            <h3 className="text-xl font-semibold text-white mb-4">Project Viability</h3>
            <ViabilityScore score={analysis.viability.score} />
            <p className="text-center text-sm text-gray-400 mt-4 italic">"{analysis.viability.assessment}"</p>
          </motion.div>
        </motion.div>

        {analysis.architectureDiagram && (
          <motion.div initial="hidden" animate="visible" variants={{ visible: { transition: { staggerChildren: 0.1 } } }}>
            <div className="flex items-center gap-3 mb-4">
              <Network className="w-6 h-6 text-purple-400" />
              <h3 className="text-2xl font-bold text-white">Architecture Diagram</h3>
            </div>
            <motion.div variants={cardVariants} custom={0}>
              <MermaidDiagram chart={analysis.architectureDiagram} />
            </motion.div>
          </motion.div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          <motion.div initial="hidden" animate="visible" variants={{ visible: { transition: { staggerChildren: 0.1 } } }}>
            <div className="flex items-center gap-3 mb-4">
              <Award className="w-6 h-6 text-green-400" />
              <h3 className="text-2xl font-bold text-white">Key Strengths</h3>
            </div>
            <ul className="space-y-3">
              {analysis.strengths.map((strength, i) => (
                <motion.li key={i} variants={cardVariants} custom={i} className="flex items-start gap-3 p-3 bg-gray-800/30 rounded-lg">
                  <Check className="w-5 h-5 text-green-500 mt-1 shrink-0" />
                  <span className="text-gray-300">{strength}</span>
                </motion.li>
              ))}
            </ul>
          </motion.div>
          <MaturityKpiCard maturity={analysis.maturity} />
        </div>

        <motion.div initial="hidden" animate="visible" variants={{ visible: { transition: { staggerChildren: 0.1 } } }}>
          <div className="flex items-center gap-3 mb-4">
            <Zap className="w-6 h-6 text-yellow-400" />
            <h3 className="text-2xl font-bold text-white">Suggested Improvements</h3>
          </div>
          <div className="space-y-4">
            {analysis.improvements.map((imp, i) => (
              <ImprovementCard key={i} item={imp} custom={i} />
            ))}
          </div>
        </motion.div>

        <motion.div initial="hidden" animate="visible" variants={{ visible: { transition: { staggerChildren: 0.1 } } }}>
          <div className="flex items-center gap-3 mb-4">
            <Target className="w-6 h-6 text-purple-400" />
            <h3 className="text-2xl font-bold text-white">Next Steps</h3>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            <div>
              <h4 className="text-lg font-semibold text-purple-300 mb-3">Short-Term</h4>
              <div className="space-y-4">
                {analysis.nextSteps.shortTerm.map((step, i) => <NextStepCard key={i} item={step} custom={i} />)}
              </div>
            </div>
            <div>
              <h4 className="text-lg font-semibold text-purple-300 mb-3">Long-Term</h4>
              <div className="space-y-4">
                {analysis.nextSteps.longTerm.map((step, i) => <NextStepCard key={i} item={step} custom={i} />)}
              </div>
            </div>
          </div>
        </motion.div>

        <SuggestionsDisplay />
      </div>
      {critique && (
        <SelfCritiqueModal
          isOpen={isCritiqueModalOpen}
          onClose={() => setIsCritiqueModalOpen(false)}
          critique={critique}
        />
      )}
    </>
  );
}

export default AnalysisResults;
