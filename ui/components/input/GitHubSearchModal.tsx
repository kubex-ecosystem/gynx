import { AnimatePresence, motion } from 'framer-motion';
import { AlertTriangle, Loader2, Search, Star, X } from 'lucide-react';
import * as React from 'react';
import { useState } from 'react';
import { useNotification } from '../../contexts/NotificationContext';
import { listUserRepos } from '../../services/integrations/github';
import { GitHubRepoListItem } from '../../types';

interface GitHubSearchModalProps {
  isOpen: boolean;
  onClose: () => void;
  onImport: (owner: string, repo: string) => void;
  githubPat: string;
}

const GitHubSearchModal: React.FC<GitHubSearchModalProps> = ({ isOpen, onClose, onImport, githubPat }) => {
  const { addNotification } = useNotification();
  const [username, setUsername] = useState('');
  const [repos, setRepos] = useState<GitHubRepoListItem[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSearch = async () => {
    if (!username.trim()) return;
    setIsLoading(true);
    setError(null);
    setRepos([]);
    try {
      const results = await listUserRepos(username, githubPat || '');
      setRepos(results);
    } catch (err: any) {
      setError(err.message);
      addNotification({ message: err.message, type: 'error' });
    } finally {
      setIsLoading(false);
    }
  };

  const handleRepoSelect = (repo: GitHubRepoListItem) => {
    onImport(repo.owner.login, repo.name);
  };

  return (
    <AnimatePresence>
      {isOpen && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          onTap={onClose}
          className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4"
        >
          <motion.div
            initial={{ scale: 0.95, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            exit={{ scale: 0.95, opacity: 0 }}
            onTap={(e) => e.stopPropagation()}
            className="bg-gray-800 border border-gray-700 rounded-xl w-full max-w-2xl flex flex-col shadow-2xl h-[70vh]"
          >
            {/* Header */}
            <div className="flex items-center justify-between p-4 border-b border-gray-700">
              <h2 className="text-xl font-bold text-white">Search GitHub Repositories</h2>
              <button title='Close' onClick={onClose} className="p-1 rounded-full text-gray-400 hover:bg-gray-700">
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Search Input */}
            <div className="p-4">
              <div className="flex gap-2">
                <input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="Enter a GitHub username or organization"
                  className="flex-grow p-2 bg-gray-900 border border-gray-600 rounded-md"
                  onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                />
                <button onClick={handleSearch} disabled={isLoading || !username.trim()} className="px-4 py-2 text-sm font-semibold text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:bg-gray-600 flex items-center gap-2">
                  {isLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Search className="w-4 h-4" />}
                  Search
                </button>
              </div>
              {!githubPat && (
                <div className="mt-3 p-2 text-xs bg-yellow-900/50 text-yellow-300 rounded-md flex items-center gap-2">
                  <AlertTriangle className="w-4 h-4" />
                  <span>Provide a GitHub PAT in Settings for private repos and higher rate limits.</span>
                </div>
              )}
            </div>

            {/* Results */}
            <div className="flex-grow p-4 overflow-y-auto">
              {isLoading && (
                <div className="flex justify-center items-center h-full">
                  <Loader2 className="w-8 h-8 text-purple-400 animate-spin" />
                </div>
              )}
              {error && (
                <div className="text-center text-red-400">{error}</div>
              )}
              {!isLoading && !error && repos.length === 0 && (
                <div className="text-center text-gray-500">Enter a username and click search to see repositories.</div>
              )}
              <div className="space-y-2">
                {repos.map(repo => (
                  <button key={repo.id} onClick={() => handleRepoSelect(repo)} className="w-full text-left p-3 bg-gray-900/50 border border-gray-700 rounded-lg hover:bg-gray-700/80 transition-colors">
                    <div className="flex justify-between items-center">
                      <p className="font-semibold text-blue-400">{repo.full_name}</p>
                      <div className="flex items-center gap-4 text-xs text-gray-400">
                        <span className="flex items-center gap-1"><Star className="w-3 h-3" /> {repo.stargazers_count}</span>
                      </div>
                    </div>
                    <p className="text-sm text-gray-400 mt-1 line-clamp-2">{repo.description}</p>
                  </button>
                ))}
              </div>
            </div>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default GitHubSearchModal;
