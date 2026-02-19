import * as React from 'react';
import { motion, Variants } from 'framer-motion';
import { EvolutionAnalysis, Improvement, AnalysisType, Priority, ViewType } from '../../types';
import { BrainCircuit, Calculator, Check, GitCompareArrows, Lightbulb, Repeat, RotateCcw, TrendingDown, TrendingUp } from 'lucide-react';
import DifficultyMeter from '../common/DifficultyMeter';
import { useProjectContext } from '../../contexts/ProjectContext';

interface EvolutionDisplayProps {
  // onNavigate is now handled by context
}

const ImprovementCard: React.FC<{ improvement: Improvement; type: 'resolved' | 'new' | 'persistent' }> = ({ improvement, type }) => {
    const typeConfig = {
        resolved: {
            icon: <Check className="w-5 h-5 text-green-400" />,
            borderColor: 'border-green-700/50',
            bgColor: 'bg-green-900/20',
            hoverBorderColor: 'hover:border-green-500/50'
        },
        new: {
            icon: <Lightbulb className="w-5 h-5 text-yellow-400" />,
            borderColor: 'border-yellow-700/50',
            bgColor: 'bg-yellow-900/20',
            hoverBorderColor: 'hover:border-yellow-500/50'
        },
        persistent: {
            icon: <Repeat className="w-5 h-5 text-red-400" />,
            borderColor: 'border-red-700/50',
            bgColor: 'bg-red-900/20',
            hoverBorderColor: 'hover:border-red-500/50'
        },
    };
    
    const config = typeConfig[type];

    return (
        <div className={`p-4 rounded-lg border bg-gray-800/40 ${config.bgColor} ${config.borderColor} ${config.hoverBorderColor} transition-all duration-300 hover:scale-[1.02]`}>
            <div className="flex items-start gap-3">
                <div className="shrink-0 mt-1">{config.icon}</div>
                <div>
                    <h4 className="font-semibold text-white">{improvement.title}</h4>
                    <p className="mt-1 text-sm text-gray-400">{improvement.description}</p>
                    <div className="mt-3 flex items-center gap-4 text-xs">
                        <DifficultyMeter difficulty={improvement.difficulty} />
                        <span className={`px-2 py-0.5 rounded-full font-mono text-xs ${
                            improvement.priority === Priority.High ? 'bg-red-900/80 text-red-300' : 
                            improvement.priority === Priority.Medium ? 'bg-yellow-900/80 text-yellow-300' : 
                            'bg-blue-900/80 text-blue-300'
                        }`}>{improvement.priority}</span>
                    </div>
                </div>
            </div>
        </div>
    );
};

const EvolutionDisplay: React.FC<EvolutionDisplayProps> = () => {
    const { evolutionAnalysis: analysis } = useProjectContext();

    if (!analysis) return null;

    const { keyMetrics: km } = analysis;

    const typeLabels: Partial<Record<AnalysisType, string>> = {
        [AnalysisType.CodeQuality]: "Code Quality",
        [AnalysisType.Security]: "Security Analysis",
        [AnalysisType.Scalability]: "Scalability Analysis",
        [AnalysisType.Compliance]: "Compliance",
        [AnalysisType.DocumentationReview]: "Documentation Review",
        [AnalysisType.Architecture]: "Architectural Review"
    };

    const cardVariants: Variants = {
        hidden: { opacity: 0, y: 20 },
        visible: (i: number) => ({
            opacity: 1,
            y: 0,
            transition: { delay: i * 0.1, duration: 0.5, ease: "easeOut" },
        }),
    };

    const MetricCard: React.FC<{ title: string; previous: number; current: number; custom: number }> = ({ title, previous, current, custom }) => {
        const change = current - previous;
        const isPositiveChange = title === 'Improvements' ? change < 0 : change > 0;
        
        return (
            <motion.div variants={cardVariants} custom={custom} className="group bg-gradient-to-br from-gray-800 to-gray-900/50 border border-gray-700 p-4 rounded-lg flex flex-col justify-between text-center h-full transition-all duration-300 hover:border-blue-500/50 hover:scale-[1.02]">
                <h4 className="text-sm font-medium text-gray-400">{title}</h4>
                <div className="flex items-baseline justify-center gap-3 my-2">
                    <span className="text-xl font-semibold text-gray-500 line-through">{previous}</span>
                    <span className="text-4xl font-bold text-white transition-colors duration-300 group-hover:text-blue-300">{current}</span>
                </div>
                <div className="h-6 flex items-center justify-center">
                    {change !== 0 && (
                        <div className={`flex items-center justify-center gap-1 text-base font-bold ${isPositiveChange ? 'text-green-400' : 'text-red-400'}`}>
                            {isPositiveChange ? <TrendingUp className="w-5 h-5" /> : <TrendingDown className="w-5 h-5" />}
                            <span>{Math.abs(change)}</span>
                        </div>
                    )}
                </div>
            </motion.div>
        );
    };
    
    const analysisTypeLabel = typeLabels[analysis.analysisType] || analysis.analysisType;

  return (
    <div className="space-y-12">
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
      >
        <div className="text-center">
            <div className="inline-flex items-center justify-center gap-3 text-purple-400">
                <GitCompareArrows className="w-8 h-8 md:w-10 md:h-10" />
                <h1 className="text-3xl md:text-4xl lg:text-5xl font-bold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-teal-400">
                    Evolution Analysis
                </h1>
                <BrainCircuit className="w-8 h-8 md:w-10 md:h-10" />
            </div>
            <p className="mt-3 text-lg text-gray-400">
                 {`Comparing analyses for ${analysis.projectName} (${analysisTypeLabel})`}
            </p>
        </div>
      </motion.div>

      <motion.div 
        className="grid grid-cols-1 lg:grid-cols-5 gap-8 bg-gray-900/30 p-6 rounded-xl border border-gray-800"
        initial="hidden"
        animate="visible"
        variants={{ visible: { transition: { staggerChildren: 0.1 } } }}
      >
        <motion.div 
            variants={cardVariants}
            className="lg:col-span-3 bg-gradient-to-br from-gray-800 to-gray-900/50 border border-gray-700 rounded-xl p-6"
        >
            <h3 className="text-xl font-semibold text-white mb-3">Evolution Summary</h3>
            <p className="text-gray-300">{analysis.evolutionSummary}</p>
        </motion.div>
        <div className="lg:col-span-2 flex flex-col gap-4">
            <MetricCard title="Viability Score" previous={km.previousScore} current={km.currentScore} custom={1} />
            <div className="grid grid-cols-2 gap-4">
                <MetricCard title="Strengths" previous={km.previousStrengths} current={km.currentStrengths} custom={2} />
                <MetricCard title="Improvements" previous={km.previousImprovements} current={km.currentImprovements} custom={3} />
            </div>
        </div>
      </motion.div>
      
      {analysis.usageMetadata && (
        <motion.div
            className="flex items-center justify-center gap-3 text-xs text-gray-400 p-2 bg-gray-800/50 border border-gray-700 rounded-lg max-w-md mx-auto"
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3, delay: 0.4 }}
            aria-label="Token usage for this comparison analysis"
        >
            <Calculator className="w-4 h-4 text-gray-500 shrink-0" />
            <div className="flex flex-wrap items-center justify-center gap-x-2 gap-y-1">
                <span className="font-semibold">Comparison Cost:</span>
                <span>{analysis.usageMetadata.totalTokenCount.toLocaleString('en-US')} Tokens</span>
            </div>
        </motion.div>
      )}

      <motion.div
        className="space-y-8"
        initial="hidden"
        animate="visible"
        variants={{ visible: { transition: { staggerChildren: 0.15 } } }}
      >
        <h3 className="text-2xl font-bold text-center text-gray-200">Improvement Breakdown</h3>
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            <motion.div variants={cardVariants} className="space-y-4">
                <div className="flex items-center gap-3">
                    <Check className="w-8 h-8 text-green-400 bg-green-900/50 p-1.5 rounded-full" />
                    <h3 className="text-2xl font-semibold text-green-400">Achievements ({analysis.resolvedImprovements.length})</h3>
                </div>
                {analysis.resolvedImprovements.length > 0 ? (
                    analysis.resolvedImprovements.map((imp, i) => <ImprovementCard key={i} improvement={imp} type="resolved" />)
                ) : <p className="text-gray-500 italic p-4 text-center">No previously identified issues were resolved.</p>}
            </motion.div>

            <motion.div variants={cardVariants} className="space-y-4">
                <div className="flex items-center gap-3">
                    <Lightbulb className="w-8 h-8 text-yellow-400 bg-yellow-900/50 p-1.5 rounded-full" />
                    <h3 className="text-2xl font-semibold text-yellow-400">New Challenges ({analysis.newImprovements.length})</h3>
                </div>
                {analysis.newImprovements.length > 0 ? (
                    analysis.newImprovements.map((imp, i) => <ImprovementCard key={i} improvement={imp} type="new" />)
                ) : <p className="text-gray-500 italic p-4 text-center">No new areas for improvement were identified.</p>}
            </motion.div>

            <motion.div variants={cardVariants} className="space-y-4">
                <div className="flex items-center gap-3">
                    <Repeat className="w-8 h-8 text-red-400 bg-red-900/50 p-1.5 rounded-full" />
                    <h3 className="text-2xl font-semibold text-red-400">Technical Debt ({analysis.persistentImprovements.length})</h3>
                </div>
                {analysis.persistentImprovements.length > 0 ? (
                    analysis.persistentImprovements.map((imp, i) => <ImprovementCard key={i} improvement={imp} type="persistent" />)
                ) : <p className="text-gray-500 italic p-4 text-center">No persistent issues were found.</p>}
            </motion.div>
        </div>
      </motion.div>
    </div>
  );
};

export default EvolutionDisplay;