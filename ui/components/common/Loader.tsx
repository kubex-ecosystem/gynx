import * as React from 'react';

import { AnimatePresence, motion } from 'framer-motion';
import { Loader2 } from 'lucide-react';
import { useEffect, useState } from 'react';

const Loader: React.FC = () => {
  const steps = [
    "Parsing file structure...",
    "Evaluating architecture...",
    "Checking code quality...",
    "Identifying potential improvements...",
    "Compiling the report..."
  ];
  const [currentStep, setCurrentStep] = useState(0);

  useEffect(() => {
    const interval = setInterval(() => {
      setCurrentStep(prev => (prev + 1) % (steps?.length || 1));
    }, 2500);
    return () => clearInterval(interval);
  }, [steps]);

  if (!steps || steps.length === 0) {
    return null; // Don't render if translations are not ready
  }

  return (
    <div
      className="fixed inset-0 bg-gray-900/80 backdrop-blur-sm z-[100] flex flex-col items-center justify-center"
      aria-label="Analyzing content, please wait."
      role="status"
    >
      <Loader2 className="w-12 h-12 text-purple-400 animate-spin" />
      <h2 className="mt-4 text-2xl font-bold text-white">Analyzing your project...</h2>
      <p className="text-gray-400">This may take a few moments.</p>
      <div className="mt-6 text-center h-6 overflow-hidden">
        <AnimatePresence mode="wait">
          <motion.p
            key={currentStep}
            initial={{ y: 20, opacity: 0 }}
            animate={{ y: 0, opacity: 1 }}
            exit={{ y: -20, opacity: 0 }}
            transition={{ duration: 0.5, ease: 'easeInOut' }}
            className="text-gray-300"
          >
            {steps[currentStep]}
          </motion.p>
        </AnimatePresence>
      </div>
    </div>
  );
};

export default Loader;
