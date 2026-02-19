import { Project, ProjectAnalysis, AnalysisType, Priority, Difficulty, Effort, MaturityLevel, KanbanState } from '../types';
import { v4 as uuidv4 } from 'uuid';

export const exampleAnalysis: ProjectAnalysis = {
  projectName: "Kortex",
  analysisType: AnalysisType.Architecture,
  summary: "Kortex is a well-architected, production-ready monitoring dashboard with a strong focus on real-time data and developer experience. It has successfully transitioned from a mock-data prototype to a fully integrated application. Key strengths include its resilient architecture, comprehensive feature set, and high code quality. The main areas for improvement involve enhancing security with authentication, expanding API integrations, and formalizing the alerting system.",
  strengths: [
    "Complete Desmocking Strategy: Successfully integrated real APIs for GitHub, Azure DevOps, and MCP, moving beyond mock data.",
    "Resilient Architecture: Features a robust service layer, WebSocket system with auto-reconnection, and graceful fallbacks for API outages.",
    "Excellent Developer Experience: 100% TypeScript with strict type safety, a complete mock API server for local development, and a modular component structure.",
    "Real-time Monitoring Capabilities: Provides live updates for dashboards, servers, and cluster management through a WebSocket system.",
    "Clear Future Roadmap: A well-defined plan for future versions, including production API connections, authentication, and multi-cloud support."
  ],
  improvements: [
    {
      title: "Implement Authentication and Authorization",
      description: "The project currently lacks a user authentication and authorization layer, which is critical for a production enterprise application. This exposes the dashboard to unauthorized access.",
      priority: Priority.High,
      difficulty: Difficulty.Medium,
      businessImpact: "Prevents unauthorized access to sensitive monitoring data and management controls, which is essential for security and compliance."
    },
    {
      title: "Add Advanced Alerting and Notification System",
      description: "While the dashboard provides real-time monitoring, it lacks a proactive alerting system to notify developers of critical issues (e.g., server downtime, build failures) via channels like Slack or email.",
      priority: Priority.Medium,
      difficulty: Difficulty.Medium,
      businessImpact: "Reduces response time to critical incidents, minimizes downtime, and improves operational efficiency."
    },
    {
      title: "Expand API Integration to Production Endpoints",
      description: "The current integration is primarily with mock and staging APIs. The immediate next step is to connect to the production StatusRafa and Kosmos APIs to reflect real-world operational data.",
      priority: Priority.High,
      difficulty: Difficulty.Low,
      businessImpact: "Provides actual, actionable insights for the development team, making the dashboard a central tool for production monitoring."
    }
  ],
  nextSteps: {
    shortTerm: [
      {
        title: "Connect to Production StatusRafa and Kosmos APIs",
        description: "Update the service layer to switch from mock/staging API endpoints to the live production endpoints.",
        difficulty: Difficulty.Low
      },
      {
        title: "Implement a Basic JWT-based Authentication System",
        description: "Add a login page and protect all routes, requiring a valid JSON Web Token for access.",
        difficulty: Difficulty.Medium
      }
    ],
    longTerm: [
      {
        title: "Develop a User-Configurable Custom Dashboard",
        description: "Allow users to create their own dashboard layouts by selecting and arranging various monitoring widgets.",
        difficulty: Difficulty.High
      },
      {
        title: "Integrate with AWS and GCP for Multi-Cloud Support",
        description: "Expand monitoring capabilities to include services and infrastructure from other major cloud providers.",
        difficulty: Difficulty.High
      }
    ]
  },
  viability: {
    score: 9,
    assessment: "The project's viability is extremely high. It addresses a clear need for a centralized, real-time monitoring dashboard. The technical execution is solid, the architecture is scalable, and the roadmap is strategic. The only factor holding it back from a perfect 10 is the current lack of production-critical features like authentication and alerting, which are already planned."
  },
  roiAnalysis: {
    assessment: "The potential ROI is significant, primarily through increased developer productivity and reduced system downtime. By centralizing monitoring and providing real-time insights, Kortex can drastically cut down the time developers spend context-switching and diagnosing issues.",
    potentialGains: [
      "Reduced Mean Time to Resolution (MTTR) for production incidents.",
      "Increased development velocity due to better visibility of build and deployment pipelines.",
      "Improved decision-making with aggregated data from multiple sources.",
      "Lower operational overhead through centralized management of clusters and releases."
    ],
    estimatedEffort: Effort.Medium
  },
  maturity: {
    level: MaturityLevel.Production,
    assessment: "The project has reached the 'Production' maturity level. It has moved beyond an MVP by replacing all mock data with real API integrations, implementing a resilient architecture, and ensuring high code quality. While it's production-ready, it's not yet 'Optimized' as it still needs features like advanced alerting and multi-cloud support."
  },
  architectureDiagram: `
graph TD
    A[Kortex Dashboard] --> B[Real-Time Hooks]
    B --> C[Resilient Service Layer]
    C --> D[Mock API Server]
    C --> E[Production APIs]
    
    D --> F[GitHub API Mock]
    D --> G[Azure DevOps Mock]
    D --> H[MCP Server Mock]
    
    E --> J[StatusRafa MCP]
    E --> K[Kosmos Backend]
    
    A --> M[WebSocket System]
    M --> N[Real-time Events]
`
};

const exampleKanban: KanbanState = {
    cards: {
        'card-1': { id: 'card-1', title: 'Implement Authentication and Authorization', description: 'The project currently lacks a user authentication and authorization layer, which is critical for a production enterprise application. This exposes the dashboard to unauthorized access.', priority: Priority.High, difficulty: Difficulty.Medium, tags: ['security'] },
        'card-2': { id: 'card-2', title: 'Add Advanced Alerting and Notification System', description: 'While the dashboard provides real-time monitoring, it lacks a proactive alerting system to notify developers of critical issues (e.g., server downtime, build failures) via channels like Slack or email.', priority: Priority.Medium, difficulty: Difficulty.Medium, tags: ['feature'] },
        'card-3': { id: 'card-3', title: 'Expand API Integration to Production Endpoints', description: 'The current integration is primarily with mock and staging APIs. The immediate next step is to connect to the production StatusRafa and Kosmos APIs to reflect real-world operational data.', priority: Priority.High, difficulty: Difficulty.Low, tags: ['integration'] },
    },
    columns: {
        backlog: { id: 'backlog', title: 'Backlog', cardIds: ['card-1', 'card-2', 'card-3'] },
        todo: { id: 'todo', title: 'To Do', cardIds: [] },
        inProgress: { id: 'inProgress', title: 'In Progress', cardIds: [] },
        done: { id: 'done', title: 'Done', cardIds: [] },
    },
    columnOrder: ['backlog', 'todo', 'inProgress', 'done'],
};


export const exampleProject: Project = {
    id: 'example-project-id',
    name: 'Kortex (Example)',
    createdAt: new Date('2024-07-01T10:00:00Z').toISOString(),
    updatedAt: new Date().toISOString(),
    history: [
        {
            id: Date.now(),
            timestamp: new Date().toISOString(),
            analysis: exampleAnalysis
        }
    ],
    kanban: exampleKanban,
    chatHistories: {
        [Date.now()]: [
            { role: 'model', parts: [{ text: "Hello! I've analyzed the Kortex project's architecture. Ask me anything about its structure, resilience, or how it uses WebSockets." }] }
        ]
    },
    contextFiles: []
};