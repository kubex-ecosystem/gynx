import { Check, Loader2, Shield, X } from 'lucide-react';
import * as React from 'react';
import { useEffect, useState } from 'react';
import { useNotification } from '../../contexts/NotificationContext';
import { useUser } from '../../contexts/UserContext';
import { Provider, unifiedAI } from '../../services/unified-ai';
import { UserSettings } from '../../types';

interface SecurityTabProps {
  // Usa contexto diretamente
}

const SecurityTab: React.FC<SecurityTabProps> = () => {
  const { addNotification } = useNotification();
  const { userSettings, updateUserSetting } = useUser();

  const [apiKey, setApiKey] = useState(userSettings.userApiKey || '');
  const [isTestingKey, setIsTestingKey] = useState(false);
  const [testStatus, setTestStatus] = useState<'success' | 'failure' | null>(null);
  const [availableProviders, setAvailableProviders] = useState<Provider[]>([]);
  const [isLoadingProviders, setIsLoadingProviders] = useState(true);

  // Carregar providers disponíveis do backend
  useEffect(() => {
    const loadProviders = async () => {
      try {
        const providers = await unifiedAI.getProviders();
        setAvailableProviders(providers);

        addNotification({
          message: `🔌 Found ${providers.length} available providers: ${providers.map(p => p.name).join(', ')}`,
          type: 'success'
        });
      } catch (error) {
        console.error('Failed to load providers:', error);
        // Fallback para providers padrão se não conseguir carregar do backend
        const fallbackProviders: Provider[] = [
          { id: 'openai', name: 'OpenAI', models: [], available: true },
          { id: 'claude', name: 'Claude', models: [], available: true },
          { id: 'gemini', name: 'Google Gemini', models: [], available: true },
          { id: 'ollama', name: 'Ollama', models: [], available: true },
          { id: 'groq', name: 'Groq', models: [], available: true }
        ];
        setAvailableProviders(fallbackProviders);

        addNotification({
          message: '⚠️ Using default providers - backend connection failed',
          type: 'info'
        });
      } finally {
        setIsLoadingProviders(false);
      }
    };

    loadProviders();
  }, [addNotification]); const handleFieldChange = (key: keyof UserSettings, value: any) => {
    updateUserSetting(key, value);

    // Feedback para mudanças de configuração
    const feedbackMessages = {
      apiProvider: `API provider changed to ${value}`,
      customApiEndpoint: `Custom API endpoint updated`,
    };

    const message = feedbackMessages[key as keyof typeof feedbackMessages];
    if (message) {
      addNotification({
        message: `🔧 ${message}`,
        type: 'success'
      });
    }
  };

  const handleApiKeyChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newKey = e.target.value;
    setApiKey(newKey);
    updateUserSetting('userApiKey', newKey);
    setTestStatus(null);

    if (newKey.length > 0) {
      addNotification({
        message: '🔑 API key updated',
        type: 'success'
      });
    }
  };

  const handleTestApiKey = async () => {
    if (!userSettings.apiProvider) {
      addNotification({
        message: '❌ Please select a provider first',
        type: 'error'
      });
      return;
    }

    setIsTestingKey(true);

    try {
      // Use o unified AI service para testar a conectividade do provider
      const result = await unifiedAI.testProvider(userSettings.apiProvider);
      setTestStatus(result ? 'success' : 'failure');

      if (result) {
        addNotification({
          message: `✅ ${userSettings.apiProvider.toUpperCase()} provider is working correctly`,
          type: 'success'
        });
      } else {
        addNotification({
          message: `❌ ${userSettings.apiProvider.toUpperCase()} provider test failed - check backend configuration`,
          type: 'error'
        });
      }
    } catch (error) {
      console.error('Provider test error:', error);
      setTestStatus('failure');
      addNotification({
        message: '❌ Failed to test provider - please check your connection',
        type: 'error'
      });
    } finally {
      setIsTestingKey(false);
    }
  }; const renderTestButton = () => {
    if (isTestingKey) {
      return <div className="flex items-center gap-2 text-blue-400"><Loader2 className="w-4 h-4 animate-spin" /> Testing...</div>;
    }
    if (testStatus === 'success') {
      return <div className="flex items-center gap-2 text-green-400"><Check className="w-4 h-4" /> Working</div>;
    }
    if (testStatus === 'failure') {
      return <div className="flex items-center gap-2 text-red-400"><X className="w-4 h-4" /> Failed</div>;
    }
    return 'Test Provider';
  };

  return (
    <div className="space-y-8">
      {/* API Configuration Section */}
      <section>
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <Shield className="w-5 h-5 text-gray-400" /> API Configuration
        </h3>
        <div className="space-y-4">
          {/* API Provider */}
          <div className="flex items-start justify-between">
            <div>
              <label htmlFor="apiProvider" className="font-medium text-gray-200">API Provider</label>
              <p className="text-sm text-gray-400">
                Choose your preferred AI provider.
                {isLoadingProviders && <span className="text-blue-400"> Loading providers...</span>}
              </p>
            </div>
            <select
              id="apiProvider"
              value={userSettings.apiProvider || 'gemini'}
              onChange={(e) => handleFieldChange('apiProvider', e.target.value as UserSettings['apiProvider'])}
              className="mt-1 px-3 py-1 bg-gray-700 border border-gray-600 rounded-md text-white"
              disabled={isLoadingProviders}
            >
              {availableProviders.map(provider => (
                <option key={provider.id} value={provider.id} disabled={!provider.available}>
                  {provider.name} {!provider.available && '(Unavailable)'}
                </option>
              ))}
              <option value="custom">Custom</option>
            </select>
          </div>

          {/* Custom API Endpoint */}
          {userSettings.apiProvider === 'custom' && (
            <div className="flex items-start justify-between">
              <div>
                <label htmlFor="customApiEndpoint" className="font-medium text-gray-200">Custom API Endpoint</label>
                <p className="text-sm text-gray-400">URL for your custom API endpoint.</p>
              </div>
              <input
                id="customApiEndpoint"
                type="url"
                value={userSettings.customApiEndpoint || ''}
                onChange={(e) => handleFieldChange('customApiEndpoint', e.target.value)}
                placeholder="https://api.example.com"
                className="mt-1 px-3 py-1 bg-gray-700 border border-gray-600 rounded-md text-white w-64"
              />
            </div>
          )}

          {/* API Key */}
          <div>
            <label htmlFor="userApiKey" className="text-sm font-medium text-gray-300">Your API Key (Optional)</label>
            <p className="text-sm text-gray-400 mb-2">
              Store your API key locally for reference. The backend uses its own configured keys for AI providers.
              Test button validates backend provider connectivity.
            </p>
            <div className="flex gap-2">
              <input
                type="password"
                id="userApiKey"
                value={apiKey}
                onChange={handleApiKeyChange}
                placeholder="Enter your API key"
                className="flex-grow p-2 bg-gray-900 border border-gray-600 rounded-md"
              />
              <button
                onClick={handleTestApiKey}
                disabled={isTestingKey}
                className="px-4 py-2 text-sm font-semibold text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:bg-gray-600"
              >
                {renderTestButton()}
              </button>
            </div>
          </div>
        </div>
      </section>

      {/* Security Information */}
      <section>
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <Shield className="w-5 h-5 text-green-400" /> Security & Privacy
        </h3>
        <div className="space-y-4">
          <div className="p-4 bg-green-900/20 border border-green-700/50 rounded-lg">
            <h4 className="text-sm font-medium text-green-300 mb-2">🔒 Your Data is Secure</h4>
            <ul className="text-xs text-gray-300 space-y-1">
              <li>• API keys are stored locally in your browser using encrypted IndexedDB</li>
              <li>• No sensitive data is transmitted to external servers</li>
              <li>• All communications use HTTPS encryption</li>
              <li>• You can clear all data at any time from the Data tab</li>
            </ul>
          </div>

          <div className="p-4 bg-blue-900/20 border border-blue-700/50 rounded-lg">
            <h4 className="text-sm font-medium text-blue-300 mb-2">🛡️ API Key Best Practices</h4>
            <ul className="text-xs text-gray-300 space-y-1">
              <li>• Use dedicated API keys for this application</li>
              <li>• Set appropriate usage limits in your provider dashboard</li>
              <li>• Regularly rotate your API keys for enhanced security</li>
              <li>• Monitor usage through your provider's console</li>
            </ul>
          </div>

          <div className="p-4 bg-purple-900/20 border border-purple-700/50 rounded-lg">
            <h4 className="text-sm font-medium text-purple-300 mb-2">📊 Telemetry & Analytics</h4>
            <p className="text-xs text-gray-300">
              When telemetry is enabled, we collect anonymous usage statistics to improve the application.
              No personal data, API keys, or content is ever shared. You can disable this at any time in General settings.
            </p>
          </div>
        </div>
      </section>
    </div>
  );
};

export default SecurityTab;
