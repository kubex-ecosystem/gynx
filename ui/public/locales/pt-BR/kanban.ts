import { KanbanTranslations } from '../types';

export const kanbanPtBR: KanbanTranslations = {
  title: "Quadro Kanban",
  addCard: "Adicionar Cartão",
  editCard: "Editar Cartão",
  exampleModeNotice: "Este é um quadro Kanban de exemplo. As alterações podem não persistir entre sessões.",
  notes: "Notas",
  notesPlaceholder: "Adicione quaisquer notas ou detalhes extras aqui...",
  deleteConfirm: {
    title: "Excluir Cartão",
    message: "Tem certeza de que deseja excluir este cartão? Esta ação não pode ser desfeita.",
    confirm: "Excluir",
  },
  columns: {
    todo: "A Fazer",
    inProgress: "Em Progresso",
    done: "Concluído",
    blocked: "Bloqueado"
  }
};