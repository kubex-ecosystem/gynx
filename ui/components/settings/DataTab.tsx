import * as React from 'react';
import { useRef } from 'react';
import { useAppContext } from '../../contexts/AppContext';
import { useConfirmation } from '../../contexts/ConfirmationContext';
import { useNotification } from '../../contexts/NotificationContext';
import { useUser } from '../../contexts/UserContext';

import { defaultSettings, defaultUserProfile } from '../../constants';
import { getAllProjects, set, setProject } from '../../lib/idb';
import { clearAllAppData } from '../../lib/storage';
import { AppSettings, Project, UserProfile } from '../../types';

interface DataTabProps {
  isExample: boolean;
}

interface BackupData {
  timestamp: string;
  version: string;
  settings: AppSettings;
  profile: UserProfile;
  projects: Project[];
}

const DataTab: React.FC<DataTabProps> = ({ isExample }) => {
  const { addNotification } = useNotification();
  const { showConfirmation } = useConfirmation();
  const { resetApplication } = useAppContext();
  const { userSettings: settings, name: userName, email: userEmail } = useUser();
  const importFileRef = useRef<HTMLInputElement>(null);

  // Criar profile baseado nos dados do usuário
  const profile: UserProfile = {
    name: userName || 'User',
    email: userEmail || '',
  };

  const handleExport = async () => {
    try {
      const projects = await getAllProjects();

      if (projects.length === 0) {
        addNotification({ message: 'No data to export.', type: 'info' });
        return;
      }

      const backupData: BackupData = {
        timestamp: new Date().toISOString(),
        version: '2.0.0', // Updated version for new project-based structure
        settings,
        profile,
        projects,
      };

      const jsonString = JSON.stringify(backupData, null, 2);
      const blob = new Blob([jsonString], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      const date = new Date().toISOString().split('T')[0];
      a.download = `gemx_backup_${date}.json`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      addNotification({ message: 'Data exported successfully.', type: 'success' });
    } catch (error: any) {
      console.error('Export failed:', error);
      addNotification({ message: 'Failed to export data.', type: 'error' });
    }
  };

  const handleImport = async (file: File) => {
    if (!file) return;

    const reader = new FileReader();
    reader.onload = async (event) => {
      try {
        const content = event.target?.result as string;
        if (!content) {
          throw new Error('The selected file is empty.');
        }
        const importedData: BackupData = JSON.parse(content);

        if (importedData.version !== '2.0.0' || !importedData.settings || !importedData.profile || !Array.isArray(importedData.projects)) {
          throw new Error('The imported file has an invalid format.');
        }

        showConfirmation({
          title: 'Confirm Import',
          message: 'Are you sure you want to import this data? All your current projects and settings will be overwritten.',
          confirmText: 'Import',
          onConfirm: async () => {
            try {
              await clearAllAppData();

              const finalSettings = { ...defaultSettings, ...importedData.settings };
              const finalProfile = { ...defaultUserProfile, ...importedData.profile };

              // Persist settings and profile to the 'keyval' store
              await set('appSettings', finalSettings);
              await set('userProfile', finalProfile);

              // Persist all projects to the 'projects' store
              for (const project of importedData.projects) {
                await setProject(project);
              }

              addNotification({ message: 'Data imported successfully. The application will now reload.', type: 'success' });
              resetApplication();

            } catch (error: any) {
              addNotification({ message: error.message, type: 'error' });
            }
          },
          onCancel: () => {
            addNotification({ message: 'Import operation was cancelled.', type: 'info' });
          }
        });

      } catch (error: any) {
        console.error('Import failed:', error);
        addNotification({ message: error.message || 'An error occurred during import.', type: 'error' });
      } finally {
        if (importFileRef.current) importFileRef.current.value = '';
      }
    };
    reader.readAsText(file);
  };

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-white">Import & Export Data</h3>
      <p className="text-sm text-gray-400">Backup your projects and settings, or import them from a file. This is useful for moving data between browsers or devices.</p>
      <div className="p-4 bg-yellow-900/30 border border-yellow-700/50 text-yellow-300 rounded-lg text-sm">
        Warning: Importing data will replace all your current projects and settings. It is recommended to export your current data first.
      </div>
      <div className="flex gap-4">
        <input title="Import JSON" type="file" ref={importFileRef} onChange={(e) => e.target.files && handleImport(e.target.files[0])} className="hidden" accept=".json" />
        <button onClick={() => importFileRef.current?.click()} className="flex-1 px-4 py-2 text-sm font-semibold text-white bg-blue-600 rounded-md hover:bg-blue-700">
          Import Data
        </button>
        <button onClick={handleExport} disabled={isExample} className="flex-1 px-4 py-2 text-sm font-semibold text-white bg-gray-700 rounded-md hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed">
          Export Data
        </button>
      </div>
    </div>
  );
};

export default DataTab;
