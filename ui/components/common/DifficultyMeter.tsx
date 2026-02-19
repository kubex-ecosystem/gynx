import * as React from 'react';
import { Difficulty } from '../../types';

interface DifficultyMeterProps {
  difficulty: Difficulty;
}

const difficultyConfig: Record<Difficulty, { label: string; color: string; level: number }> = {
  [Difficulty.Low]: { label: 'Low', color: 'bg-green-500', level: 1 },
  [Difficulty.Medium]: { label: 'Medium', color: 'bg-yellow-500', level: 2 },
  [Difficulty.High]: { label: 'High', color: 'bg-red-500', level: 3 },
};

const DifficultyMeter: React.FC<DifficultyMeterProps> = ({ difficulty }) => {
  const config = difficultyConfig[difficulty];

  if (!config) {
    return null;
  }

  return (
    <div className="flex items-center gap-2" title={`Difficulty: ${config.label}`}>
      <div className="flex items-center gap-1">
        {[1, 2, 3].map(level => (
          <div
            key={level}
            className={`w-2 h-2 rounded-full ${level <= config.level ? config.color : 'bg-gray-600'}`}
          />
        ))}
      </div>
      <span className="text-xs text-gray-300">{config.label}</span>
    </div>
  );
};

export default DifficultyMeter;