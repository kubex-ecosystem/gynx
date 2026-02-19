import * as React from 'react';
import { useState } from 'react';
import { useNotification } from '../../contexts/NotificationContext';
import { useProjectContext } from '../../contexts/ProjectContext';
import { useUser } from '../../contexts/UserContext';
import { UserProfile } from '../../types';

interface ProfileModalProps {
  // Empty for now, uses context
}

const ProfileModal: React.FC<ProfileModalProps> = () => {
  const { isExample } = useProjectContext();
  const { name: userName, email: userEmail, setName, setEmail } = useUser();
  const { addNotification } = useNotification();

  // Criar um profile baseado nos dados do usuário
  const userProfile: UserProfile = {
    name: userName || '',
    email: userEmail || '',
  };

  const [profile, setProfile] = useState<UserProfile>(userProfile);

  const handleSave = () => {
    if (isExample) {
      addNotification({ message: 'Saving is disabled in example mode.', type: 'info' });
      return;
    }
    // Atualizar os dados do usuário
    setName(profile.name);
    setEmail(profile.email || null);
    addNotification({ message: 'Profile saved successfully.', type: 'success' });
  };

  return (
    <div className="space-y-6">
      <h3 className="text-lg font-semibold text-white">User Profile</h3>
      <div>
        <label htmlFor="name" className="text-sm font-medium text-gray-300">Name</label>
        <input
          type="text"
          id="name"
          value={profile.name}
          onChange={(e) => setProfile({ ...profile, name: e.target.value })}
          placeholder="Enter your name"
          className="w-full p-2 mt-1 bg-gray-900 border border-gray-600 rounded-md"
        />
      </div>
      <div>
        <label htmlFor="email" className="text-sm font-medium text-gray-300">Email</label>
        <input
          type="email"
          id="email"
          value={profile.email}
          onChange={(e) => setProfile({ ...profile, email: e.target.value })}
          placeholder="Enter your email"
          className="w-full p-2 mt-1 bg-gray-900 border border-gray-600 rounded-md"
        />
      </div>
      <div className="flex justify-end">
        <button
          onClick={handleSave}
          className="px-4 py-2 text-sm font-semibold text-white bg-purple-600 rounded-md hover:bg-purple-700 disabled:opacity-50"
          disabled={isExample}
        >
          Save
        </button>
      </div>
    </div>
  );
};

export default ProfileModal;
