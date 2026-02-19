import * as React from 'react';

import { AnimatePresence, motion } from 'framer-motion';
import { GitCompareArrows, History as HistoryIcon, Trash2, X } from 'lucide-react';
import { useState } from 'react';
import { useProjectContext } from '../../contexts/ProjectContext';

const HistoryPanel: React.FC = () => {
  const {
    handleImportHistory,
    handleExportHistory,

    isHistoryPanelOpen,
    setIsHistoryPanelOpen,
    activeProject,
    handleSelectHistoryItem,
    handleCompareHistoryItems,
    handleDeleteHistoryItem,
  } = useProjectContext();

  const [selectedForCompare, setSelectedForCompare] = useState<number[]>([]);

  const history = activeProject?.history || [];

  const toggleCompareSelection = (id: number) => {
    setSelectedForCompare(prev => {
      if (prev.includes(id)) {
        return prev.filter(item => item !== id);
      }
      if (prev.length < 2) {
        return [...prev, id];
      }
      return [prev[1], id];
    });
  };

  const canCompare = selectedForCompare.length === 2;

  const handleCompareClick = () => {
    if (canCompare) {
      handleCompareHistoryItems(selectedForCompare[0], selectedForCompare[1]);
      setSelectedForCompare([]);
    }
  }

  const handleImportHistoryClick = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e) return;
    if (e.target.files && e.target.files.length > 0) {
      const file = e.target.files[0];
      try {
        await handleImportHistory(file);
        e.target.value = ''; // Reset the input
      } catch (error) {
        console.error('Error importing history:', error);
        alert('Failed to import history. Please check the file format.');
      }
    }
  }

  const handleExportHistoryClick = async (e: any) => {
    if (!e) return;
    if (e.activeProject && e.activeProject.id) {
      try {
        const exported = await handleExportHistory(e.activeProject.id);
      } catch (error) {
        console.error('Error exporting history:', error);
        alert('Failed to export history. Please try again later.');
      }
    }
  };

  return (
    <AnimatePresence>
      {isHistoryPanelOpen && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          onClick={() => setIsHistoryPanelOpen(false)}
          className="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 flex justify-end"
        >
          <motion.div
            initial={{ x: '100%' }}
            animate={{ x: '0%' }}
            exit={{ x: '100%' }}
            transition={{ type: 'spring', stiffness: 300, damping: 30 }}
            onClick={(e) => e.stopPropagation()}
            className="bg-gray-800 border-l border-gray-700 w-full max-w-md h-full flex flex-col shadow-2xl"
          >
            {/* Header */}
            <div className="flex items-center justify-between p-4 border-b border-gray-700">
              <div className="flex items-center gap-3">
                <HistoryIcon className="w-6 h-6 text-blue-400" />
                <h2 className="text-xl font-bold text-white">{`History for ${activeProject?.name}`}</h2>
              </div>
              <button title='Close history panel' onClick={() => setIsHistoryPanelOpen(false)} className="p-1 rounded-full text-gray-400 hover:bg-gray-700">
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Content */}
            <div className="flex-grow p-4 overflow-y-auto space-y-3">
              {history.length > 0 ? (
                history.map(item => {
                  const isSelected = selectedForCompare.includes(item.id);
                  return (
                    <div key={item.id} className={`p-3 rounded-lg flex items-center gap-3 transition-colors duration-200 ${isSelected ? 'bg-purple-900/50 ring-2 ring-purple-500' : 'bg-gray-900/50'}`}>
                      <input
                        title='Select for comparison'
                        type="checkbox"
                        checked={isSelected}
                        onChange={() => toggleCompareSelection(item.id)}
                        className="w-4 h-4 rounded bg-gray-700 border-gray-600 text-purple-600 focus:ring-purple-500 shrink-0"
                      />
                      <div className="flex-grow cursor-pointer" onClick={() => handleSelectHistoryItem(item.id)}>
                        <p className="font-semibold text-white truncate">{item.analysis.projectName}</p>
                        <p className="text-xs text-gray-400">
                          {new Date(item.timestamp).toLocaleString('en-US', { dateStyle: 'short', timeStyle: 'short' })} - {item.analysis.analysisType}
                        </p>
                      </div>
                      <button
                        onClick={() => handleDeleteHistoryItem(item.id)}
                        className="p-2 text-gray-500 hover:text-red-400 hover:bg-red-900/30 rounded-full"
                        aria-label="Delete history item"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  )
                })
              ) : (
                <div className="text-center text-gray-500 pt-10">
                  <HistoryIcon className="w-12 h-12 mx-auto mb-4" />
                  <p>No history for this project yet.</p>
                </div>
              )}
            </div>

            {/* Footer */}
            {/* {history.length > 0 && ( */}
            <div className="p-4 bg-gray-900/50 border-t border-gray-700 space-y-3">
              <div className="flex gap-2">
                <label
                  htmlFor="import-history"
                  className="w-full cursor-pointer px-4 py-2 text-sm font-semibold text-white bg-gray-700 border border-gray-600 rounded-md hover:bg-gray-700/80 disabled:opacity-50 disabled:cursor-not-allowed align-middle text-center"
                >
                  Import History
                  <input
                    id="import-history"
                    type="file"
                    accept=".json"
                    onChange={handleImportHistoryClick}
                    className="hidden"
                  />
                </label>

                <button
                  title='Export History'
                  onClick={handleExportHistoryClick}
                  disabled={!activeProject || history.length === 0}
                  className="w-full px-4 py-2 text-sm font-semibold text-white bg-gray-700 border border-gray-600 rounded-md hover:bg-gray-700/80 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Export History
                </button>
                <button
                  type='button'
                  title='Clear Selection'
                  onClick={() => setSelectedForCompare([])}
                  disabled={selectedForCompare.length === 0}
                  className="w-full px-4 py-2 text-sm font-semibold text-white bg-gray-700 border border-gray-600 rounded-md hover:bg-gray-700/80 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Clear Selection
                </button>
              </div>

              <button
                onClick={handleCompareClick}
                disabled={!canCompare}
                className="w-full flex items-center justify-center gap-2 px-4 py-2 text-sm font-semibold text-white bg-purple-600 rounded-md hover:bg-purple-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <GitCompareArrows className="w-4 h-4" /> {`Compare Selected (${selectedForCompare.length})`}
              </button>
            </div>
            {/* )} */}
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default HistoryPanel;
