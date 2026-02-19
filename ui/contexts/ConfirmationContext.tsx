import * as React from 'react';
import { createContext, useState, useContext, ReactNode, useCallback } from 'react';

interface ConfirmationOptions {
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  onConfirm: () => void;
  onCancel?: () => void;
}

interface ConfirmationContextType {
  showConfirmation: (options: ConfirmationOptions) => void;
  hideConfirmation: () => void;
  options: ConfirmationOptions | null;
  isOpen: boolean;
}

const ConfirmationContext = createContext<ConfirmationContextType | undefined>(undefined);

export const ConfirmationProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [options, setOptions] = useState<ConfirmationOptions | null>(null);
  const [isOpen, setIsOpen] = useState(false);

  const showConfirmation = useCallback((opts: ConfirmationOptions) => {
    setOptions(opts);
    setIsOpen(true);
  }, []);
  
  const hideConfirmation = useCallback(() => {
    setIsOpen(false);
    // Give time for animation before clearing options
    setTimeout(() => setOptions(null), 300);
  }, []);

  const value = {
    showConfirmation,
    hideConfirmation,
    options,
    isOpen,
  };

  return (
    <ConfirmationContext.Provider value={value}>
      {children}
    </ConfirmationContext.Provider>
  );
};

export const useConfirmation = (): ConfirmationContextType => {
  const context = useContext(ConfirmationContext);
  if (context === undefined) {
    throw new Error('useConfirmation must be used within a ConfirmationProvider');
  }
  return context;
};