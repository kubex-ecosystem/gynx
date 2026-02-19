import { InputTranslations } from '../types';

export const inputEnUS: InputTranslations = {
  title: "Project Context",
  projectName: "Project Name",
  projectNamePlaceholder: "e.g., Kortex Project",
  importFromGithub: "Import from GitHub",
  description: "Provide the project context below. You can paste documentation, READMEs, or any relevant text.",
  placeholder: "Paste your project documentation here...\n\n# Kortex Project\n## Overview\nKortex is a real-time monitoring dashboard...",
  useExample: "Or use an example",
  analysisTypeTitle: "Analysis Type",
  analysisTypes: {
    GENERAL: {
      label: "General Analysis",
      description: "Comprehensive evaluation of architecture, quality, and project viability"
    },
    SECURITY: {
      label: "Security Analysis",
      description: "Focus on vulnerabilities, security practices, and compliance"
    },
    SCALABILITY: {
      label: "Scalability Analysis",
      description: "Assessment of system growth capacity and performance"
    },
    CODEQUALITY: {
      label: "Code Quality",
      description: "Analysis of patterns, maintainability, and development best practices"
    },
    DOCUMENTATIONREVIEW: {
      label: "Documentation Review",
      description: "Analysis of clarity, completeness, and structure of project documentation"
    }
  }
};