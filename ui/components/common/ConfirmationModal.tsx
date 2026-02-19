import { AnimatePresence, motion } from 'framer-motion';
import { AlertTriangle, X } from 'lucide-react';
import * as React from 'react';
import { useConfirmation } from '../../contexts/ConfirmationContext';

const ConfirmationModal: React.FC = () => {
  const { isOpen, options, hideConfirmation } = useConfirmation();

  const handleConfirm = () => {
    options?.onConfirm();
    hideConfirmation();
  };

  const handleCancel = () => {
    options?.onCancel?.();
    hideConfirmation();
  };

  return (
    <AnimatePresence>
      {isOpen && options && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          onClick={handleCancel}
          className="fixed inset-0 bg-black/70 backdrop-blur-sm z-50 flex items-center justify-center p-4"
        >
          <motion.div
            initial={{ scale: 0.9, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            exit={{ scale: 0.9, opacity: 0 }}
            transition={{ type: 'spring', stiffness: 300, damping: 25 }}
            onClick={(e) => e.stopPropagation()}
            className="bg-gray-800 border border-gray-700 rounded-xl w-full max-w-md flex flex-col shadow-2xl relative"
          >
            <div className="p-6 flex items-start gap-4">
              <div className="w-10 h-10 bg-red-900/50 rounded-full flex items-center justify-center shrink-0 border border-red-800">
                <AlertTriangle className="w-5 h-5 text-red-400" />
              </div>
              <div className="flex-grow">
                <h2 className="text-xl font-bold text-white">{options.title}</h2>
                <p className="mt-2 text-gray-300">{options.message}</p>
              </div>
              <button title='Close' onClick={handleCancel} className="p-1 rounded-full text-gray-400 hover:bg-gray-700 transition-colors absolute top-4 right-4">
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="p-4 bg-gray-900/50 flex justify-end gap-3 rounded-b-xl">
              <button
                onClick={handleCancel}
                className="px-4 py-2 text-sm font-semibold text-gray-200 bg-gray-700 rounded-md hover:bg-gray-600"
              >
                {options.cancelText || 'Cancel'}
              </button>
              <button
                onClick={handleConfirm}
                className="px-4 py-2 text-sm font-semibold text-white bg-red-600 rounded-md hover:bg-red-700"
              >
                {options.confirmText || 'Confirm'}
              </button>
            </div>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default ConfirmationModal;
