import { LandingTranslations } from '../types';

export const landingEnUS: LandingTranslations = {
  cta: "Start Analysis",
  featuresTitle: "Features",
  featuresSubtitle: "Discover what makes our tool unique",
  dynamicPhrases: [
    "complex architectures",
    "legacy code",
    "microservices",
    "RESTful APIs",
    "databases",
    "cloud infrastructure",
    "web applications",
    "distributed systems"
  ],
  hero: {
    title: {
      static: "Transform Documentation into"
    },
    subtitle: "Analyze your projects with AI and get actionable insights",
    cta: "Start Analysis"
  },
  features: {
    title: "Features",
    aiDriven: {
      title: "AI-Driven",
      description: "Intelligent analysis using advanced algorithms"
    },
    comprehensive: {
      title: "Comprehensive",
      description: "Complete analysis of all project aspects"
    },
    actionable: {
      title: "Actionable",
      description: "Practical insights and clear next steps"
    }
  },
  howItWorks: {
    title: "How It Works",
    step1: {
      title: "Provide Context",
      description: "Describe your project or upload documents"
    },
    step2: {
      title: "AI Analysis",
      description: "Our AI analyzes and processes the information"
    },
    step3: {
      title: "Get Insights",
      description: "Receive a detailed report with recommendations"
    }
  },
  featureDetails: {
    general: "Get a 360-degree view of your project. This analysis dives into your architecture, code quality, developer experience, and future roadmap to provide a holistic assessment of its viability and maturity. It's the perfect starting point to understand the overall health of your codebase.",
    security: "Put on your white hat. The security analysis acts as an automated cybersecurity expert, scanning your documentation for potential vulnerabilities, insecure practices, and missing security layers like authentication. It helps you identify and prioritize risks before they become critical.",
    scalability: "Will your project handle success? This review focuses on your architecture's ability to scale. It looks for performance bottlenecks, single points of failure, and inefficient data handling, providing recommendations to ensure your application can grow with your user base.",
    codeQuality: "Promote a healthy and maintainable codebase. This analysis evaluates your project's structure, adherence to best practices, modularity, and error handling. It's like having a principal engineer review your documentation to improve long-term developer experience.",
    documentation: "How good is your project's first impression? This review analyzes your documentation itself for clarity, completeness, and ease of use for a new developer. It provides suggestions to make your READMEs, guides, and comments more effective and welcoming."
  }
};