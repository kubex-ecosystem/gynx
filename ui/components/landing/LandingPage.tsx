import * as React from 'react';
import { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { LucideIcon, FileCode, ShieldCheck, BarChart, Scale, BookOpen, Network } from 'lucide-react';
import { useAuth } from '../../contexts/AuthContext';
import FeatureDetailModal from './FeatureDetailModal';

export interface AnalysisFeature {
  title: string;
  description: string;
  detailKey: string;
  icon: LucideIcon;
  color: 'blue' | 'red' | 'purple' | 'teal' | 'amber' | 'green';
}

const features: AnalysisFeature[] = [
  { title: 'Architectural Review', description: 'Analyzes high-level design and generates a visual diagram', detailKey: 'architecture', icon: Network, color: 'purple' },
  { title: 'Code Quality', description: 'Evaluates patterns, maintainability, and adherence to principles like SOLID', detailKey: 'codeQuality', icon: FileCode, color: 'teal' },
  { title: 'Security Analysis', description: 'Focus on vulnerabilities, security practices, and compliance', detailKey: 'security', icon: ShieldCheck, color: 'red' },
  { title: 'Scalability Analysis', description: 'Assessment of system growth capacity and performance', detailKey: 'scalability', icon: BarChart, color: 'blue' },
  { title: 'Compliance & Practices', description: 'Focus on accessibility (WCAG), data privacy, and industry standards', detailKey: 'compliance', icon: Scale, color: 'green' },
  { title: 'Documentation Review', description: 'Analysis of clarity, completeness, and structure of project documentation', detailKey: 'documentation', icon: BookOpen, color: 'amber' },
];

const colorMap = {
    blue: { text: 'text-blue-400', border: 'border-blue-500/40', hoverBorder: 'hover:border-blue-500/80' },
    red: { text: 'text-red-400', border: 'border-red-500/40', hoverBorder: 'hover:border-red-500/80' },
    purple: { text: 'text-purple-400', border: 'border-purple-500/40', hoverBorder: 'hover:border-purple-500/80' },
    teal: { text: 'text-teal-400', border: 'border-teal-500/40', hoverBorder: 'hover:border-teal-500/80' },
    amber: { text: 'text-amber-400', border: 'border-amber-500/40', hoverBorder: 'hover:border-amber-500/80' },
    green: { text: 'text-green-400', border: 'border-green-500/40', hoverBorder: 'hover:border-green-500/80' },
};


const LandingPage: React.FC = () => {
  const { login } = useAuth();
  const dynamicPhrases = [
    "complex architectures",
    "legacy code",
    "microservices",
    "RESTful APIs",
    "databases",
    "cloud infrastructure",
    "web applications",
    "distributed systems"
  ];
  const [phraseIndex, setPhraseIndex] = useState(0);
  const [selectedFeature, setSelectedFeature] = useState<AnalysisFeature | null>(null);

  useEffect(() => {
    const interval = setInterval(() => {
      setPhraseIndex(prev => (prev + 1) % (dynamicPhrases?.length || 1));
    }, 2500);
    return () => clearInterval(interval);
  }, [dynamicPhrases]);

  return (
    <>
      <div className="min-h-screen font-sans selection:bg-purple-500/30 overflow-x-hidden">
        <main className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16 text-center">
          <motion.h1
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="text-4xl sm:text-5xl md:text-6xl font-extrabold tracking-tight"
          >
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-teal-400">Transform Documentation into</span>{' '}
            <span className="relative inline-flex h-[1.3em] overflow-hidden align-bottom">
              <motion.span
                key={phraseIndex}
                initial={{ y: '100%' }}
                animate={{ y: '0%' }}
                exit={{ y: '-100%' }}
                transition={{ duration: 0.5, ease: 'easeInOut' }}
                className="absolute inset-0 bg-clip-text text-transparent bg-gradient-to-r from-purple-400 to-teal-400"
              >
                {dynamicPhrases[phraseIndex]}
              </motion.span>
            </span>
          </motion.h1>

          <motion.p
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2 }}
            className="mt-6 max-w-2xl mx-auto text-lg text-gray-400"
          >
            Transform your project documentation into actionable insights with AI-driven analysis.
          </motion.p>

          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: 0.4 }}
            className="mt-10"
          >
            <motion.button
              onTap={login}
              className="px-8 py-3 bg-gradient-to-r from-purple-600 to-blue-500 text-white rounded-lg font-semibold text-lg shadow-lg"
              whileHover={{ scale: 1.05, boxShadow: "0px 10px 30px rgba(59, 130, 246, 0.4)" }}
              whileTap={{ scale: 0.95 }}
            >
              Start Analysis
            </motion.button>
          </motion.div>

          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.6 }}
            className="mt-24"
          >
            <h2 className="text-3xl font-bold">Features</h2>
            <div className="mt-12 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
              {features.map((feature, i) => (
                <motion.div
                  key={feature.title}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.8 + i * 0.1 }}
                  onClick={() => setSelectedFeature(feature)}
                  className={`bg-gray-800/50 border p-6 rounded-xl text-left ${colorMap[feature.color].border} ${colorMap[feature.color].hoverBorder} transition-all duration-300 cursor-pointer transform hover:scale-105`}
                >
                  <div className="flex items-center gap-4">
                    <feature.icon className={`w-8 h-8 ${colorMap[feature.color].text}`} />
                    <h3 className="text-xl font-bold">{feature.title}</h3>
                  </div>
                  <p className="mt-4 text-gray-400">{feature.description}</p>
                </motion.div>
              ))}
            </div>
          </motion.div>
        </main>
      </div>
      <FeatureDetailModal feature={selectedFeature} onClose={() => setSelectedFeature(null)} />
    </>
  );
};

export default LandingPage;