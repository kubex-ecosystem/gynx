import * as React from 'react';
import { useState, useEffect, useRef } from 'react';
import { motion, PanInfo } from 'framer-motion';
import { Plus, Kanban, Info } from 'lucide-react';
import { useProjectContext } from '../../contexts/ProjectContext';
import KanbanCardComponent from './KanbanCardComponent';
import { KanbanCard, KanbanColumn, KanbanColumnId, Priority, Difficulty, KanbanState } from '../../types';
import EditCardModal from './EditCardModal';
import { v4 as uuidv4 } from 'uuid';

interface KanbanColumnProps {
  column: KanbanColumn;
  cards: KanbanCard[];
  onCardEdit: (card: KanbanCard) => void;
  // FIX: Corrected prop name to match KanbanCardComponent
  onDragStart: (cardId: string, columnId: KanbanColumnId) => void;
  onDragMotion: (event: MouseEvent | TouchEvent | PointerEvent, info: PanInfo) => void;
  onDragMotionEnd: (event: MouseEvent | TouchEvent | PointerEvent, info: PanInfo) => void;
}

const KanbanColumnComponent: React.FC<KanbanColumnProps> = ({ column, cards, onCardEdit, onDragStart, onDragMotion, onDragMotionEnd }) => {
    return (
        // FIX: Removed invalid `ref` callback return value
        <div data-kanban-column-id={column.id} className="w-72 bg-gray-900/50 border border-gray-800 rounded-lg p-2 flex flex-col shrink-0 h-full">
            <h3 className="text-md font-semibold text-gray-300 px-2 py-1 mb-2">{column.title} ({cards.length})</h3>
            <div className="flex-grow min-h-[100px] space-y-2 overflow-y-auto pr-1">
                {cards.map(card => (
                    <KanbanCardComponent
                        key={card.id}
                        card={card}
                        onEdit={() => onCardEdit(card)}
                        // FIX: Corrected prop names to match component
                        onDragStart={() => onDragStart(card.id, column.id)}
                        onDrag={onDragMotion}
                        onDragEnd={onDragMotionEnd}
                    />
                ))}
            </div>
        </div>
    );
};


const KanbanBoard: React.FC = () => {
    const { kanbanState, setKanbanState, isExample } = useProjectContext();

    const [isEditingCard, setIsEditingCard] = useState<KanbanCard | Omit<KanbanCard, 'id'> | null>(null);
    const [isClient, setIsClient] = useState(false);

    const [draggedItem, setDraggedItem] = useState<{cardId: string, sourceColumnId: KanbanColumnId} | null>(null);
    const columnRefs = useRef<Record<string, HTMLDivElement | null>>({});

    useEffect(() => {
        setIsClient(true);
    }, []);

    const handleSaveCard = (cardToSave: KanbanCard | Omit<KanbanCard, 'id'>) => {
        // FIX: Refactored to not use function updater with setKanbanState
        if (!kanbanState) return;

        if (!('id' in cardToSave) || !cardToSave.id) {
            // New Card
            const newId = uuidv4();
            const newCard: KanbanCard = {
                ...(cardToSave as Omit<KanbanCard, 'id'>),
                id: newId,
                tags: (cardToSave as KanbanCard).tags || [],
                priority: (cardToSave as KanbanCard).priority || Priority.Medium,
                difficulty: (cardToSave as KanbanCard).difficulty || Difficulty.Medium,
            };
            
            const newBacklog = { ...kanbanState.columns.backlog };
            newBacklog.cardIds = [newId, ...newBacklog.cardIds];
            
            const newState: KanbanState = {
                ...kanbanState,
                cards: { ...kanbanState.cards, [newId]: newCard },
                columns: { ...kanbanState.columns, backlog: newBacklog },
            };
            setKanbanState(newState);
        } else {
            // Existing Card
            const newState: KanbanState = {
                ...kanbanState,
                cards: { ...kanbanState.cards, [cardToSave.id]: cardToSave as KanbanCard },
            };
            setKanbanState(newState);
        }
        setIsEditingCard(null);
    };

    const handleDeleteCard = (cardId: string) => {
        // FIX: Refactored to not use function updater with setKanbanState
        if (!kanbanState) return;

        const newCards = { ...kanbanState.cards };
        delete newCards[cardId];
        const newColumns = { ...kanbanState.columns };
        Object.keys(newColumns).forEach(key => {
            const colId = key as KanbanColumnId;
            newColumns[colId].cardIds = newColumns[colId].cardIds.filter(id => id !== cardId);
        });
        const newState: KanbanState = { ...kanbanState, cards: newCards, columns: newColumns };
        setKanbanState(newState);
        setIsEditingCard(null);
    };
    
    const handleDragStart = (cardId: string, sourceColumnId: KanbanColumnId) => {
        setDraggedItem({ cardId, sourceColumnId });
    };

    const handleDragEnd = (event: MouseEvent | TouchEvent | PointerEvent, info: PanInfo) => {
        if (!draggedItem || !kanbanState) return;

        const pointer = { x: info.point.x, y: info.point.y };
        let targetColumnId: KanbanColumnId | null = null;
        
        kanbanState.columnOrder.forEach(colId => {
            const colElement = columnRefs.current[colId];
            if (colElement) {
                const rect = colElement.getBoundingClientRect();
                if (pointer.x > rect.left && pointer.x < rect.right && pointer.y > rect.top && pointer.y < rect.bottom) {
                    targetColumnId = colId as KanbanColumnId;
                }
            }
        });
        
        if (targetColumnId && targetColumnId !== draggedItem.sourceColumnId) {
            // FIX: Refactored to not use function updater with setKanbanState
            const sourceCol = { ...kanbanState.columns[draggedItem.sourceColumnId] };
            const targetCol = { ...kanbanState.columns[targetColumnId] };
            
            sourceCol.cardIds = sourceCol.cardIds.filter(id => id !== draggedItem.cardId);
            // This logic is a bit naive, should insert at a specific index based on pointer.y
            // For now, just adding to the end is fine.
            targetCol.cardIds.push(draggedItem.cardId);
            
            const newState: KanbanState = {
                ...kanbanState,
                columns: {
                    ...kanbanState.columns,
                    [draggedItem.sourceColumnId]: sourceCol,
                    [targetColumnId]: targetCol
                }
            };
            setKanbanState(newState);
        }

        setDraggedItem(null);
    };

    if (!isClient || !kanbanState) return null;

    return (
        <div className="h-full flex flex-col">
            <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-3">
                    <Kanban className="w-7 h-7 text-teal-400" />
                    <h2 className="text-2xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-teal-400">Kanban Board</h2>
                </div>
                <button
                    onClick={() => setIsEditingCard({} as Omit<KanbanCard, 'id'>)}
                    className="flex items-center gap-2 px-3 py-2 text-sm bg-purple-600 text-white rounded-lg hover:bg-purple-700"
                >
                    <Plus className="w-4 h-4" /> Add Card
                </button>
            </div>
            {isExample && (
                 <div className="p-3 mb-4 bg-purple-900/50 border border-purple-700 text-purple-300 rounded-lg flex items-center gap-3 text-sm">
                    <Info className="w-5 h-5 shrink-0" />
                    <p>This is an example Kanban board. Changes may not persist across sessions.</p>
                </div>
            )}
            <div className="flex-grow overflow-x-auto pb-4">
                <div className="flex gap-4 h-full">
                    {kanbanState.columnOrder.map(columnId => {
                        const column = kanbanState.columns[columnId];
                        const cards = column.cardIds.map(cardId => kanbanState.cards[cardId]).filter(Boolean);
                        return (
                            // FIX: Corrected ref callback to not return a value
                            <div key={column.id} ref={(el): void => { columnRefs.current[column.id] = el; }} className="h-full">
                                <KanbanColumnComponent
                                    column={column}
                                    cards={cards}
                                    onCardEdit={(card) => setIsEditingCard(card)}
                                    onDragStart={handleDragStart}
                                    onDragMotion={() => {}} // onDrag logic can be added here if needed
                                    onDragMotionEnd={handleDragEnd}
                                />
                            </div>
                        );
                    })}
                </div>
            </div>
            <EditCardModal
                isOpen={!!isEditingCard}
                onClose={() => setIsEditingCard(null)}
                card={isEditingCard}
                onSave={handleSaveCard}
                onDelete={handleDeleteCard}
                isExample={isExample}
            />
        </div>
    );
};

export default KanbanBoard;