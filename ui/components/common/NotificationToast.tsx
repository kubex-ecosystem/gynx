import * as React from 'react';
import { useEffect } from 'react';
import { motion } from 'framer-motion';
import { CheckCircle, AlertTriangle, Info, X } from 'lucide-react';
// FIX: Corrected import path for types
import { Notification, NotificationType } from '../../types';

interface NotificationToastProps {
  notification: Notification;
  onDismiss: () => void;
}

const icons: Record<NotificationType, React.ElementType> = {
  success: CheckCircle,
  error: AlertTriangle,
  info: Info,
};

const theme: Record<NotificationType, { bg: string; border: string; icon: string }> = {
  success: {
    bg: 'bg-green-900/50',
    border: 'border-green-700',
    icon: 'text-green-400',
  },
  error: {
    bg: 'bg-red-900/50',
    border: 'border-red-700',
    icon: 'text-red-400',
  },
  info: {
    bg: 'bg-blue-900/50',
    border: 'border-blue-700',
    icon: 'text-blue-400',
  },
};

const NotificationToast: React.FC<NotificationToastProps> = ({ notification, onDismiss }) => {
  const { message, type, duration = 5000 } = notification;

  useEffect(() => {
    const timer = setTimeout(() => {
      onDismiss();
    }, duration);

    return () => clearTimeout(timer);
  }, [notification, duration, onDismiss]);

  const Icon = icons[type];
  const colors = theme[type];

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 50, scale: 0.5 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      exit={{ opacity: 0, y: 20, scale: 0.8 }}
      transition={{ type: 'spring', stiffness: 300, damping: 25 }}
      className={`p-4 w-full ${colors.bg} border ${colors.border} rounded-xl shadow-lg flex items-start gap-3 backdrop-blur-md`}
    >
      <div className={`shrink-0 ${colors.icon}`}>
        <Icon className="w-6 h-6" />
      </div>
      <div className="flex-grow text-sm text-gray-200">
        <p>{message}</p>
      </div>
      <button
        onClick={onDismiss}
        className="p-1 rounded-full text-gray-400 hover:bg-gray-700 transition-colors"
        aria-label="Dismiss notification"
      >
        <X className="w-4 h-4" />
      </button>
    </motion.div>
  );
};

export default NotificationToast;