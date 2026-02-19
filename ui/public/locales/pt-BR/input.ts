import { InputTranslations } from '../types';

export const inputPtBR: InputTranslations = {
  title: "Contexto do Projeto",
  projectName: "Nome do Projeto",
  projectNamePlaceholder: "Ex: Projeto Kortex",
  importFromGithub: "Importar do GitHub",
  description: "Forneça o contexto do projeto abaixo. Você pode colar documentação, READMEs ou qualquer texto relevante.",
  placeholder: "Cole a documentação do seu projeto aqui...\n\n# Projeto Kortex\n## Visão Geral\nKortex é um painel de monitoramento em tempo real...",
  useExample: "Ou use um exemplo",
  analysisTypeTitle: "Tipo de Análise",
  analysisTypes: {
    GENERAL: {
      label: "Análise Geral",
      description: "Avaliação abrangente de arquitetura, qualidade e viabilidade do projeto"
    },
    SECURITY: {
      label: "Análise de Segurança",
      description: "Foco em vulnerabilidades, práticas de segurança e conformidade"
    },
    SCALABILITY: {
      label: "Análise de Escalabilidade",
      description: "Avaliação da capacidade de crescimento e performance do sistema"
    },
    CODEQUALITY: {
      label: "Qualidade de Código",
      description: "Análise de padrões, manutenibilidade e boas práticas de desenvolvimento"
    },
    DOCUMENTATIONREVIEW: {
        label: "Revisão de Documentação",
        description: "Análise de clareza, completude e estrutura da documentação do projeto"
    }
  }
};