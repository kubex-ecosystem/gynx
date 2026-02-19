import * as React from 'react';
import { motion } from 'framer-motion';

interface ViabilityScoreProps {
  score: number; // score out of 10
}

const ViabilityScore: React.FC<ViabilityScoreProps> = ({ score }) => {
  const size = 120;
  const strokeWidth = 10;
  const center = size / 2;
  const radius = center - strokeWidth / 2;
  const circumference = 2 * Math.PI * radius;

  const scorePercentage = score / 10;
  const strokeDashoffset = circumference * (1 - scorePercentage);

  const getColor = (s: number) => {
    if (s <= 3) return '#ef4444'; // red-500
    if (s <= 6) return '#f59e0b'; // amber-500
    return '#22c55e'; // green-500
  };

  const color = getColor(score);

  return (
    <div className="relative" style={{ width: size, height: size }}>
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`} className="-rotate-90">
        {/* Background circle */}
        <circle
          cx={center}
          cy={center}
          r={radius}
          fill="transparent"
          stroke="rgba(255, 255, 255, 0.1)"
          strokeWidth={strokeWidth}
        />
        {/* Progress circle */}
        <motion.circle
          cx={center}
          cy={center}
          r={radius}
          fill="transparent"
          stroke={color}
          strokeWidth={strokeWidth}
          strokeDasharray={circumference}
          strokeLinecap="round"
          initial={{ strokeDashoffset: circumference }}
          animate={{ strokeDashoffset }}
          transition={{ duration: 1.5, ease: "easeOut" }}
        />
      </svg>
      <div className="absolute inset-0 flex flex-col items-center justify-center">
        <motion.span 
          className="text-4xl font-bold text-white"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.5, duration: 1 }}
        >
          {score}
        </motion.span>
        <span className="text-sm text-gray-400">/10</span>
      </div>
    </div>
  );
};

export default ViabilityScore;