import * as React from 'react';
import { Calculator } from 'lucide-react';
import { UsageMetadata } from '../../types';

interface SubtleTokenUsageProps {
  usageMetadata?: UsageMetadata;
  label: string;
}

const SubtleTokenUsage: React.FC<SubtleTokenUsageProps> = ({ usageMetadata, label }) => {

  if (!usageMetadata) {
    return null;
  }

  return (
    <div
      className="flex items-center justify-center gap-3 text-xs text-gray-400 p-2 bg-gray-800/50 border border-gray-700 rounded-lg max-w-md mx-auto"
      aria-label="Token usage metadata for the last analysis"
    >
      <Calculator className="w-4 h-4 text-gray-500 shrink-0" />
      <div className="flex flex-wrap items-center justify-center gap-x-2 gap-y-1">
        <span className="font-semibold">{label}:</span>
        <span>{usageMetadata.totalTokenCount.toLocaleString('en-US')} Tokens</span>
      </div>
    </div>
  );
};

export default SubtleTokenUsage;