import { AlertTriangle, Beaker, Zap } from 'lucide-react';
import * as React from 'react';
import { useNotification } from '../../contexts/NotificationContext';
import { useUser } from '../../contexts/UserContext';
import { UserSettings } from '../../types';

interface AdvancedTabProps {
  // Usa contexto diretamente
}

const AdvancedTab: React.FC<AdvancedTabProps> = () => {
  const { userSettings, updateUserSetting } = useUser();
  const { addNotification } = useNotification();

  const handleFieldChange = (key: keyof UserSettings, value: any) => {
    updateUserSetting(key, value);

    // Feedback para mudanças de funcionalidades experimentais
    const feedbackMessages = {
      enableBetaFeatures: value ? 'Beta features enabled' : 'Beta features disabled',
      enableExperimentalFeatures: value ? 'Experimental features enabled' : 'Experimental features disabled',
    };

    const message = feedbackMessages[key as keyof typeof feedbackMessages];
    if (message) {
      addNotification({
        message: `⚗️ ${message}`,
        type: 'success'
      });
    }
  };

  return (
    <div className="space-y-8">
      {/* Experimental Features Section */}
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

      {/* Feature Information */}
      <section>
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <Zap className="w-5 h-5 text-yellow-400" /> Feature Details
        </h3>
        <div className="space-y-4">

          {/* Beta Features Info */}
          <div className="p-4 bg-blue-900/20 border border-blue-700/50 rounded-lg">
            <h4 className="text-sm font-medium text-blue-300 mb-2 flex items-center gap-2">
              <Beaker className="w-4 h-4" />
              Beta Features
            </h4>
            <div className="text-xs text-gray-300 space-y-1">
              <p>• <strong>Enhanced Analytics:</strong> Advanced usage charts and AI insights</p>
              <p>• <strong>Smart Suggestions:</strong> AI-powered project optimization tips</p>
              <p>• <strong>Collaboration Tools:</strong> Team sharing and commenting features</p>
              <p>• <strong>Export Enhancements:</strong> Multiple format exports with custom templates</p>
            </div>
          </div>

          {/* Experimental Features Info */}
          <div className="p-4 bg-orange-900/20 border border-orange-700/50 rounded-lg">
            <h4 className="text-sm font-medium text-orange-300 mb-2 flex items-center gap-2">
              <AlertTriangle className="w-4 h-4" />
              Experimental Features
            </h4>
            <div className="text-xs text-gray-300 space-y-1">
              <p>• <strong>AI Code Generation:</strong> Automatic code suggestions and generation</p>
              <p>• <strong>Voice Commands:</strong> Voice-controlled interface (browser support required)</p>
              <p>• <strong>Real-time Collaboration:</strong> Live editing with team members</p>
              <p>• <strong>Advanced AI Models:</strong> Access to latest cutting-edge models</p>
            </div>
            <div className="mt-3 p-2 bg-orange-800/30 rounded border border-orange-600/30">
              <p className="text-xs text-orange-200 flex items-center gap-1">
                <AlertTriangle className="w-3 h-3" />
                <strong>Warning:</strong> Experimental features may be unstable and could affect performance.
              </p>
            </div>
          </div>

          {/* Feature Roadmap */}
          <div className="p-4 bg-purple-900/20 border border-purple-700/50 rounded-lg">
            <h4 className="text-sm font-medium text-purple-300 mb-2">🚀 Coming Soon</h4>
            <div className="text-xs text-gray-300 space-y-1">
              <p>• <strong>Plugin System:</strong> Custom extensions and integrations</p>
              <p>• <strong>Mobile Apps:</strong> Native iOS and Android applications</p>
              <p>• <strong>Enterprise SSO:</strong> Single sign-on for organizations</p>
              <p>• <strong>Advanced Security:</strong> 2FA, audit logs, and compliance features</p>
            </div>
          </div>
        </div>
      </section>

      {/* Developer Options */}
      <section>
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <Beaker className="w-5 h-5 text-green-400" /> Developer Options
        </h3>
        <div className="space-y-4">
          <div className="p-4 bg-gray-900/50 border border-gray-700 rounded-lg">
            <p className="text-sm text-gray-300 mb-3">Debug and development tools for advanced users:</p>
            <div className="flex flex-wrap gap-2">
              <button className="px-3 py-1 text-xs bg-gray-700 hover:bg-gray-600 text-gray-300 rounded">
                View Console Logs
              </button>
              <button className="px-3 py-1 text-xs bg-gray-700 hover:bg-gray-600 text-gray-300 rounded">
                Export Debug Info
              </button>
              <button className="px-3 py-1 text-xs bg-gray-700 hover:bg-gray-600 text-gray-300 rounded">
                Performance Monitor
              </button>
              <button className="px-3 py-1 text-xs bg-gray-700 hover:bg-gray-600 text-gray-300 rounded">
                Feature Flags
              </button>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
};

export default AdvancedTab;
