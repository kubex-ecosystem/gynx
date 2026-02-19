import '@/types/MotionExtendedProps';
import * as React from 'react';

import '@/types/MotionExtendedProps';
import { motion } from 'framer-motion';
import { FileText, Github, Loader2, MessageSquareQuote, Wand2 } from 'lucide-react';
import { useEffect, useState } from 'react';
import { initialProjectContext } from '../../constants';
import { useNotification } from '../../contexts/NotificationContext';
import { useProjectContext } from '../../contexts/ProjectContext';
import { useUser } from '../../contexts/UserContext';
import { exampleProject } from '../../data/exampleAnalysis';
import { fetchRepoContents } from '../../services/integrations/github';
import { AnalysisType } from '../../types';
import GitHubSearchModal from './GitHubSearchModal';

const colorMap: Record<string, { border: string; bg: string; hoverBorder: string }> = {
  blue: { border: 'border-blue-600', bg: 'bg-blue-900/50', hoverBorder: 'hover:border-blue-500/80' },
  red: { border: 'border-red-600', bg: 'bg-red-900/50', hoverBorder: 'hover:border-red-500/80' },
  purple: { border: 'border-purple-600', bg: 'bg-purple-900/50', hoverBorder: 'hover:border-purple-500/80' },
  teal: { border: 'border-teal-600', bg: 'bg-teal-900/50', hoverBorder: 'hover:border-teal-500/80' },
  amber: { border: 'border-amber-600', bg: 'bg-amber-900/50', hoverBorder: 'hover:border-amber-500/80' },
  green: { border: 'border-green-600', bg: 'bg-green-900/50', hoverBorder: 'hover:border-green-500/80' },
  pink: { border: 'border-pink-600', bg: 'bg-pink-900/50', hoverBorder: 'hover:border-pink-500/80' },
};

const AnalysisTypeButton: React.FC<{
  type: AnalysisType;
  label: string;
  description: string;
  color: string;
  isSelected: boolean;
  onClick: () => void;
  disabled?: boolean;
}> = ({ type, label, description, color, isSelected, onClick, disabled = false }) => (
  <motion.button
    onTap={onClick}

    className={`p-4 text-left border rounded-lg transition-all w-full relative ${isSelected
      ? `${colorMap[color].bg} ${colorMap[color].border}`
      : `bg-gray-800/50 border-gray-700 ${!disabled ? colorMap[color].hoverBorder : ''}`
      } ${disabled ? 'opacity-50 cursor-not-allowed' : ''}`}
    whileHover={{ scale: disabled ? 1 : 1.02, zIndex: 1 }}
    whileTap={{ scale: disabled ? 1 : 0.98 }}
    transition={{ type: 'spring', stiffness: 400, damping: 17 }}
    style={{ transformOrigin: 'center' }}
  >
    {isSelected && (
      <motion.div
        layoutId="analysis-type-selector"
        className={`absolute inset-0 ${colorMap[color].bg.replace('/50', '/20')} rounded-lg`}
        style={{ zIndex: -1 }}
      />
    )}
    <h4 className="font-semibold text-white">{label}</h4>
    <p className="text-sm text-gray-400 mt-1">{description}</p>
  </motion.button>
);

const ProjectInput: React.FC = () => {
  const {
    handleAnalyze,
    isAnalyzing,
    activeProject,
  } = useProjectContext();

  const { userSettings, integrations } = useUser();

  const [projectContext, setProjectContext] = useState('');
  const [projectName, setProjectName] = useState('');
  const [analysisType, setAnalysisType] = useState<AnalysisType>(AnalysisType.Architecture);
  const [isGithubModalOpen, setIsGithubModalOpen] = useState(false);
  const [isFetchingRepo, setIsFetchingRepo] = useState(false);
  const { addNotification } = useNotification();

  const hasPreviousAnalysis = !!activeProject && activeProject.history.length > 0;

  useEffect(() => {
    if (activeProject) {
      setProjectName(activeProject.name);
      // Maybe load last context file? For now, keep it simple.
      setProjectContext('');
    } else {
      setProjectName('');
      setProjectContext('');
    }
  }, [activeProject]);

  const analysisTypes = [
    { type: AnalysisType.Architecture, color: 'purple', label: "Architectural Review", description: "Analyzes high-level design and generates a visual diagram" },
    { type: AnalysisType.CodeQuality, color: 'teal', label: "Code Quality", description: "Evaluates patterns, maintainability, and adherence to principles like SOLID" },
    { type: AnalysisType.Security, color: 'red', label: "Security Analysis", description: "Focus on vulnerabilities, security practices, and compliance" },
    { type: AnalysisType.Scalability, color: 'blue', label: "Scalability Analysis", description: "Assessment of system growth capacity and performance" },
    { type: AnalysisType.Compliance, color: 'green', label: "Compliance & Best Practices", description: "Focus on accessibility (WCAG), data privacy, and industry standards" },
    { type: AnalysisType.DocumentationReview, color: 'amber', label: "Documentation Review", description: "Analysis of clarity, completeness, and structure of project documentation" },
    { type: AnalysisType.SelfCritique, color: 'pink', label: "Self-Critique (BETA)", description: "The AI reviews its own last analysis for quality and consistency.", disabled: !hasPreviousAnalysis },
  ];

  const handleImportFromGithub = async (owner: string, repo: string) => {
    setIsGithubModalOpen(false);
    setIsFetchingRepo(true);
    if (!activeProject) {
      setProjectName(`${owner}/${repo}`);
    }
    try {
      const githubPat = integrations?.github?.githubPat;
      if (!githubPat) {
        addNotification({ message: 'GitHub PAT is required for fetching repositories', type: 'error' });
        return;
      }
      const content = await fetchRepoContents(`https://github.com/${owner}/${repo}`, githubPat);
      setProjectContext(content);
      addNotification({ message: `Successfully imported repository: ${owner}/${repo}`, type: 'success' });
    } catch (error: any) {
      addNotification({ message: error.message, type: 'error' });
    } finally {
      setIsFetchingRepo(false);
    }
  };

  const handleUseExample = () => {
    setProjectName(exampleProject.name);
    setProjectContext(initialProjectContext);
    addNotification({ message: 'Example project context has been loaded into the form.', type: 'info' });
  }

  const handleTriggerAnalysis = () => {
    // For self-critique, the context is the previous analysis, handled in the context provider.
    // We pass an empty string for context here.
    const contextToSend = analysisType === AnalysisType.SelfCritique ? '' : projectContext;
    handleAnalyze(projectName, contextToSend, analysisType);
  }

  const isSelfCritique = analysisType === AnalysisType.SelfCritique;
  const canAnalyze = (isSelfCritique && hasPreviousAnalysis) ||
    (!isSelfCritique && projectContext.trim().length > 100 && (!activeProject ? projectName.trim().length > 2 : true))
    && !isAnalyzing;


  const placeholderText = isSelfCritique
    ? `The AI will now critique its latest analysis for the project "${activeProject?.name}". No context input is needed.`
    : "Paste your project documentation here...\n\n# Kortex Project\n## Overview\nKortex is a real-time monitoring dashboard...";

  return (
    <>
      <div className="h-full flex flex-col lg:flex-row gap-8 overflow-hidden">
        {/* Left Side: Input */}
        <motion.div
          className="lg:w-1/2 flex flex-col"
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
        >
          {!activeProject && (
            <div className="mb-4">
              <label htmlFor="projectName" className="text-lg font-semibold text-gray-300">Project Name</label>
              <input
                type="text"
                id="projectName"
                value={projectName}
                onChange={(e) => setProjectName(e.target.value)}
                placeholder="e.g., Kortex Project"
                className="w-full p-2 mt-1 bg-gray-900 border border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
              />
            </div>
          )}

          <div className="flex justify-between items-center mb-4">
            <h2 className="text-2xl font-bold flex items-center gap-3">
              {isSelfCritique ? <MessageSquareQuote className="text-pink-400" /> : <FileText className="text-blue-400" />}
              {isSelfCritique ? 'Critique Target' : 'Project Context'}
            </h2>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setIsGithubModalOpen(true)}
                disabled={isFetchingRepo || isSelfCritique}
                className="flex items-center gap-2 px-3 py-2 text-sm bg-gray-700/80 border border-gray-600 rounded-lg hover:bg-gray-700 disabled:opacity-50 transition-colors"
              >
                {isFetchingRepo ? <Loader2 className="w-4 h-4 animate-spin" /> : <Github className="w-4 h-4" />}
                Import from GitHub
              </button>
            </div>
          </div>
          <p className="text-gray-400 mb-4 text-sm">
            {isSelfCritique
              ? "The AI will analyze its own previous output for quality and consistency."
              : "Provide the project context below. You can paste documentation, READMEs, or any relevant text."
            }
          </p>
          <div className="flex-grow relative">
            <textarea
              ref={(el) => {
                if (el) {
                  el.style.height = '0px';
                  const scrollHeight = el.scrollHeight;
                  el.style.height = scrollHeight + 'px';
                }
              }}
              value={projectContext}
              onChange={(e) => setProjectContext(e.target.value)}
              placeholder={placeholderText}
              className="w-full h-full p-4 bg-gray-900/50 border border-gray-700 rounded-lg resize-none focus:outline-none focus:ring-2 focus:ring-purple-500 disabled:bg-gray-800/60"
              disabled={isSelfCritique}
            />
          </div>
          <button onClick={handleUseExample} disabled={isSelfCritique} className="text-sm text-blue-400 hover:underline mt-2 self-start disabled:opacity-50">Or use an example</button>
        </motion.div>

        {/* Right Side: Options */}
        <motion.div
          className="lg:w-1/2 flex flex-col"
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
        >
          <h2 className="text-2xl font-bold mb-4 flex items-center gap-3">
            <Wand2 className="text-purple-400" />
            Analysis Type
          </h2>
          <div className="space-y-3 flex-grow overflow-y-auto px-2">
            {analysisTypes.map(at => (
              <AnalysisTypeButton
                key={at.type}
                {...at}
                isSelected={analysisType === at.type}
                onClick={() => setAnalysisType(at.type)}
              />
            ))}
          </div>
          <motion.button
            onTap={!canAnalyze ? handleTriggerAnalysis : undefined}
            whileTap={{ scale: canAnalyze ? 0.95 : 1 }}
            className="w-full mt-4 py-3 px-6 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-lg font-semibold text-lg flex items-center justify-center gap-3 transition-all disabled:opacity-50 disabled:cursor-not-allowed hover:shadow-2xl hover:shadow-blue-500/30"
            whileHover={{ scale: canAnalyze ? 1.05 : 1 }}
          // disabled={!canAnalyze}
          >
            {isAnalyzing ? (
              <>
                <Loader2 className="w-6 h-6 animate-spin" />
                Analyzing...
              </>
            ) : (
              'Analyze Project'
            )}
          </motion.button>
        </motion.div>
      </div>
      <GitHubSearchModal
        isOpen={isGithubModalOpen}
        onClose={() => setIsGithubModalOpen(false)}
        onImport={handleImportFromGithub}
        githubPat={integrations?.github?.githubPat || ''}
      />
    </>
  );
};

export default ProjectInput;

// const colorMap: Record<string, { border: string; bg: string; hoverBorder: string }> = {
//   blue: { border: 'border-blue-600', bg: 'bg-blue-900/50', hoverBorder: 'hover:border-blue-500/80' },
//   red: { border: 'border-red-600', bg: 'bg-red-900/50', hoverBorder: 'hover:border-red-500/80' },
//   purple: { border: 'border-purple-600', bg: 'bg-purple-900/50', hoverBorder: 'hover:border-purple-500/80' },
//   teal: { border: 'border-teal-600', bg: 'bg-teal-900/50', hoverBorder: 'hover:border-teal-500/80' },
//   amber: { border: 'border-amber-600', bg: 'bg-amber-900/50', hoverBorder: 'hover:border-amber-500/80' },
//   green: { border: 'border-green-600', bg: 'bg-green-900/50', hoverBorder: 'hover:border-green-500/80' },
//   pink: { border: 'border-pink-600', bg: 'bg-pink-900/50', hoverBorder: 'hover:border-pink-500/80' },
// };

// const AnalysisTypeButton: React.FC<{
//   type: AnalysisType;
//   label: string;
//   description: string;
//   color: string;
//   isSelected: boolean;
//   onClick: () => void;
//   disabled?: boolean;
// }> = ({ type, label, description, color, isSelected, onClick, disabled = false }) => (
//   <motion.button
//     onTap={onClick}
//     className={`p-4 text-left border rounded-lg transition-all w-full relative ${isSelected
//       ? `${colorMap[color].bg} ${colorMap[color].border}`
//       : `bg-gray-800/50 border-gray-700 ${!disabled ? colorMap[color].hoverBorder : ''}`
//       } ${disabled ? 'opacity-50 cursor-not-allowed' : ''}`}
//     whileHover={{ scale: disabled ? 1 : 1.02, zIndex: 1 }}
//     whileTap={{ scale: disabled ? 1 : 0.98 }}
//     transition={{ type: 'spring', stiffness: 400, damping: 17 }}
//     style={{ transformOrigin: 'center' }}
//   >
//     {isSelected && (
//       <motion.div
//         layoutId="analysis-type-selector"
//         className={`absolute inset-0 ${colorMap[color].bg.replace('/50', '/20')} rounded-lg`}
//         style={{ zIndex: -1 }}
//       />
//     )}
//     <h4 className="font-semibold text-white">{label}</h4>
//     <p className="text-sm text-gray-400 mt-1">{description}</p>
//   </motion.button>
// );

// const ProjectInput: React.FC = () => {
//   const {
//     handleAnalyze,
//     isAnalyzing,
//     activeProject,
//   } = useProjectContext();

//   const { userSettings, integrations } = useUser();

//   const [projectContext, setProjectContext] = useState('');
//   const [projectContextFile, setProjectContextFile] = useState<File | null>(null);
//   const [projectName, setProjectName] = useState('');
//   const [analysisType, setAnalysisType] = useState<AnalysisType>(AnalysisType.Architecture);
//   const [isGithubModalOpen, setIsGithubModalOpen] = useState(false);
//   const [isFetchingRepo, setIsFetchingRepo] = useState(false);
//   const { addNotification } = useNotification();

//   const hasPreviousAnalysis = !!activeProject && activeProject.history.length > 0;

//   useEffect(() => {
//     if (activeProject) {
//       setProjectName(activeProject.name);
//       // Maybe load last context file? For now, keep it simple.
//       setProjectContext('');
//     } else {
//       setProjectName('');
//       setProjectContext('');
//     }
//   }, [activeProject]);

//   const analysisTypes = [
//     { type: AnalysisType.Architecture, color: 'purple', label: "Architectural Review", description: "Analyzes high-level design and generates a visual diagram" },
//     { type: AnalysisType.CodeQuality, color: 'teal', label: "Code Quality", description: "Evaluates patterns, maintainability, and adherence to principles like SOLID" },
//     { type: AnalysisType.Security, color: 'red', label: "Security Analysis", description: "Focus on vulnerabilities, security practices, and compliance" },
//     { type: AnalysisType.Scalability, color: 'blue', label: "Scalability Analysis", description: "Assessment of system growth capacity and performance" },
//     { type: AnalysisType.Compliance, color: 'green', label: "Compliance & Best Practices", description: "Focus on accessibility (WCAG), data privacy, and industry standards" },
//     { type: AnalysisType.DocumentationReview, color: 'amber', label: "Documentation Review", description: "Analysis of clarity, completeness, and structure of project documentation" },
//     { type: AnalysisType.Scalability, color: 'blue', label: "Scalability Analysis", description: "Assessment of system growth capacity and performance" },
//     { type: AnalysisType.Compliance, color: 'green', label: "Compliance & Best Practices", description: "Focus on accessibility (WCAG), data privacy, and industry standards" },
//     { type: AnalysisType.DocumentationReview, color: 'amber', label: "Documentation Review", description: "Analysis of clarity, completeness, and structure of project documentation" },
//     { type: AnalysisType.SelfCritique, color: 'pink', label: "Self-Critique (BETA)", description: "The AI reviews its own last analysis for quality and consistency.", disabled: !hasPreviousAnalysis },
//   ];

//   const handleImportFromGithub = async (owner: string, repo: string) => {
//     setIsGithubModalOpen(false);
//     setIsFetchingRepo(true);
//     if (!activeProject) {
//       setProjectName(`${owner}/${repo}`);
//     }
//     try {
//       const githubPat = integrations?.github?.githubPat;
//       if (!githubPat) {
//         addNotification({ message: 'GitHub PAT is required for fetching repositories', type: 'error' });
//         return;
//       }
//       const content = await fetchRepoContents(`
//         {
//           "owner": "${owner}",
//           "repo": "${repo}",
//           "path": "context.json"
//         }
//       `, githubPat);
//       if (!content) {
//         addNotification({ message: 'Failed to fetch repository contents', type: 'error' });
//         return;
//       }
//       setProjectContext(content);
//     } catch (error) {
//       console.error('Error importing from GitHub:', error);
//       addNotification({ message: 'Error importing from GitHub', type: 'error' });
//     } finally {
//       setIsFetchingRepo(false);
//     }
//   };
//   const handleUseExample = () => {
//     setProjectName(exampleProject.name);
//     setProjectContext(initialProjectContext);
//     addNotification({ message: 'Example project context has been loaded into the form.', type: 'info' });
//   }

//   const handleTriggerAnalysis = () => {
//     // For self-critique, the context is the previous analysis, handled in the context provider.
//     // We pass an empty string for context here.
//     const contextToSend = analysisType === AnalysisType.SelfCritique ? '' : projectContext;
//     handleAnalyze(projectName, contextToSend, analysisType);
//   }

//   const isSelfCritique = analysisType === AnalysisType.SelfCritique;
//   const canAnalyze = (isSelfCritique && hasPreviousAnalysis) ||
//     (!isSelfCritique && projectContext.trim().length > 100 && (!activeProject ? projectName.trim().length > 2 : true))
//     && !isAnalyzing;

//   const placeholderText = isSelfCritique
//     ? `The AI will now critique its latest analysis for the project "${activeProject?.name}". No context input is needed.`
//     : "Paste your project documentation here...\n\n# Kortex Project\n## Overview\nKortex is a real-time monitoring dashboard...";

//   return (
//     <>
//       <div className="h-full flex flex-col lg:flex-row gap-8 overflow-hidden">
//         {/* Left Side: Input */}
//         <motion.div
//           className="lg:w-1/2 flex flex-col"
//           initial={{ opacity: 0, x: -20 }}
//           animate={{ opacity: 1, x: 0 }}
//         >
//           {!activeProject && (
//             <div className="mb-4">
//               <label htmlFor="projectName" className="text-lg font-semibold text-gray-300">Project Name</label>
//               <input
//                 type="text"
//                 id="projectName"
//                 value={projectName}
//                 onChange={(e) => setProjectName(e.target.value)}
//                 placeholder="e.g., Kortex Project"
//                 className="w-full p-2 mt-1 bg-gray-900 border border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
//               />
//             </div>
//           )}

//           <div className="flex justify-between items-center mb-4">
//             <h2 className="text-2xl font-bold flex items-center gap-3">
//               {isSelfCritique ? <MessageSquareQuote className="text-pink-400" /> : <FileText className="text-blue-400" />}
//               {isSelfCritique ? 'Critique Target' : 'Project Context'}
//             </h2>
//             <div className="flex items-center gap-2">
//               <button
//                 onClick={() => setIsGithubModalOpen(true)}
//                 disabled={isFetchingRepo || isSelfCritique}
//                 className="flex items-center gap-2 px-3 py-2 text-sm bg-gray-700/80 border border-gray-600 rounded-lg hover:bg-gray-700 disabled:opacity-50 transition-colors"
//               >
//                 {isFetchingRepo ? <Loader2 className="w-4 h-4 animate-spin" /> : <Github className="w-4 h-4" />}
//                 Import from <GitHub className="text-blue-400" /> GitHub
//               </button>
//             </div>
//           </div>
//           <p className="text-gray-400 mb-4 text-sm">
//             {isSelfCritique
//               ? "The AI will analyze its own previous output for quality and consistency."
//               : "Provide the project context below. You can paste documentation, READMEs, or any relevant text."
//             }
//           </p>
//           <div className="flex-grow relative">
//             <textarea
//               ref={(el) => {
//                 if (el) {
//                   el.style.height = '0px';
//                   const scrollHeight = el.scrollHeight;
//                   el.style.height = scrollHeight + 'px';
//                 }
//               }}
//               value={projectContext}
//               onChange={(e) => setProjectContext(e.target.value)}
//               placeholder={placeholderText}
//               className="w-full h-full p-4 bg-gray-900/50 border border-gray-700 rounded-lg resize-none focus:outline-none focus:ring-2 focus:ring-purple-500 disabled:bg-gray-800/60"
//               disabled={isSelfCritique}
//             />
//           </div>
//           <button onClick={handleUseExample} disabled={isSelfCritique} className="text-sm text-blue-400 hover:underline mt-2 self-start disabled:opacity-50">Or use an example</button>
//         </motion.div>

//         {/* Right Side: Options */}
//         <motion.div
//           className="lg:w-1/2 flex flex-col"
//           initial={{ opacity: 0, x: 20 }}
//           animate={{ opacity: 1, x: 0 }}
//         >
//           <h2 className="text-2xl font-bold mb-4 flex items-center gap-3">
//             <Wand2 className="text-purple-400" />
//             Analysis Type
//           </h2>
//           <div className="space-y-3 flex-grow overflow-y-auto px-2">
//             {analysisTypes.map(at => (
//               <AnalysisTypeButton
//                 key={at.type}
//                 {...at}
//                 isSelected={analysisType === at.type}
//                 onClick={() => setAnalysisType(at.type)}
//               />
//             ))}
//           </div>
//           <motion.button
//             onTap={!canAnalyze ? handleTriggerAnalysis : undefined}
//             whileTap={{ scale: canAnalyze ? 0.95 : 1 }}
//             className="w-full mt-4 py-3 px-6 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-lg font-semibold text-lg flex items-center justify-center gap-3 transition-all disabled:opacity-50 disabled:cursor-not-allowed hover:shadow-2xl hover:shadow-blue-500/30"
//             whileHover={{ scale: canAnalyze ? 1.05 : 1 }}
//           // disabled={!canAnalyze}
//           >
//             {isAnalyzing ? (
//               <>
//                 <Loader2 className="w-6 h-6 animate-spin" />
//                 Analyzing...
//               </>
//             ) : (
//               'Analyze Project'
//             )}
//           </motion.button>
//         </motion.div>
//       </div>
//       <GitHubSearchModal
//         isOpen={isGithubModalOpen}
//         onClose={() => setIsGithubModalOpen(false)}
//         onImport={handleImportFromGithub}
//         githubPat={integrations?.github?.githubPat || ''}
//       />
//     </>
//   );
// };

// export default ProjectInput;
