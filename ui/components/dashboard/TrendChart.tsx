import * as React from 'react';
import Sparkline from '../common/Sparkline';

interface TrendChartProps {
  data: number[];
  className?: string;
}

const TrendChart: React.FC<TrendChartProps> = ({ data, className = '' }) => {
  if (data.length < 2) {
    return (
      <div className="h-full w-full flex items-center justify-center bg-gray-900/30 rounded-lg">
        <p className="text-gray-500 text-sm">Dados insuficientes para exibir tendência</p>
      </div>
    );
  }

  return (
    <div className={`relative h-full w-full flex items-center ${className}`}>
      <div className="h-full flex flex-col justify-between text-xs text-gray-500 py-1 pr-2">
        <span>10</span>
        <span>5</span>
        <span>0</span>
      </div>
      <div className="relative h-full flex-grow">
        <div className="absolute top-0 left-0 w-full h-full border-b border-l border-gray-700/50">
          <div className="absolute top-1/2 left-0 w-full border-t border-dashed border-gray-700/50"></div>
        </div>
        <Sparkline
          data={data}
          width={400} // width/height will be controlled by parent's flex/size
          height={100}
          className="w-full h-full"
          stroke="rgb(96 165 250)"
          strokeWidth={2}
        />
      </div>
    </div>
  );
};

export default TrendChart;