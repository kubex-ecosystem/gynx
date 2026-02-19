import { AnimatePresence, motion } from 'framer-motion';
import { BarChart3, Beaker, Database, Link as LinkIcon, Settings as SettingsIcon, Shield, User, X } from 'lucide-react';
import * as React from 'react';
import { useState } from 'react';
import { useProjectContext } from '../../contexts/ProjectContext';
import { useUser } from '../../contexts/UserContext';
import ProfileModal from '../user/ProfileModal';
import AdvancedTab from './AdvancedTab';
import AnalyticsTab from './AnalyticsTab';
import DataTab from './DataTab';
import GeneralTab from './GeneralTab';
import IntegrationsTab from './IntegrationsTab';
import SecurityTab from './SecurityTab';

type Tab = 'profile' | 'general' | 'analytics' | 'security' | 'integrations' | 'advanced' | 'data';

const UserSettingsModal: React.FC = () => {
  const { isExample } = useProjectContext();
  const { isUserSettingsModalOpen, setIsUserSettingsModalOpen } = useUser();
  const [activeTab, setActiveTab] = useState<Tab>('profile');

  const tabs = [
    { id: 'profile', label: 'Profile', icon: User },
    { id: 'general', label: 'General', icon: SettingsIcon },
    { id: 'analytics', label: 'Analytics', icon: BarChart3 },
    { id: 'security', label: 'Security', icon: Shield },
    { id: 'integrations', label: 'Integrations', icon: LinkIcon },
    { id: 'advanced', label: 'Advanced', icon: Beaker },
    { id: 'data', label: 'Data', icon: Database },
  ];

  return (
    <AnimatePresence>
      {isUserSettingsModalOpen && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          onClick={() => setIsUserSettingsModalOpen(false)}
          className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4"
        >
          <motion.div
            initial={{ scale: 0.95, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            exit={{ scale: 0.95, opacity: 0 }}
            onClick={(e) => e.stopPropagation()}
            className="bg-gray-800 border border-gray-700 rounded-xl w-full max-w-3xl flex flex-col shadow-2xl h-[80vh]"
          >
            {/* Header */}
            <div className="flex items-center justify-between p-4 border-b border-gray-700">
              <h2 className="text-xl font-bold text-white">Settings</h2>
              <button title='Close' onClick={() => setIsUserSettingsModalOpen(false)} className="p-1 rounded-full text-gray-400 hover:bg-gray-700">
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="flex flex-grow overflow-hidden">
              {/* Sidebar */}
              <div className="w-1/4 p-4 border-r border-gray-700">
                <nav className="space-y-1">
                  {tabs.map(tab => (
                    <button
                      key={tab.id}
                      onClick={() => setActiveTab(tab.id as Tab)}
                      className={`w-full flex items-center gap-3 px-3 py-2 text-sm font-medium rounded-md ${activeTab === tab.id
                        ? 'bg-purple-900/50 text-white'
                        : 'text-gray-400 hover:bg-gray-700/50 hover:text-white'
                        }`}
                    >
                      <tab.icon className="w-4 h-4" />
                      {tab.label}
                    </button>
                  ))}
                </nav>
              </div>

              {/* Content */}
              <div className="w-3/4 p-6 overflow-y-auto">
                {activeTab === 'profile' && <ProfileModal />}
                {activeTab === 'general' && <GeneralTab />}
                {activeTab === 'analytics' && <AnalyticsTab />}
                {activeTab === 'security' && <SecurityTab />}
                {activeTab === 'integrations' && <IntegrationsTab />}
                {activeTab === 'advanced' && <AdvancedTab />}
                {activeTab === 'data' && <DataTab isExample={isExample} />}
              </div>
            </div>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default UserSettingsModal;
