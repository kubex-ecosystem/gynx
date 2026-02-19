// FIX: Corrected import to use the new generic 'clear' function
import { clear } from './idb';

// List of all keys managed by usePersistentState or stored directly
const APP_STORAGE_KEYS = [
    'projectFiles',
    'analysisHistory',
    'kanbanState',
    'appSettings',
    'userProfile',
    'usageTracking',
    'allChatHistories',
    'locale', // From LanguageContext
    'analysisFeedback' // From SuggestionsDisplay
];

/**
 * Clears all application data from both IndexedDB and localStorage.
 * This is a destructive operation used for data import or a hard reset.
 */
export const clearAllAppData = async (): Promise<void> => {
    try {
        // Clear IndexedDB store
        await clear();
        console.log('IndexedDB store cleared.');

        // Clear localStorage keys
        APP_STORAGE_KEYS.forEach(key => {
            localStorage.removeItem(key);
        });
        console.log('LocalStorage app keys cleared.');
        
    } catch (error) {
        console.error('Failed to clear all application data:', error);
        throw new Error('Could not clear existing application data. Import aborted.');
    }
};