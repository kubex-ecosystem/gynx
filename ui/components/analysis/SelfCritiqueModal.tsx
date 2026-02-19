import { AnimatePresence, motion } from 'framer-motion';
import { CheckCircle, Edit, MessageSquareQuote, X } from 'lucide-react';
import * as React from 'react';
import { SelfCritiqueAnalysis } from '../../types';
import SubtleTokenUsage from '../common/SubtleTokenUsage';

interface SelfCritiqueModalProps {
  isOpen: boolean;
  onClose: () => void;
  critique: SelfCritiqueAnalysis;
}

const SelfCritiqueModal: React.FC<SelfCritiqueModalProps> = ({ isOpen, onClose, critique }) => {

  return (
    <AnimatePresence>
      {isOpen && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          onClick={onClose}
          className="fixed inset-0 bg-black/70 backdrop-blur-sm z-50 flex items-center justify-center p-4"
        >
          <motion.div
            initial={{ scale: 0.9, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            exit={{ scale: 0.9, opacity: 0 }}
            transition={{ type: 'spring', stiffness: 300, damping: 25 }}
            onClick={(e) => e.stopPropagation()}
            className="bg-gray-800 border border-pink-700/60 rounded-xl w-full max-w-2xl max-h-[80vh] flex flex-col shadow-2xl shadow-pink-500/10"
          >
            {/* Header */}
            <div className="flex items-start justify-between p-6 border-b border-gray-700">
              <div className="flex items-center gap-4">
                <div className="bg-pink-900/50 p-3 rounded-full border border-pink-800">
                  <MessageSquareQuote className="w-7 h-7 text-pink-400" />
                </div>
                <div>
                  <h2 className="text-2xl font-bold text-white">AI Self-Critique</h2>
                  <p className="text-gray-400">An assessment of the previous analysis's quality.</p>
                </div>
              </div>
              <button title="Close" onClick={onClose} className="p-1 rounded-full text-gray-400 hover:bg-gray-700 transition-colors absolute top-4 right-4">
                <X className="w-6 h-6" />
              </button>
            </div>

            {/* Content */}
            <div className="p-6 overflow-y-auto space-y-6">
              <div className="flex flex-col sm:flex-row items-center gap-6 p-4 bg-gray-900/50 rounded-lg">
                <div className="flex flex-col items-center">
                  <p className="text-sm text-gray-400 mb-1">Confidence Score</p>
                  <p className="text-5xl font-bold text-pink-300">{critique.confidenceScore}<span className="text-3xl text-gray-500">/10</span></p>
                </div>
                <div className="flex-grow text-center sm:text-left">
                  <p className="text-gray-300 italic">"{critique.overallAssessment}"</p>
                </div>
              </div>

              <div>
                <h3 className="text-lg font-semibold text-green-400 flex items-center gap-2 mb-3">
                  <CheckCircle className="w-5 h-5" /> Positive Points
                </h3>
                <ul className="space-y-2 list-disc list-inside text-gray-300">
                  {critique.positivePoints.map((point, i) => (
                    <li key={i}>{point}</li>
                  ))}
                </ul>
              </div>

              <div>
                <h3 className="text-lg font-semibold text-yellow-400 flex items-center gap-2 mb-3">
                  <Edit className="w-5 h-5" /> Areas for Refinement
                </h3>
                <ul className="space-y-2 list-disc list-inside text-gray-300">
                  {critique.areasForRefinement.map((point, i) => (
                    <li key={i}>{point}</li>
                  ))}
                </ul>
              </div>

              <div className="pt-4">
                <SubtleTokenUsage usageMetadata={critique.usageMetadata} label="Critique Cost" />
              </div>
            </div>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default SelfCritiqueModal;
