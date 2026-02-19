import { Palette, Settings as SettingsIcon } from 'lucide-react';
import * as React from 'react';
import { useConfirmation } from '../../contexts/ConfirmationContext';
import { useNotification } from '../../contexts/NotificationContext';
import { useProjectContext } from '../../contexts/ProjectContext';
import { useUser } from '../../contexts/UserContext';
import { Theme, UserSettings } from '../../types';

interface GeneralTabProps {
  // Usa contexto diretamente
}

const GeneralTab: React.FC<GeneralTabProps> = () => {
  const { addNotification } = useNotification();
  const { showConfirmation } = useConfirmation();
  const { handleClearHistory } = useProjectContext();
  const { userSettings, updateUserSetting } = useUser();

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

      // Feedback específico para cada configuração
      const feedbackMessages = {
        theme: `Theme changed to ${value === 'system' ? 'system default' : value} mode`,
        autoAnalyze: `Auto analyze ${value ? 'enabled' : 'disabled'}`,
        enableDashboardInsights: `Dashboard insights ${value ? 'enabled' : 'disabled'}`,
        enableTelemetry: `Telemetry ${value ? 'enabled' : 'disabled'}`,
      };

      const message = feedbackMessages[key as keyof typeof feedbackMessages];
      if (message) {
        addNotification({
          message: `✅ ${message}`,
          type: 'success'
        });
      }
    }
  };

  return (
    <div className="space-y-8">
      {/* Appearance Section */}
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

      {/* General Preferences Section */}
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
    </div>
  );
};

export default GeneralTab;
