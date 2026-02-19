import { KanbanTranslations } from '../types';

export const kanbanEnUS: KanbanTranslations = {
  title: "Kanban Board",
  addCard: "Add Card",
  editCard: "Edit Card",
  exampleModeNotice: "This is an example Kanban board. Changes may not persist across sessions.",
  notes: "Notes",
  notesPlaceholder: "Add any extra notes or details here...",
  deleteConfirm: {
    title: "Delete Card",
    message: "Are you sure you want to delete this card? This action cannot be undone.",
    confirm: "Delete",
  },
  columns: {
    todo: "TODO",
    inProgress: "In Progress",
    done: "Done",
    blocked: "Blocked"
  }
};