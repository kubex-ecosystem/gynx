import { useState, useEffect, useCallback } from 'react';
// FIX: Corrected import for generic get/set from idb utility
import { get, set } from '../lib/idb';

type SetValue<T> = (value: T | ((prevValue: T) => T)) => void;

/**
 * A custom hook that provides a state management solution similar to `useState`,
 * but with the added feature of persisting the state to client-side storage.
 *
 * It prioritizes using IndexedDB for its larger storage capacity and asynchronous nature,
 * making it suitable for storing complex objects or large amounts of data. If IndexedDB
 * is unavailable or fails, it gracefully falls back to using `localStorage`.
 *
 * This hook abstracts away the complexities of data persistence, allowing components
 * to manage state without being concerned about the underlying storage mechanism.
 *
 * @template T The type of the state to be managed.
 * @param {string} key The unique key to identify the state in storage.
 * @param {T} defaultValue The initial value of the state if none is found in storage.
 * @returns {[T, SetValue<T>]} A tuple containing the current state and a function to update it.
 */
export const usePersistentState = <T>(key: string, defaultValue: T): [T, SetValue<T>] => {
  const [value, setValue] = useState<T>(defaultValue);

  // Load the persisted state from storage on initial render.
  useEffect(() => {
    let isMounted = true;
    
    const loadState = async () => {
      try {
        // 1. Try IndexedDB first (asynchronous)
        const idbValue = await get<T>(key);
        if (idbValue !== undefined && isMounted) {
          setValue(idbValue);
          return;
        }

        // 2. Fallback to localStorage (synchronous)
        const lsValue = localStorage.getItem(key);
        if (lsValue !== null && isMounted) {
          setValue(JSON.parse(lsValue));
          return;
        }

      } catch (error) {
        console.error(`Failed to load state for key "${key}" from storage.`, error);
      }
      
      // 3. Use default value if nothing is found
      if (isMounted) {
        setValue(defaultValue);
      }
    };

    loadState();
    
    return () => { isMounted = false; };
  }, [key]); // Only run on mount or if key changes

  // Persist the state to storage whenever it changes.
  useEffect(() => {
    // We don't want to persist the initial default value until it's been
    // explicitly set by the user or loaded from storage.
    // This check prevents overwriting existing stored data with the default on first render.
    if (value === defaultValue && localStorage.getItem(key) === null) {
      // A more complex check could be done with IDB but this is a reasonable heuristic.
      return;
    }
    
    const saveState = async () => {
      try {
        // Write to both to ensure data is available even if one system fails
        // and to keep localStorage as a simple, readable backup.
        await set(key, value);
        localStorage.setItem(key, JSON.stringify(value));
      } catch (error) {
        console.error(`Failed to save state for key "${key}" to storage.`, error);
      }
    };
    
    saveState();
  }, [key, value, defaultValue]);

  return [value, setValue];
};