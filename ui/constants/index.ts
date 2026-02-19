// FIX: Corrected import path for types
import { AppSettings, UserProfile } from '../types';

export const initialProjectContext = `# LookAtni Code - Gerado automaticamente
# Data: 2025-09-09T01:43:34.950Z
# Fonte: ./
# Total de arquivos: 186

// / RELEASE_NOTES.md / //
# 🚀 Kortex v0.1.0 - Release Notes

**Release Date:** July 25, 2025
**Version:** 0.1.0
**Status:** Production Ready

---

## 🎉 Major Milestone: Complete Desmocking Strategy

This release marks the successful completion of the **desmocking strategy**, transforming Kortex from a prototype with mock data into a production-ready monitoring dashboard with real API integrations.

---

## ✨ What's New

### 🔄 Real Data Integration

- **Dashboard**: Live GitHub and Azure DevOps data integration
- **Servers Page**: Real-time MCP server monitoring and management
- **Analytics**: Comprehensive data aggregation from multiple sources
- **Helm/Kubernetes**: Full cluster and release management capabilities
- **API Configuration**: Dynamic API endpoint management

### 🚀 Performance & Reliability

- **WebSocket System**: Real-time updates with auto-reconnection
- **Resilient Fallbacks**: Graceful degradation when APIs are unavailable
- **Visual Indicators**: Clear data source status (Real Data vs Demo Mode)
- **Auto-refresh**: Intelligent background data refresh (3-5 minute intervals)
- **Error Handling**: Comprehensive error boundaries and retry mechanisms

### 🛠️ Developer Experience

- **TypeScript 100%**: Strict type safety with zero compilation errors
- **Mock API Server**: Complete development environment with 10 endpoints
- **Modular Architecture**: Clean separation of concerns and reusable components
- **Hot Reload**: Fast development cycle with instant updates
- **Build Optimization**: Static site generation for fast deployments

---

## 🏗️ Technical Achievements

### Architecture Overhaul

\`\`\`
BEFORE: Static mock data → Simple UI rendering
AFTER:  Real APIs → Resilient Service Layer → WebSocket Updates → UI with Fallbacks
\`\`\`

### Performance Metrics

- **Build Success**: 14/14 pages compiling successfully
- **TypeScript Errors**: 0 compilation errors
- **API Coverage**: 10 endpoints fully implemented and tested
- **Real Data Coverage**: 5/5 major pages fully desmocked

### Quality Improvements

- **Code Coverage**: Comprehensive error handling and edge cases
- **Documentation**: Complete technical documentation and guides
- **Standards Compliance**: Follows TypeScript and Markdown best practices
- **Accessibility**: Responsive design with dark mode support

---

## 🔮 Future Roadmap

### Immediate Next Steps (v0.2.0)
- Connect to production StatusRafa and Kosmos APIs
- Implement authentication and authorization
- Add advanced alerting and notification systems
- Expand monitoring capabilities

### Planned Enhancements
- **Multi-cloud Support**: AWS, GCP integration
- **Advanced Analytics**: Machine learning insights
- **Custom Dashboards**: User-configurable interfaces
- **Mobile Application**: React Native companion app

---

// / docs/README.md / //
# Kortex Documentation

This directory contains the complete documentation for Kortex, built with MkDocs Material.

[[[[[   ## 🚀 Quick Start

### Prerequisites

- Python 3.8+
- UV package manager installed

### Setup

1. **Install dependencies**:

   \`\`\`bash
   uv sync
   \`\`\`

2. **Activate virtual environment**:

   \`\`\`bash
   source .venv/bin/activate
   \`\`\`

3. **Start development server**:

   \`\`\`bash
   mkdocs serve
   \`\`\`

   Or use the helper script:

   \`\`\`bash
   ./docs-dev.sh serve
   \`\`\`

4. **Open in browser**: <http://localhost:8000>

## 🌐 Real-Time DevOps & AI Monitoring Dashboard

**Kortex** is a production-ready, enterprise-grade monitoring dashboard designed for modern development teams. It provides real-time insights into API usage, system health, and development workflows across GitHub, Azure DevOps, Kubernetes, and AI infrastructure.

Built with **Next.js 15**, **TypeScript**, and **Tailwind CSS**, Kortex offers a responsive, real-time interface powered by WebSocket connections and resilient API integrations.

---

## 🏗️ Architecture

\`\`\`mermaid
graph TD
    A[Kortex Dashboard] --> B[Real-Time Hooks]
    B --> C[Resilient Service Layer]
    C --> D[Mock API Server]
    C --> E[Production APIs]

    D --> F[GitHub API Mock]
    D --> G[Azure DevOps Mock]
    D --> H[MCP Server Mock]
    D --> I[Helm/K8s Mock]

    E --> J[StatusRafa MCP]
    E --> K[Kosmos Backend]
    E --> L[External APIs]

    A --> M[WebSocket System]
    M --> N[Real-time Events]
    M --> O[Auto-reconnect]
\`\`\`

### Core Components

- **Frontend**: Next.js 15 with TypeScript and Tailwind CSS
- **State Management**: React Context API with custom hooks
- **Real-time**: WebSocket connections with automatic reconnection
- **API Layer**: Resilient service layer with fallback mechanisms
- **Development**: Mock API server for local development
- **Production**: Integration with StatusRafa MCP and Kosmos backends
`;

export const defaultSettings: AppSettings = {
  saveHistory: true,
  theme: 'dark',
  enableTelemetry: false,
  autoAnalyze: false,
  tokenLimit: 1000000,
  userApiKey: '',
  githubPat: '',
  jiraInstanceUrl: '',
  jiraUserEmail: '',
  jiraApiToken: '',
  enableDashboardInsights: true,
};

export const defaultUserProfile: UserProfile = {
  name: 'GemX User',
  email: '',
  avatar: '',
};

