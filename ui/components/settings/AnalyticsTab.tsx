import { BarChart3, Clock, RefreshCw, TrendingUp } from 'lucide-react';
import * as React from 'react';
import { useConfirmation } from '../../contexts/ConfirmationContext';
import { useNotification } from '../../contexts/NotificationContext';
import { useUser } from '../../contexts/UserContext';
import { UserSettings } from '../../types';

interface AnalyticsTabProps {
  // Usa contexto diretamente
}

const AnalyticsTab: React.FC<AnalyticsTabProps> = () => {
  const { addNotification } = useNotification();
  const { showConfirmation } = useConfirmation();
  const { userSettings, updateUserSetting, usageTracking, resetDailyUsage, resetMonthlyUsage } = useUser();

  const handleFieldChange = (key: keyof UserSettings, value: any) => {
    updateUserSetting(key, value);
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

  return (
    <div className="space-y-8">
      {/* Token Usage Dashboard Section */}
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
                  style={{
                    width: `${Math.min(calculatePercentage(usageTracking.dailyTokens, userSettings.dailyTokenLimit), 100)}%`,
                    maxWidth: '100%'
                  }}
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
                  style={{
                    width: `${Math.min(calculatePercentage(usageTracking.monthlyTokens, userSettings.monthlyTokenLimit), 100)}%`,
                    maxWidth: '100%'
                  }}
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

      {/* Token Limits Section */}
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
    </div>
  );
};

export default AnalyticsTab;
