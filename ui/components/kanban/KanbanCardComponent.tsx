import { motion, PanInfo } from 'framer-motion';
import { Edit2 } from 'lucide-react';
import * as React from 'react';
import { KanbanCard, KanbanColumnId, Priority } from '../../types';
import DifficultyMeter from '../common/DifficultyMeter';

interface KanbanCardProps {
  card: KanbanCard;
  onEdit: () => void;
  onDragStart: (cardId: string, columnId: KanbanColumnId) => void;
  onDrag: (event: MouseEvent | TouchEvent | PointerEvent, info: PanInfo) => void;
  onDragEnd: (event: MouseEvent | TouchEvent | PointerEvent, info: PanInfo) => void;
}

const priorityColors: Record<Priority, string> = {
  [Priority.High]: 'bg-red-500',
  [Priority.Medium]: 'bg-yellow-500',
  [Priority.Low]: 'bg-blue-500',
};

const KanbanCardComponent: React.FC<KanbanCardProps> = ({ card, onEdit, onDragStart, onDrag, onDragEnd }) => {
  return (
    <motion.div
      layout
      drag
      dragConstraints={{ top: 0, left: 0, right: 0, bottom: 0 }}
      dragElastic={1}
      onDragStart={() => onDragStart(card.id, 'backlog' /* This is a placeholder, context should provide column */)}
      onDrag={onDrag}
      onDragEnd={onDragEnd}
      className="p-3 bg-gray-800 border border-gray-700 rounded-md cursor-grab active:cursor-grabbing"
    >
      <div className="flex justify-between items-start">
        <h4 className="text-sm font-semibold text-gray-200">{card.title}</h4>
        <button title='Edit' onClick={onEdit} className="p-1 text-gray-500 hover:text-white">
          <Edit2 className="w-3 h-3" />
        </button>
      </div>
      <p className="text-xs text-gray-400 mt-1 line-clamp-2">{card.description}</p>
      <div className="mt-3 flex items-center justify-between">
        <DifficultyMeter difficulty={card.difficulty} />
        <div className={`w-3 h-3 rounded-full ${priorityColors[card.priority]}`} title={`Priority: ${card.priority}`} />
      </div>
      {card.tags && card.tags.length > 0 && (
        <div className="mt-2 flex flex-wrap gap-1">
          {card.tags.map(tag => (
            <span key={tag} className="px-1.5 py-0.5 text-xs bg-gray-700 text-gray-300 rounded">
              {tag}
            </span>
          ))}
        </div>
      )}
    </motion.div>
  );
};

export default KanbanCardComponent;
