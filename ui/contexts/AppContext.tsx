import * as React from 'react';
import { createContext, useState, useContext, ReactNode, useCallback } from 'react';

interface AppContextType {
  resetApplication: () => void;
}

const AppContext = createContext<AppContextType | undefined>(undefined);

export const AppProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [appKey, setAppKey] = useState(0);

  const resetApplication = useCallback(() => {
    setAppKey(prevKey => prevKey + 1);
  }, []);

  // By passing the key down, we allow the consumer to force a remount
  // of any component that uses this key.
  return (
    <AppContext.Provider value={{ resetApplication }}>
      {React.cloneElement(children as React.ReactElement, { key: appKey })}
    </AppContext.Provider>
  );
};

export const useAppContext = (): AppContextType => {
  const context = useContext(AppContext);
  if (context === undefined) {
    throw new Error('useAppContext must be used within an AppProvider');
  }
  return context;
};