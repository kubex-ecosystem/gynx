import * as React from 'react';
import { motion } from 'framer-motion';
import { LayoutDashboard, FileText, BarChart3, Bot, KanbanSquare, GitCompareArrows } from 'lucide-react';
import { ViewType } from '../../types';

interface NavigationBarProps {
  currentView: ViewType;
  onNavigate: (view: ViewType) => void;
  hasAnalysis: boolean;
  isAnalysisOpen: boolean;
}

const NavItem: React.FC<{
  label: string;
  icon: React.ReactNode;
  isActive: boolean;
  onClick: () => void;
  disabled?: boolean;
}> = ({ label, icon, isActive, onClick, disabled }) => {
  return (
    <motion.button
      onTap={onClick}
      disabled={disabled}
      className={`relative px-3 py-2 text-sm font-medium rounded-md transition-colors flex items-center gap-2 ${
        isActive ? 'text-white' : 'text-gray-400 hover:bg-gray-700/50 hover:text-white'
      } ${disabled ? 'opacity-50 cursor-not-allowed' : ''}`}
      whileHover={{ scale: disabled || isActive ? 1 : 1.05 }}
      whileTap={{ scale: disabled ? 1 : 0.95 }}
    >
      {isActive && (
        <motion.div
          layoutId="active-nav-indicator"
          className="absolute inset-0 bg-purple-900/40 rounded-md z-0"
          style={{ originY: "0px" }}
          transition={{ type: "spring", stiffness: 300, damping: 25 }}
        />
      )}
      <span className="relative z-10">{icon}</span>
      <span className="relative z-10">{label}</span>
    </motion.button>
  );
};

const NavigationBar: React.FC<NavigationBarProps> = ({
  currentView,
  onNavigate,
  hasAnalysis,
  isAnalysisOpen
}) => {
  const navItems = [
    { view: ViewType.Dashboard, label: "Dashboard", icon: <LayoutDashboard className="w-4 h-4" />, disabled: false },
    { view: ViewType.Input, label: "New Analysis", icon: <FileText className="w-4 h-4" />, disabled: false },
    { view: ViewType.Analysis, label: "Current Analysis", icon: <BarChart3 className="w-4 h-4" />, disabled: !hasAnalysis },
    { view: ViewType.Chat, label: "Chat", icon: <Bot className="w-4 h-4" />, disabled: !isAnalysisOpen },
    { view: ViewType.Kanban, label: "Kanban", icon: <KanbanSquare className="w-4 h-4" />, disabled: !isAnalysisOpen },
    { view: ViewType.Evolution, label: "Evolution", icon: <GitCompareArrows className="w-4 h-4" />, disabled: !hasAnalysis },
  ];

  return (
    <div className="p-1.5 bg-gray-800/60 border border-gray-700 rounded-lg flex items-center justify-center space-x-2 relative backdrop-blur-sm">
      {navItems.map(item => (
        <NavItem
          key={item.view}
          label={item.label}
          icon={item.icon}
          isActive={currentView === item.view}
          onClick={() => onNavigate(item.view)}
          disabled={item.disabled}
        />
      ))}
    </div>
  );
};

export default NavigationBar;