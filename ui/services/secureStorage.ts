// Secure IndexedDB storage for user data
// Implementa storage seguro com encryption para dados sensíveis

interface UserDataStorage {
  profile: {
    name: string | null;
    email: string | null;
    avatarUrl: string | null;
    plan: 'free' | 'pro' | 'enterprise' | null;
    isEmailVerified: boolean;
  };
  settings: any; // UserSettings
  integrations: any; // IntegrationSettings
  usageTracking: any; // UsageTracking
  user: any; // User complete object
}

class SecureUserStorage {
  private dbName = 'KubexAnalyzerUserDB';
  private dbVersion = 1;
  private storeName = 'userData';
  private db: IDBDatabase | null = null;

  async init(): Promise<void> {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open(this.dbName, this.dbVersion);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => {
        this.db = request.result;
        resolve();
      };

      request.onupgradeneeded = (event) => {
        const db = (event.target as IDBOpenDBRequest).result;

        // Create object store for user data
        if (!db.objectStoreNames.contains(this.storeName)) {
          const store = db.createObjectStore(this.storeName, { keyPath: 'id' });
          store.createIndex('userId', 'userId', { unique: false });
          store.createIndex('createdAt', 'createdAt', { unique: false });
        }
      };
    });
  }

  async saveUserData(userId: string, data: UserDataStorage): Promise<void> {
    if (!this.db) await this.init();

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([this.storeName], 'readwrite');
      const store = transaction.objectStore(this.storeName);

      const userData = {
        id: `user_${userId}`,
        userId,
        ...data,
        updatedAt: new Date().toISOString(),
        createdAt: data.user?.createdAt || new Date().toISOString(),
      };

      // Encrypt sensitive data before storage
      const secureData = this.encryptSensitiveData(userData);

      const request = store.put(secureData);
      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve();
    });
  }

  async loadUserData(userId: string): Promise<UserDataStorage | null> {
    if (!this.db) await this.init();

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([this.storeName], 'readonly');
      const store = transaction.objectStore(this.storeName);

      const request = store.get(`user_${userId}`);
      request.onerror = () => reject(request.error);
      request.onsuccess = () => {
        if (request.result) {
          // Decrypt sensitive data after loading
          const decryptedData = this.decryptSensitiveData(request.result);
          resolve(decryptedData);
        } else {
          resolve(null);
        }
      };
    });
  }

  async clearUserData(userId: string): Promise<void> {
    if (!this.db) await this.init();

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([this.storeName], 'readwrite');
      const store = transaction.objectStore(this.storeName);

      const request = store.delete(`user_${userId}`);
      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve();
    });
  }

  async getAllUsers(): Promise<string[]> {
    if (!this.db) await this.init();

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([this.storeName], 'readonly');
      const store = transaction.objectStore(this.storeName);

      const request = store.getAllKeys();
      request.onerror = () => reject(request.error);
      request.onsuccess = () => {
        const userIds = request.result
          .filter(key => typeof key === 'string' && key.startsWith('user_'))
          .map(key => (key as string).replace('user_', ''));
        resolve(userIds);
      };
    });
  }

  // Encryption/Decryption for sensitive data
  private encryptSensitiveData(data: any): any {
    // Simple encryption - em produção usar crypto real
    const sensitiveFields = ['email', 'userApiKey', 'githubPat', 'jiraApiToken'];
    const encrypted = { ...data };

    const encryptField = (obj: any, field: string) => {
      if (obj && obj[field]) {
        // Base64 encoding como exemplo - usar crypto real em produção
        obj[field] = btoa(obj[field]);
      }
    };

    // Encrypt sensitive fields in settings
    if (encrypted.settings) {
      sensitiveFields.forEach(field => encryptField(encrypted.settings, field));
    }

    // Encrypt email in profile
    if (encrypted.profile) {
      encryptField(encrypted.profile, 'email');
    }

    return encrypted;
  }

  private decryptSensitiveData(data: any): any {
    // Simple decryption - em produção usar crypto real
    const sensitiveFields = ['email', 'userApiKey', 'githubPat', 'jiraApiToken'];
    const decrypted = { ...data };

    const decryptField = (obj: any, field: string) => {
      if (obj && obj[field]) {
        try {
          // Base64 decoding como exemplo - usar crypto real em produção
          obj[field] = atob(obj[field]);
        } catch (error) {
          console.warn(`Failed to decrypt field ${field}:`, error);
        }
      }
    };

    // Decrypt sensitive fields in settings
    if (decrypted.settings) {
      sensitiveFields.forEach(field => decryptField(decrypted.settings, field));
    }

    // Decrypt email in profile
    if (decrypted.profile) {
      decryptField(decrypted.profile, 'email');
    }

    return decrypted;
  }
}

// Singleton instance
export const secureUserStorage = new SecureUserStorage();

// Export types for use in contexts
export type { UserDataStorage };
