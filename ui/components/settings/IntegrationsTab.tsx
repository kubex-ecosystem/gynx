import { AtSign, ExternalLink, Github, Trello } from 'lucide-react';
import * as React from 'react';
import { useState } from 'react';
import { useNotification } from '../../contexts/NotificationContext';
import { useUser } from '../../contexts/UserContext';
import { JiraIntegrationSettings, TrelloIntegrationSettings } from '../../types';

interface IntegrationsTabProps {
  // Agora usa contexto diretamente
}

const IntegrationsTab: React.FC<IntegrationsTabProps> = () => {
  const { integrations, setIntegrations } = useUser();
  const { addNotification } = useNotification();
  const [testingConnection, setTestingConnection] = useState<string | null>(null);

  const getDefaultIntegrations = () => ({
    github: { githubIntegrationEnabled: false },
    jira: { jiraIntegrationEnabled: false },
    gitlab: { gitlabIntegrationEnabled: false },
    trello: { trelloIntegrationEnabled: false }
  });

  const handleGithubPatChange = (value: string) => {
    const currentIntegrations = integrations || getDefaultIntegrations();
    const updatedIntegrations = {
      ...currentIntegrations,
      github: {
        ...currentIntegrations.github,
        githubPat: value,
        githubIntegrationEnabled: value.length > 0,
      }
    };
    setIntegrations(updatedIntegrations);
  };

  const handleTrelloChange = (field: 'trelloApiKey' | 'trelloToken', value: string) => {
    const currentIntegrations = integrations || getDefaultIntegrations();
    const currentTrello = currentIntegrations.trello as TrelloIntegrationSettings;
    const updatedTrello = {
      ...currentTrello,
      [field]: value,
    };

    // Ativar integração se ambos os campos estão preenchidos
    const wasEnabled = updatedTrello.trelloIntegrationEnabled;
    updatedTrello.trelloIntegrationEnabled = Boolean(
      updatedTrello.trelloApiKey && updatedTrello.trelloToken
    );

    const updatedIntegrations = {
      ...currentIntegrations,
      trello: updatedTrello
    };
    setIntegrations(updatedIntegrations);

    // Feedback para mudanças na integração do Trello
    if (value.length > 0) {
      addNotification({
        message: `🔗 Trello ${field === 'trelloApiKey' ? 'API key' : 'token'} updated`,
        type: 'success'
      });
    }

    if (!wasEnabled && updatedTrello.trelloIntegrationEnabled) {
      addNotification({
        message: '🎯 Trello integration activated! Both credentials provided.',
        type: 'success'
      });
    }
  };

  const handleJiraChange = (field: 'jiraInstanceUrl' | 'jiraUserEmail' | 'jiraApiToken', value: string) => {
    const currentIntegrations = integrations || getDefaultIntegrations();
    const currentJira = currentIntegrations.jira as JiraIntegrationSettings;
    const updatedJira = {
      ...currentJira,
      [field]: value,
    };

    // Ativar integração se todos os campos essenciais estão preenchidos
    const wasEnabled = updatedJira.jiraIntegrationEnabled;
    updatedJira.jiraIntegrationEnabled = Boolean(
      updatedJira.jiraInstanceUrl && updatedJira.jiraUserEmail && updatedJira.jiraApiToken
    );

    const updatedIntegrations = {
      ...currentIntegrations,
      jira: updatedJira
    };
    setIntegrations(updatedIntegrations);

    // Feedback para mudanças na integração do Jira
    if (value.length > 0) {
      const fieldNames = {
        jiraInstanceUrl: 'instance URL',
        jiraUserEmail: 'user email',
        jiraApiToken: 'API token'
      };
      addNotification({
        message: `🔗 Jira ${fieldNames[field]} updated`,
        type: 'success'
      });
    }

    if (!wasEnabled && updatedJira.jiraIntegrationEnabled) {
      addNotification({
        message: '🎯 Jira integration activated! All credentials provided.',
        type: 'success'
      });
    }
  }; const testConnection = async (service: 'github' | 'trello' | 'jira') => {
    setTestingConnection(service);

    // Simulação de teste de conexão
    await new Promise(resolve => setTimeout(resolve, 2000));

    // Aqui você implementaria os testes reais
    const success = Math.random() > 0.3; // 70% de chance de sucesso para demo

    if (success) {
      addNotification({
        message: `✅ ${service.charAt(0).toUpperCase() + service.slice(1)} connection successful!`,
        type: 'success'
      });
    } else {
      addNotification({
        message: `❌ Failed to connect to ${service}. Please check your credentials.`,
        type: 'error'
      });
    }

    setTestingConnection(null);
  };

  return (
    <section>
      <p className="text-sm text-gray-400 mb-4">Connect your accounts to enable additional features, like importing from private GitHub repositories.</p>

      <div className="space-y-6">
        {/* GitHub */}
        <div className="p-4 bg-gray-900/50 border border-gray-700 rounded-lg">
          <div className="flex items-center justify-between mb-3">
            <h4 className="text-md font-semibold text-white flex items-center gap-2">
              <Github className="w-5 h-5 text-gray-400" /> GitHub
            </h4>
            {integrations?.github?.githubIntegrationEnabled && (
              <span className="text-xs px-2 py-1 bg-green-600 text-white rounded-full">Connected</span>
            )}
          </div>
          <p className="text-sm text-gray-400 mb-4">
            Provide a Personal Access Token (PAT) to access private repositories and increase API rate limits.
          </p>
          <div className="space-y-3">
            <div>
              <label htmlFor="githubPat" className="text-sm font-medium text-gray-300">Personal Access Token (PAT)</label>
              <input
                type="password"
                id="githubPat"
                value={integrations?.github?.githubPat || ''}
                onChange={(e) => handleGithubPatChange(e.target.value)}
                placeholder="Enter your GitHub PAT (e.g., ghp_xxxxxxxxxxxx)"
                className="w-full p-2 mt-1 bg-gray-900 border border-gray-600 rounded-md"
              />
            </div>
            <div className="text-xs text-gray-500">
              <p>Required scopes: <code className="bg-gray-800 px-1 rounded">repo</code>, <code className="bg-gray-800 px-1 rounded">read:user</code></p>
              <a
                href="https://github.com/settings/tokens/new"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1 text-blue-400 hover:text-blue-300 mt-1"
              >
                Create a new token <ExternalLink className="w-3 h-3" />
              </a>
            </div>
          </div>
        </div>

        {/* Trello */}
        <div className="p-4 bg-gray-900/50 border border-gray-700 rounded-lg">
          <div className="flex items-center justify-between mb-3">
            <h4 className="text-md font-semibold text-white flex items-center gap-2">
              <Trello className="w-5 h-5 text-blue-400" /> Trello
            </h4>
            <div className="flex items-center gap-2">
              {integrations?.trello?.trelloIntegrationEnabled && (
                <span className="text-xs px-2 py-1 bg-green-600 text-white rounded-full">Connected</span>
              )}
              <button
                onClick={() => testConnection('trello')}
                disabled={testingConnection === 'trello'}
                className="px-3 py-1 text-xs bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 text-white rounded-md"
              >
                {testingConnection === 'trello' ? 'Testing...' : 'Test'}
              </button>
            </div>
          </div>
          <p className="text-sm text-gray-400 mb-4">
            Connect your Trello account to import boards and cards into your project management.
          </p>
          <div className="space-y-3">
            <div>
              <label htmlFor="trelloApiKey" className="text-sm font-medium text-gray-300">API Key</label>
              <input
                type="password"
                id="trelloApiKey"
                value={(integrations?.trello as TrelloIntegrationSettings)?.trelloApiKey || ''}
                onChange={(e) => handleTrelloChange('trelloApiKey', e.target.value)}
                placeholder="Enter your Trello API Key"
                className="w-full p-2 mt-1 bg-gray-900 border border-gray-600 rounded-md"
              />
            </div>
            <div>
              <label htmlFor="trelloToken" className="text-sm font-medium text-gray-300">Token</label>
              <input
                type="password"
                id="trelloToken"
                value={(integrations?.trello as TrelloIntegrationSettings)?.trelloToken || ''}
                onChange={(e) => handleTrelloChange('trelloToken', e.target.value)}
                placeholder="Enter your Trello Token"
                className="w-full p-2 mt-1 bg-gray-900 border border-gray-600 rounded-md"
              />
            </div>
            <div className="text-xs text-gray-500">
              <a
                href="https://trello.com/app-key"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1 text-blue-400 hover:text-blue-300"
              >
                Get your API Key & Token <ExternalLink className="w-3 h-3" />
              </a>
            </div>
          </div>
        </div>

        {/* Jira */}
        <div className="p-4 bg-gray-900/50 border border-gray-700 rounded-lg">
          <div className="flex items-center justify-between mb-3">
            <h4 className="text-md font-semibold text-white flex items-center gap-2">
              <AtSign className="w-5 h-5 text-blue-500" /> Jira
            </h4>
            <div className="flex items-center gap-2">
              {integrations?.jira?.jiraIntegrationEnabled && (
                <span className="text-xs px-2 py-1 bg-green-600 text-white rounded-full">Connected</span>
              )}
              <button
                onClick={() => testConnection('jira')}
                disabled={testingConnection === 'jira'}
                className="px-3 py-1 text-xs bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 text-white rounded-md"
              >
                {testingConnection === 'jira' ? 'Testing...' : 'Test'}
              </button>
            </div>
          </div>
          <p className="text-sm text-gray-400 mb-4">
            Connect your Jira instance to sync issues and project data with your analysis workflow.
          </p>
          <div className="space-y-3">
            <div>
              <label htmlFor="jiraInstanceUrl" className="text-sm font-medium text-gray-300">Instance URL</label>
              <input
                type="url"
                id="jiraInstanceUrl"
                value={(integrations?.jira as JiraIntegrationSettings)?.jiraInstanceUrl || ''}
                onChange={(e) => handleJiraChange('jiraInstanceUrl', e.target.value)}
                placeholder="https://yourcompany.atlassian.net"
                className="w-full p-2 mt-1 bg-gray-900 border border-gray-600 rounded-md"
              />
            </div>
            <div>
              <label htmlFor="jiraUserEmail" className="text-sm font-medium text-gray-300">Email</label>
              <input
                type="email"
                id="jiraUserEmail"
                value={(integrations?.jira as JiraIntegrationSettings)?.jiraUserEmail || ''}
                onChange={(e) => handleJiraChange('jiraUserEmail', e.target.value)}
                placeholder="your.email@company.com"
                className="w-full p-2 mt-1 bg-gray-900 border border-gray-600 rounded-md"
              />
            </div>
            <div>
              <label htmlFor="jiraApiToken" className="text-sm font-medium text-gray-300">API Token</label>
              <input
                type="password"
                id="jiraApiToken"
                value={(integrations?.jira as JiraIntegrationSettings)?.jiraApiToken || ''}
                onChange={(e) => handleJiraChange('jiraApiToken', e.target.value)}
                placeholder="Enter your Jira API Token"
                className="w-full p-2 mt-1 bg-gray-900 border border-gray-600 rounded-md"
              />
            </div>
            <div className="text-xs text-gray-500">
              <a
                href="https://id.atlassian.com/manage-profile/security/api-tokens"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1 text-blue-400 hover:text-blue-300"
              >
                Create API Token <ExternalLink className="w-3 h-3" />
              </a>
            </div>
          </div>
        </div>

        {/* Future integrations placeholder */}
        <div className="p-4 bg-gray-900/30 border border-gray-700 border-dashed rounded-lg">
          <p className="text-center text-gray-500 text-sm">More integrations (GitLab, Slack, Discord) coming soon...</p>
        </div>
      </div>
    </section>
  );
};

export default IntegrationsTab;
