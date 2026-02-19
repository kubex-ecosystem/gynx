import { AnimatePresence, motion } from 'framer-motion';
import { X } from 'lucide-react';
import * as React from 'react';
import { AnalysisFeature } from './LandingPage';

interface FeatureDetailModalProps {
  feature: AnalysisFeature | null;
  onClose: () => void;
}

const featureDetails: Record<string, string> = {
  architecture: "Think like an architect. This review analyzes your project's high-level design, identifying its architectural style (e.g., microservices, monolith) and evaluating adherence to fundamental principles. As a key feature, it automatically generates a visual diagram of your architecture, providing instant clarity.",
  security: "Put on your white hat. The security analysis acts as an automated cybersecurity expert, scanning your documentation for potential vulnerabilities, insecure practices, and missing security layers like authentication. It helps you identify and prioritize risks before they become critical.",
  scalability: "Will your project handle success? This review focuses on your architecture's ability to scale. It looks for performance bottlenecks, single points of failure, and inefficient data handling, providing recommendations to ensure your application can grow with your user base.",
  codeQuality: "Promote a healthy and maintainable codebase. This analysis evaluates your project's structure, adherence to best practices, modularity, and principles like SOLID. It's like having a principal engineer review your documentation to improve long-term developer experience.",
  compliance: "Ensure your project is responsible and accessible. This analysis focuses on compliance with accessibility guidelines (WCAG), data privacy regulations (like GDPR/LGPD), and other industry best practices. It helps you build a more inclusive and trustworthy application.",
  documentation: "How good is your project's first impression? This review analyzes your documentation itself for clarity, completeness, and ease of use for a new developer. It provides suggestions to make your READMEs, guides, and comments more effective and welcoming.",
};

const colorMap = {
  blue: { text: 'text-blue-400', border: 'border-blue-600/60', shadowRgb: '96, 165, 250' },
  red: { text: 'text-red-400', border: 'border-red-600/60', shadowRgb: '248, 113, 113' },
  purple: { text: 'text-purple-400', border: 'border-purple-600/60', shadowRgb: '192, 132, 252' },
  teal: { text: 'text-teal-400', border: 'border-teal-600/60', shadowRgb: '45, 212, 191' },
  amber: { text: 'text-amber-400', border: 'border-amber-600/60', shadowRgb: '251, 191, 36' },
  green: { text: 'text-green-400', border: 'border-green-600/60', shadowRgb: '74, 222, 128' },
};

const FeatureDetailModal: React.FC<FeatureDetailModalProps> = ({ feature, onClose }) => {

  return (
    <AnimatePresence>
      {feature && (
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
            style={{ '--shadow-rgb': colorMap[feature.color].shadowRgb } as React.CSSProperties}
            className={`bg-gray-800 border ${colorMap[feature.color].border} rounded-xl w-full max-w-2xl max-h-[80vh] flex flex-col relative shadow-[0_4px_30px_rgba(var(--shadow-rgb),0.2)]`}
          >
            {/* Header */}
            <div className="flex items-start justify-between p-6 border-b border-gray-700">
              <div className="flex items-center gap-4">
                <div className="bg-gray-900/50 p-3 rounded-full">
                  <feature.icon className={`w-7 h-7 ${colorMap[feature.color].text}`} />
                </div>
                <div>
                  <h2 className="text-2xl font-bold text-white">{feature.title}</h2>
                  <p className="text-gray-400">{feature.description}</p>
                </div>
              </div>
              <button title='Close' onClick={onClose} className="p-1 rounded-full text-gray-400 hover:bg-gray-700 transition-colors absolute top-4 right-4">
                <X className="w-6 h-6" />
              </button>
            </div>

            {/* Content */}
            <div className="p-6 overflow-y-auto">
              <p className="text-gray-300 whitespace-pre-line leading-relaxed">
                {featureDetails[feature.detailKey]}
              </p>
            </div>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default FeatureDetailModal;
