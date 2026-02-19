import * as React from 'react';
import { TrendingUp, AlertCircle } from 'lucide-react';

interface TokenUsageAlertProps {
  consumed: number;
  limit: number;
}

const TokenUsageAlert: React.FC<TokenUsageAlertProps> = ({ consumed, limit }) => {
  const percentage = limit > 0 ? Math.round((consumed / limit) * 100) : 0;

  const getBarColor = () => {
    if (percentage > 90) return 'bg-red-500';
    if (percentage > 70) return 'bg-yellow-500';
    return 'bg-blue-500';
  };

  return (
    <div className="p-4 bg-gray-800/50 border border-gray-700 rounded-lg">
      <h3 className="text-md font-semibold text-white flex items-center gap-2">
        <TrendingUp className="w-5 h-5 text-gray-400" />
        Monthly Usage
      </h3>
      <div className="mt-3">
        <p className="text-sm text-gray-400">
          {`You have used ${consumed.toLocaleString()} of ${limit.toLocaleString()} tokens (${percentage}%).`}
        </p>
        <div className="relative w-full bg-gray-700 rounded-full h-2 mt-2">
          <div
            className={`h-2 rounded-full transition-all duration-500 ${getBarColor()}`}
            style={{ width: `${percentage}%` }}
          />
        </div>
      </div>
      {percentage > 90 && (
        <div className="mt-3 text-xs text-yellow-400 flex items-center gap-2 p-2 bg-yellow-900/30 rounded-md">
          <AlertCircle className="w-4 h-4" />
          <span>You are approaching your token limit.</span>
        </div>
      )}
    </div>
  );
};

export default TokenUsageAlert;