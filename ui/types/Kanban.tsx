import { Difficulty, Priority } from "./Enums";

// Kanban types
export type KanbanColumnId = 'backlog' | 'todo' | 'inProgress' | 'done';

export interface KanbanCard {
  id: string;
  title: string;
  description: string;
  priority: Priority;
  difficulty: Difficulty;
  tags?: string[];
  notes?: string;
}

export interface KanbanColumn {
  id: KanbanColumnId;
  title: string;
  cardIds: string[];
}

export interface KanbanState {
  cards: Record<string, KanbanCard>;
  columns: Record<KanbanColumnId, KanbanColumn>;
  columnOrder: KanbanColumnId[];
}

export interface KanbanTaskSuggestion {
  title: string;
  description: string;
  priority: Priority;
  difficulty: Difficulty;
  tags: string[];
}
