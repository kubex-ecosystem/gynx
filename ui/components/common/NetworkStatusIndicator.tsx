import * as React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { WifiOff } from 'lucide-react';
import { useNetworkStatus } from '../../hooks/useNetworkStatus';

const NetworkStatusIndicator: React.FC = () => {
  const isOnline = useNetworkStatus();

  return (
    <AnimatePresence>
      {!isOnline && (
        <motion.div
          initial={{ opacity: 0, y: 50 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: 50 }}
          className="fixed bottom-4 left-4 z-[100] p-3 bg-red-900/70 border border-red-700 text-red-300 rounded-lg shadow-lg flex items-center gap-3 backdrop-blur-md"
        >
          <WifiOff className="w-5 h-5" />
          <span className="text-sm font-medium">You are offline</span>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default NetworkStatusIndicator;