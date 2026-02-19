import { useEffect, useMemo } from 'react';
import { useLanguage } from '../contexts/LanguageContext';
import { TranslationNamespace } from '../public/locales/types';

const getDeepValue = (obj: any, path: string[]): any => {
  let current = obj;
  for (let i = 0; i < path.length; i++) {
    const key = path[i];

    if (current === null || current === undefined) {
      return undefined;
    }

    if (typeof current !== 'object') {
      return undefined;
    }
    
    // Find a key in the current object that matches the path segment case-insensitively.
    const currentKeys = Object.keys(current);
    const foundKey = currentKeys.find(k => k.toLowerCase() === key.toLowerCase());

    if (foundKey === undefined) {
      return undefined;
    }

    current = current[foundKey];
  }

  return current;
};


export const useTranslation = (namespaces: TranslationNamespace | TranslationNamespace[] = 'common') => {
  const { translations, loadNamespace } = useLanguage();
  const nsArray = Array.isArray(namespaces) ? namespaces : [namespaces];

  useEffect(() => {
    nsArray.forEach(ns => {
      loadNamespace(ns as string);
    });
  }, [nsArray, loadNamespace]);

  const isLoading = useMemo(() => {
    return nsArray.some(ns => translations[ns as string] === undefined);
  }, [nsArray, translations]);

  const t = (key: string, options?: Record<string, string | number>): string => {
    const keyParts = key.split(':');
    let result: any;

    const validNamespaces = ['common', 'analysis', 'auth', 'chat', 'dashboard', 'dataSources', 'example', 'files', 'githubSearch', 'history', 'importExport', 'input', 'kanban', 'landing', 'notifications', 'profile', 'settings', 'tabs', 'tokenUsage'];

    if (keyParts.length > 1) {
      // Explicit namespace syntax: "namespace:key.subkey"
      const [ns, lookupKey] = keyParts;
      const path = lookupKey.split('.');

      if (validNamespaces.includes(ns) && translations[ns] === undefined) {
        loadNamespace(ns);
        return ''; // Return empty while loading
      }

      result = getDeepValue(translations[ns], path);
    } else {
      // Implicit or no namespace
      const path = key.split('.');
      const potentialNs = path[0];

      // Check for implicit namespace: "namespace.key.subkey"
      if (path.length > 1 && validNamespaces.includes(potentialNs)) {
        if (translations[potentialNs] === undefined) {
          loadNamespace(potentialNs);
          return ''; // Return empty while loading
        }
        // Found implicit namespace, adjust path and search
        const namespaceKey = path.slice(1);
        result = getDeepValue(translations[potentialNs], namespaceKey);
      }
      
      // If not found via implicit, search in loaded namespaces: "key.subkey"
      if (result === undefined) {
        for (const searchNs of nsArray) {
          if (translations[searchNs as string] !== undefined) {
            const found = getDeepValue(translations[searchNs as string], path);
            if (found !== undefined) {
              result = found;
              break;
            }
          }
        }
      }
    }

    if (result === undefined) {
      console.warn(`Translation key not found: ${key}`);
      return key;
    }

    if (options && typeof result === 'string') {
      return Object.keys(options).reduce((acc, optionKey) => {
        const regex = new RegExp(`{${optionKey}}`, 'g');
        return acc.replace(regex, String(options[optionKey]));
      }, result);
    }

    return result;
  };

  return { t, isLoading };
};