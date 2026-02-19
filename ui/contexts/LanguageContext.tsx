import * as React from 'react';
import { createContext, useState, useEffect, useContext, ReactNode, useCallback } from 'react';

type Locale = 'en-US' | 'pt-BR';
type Translations = Record<string, any>;

interface LanguageContextType {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  translations: Translations;
  loadNamespace: (namespace: string) => Promise<void>;
}

const LanguageContext = createContext<LanguageContextType | undefined>(undefined);

// Helper function to explicitly handle dynamic imports, avoiding bundler-specific features like import.meta.glob.
// This makes the code compatible with environments that execute modules directly in the browser.
const loadTranslationModule = (locale: Locale, namespace: string): Promise<{ default: any }> => {
  const path = `${locale}/${namespace}`;
  switch (path) {
    // English (en-US)
    // FIX: Wrap named exports in a `default` property to match expected return type
    case 'en-US/analysis': return import('../public/locales/en-US/analysis.ts').then(m => ({ default: m.analysisEnUS }));
    case 'en-US/auth': return import('../public/locales/en-US/auth.ts').then(m => ({ default: m.authEnUS }));
    case 'en-US/chat': return import('../public/locales/en-US/chat.ts').then(m => ({ default: m.chatEnUS }));
    case 'en-US/common': return import('../public/locales/en-US/common.ts').then(m => ({ default: m.commonEnUS }));
    case 'en-US/dashboard': return import('../public/locales/en-US/dashboard.ts').then(m => ({ default: m.dashboardEnUS }));
    case 'en-US/dataSources': return import('../public/locales/en-US/dataSources.ts').then(m => ({ default: m.dataSources }));
    case 'en-US/example': return import('../public/locales/en-US/example.ts').then(m => ({ default: m.exampleEnUS }));
    case 'en-US/files': return import('../public/locales/en-US/files.ts').then(m => ({ default: m.files }));
    case 'en-US/githubSearch': return import('../public/locales/en-US/githubSearch.ts').then(m => ({ default: m.githubSearch }));
    case 'en-US/history': return import('../public/locales/en-US/history.ts').then(m => ({ default: m.historyEnUS }));
    case 'en-US/importExport': return import('../public/locales/en-US/importExport.ts').then(m => ({ default: m.importExport }));
    case 'en-US/input': return import('../public/locales/en-US/input.ts').then(m => ({ default: m.inputEnUS }));
    case 'en-US/kanban': return import('../public/locales/en-US/kanban.ts').then(m => ({ default: m.kanbanEnUS }));
    case 'en-US/landing': return import('../public/locales/en-US/landing.ts').then(m => ({ default: m.landingEnUS }));
    case 'en-US/notifications': return import('../public/locales/en-US/notifications.ts').then(m => ({ default: m.notifications }));
    case 'en-US/profile': return import('../public/locales/en-US/profile.ts').then(m => ({ default: m.profileEnUS }));
    case 'en-US/settings': return import('../public/locales/en-US/settings.ts').then(m => ({ default: m.settingsEnUS }));
    case 'en-US/tabs': return import('../public/locales/en-US/tabs.ts').then(m => ({ default: m.tabsEnUS }));
    case 'en-US/tokenUsage': return import('../public/locales/en-US/tokenUsage.ts').then(m => ({ default: m.tokenUsage }));

    // Portuguese (pt-BR)
    case 'pt-BR/analysis': return import('../public/locales/pt-BR/analysis.ts').then(m => ({ default: m.analysisPtBR }));
    case 'pt-BR/auth': return import('../public/locales/pt-BR/auth.ts').then(m => ({ default: m.authPtBR }));
    case 'pt-BR/chat': return import('../public/locales/pt-BR/chat.ts').then(m => ({ default: m.chatPtBR }));
    case 'pt-BR/common': return import('../public/locales/pt-BR/common.ts').then(m => ({ default: m.commonPtBR }));
    case 'pt-BR/dashboard': return import('../public/locales/pt-BR/dashboard.ts').then(m => ({ default: m.dashboardPtBR }));
    case 'pt-BR/dataSources': return import('../public/locales/pt-BR/dataSources.ts').then(m => ({ default: m.dataSources }));
    case 'pt-BR/example': return import('../public/locales/pt-BR/example.ts').then(m => ({ default: m.examplePtBR }));
    case 'pt-BR/files': return import('../public/locales/pt-BR/files.ts').then(m => ({ default: m.files }));
    case 'pt-BR/githubSearch': return import('../public/locales/pt-BR/githubSearch.ts').then(m => ({ default: m.githubSearch }));
    case 'pt-BR/history': return import('../public/locales/pt-BR/history.ts').then(m => ({ default: m.historyPtBR }));
    case 'pt-BR/importExport': return import('../public/locales/pt-BR/importExport.ts').then(m => ({ default: m.importExport }));
    case 'pt-BR/input': return import('../public/locales/pt-BR/input.ts').then(m => ({ default: m.inputPtBR }));
    case 'pt-BR/kanban': return import('../public/locales/pt-BR/kanban.ts').then(m => ({ default: m.kanbanPtBR }));
    case 'pt-BR/landing': return import('../public/locales/pt-BR/landing.ts').then(m => ({ default: m.landingPtBR }));
    case 'pt-BR/notifications': return import('../public/locales/pt-BR/notifications.ts').then(m => ({ default: m.notifications }));
    case 'pt-BR/profile': return import('../public/locales/pt-BR/profile.ts').then(m => ({ default: m.profilePtBR }));
    case 'pt-BR/settings': return import('../public/locales/pt-BR/settings.ts').then(m => ({ default: m.settingsPtBR }));
    case 'pt-BR/tabs': return import('../public/locales/pt-BR/tabs.ts').then(m => ({ default: m.tabsPtBR }));
    case 'pt-BR/tokenUsage': return import('../public/locales/pt-BR/tokenUsage.ts').then(m => ({ default: m.tokenUsage }));

    default:
      return Promise.reject(new Error(`Translation module not found for path: ${path}`));
  }
};


const getInitialLocale = (): Locale => {
    const storedLocale = localStorage.getItem('locale') as Locale;
    if (storedLocale && ['en-US', 'pt-BR'].includes(storedLocale)) {
        return storedLocale;
    }
    const browserLang = navigator.language;
    if (browserLang.startsWith('pt')) {
        return 'pt-BR';
    }
    return 'en-US';
};

export const LanguageProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [locale, setLocaleState] = useState<Locale>(getInitialLocale());
  const [translations, setTranslations] = useState<Translations>({});
  const [loadedNamespaces, setLoadedNamespaces] = useState<Record<string, boolean>>({});
  const [loadingNamespaces, setLoadingNamespaces] = useState<Record<string, boolean>>({});
  const [isInitialLoad, setIsInitialLoad] = useState(true);

  const setLocale = (newLocale: Locale) => {
    setLocaleState(newLocale);
    setTranslations({});
    setLoadedNamespaces({});
    setLoadingNamespaces({});
    setIsInitialLoad(true);
    localStorage.setItem('locale', newLocale);
  };

  const loadNamespace = useCallback(async (namespace: string) => {
    const namespaceKey = `${locale}-${namespace}`;
    if (loadedNamespaces[namespaceKey] || loadingNamespaces[namespaceKey]) {
      return;
    }

    setLoadingNamespaces(prev => ({ ...prev, [namespaceKey]: true }));
    try {
      // Use the new explicit loader function
      const mod = await loadTranslationModule(locale, namespace);
      const data = mod.default;
      
      setTranslations(prev => ({
        ...prev,
        [namespace]: data,
      }));
      setLoadedNamespaces(prev => ({ ...prev, [namespaceKey]: true }));

    } catch (error) {
      console.error(`Failed to load translations for ${locale}/${namespace}`, error);
    } finally {
      setLoadingNamespaces(prev => ({ ...prev, [namespaceKey]: false }));
    }
  }, [locale, loadedNamespaces, loadingNamespaces]);
  
  useEffect(() => {
    if (isInitialLoad) {
        // Load essential namespaces for the initial render
        Promise.all([
          loadNamespace('common'),
          loadNamespace('landing')
        ]).finally(() => {
            setIsInitialLoad(false);
        });
    }
  }, [locale, isInitialLoad, loadNamespace]);
  
  const value = {
    locale,
    setLocale,
    translations,
    loadNamespace,
  };

  return (
    <LanguageContext.Provider value={value}>
      {!isInitialLoad ? children : null}
    </LanguageContext.Provider>
  );
};

export const useLanguage = (): LanguageContextType => {
  const context = useContext(LanguageContext);
  if (context === undefined) {
    throw new Error('useLanguage must be used within a LanguageProvider');
  }
  return context;
};
