
// Integration settings interface

export interface IntegrationSettings {
  github: GitHubIntegrationSettings;
  jira: JiraIntegrationSettings;
  gitlab: GitLabIntegrationSettings;
  trello: TrelloIntegrationSettings;
}

// Trello Board type

export interface TrelloIntegrationSettings {
  // trello
  trelloIntegrationEnabled: boolean;

  trelloApiKey?: string;
  trelloToken?: string;
  trelloBoardIds?: string[];
}

export interface TrelloBoardListItem {
  id: string;
  name: string;
  url: string;
}

export interface TrelloList {
  id: string;
  name: string;
  closed: boolean;
  idBoard: string;
  pos: number;
}

export interface TrelloCard {
  id: string;
  name: string;
  desc: string;
  closed: boolean;
  idList: string;
  url: string;
}

// GitLab Repository type

export interface GitLabIntegrationSettings {
  // gitlab
  gitlabIntegrationEnabled: boolean;

  gitlabPat?: string;
  gitlabUsername?: string;
  gitlabOAuthToken?: string;
  gitlabInstanceUrl?: string;
  gitlabProjects?: string[];
  gitlabGroups?: string[];
}

export interface GitLabRepoListItem {
  id: number;
  name: string;
  path_with_namespace: string;
  description: string | null;
  star_count: number;
  web_url: string;
  owner: {
    username: string;
  };
}

export interface GitLabFileContent {
  file_name: string;
  file_path: string;
  size: number;
  encoding: string;
  content: string;
  ref: string;
  blob_id: string;
  commit_id: string;
  last_commit_id: string;
}

export interface GitLabTreeItem {
  id: string;
  name: string;
  type: 'blob' | 'tree' | 'commit';
  size: number;
  sha: string;
  url: string;
}

// GitHub Repository type

export interface GitHubIntegrationSettings {
  // github
  githubIntegrationEnabled: boolean;

  githubPat?: string;
  githubUsername?: string;
  githubOAuthToken?: string;
  githubEnterpriseUrl?: string;
  githubRepositories?: string[];
  githubOrganizations?: string[];

  githubTeams?: string[];
  githubTeamRepos?: string[];
  githubTeamMembers?: string[];
  githubTeamSlug?: string;
  githubRole?: 'member' | 'admin' | 'maintain' | 'write' | 'triage';
}

export interface GitHubRepoListItem {
  id: number;
  name: string;
  full_name: string;
  description: string | null;
  stargazers_count: number;
  html_url: string;
  owner: {
    login: string;
  };
}

export interface GitHubFileContent {
  name: string;
  path: string;
  sha: string;
  size: number;
  url: string;
  html_url: string;
  git_url: string;
  download_url: string;
  type: string;
  content: string;
  encoding: string;
}

export interface GitHubTreeItem {
  path: string;
  mode: string;
  type: string;
  size: number;
  sha: string;
  url: string;
}

// Jira Project type

export interface JiraIntegrationSettings {
  // jira
  jiraIntegrationEnabled: boolean;
  jiraPat?: string;
  jiraInstanceUrl?: string;
  jiraUserEmail?: string;
  jiraApiToken?: string;
  jiraProjects?: string[];
}

export interface JiraProjectListItem {
  id: string;
  key: string;
  name: string;
}

export interface JiraIssue {
  id: string;
  key: string;
  fields: {
    summary: string;
    status: {
      name: string;
    };
    issuetype: {
      name: string;
    };
    priority: {
      name: string;
    };
    description: string | null;
  };
}

export interface JiraIssuesResponse {
  issues: JiraIssue[];
  total: number;
  startAt: number;
  maxResults: number;
}

export interface JiraCreateIssuePayload {
  fields: {
    project: {
      key: string;
    };
    summary: string;
    description: string;
    issuetype: {
      name: string;
    };
    priority?: {
      name: string;
    };
  };
}

export interface JiraCreateIssueResponse {
  id: string;
  key: string;
  self: string;
}

// Rastreabilidade segura - sem dados sensíveis

export interface IntegrationTrackingMetadata {
  userId: string;
  userName: string; // Para display apenas
  createdAt: string;
}

// Não incluir email, API keys, ou outros dados sensíveis
