import { BarChart3, Beaker, Check, Clock, Loader2, Palette, RefreshCw, Settings as SettingsIcon, Shield, TrendingUp, X } from 'lucide-react';
import * as React from 'react';
import { useState } from 'react';
import { useConfirmation } from '../../contexts/ConfirmationContext';
import { useNotification } from '../../contexts/NotificationContext';
import { useProjectContext } from '../../contexts/ProjectContext';
import { useUser } from '../../contexts/UserContext';
import { testApiKey } from '../../services/gemini/api';
import { Theme, UserSettings } from '../../types';

interface PreferencesTabProps {
  // Remover props e usar contexto diretamente
}

const PreferencesTab: React.FC<PreferencesTabProps> = () => {
  const { addNotification } = useNotification();
  const { showConfirmation } = useConfirmation();
  const { handleClearHistory } = useProjectContext();
  const { userSettings, updateUserSetting, usageTracking, resetDailyUsage, resetMonthlyUsage } = useUser();

  const [apiKey, setApiKey] = useState(userSettings.userApiKey || '');
  const [isTestingKey, setIsTestingKey] = useState(false);
  const [testStatus, setTestStatus] = useState<'success' | 'failure' | null>(null);

  const handleFieldChange = (key: keyof UserSettings, value: any) => {
    if (key === 'saveHistory' && value === false && userSettings.saveHistory === true) {
      showConfirmation({
        title: 'Disable History Saving',
        message: 'Disabling this option will also clear the current analysis history for this project. Are you sure you want to continue?',
        confirmText: 'Confirm',
        cancelText: 'Cancel',
        onConfirm: () => {
          updateUserSetting(key, value);
          handleClearHistory();
          addNotification({ message: 'History saving disabled and history cleared.', type: 'info' });
        },
      });
    } else {
      updateUserSetting(key, value);
    }
  };

  const handleApiKeyChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newKey = e.target.value;
    setApiKey(newKey);
    updateUserSetting('userApiKey', newKey);
    setTestStatus(null);
  };

  const handleTestApiKey = async () => {
    setIsTestingKey(true);
    const isValid = await testApiKey(apiKey);
    setTestStatus(isValid ? 'success' : 'failure');
    setIsTestingKey(false);
  };

  const handleResetDailyUsage = () => {
    showConfirmation({
      title: 'Reset Daily Token Usage',
      message: 'This will reset your daily token counter to zero. Are you sure?',
      confirmText: 'Reset',
      cancelText: 'Cancel',
      onConfirm: () => {
        resetDailyUsage();
        addNotification({ message: 'Daily token usage reset successfully.', type: 'success' });
      },
    });
  };

  const handleResetMonthlyUsage = () => {
    showConfirmation({
      title: 'Reset Monthly Token Usage',
      message: 'This will reset your monthly token counter to zero. Are you sure?',
      confirmText: 'Reset',
      cancelText: 'Cancel',
      onConfirm: () => {
        resetMonthlyUsage();
        addNotification({ message: 'Monthly token usage reset successfully.', type: 'success' });
      },
    });
  };

  const calculatePercentage = (used: number, limit: number) => {
    return Math.min(Math.round((used / limit) * 100), 100);
  };

  const getUsageColor = (percentage: number) => {
    if (percentage >= 90) return 'bg-red-500';
    if (percentage >= 70) return 'bg-yellow-500';
    return 'bg-green-500';
  };

  const renderTestButton = () => {
    if (isTestingKey) {
      return <div className="flex items-center gap-2"><Loader2 className="w-4 h-4 animate-spin" /> Testing...</div>;
    }
    if (testStatus === 'success') {
      return <div className="flex items-center gap-2 text-green-400"><Check className="w-4 h-4" /> Valid</div>;
    }
    if (testStatus === 'failure') {
      return <div className="flex items-center gap-2 text-red-400"><X className="w-4 h-4" /> Invalid</div>;
    }
    return 'Test Key';
  };

  return (
    <div className="space-y-8">
      <section>
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <Palette className="w-5 h-5 text-gray-400" /> Appearance
        </h3>
        <div className="space-y-4">
          {/* Theme Selection */}
          <div className="flex items-start justify-between">
            <div>
              <label htmlFor="theme" className="font-medium text-gray-200">Theme</label>
              <p className="text-sm text-gray-400">Choose your preferred theme for the interface.</p>
            </div>
            <select
              id="theme"
              value={userSettings.theme}
              onChange={(e) => handleFieldChange('theme', e.target.value as Theme)}
              className="mt-1 px-3 py-1 bg-gray-700 border border-gray-600 rounded-md text-white"
            >
              <option value="light">Light</option>
              <option value="dark">Dark</option>
              <option value="system">System</option>
            </select>
          </div>
        </div>
      </section>

      <section>
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <SettingsIcon className="w-5 h-5 text-gray-400" /> General Preferences
        </h3>
        <div className="space-y-4">
          {/* Auto Analyze */}
          <div className="flex items-start justify-between">
            <div>
              <label htmlFor="autoAnalyze" className="font-medium text-gray-200">Auto Analyze</label>
              <p className="text-sm text-gray-400">Automatically start analysis when files are uploaded.</p>
            </div>
            <input
              id="autoAnalyze"
              type="checkbox"
              checked={userSettings.autoAnalyze}
              onChange={(e) => handleFieldChange('autoAnalyze', e.target.checked)}
              className="mt-1 w-4 h-4 rounded bg-gray-700 border-gray-600 text-purple-600 focus:ring-purple-500"
            />
          </div>

          {/* Save History */}
          <div className="flex items-start justify-between">
            <div>
              <label htmlFor="saveHistory" className="font-medium text-gray-200">Save Analysis History</label>
              <p className="text-sm text-gray-400">Automatically save each analysis to the project's history.</p>
            </div>
            <input
              id="saveHistory"
              type="checkbox"
              checked={userSettings.saveHistory}
              onChange={(e) => handleFieldChange('saveHistory', e.target.checked)}
              className="mt-1 w-4 h-4 rounded bg-gray-700 border-gray-600 text-purple-600 focus:ring-purple-500"
            />
          </div>

          {/* Enable Dashboard Insights */}
          <div className="flex items-start justify-between">
            <div>
              <label htmlFor="enableDashboardInsights" className="font-medium text-gray-200">Enable Dashboard Insights</label>
              <p className="text-sm text-gray-400">Allow the AI to generate personalized insights on your dashboard.</p>
            </div>
            <input
              id="enableDashboardInsights"
              type="checkbox"
              checked={userSettings.enableDashboardInsights}
              onChange={(e) => handleFieldChange('enableDashboardInsights', e.target.checked)}
              className="mt-1 w-4 h-4 rounded bg-gray-700 border-gray-600 text-purple-600 focus:ring-purple-500"
            />
          </div>

          {/* Enable Telemetry */}
          <div className="flex items-start justify-between">
            <div>
              <label htmlFor="enableTelemetry" className="font-medium text-gray-200">Enable Telemetry</label>
              <p className="text-sm text-gray-400">Help improve the tool by sharing anonymous usage data.</p>
            </div>
            <input
              id="enableTelemetry"
              type="checkbox"
              checked={userSettings.enableTelemetry}
              onChange={(e) => handleFieldChange('enableTelemetry', e.target.checked)}
              className="mt-1 w-4 h-4 rounded bg-gray-700 border-gray-600 text-purple-600 focus:ring-purple-500"
            />
          </div>
        </div>
      </section>

      <section>
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <Clock className="w-5 h-5 text-gray-400" /> Token Limits
        </h3>
        <div className="space-y-4">
          {/* Token Limit */}
          <div className="flex items-start justify-between">
            <div>
              <label htmlFor="tokenLimit" className="font-medium text-gray-200">General Token Limit</label>
              <p className="text-sm text-gray-400">Maximum tokens per request.</p>
            </div>
            <input
              id="tokenLimit"
              type="number"
              value={userSettings.tokenLimit}
              onChange={(e) => handleFieldChange('tokenLimit', Number(e.target.value))}
              className="mt-1 px-3 py-1 bg-gray-700 border border-gray-600 rounded-md text-white w-24"
              min="1000"
              max="1000000"
            />
          </div>

          {/* Daily Token Limit */}
          <div className="flex items-start justify-between">
            <div>
              <label htmlFor="dailyTokenLimit" className="font-medium text-gray-200">Daily Token Limit</label>
              <p className="text-sm text-gray-400">Maximum tokens per day.</p>
            </div>
            <input
              id="dailyTokenLimit"
              type="number"
              value={userSettings.dailyTokenLimit}
              onChange={(e) => handleFieldChange('dailyTokenLimit', Number(e.target.value))}
              className="mt-1 px-3 py-1 bg-gray-700 border border-gray-600 rounded-md text-white w-24"
              min="1000"
              max="100000"
            />
          </div>

          {/* Monthly Token Limit */}
          <div className="flex items-start justify-between">
            <div>
              <label htmlFor="monthlyTokenLimit" className="font-medium text-gray-200">Monthly Token Limit</label>
              <p className="text-sm text-gray-400">Maximum tokens per month.</p>
            </div>
            <input
              id="monthlyTokenLimit"
              type="number"
              value={userSettings.monthlyTokenLimit}
              onChange={(e) => handleFieldChange('monthlyTokenLimit', Number(e.target.value))}
              className="mt-1 px-3 py-1 bg-gray-700 border border-gray-600 rounded-md text-white w-32"
              min="10000"
              max="10000000"
            />
          </div>
        </div>
      </section>

      <section>
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <BarChart3 className="w-5 h-5 text-gray-400" /> Token Usage Dashboard
        </h3>
        <div className="space-y-6">
          {/* Usage Overview */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {/* Daily Usage */}
            <div className="p-4 bg-gray-900/50 border border-gray-700 rounded-lg">
              <div className="flex items-center justify-between mb-2">
                <h4 className="text-sm font-medium text-gray-200">Daily Usage</h4>
                <button
                  onClick={handleResetDailyUsage}
                  className="p-1 text-gray-400 hover:text-white hover:bg-gray-700 rounded"
                  title="Reset daily usage"
                >
                  <RefreshCw className="w-4 h-4" />
                </button>
              </div>
              <p className="text-lg font-bold text-white">
                {usageTracking.dailyTokens.toLocaleString()} / {userSettings.dailyTokenLimit.toLocaleString()}
              </p>
              <div className="mt-2 bg-gray-700 rounded-full h-2 overflow-hidden">
                <div
                  className={`h-2 rounded-full transition-all duration-300 ${getUsageColor(calculatePercentage(usageTracking.dailyTokens, userSettings.dailyTokenLimit))}`}
                  style={{ width: `${Math.min(calculatePercentage(usageTracking.dailyTokens, userSettings.dailyTokenLimit), 100)}%` }}
                />
              </div>
              <p className="text-xs text-gray-400 mt-1">
                {calculatePercentage(usageTracking.dailyTokens, userSettings.dailyTokenLimit)}% used
              </p>
            </div>

            {/* Monthly Usage */}
            <div className="p-4 bg-gray-900/50 border border-gray-700 rounded-lg">
              <div className="flex items-center justify-between mb-2">
                <h4 className="text-sm font-medium text-gray-200">Monthly Usage</h4>
                <button
                  onClick={handleResetMonthlyUsage}
                  className="p-1 text-gray-400 hover:text-white hover:bg-gray-700 rounded"
                  title="Reset monthly usage"
                >
                  <RefreshCw className="w-4 h-4" />
                </button>
              </div>
              <p className="text-lg font-bold text-white">
                {usageTracking.monthlyTokens.toLocaleString()} / {userSettings.monthlyTokenLimit.toLocaleString()}
              </p>
              <div className="mt-2 bg-gray-700 rounded-full h-2 overflow-hidden">
                <div
                  className={`h-2 rounded-full transition-all duration-300 ${getUsageColor(calculatePercentage(usageTracking.monthlyTokens, userSettings.monthlyTokenLimit))}`}
                  style={{ width: `${Math.min(calculatePercentage(usageTracking.monthlyTokens, userSettings.monthlyTokenLimit), 100)}%` }}
                />
              </div>
              <p className="text-xs text-gray-400 mt-1">
                {calculatePercentage(usageTracking.monthlyTokens, userSettings.monthlyTokenLimit)}% used
              </p>
            </div>

            {/* Total Usage */}
            <div className="p-4 bg-gray-900/50 border border-gray-700 rounded-lg">
              <div className="flex items-center justify-between mb-2">
                <h4 className="text-sm font-medium text-gray-200">Total Usage</h4>
                <TrendingUp className="w-4 h-4 text-green-400" />
              </div>
              <p className="text-lg font-bold text-white">
                {usageTracking.totalTokens.toLocaleString()}
              </p>
              <p className="text-xs text-gray-400 mt-1">
                All time tokens used
              </p>
            </div>
          </div>

          {/* Usage Statistics */}
          {usageTracking.analysisCount !== undefined && (
            <div className="p-4 bg-gray-900/30 border border-gray-700 border-dashed rounded-lg">
              <h4 className="text-sm font-medium text-gray-200 mb-3">Activity Statistics</h4>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-center">
                <div>
                  <p className="text-xl font-bold text-purple-400">{usageTracking.analysisCount || 0}</p>
                  <p className="text-xs text-gray-400">Analyses</p>
                </div>
                <div>
                  <p className="text-xl font-bold text-blue-400">{usageTracking.projectCount || 0}</p>
                  <p className="text-xs text-gray-400">Projects</p>
                </div>
                <div>
                  <p className="text-xl font-bold text-green-400">{usageTracking.kanbanBoardCount || 0}</p>
                  <p className="text-xs text-gray-400">Kanban Boards</p>
                </div>
                <div>
                  <p className="text-xl font-bold text-orange-400">{usageTracking.chatSessionCount || 0}</p>
                  <p className="text-xs text-gray-400">Chat Sessions</p>
                </div>
              </div>
            </div>
          )}

          {/* Last Reset Info */}
          <div className="text-xs text-gray-500">
            <p>Last daily reset: {usageTracking.lastResetDate}</p>
            <p>Usage tracking helps you monitor your API consumption and stay within limits.</p>
          </div>
        </div>
      </section>

      <section>
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <Shield className="w-5 h-5 text-gray-400" /> API Configuration
        </h3>
        <div className="space-y-4">
          {/* API Provider */}
          <div className="flex items-start justify-between">
            <div>
              <label htmlFor="apiProvider" className="font-medium text-gray-200">API Provider</label>
              <p className="text-sm text-gray-400">Choose your preferred AI provider.</p>
            </div>
            <select
              id="apiProvider"
              value={userSettings.apiProvider || 'gemini'}
              onChange={(e) => handleFieldChange('apiProvider', e.target.value as UserSettings['apiProvider'])}
              className="mt-1 px-3 py-1 bg-gray-700 border border-gray-600 rounded-md text-white"
            >
              <option value="openai">OpenAI</option>
              <option value="claude">Claude</option>
              <option value="gemini">Google Gemini</option>
              <option value="ollama">Ollama</option>
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
            <label htmlFor="userApiKey" className="text-sm font-medium text-gray-300">Your API Key</label>
            <p className="text-sm text-gray-400 mb-2">Provide your own API key. Stored locally in your browser.</p>
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

      <section>
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <Beaker className="w-5 h-5 text-gray-400" /> Experimental Features
        </h3>
        <div className="space-y-4">
          {/* Beta Features */}
          <div className="flex items-start justify-between">
            <div>
              <label htmlFor="enableBetaFeatures" className="font-medium text-gray-200">Beta Features</label>
              <p className="text-sm text-gray-400">Enable access to beta features (stable but new).</p>
            </div>
            <input
              id="enableBetaFeatures"
              type="checkbox"
              checked={userSettings.enableBetaFeatures}
              onChange={(e) => handleFieldChange('enableBetaFeatures', e.target.checked)}
              className="mt-1 w-4 h-4 rounded bg-gray-700 border-gray-600 text-purple-600 focus:ring-purple-500"
            />
          </div>

          {/* Experimental Features */}
          <div className="flex items-start justify-between">
            <div>
              <label htmlFor="enableExperimentalFeatures" className="font-medium text-gray-200">Experimental Features</label>
              <p className="text-sm text-gray-400">Enable cutting-edge features (may be unstable).</p>
            </div>
            <input
              id="enableExperimentalFeatures"
              type="checkbox"
              checked={userSettings.enableExperimentalFeatures}
              onChange={(e) => handleFieldChange('enableExperimentalFeatures', e.target.checked)}
              className="mt-1 w-4 h-4 rounded bg-gray-700 border-gray-600 text-purple-600 focus:ring-purple-500"
            />
          </div>
        </div>
      </section>
    </div>
  );
};

export default PreferencesTab;
