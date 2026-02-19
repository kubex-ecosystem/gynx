import { AnimatePresence, motion } from 'framer-motion';
import { Trash2, X } from 'lucide-react';
import * as React from 'react';
import { useEffect, useState } from 'react';
import { useConfirmation } from '../../contexts/ConfirmationContext';
import { Difficulty, KanbanCard, Priority } from '../../types';

interface EditCardModalProps {
  isOpen: boolean;
  onClose: () => void;
  card: KanbanCard | Omit<KanbanCard, 'id'> | null;
  onSave: (card: KanbanCard | Omit<KanbanCard, 'id'>) => void;
  onDelete: (cardId: string) => void;
  isExample: boolean;
}

const EditCardModal: React.FC<EditCardModalProps> = ({ isOpen, onClose, card, onSave, onDelete, isExample }) => {
  const { showConfirmation } = useConfirmation();

  const [formData, setFormData] = useState<KanbanCard | Omit<KanbanCard, 'id'>>({
    title: '',
    description: '',
    priority: Priority.Medium,
    difficulty: Difficulty.Medium,
    tags: [],
    notes: ''
  });

  useEffect(() => {
    if (card) {
      setFormData(card);
    }
  }, [card]);

  const isNewCard = !('id' in (card || {}));

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleSave = () => {
    onSave(formData);
  };

  const handleDelete = () => {
    if ('id' in formData) {
      showConfirmation({
        title: "Delete Card",
        message: "Are you sure you want to delete this card? This action cannot be undone.",
        confirmText: "Delete",
        onConfirm: () => onDelete(formData.id!),
      });
    }
  };

  return (
    <AnimatePresence>
      {isOpen && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          onClick={onClose}
          className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4"
        >
          <motion.div
            initial={{ scale: 0.95, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            exit={{ scale: 0.95, opacity: 0 }}
            onClick={(e) => e.stopPropagation()}
            className="bg-gray-800 border border-gray-700 rounded-xl w-full max-w-lg flex flex-col shadow-2xl"
          >
            {/* Header */}
            <div className="flex items-center justify-between p-4 border-b border-gray-700">
              <h2 className="text-xl font-bold text-white">{isNewCard ? 'Add Card' : 'Edit Card'}</h2>
              <button title='Close' onClick={onClose} className="p-1 rounded-full text-gray-400 hover:bg-gray-700">
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Form */}
            <div className="p-6 space-y-4 overflow-y-auto">
              <div>
                <label htmlFor="title" className="text-sm font-medium text-gray-300">Title</label>
                <input title='Title' type="text" name="title" value={formData.title} onChange={handleChange} className="w-full p-2 mt-1 bg-gray-900 border border-gray-600 rounded-md" />
              </div>
              <div>
                <label className="text-sm font-medium text-gray-300">Description</label>
                <p className="text-xs text-gray-500 p-2 bg-gray-900 rounded-md mt-1">{formData.description}</p>
              </div>
              <div>
                <label htmlFor="notes" className="text-sm font-medium text-gray-300">Notes</label>
                <textarea name="notes" value={formData.notes || ''} onChange={handleChange} placeholder="Add any extra notes or details here..." rows={4} className="w-full p-2 mt-1 bg-gray-900 border border-gray-600 rounded-md" />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label htmlFor="priority" className="text-sm font-medium text-gray-300">Priority</label>
                  <select title='Priority' name="priority" value={formData.priority} onChange={handleChange} className="w-full p-2 mt-1 bg-gray-900 border border-gray-600 rounded-md">
                    {/* FIX: Explicitly cast enum values to prevent type errors in .map() */}
                    {Object.values(Priority).map(p => <option key={p as string} value={p as string}>{p as string}</option>)}
                  </select>
                </div>
                <div>
                  <label htmlFor="difficulty" className="text-sm font-medium text-gray-300">Difficulty</label>
                  <select title='Difficulty' name="difficulty" value={formData.difficulty} onChange={handleChange} className="w-full p-2 mt-1 bg-gray-900 border border-gray-600 rounded-md">
                    {/* FIX: Explicitly cast enum values to prevent type errors in .map() */}
                    {Object.values(Difficulty).map(d => <option key={d as string} value={d as string}>{d as string}</option>)}
                  </select>
                </div>
              </div>
            </div>

            {/* Footer */}
            <div className="p-4 bg-gray-900/50 flex justify-between items-center rounded-b-xl">
              <div>
                {!isNewCard && (
                  <button title='Delete' onClick={handleDelete} className="p-2 text-gray-400 hover:text-red-400 hover:bg-red-900/30 rounded-full" disabled={isExample}>
                    <Trash2 className="w-5 h-5" />
                  </button>
                )}
              </div>
              <div className="flex gap-3">
                <button onClick={onClose} className="px-4 py-2 text-sm font-semibold text-gray-200 bg-gray-700 rounded-md hover:bg-gray-600">Cancel</button>
                <button onClick={handleSave} className="px-4 py-2 text-sm font-semibold text-white bg-purple-600 rounded-md hover:bg-purple-700" disabled={isExample}>Save</button>
              </div>
            </div>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default EditCardModal;
