import { openDB, DBSchema } from 'idb';
import { Project } from '../types';

const DB_NAME = 'gemx-db';
const KEYVAL_STORE_NAME = 'keyval';
const PROJECTS_STORE_NAME = 'projects';
const DB_VERSION = 2;

interface GemXDB extends DBSchema {
  [KEYVAL_STORE_NAME]: {
    key: string;
    value: any;
  };
  [PROJECTS_STORE_NAME]: {
    key: string;
    value: Project;
    indexes: { 'by-name': string };
  };
}

const dbPromise = openDB<GemXDB>(DB_NAME, DB_VERSION, {
  upgrade(db, oldVersion) {
    if (oldVersion < 1) {
      db.createObjectStore(KEYVAL_STORE_NAME);
    }
    if (oldVersion < 2) {
      const projectStore = db.createObjectStore(PROJECTS_STORE_NAME, { keyPath: 'id' });
      projectStore.createIndex('by-name', 'name');
    }
  },
});

// Generic Key-Value Store Functions
export async function get<T>(key: string): Promise<T | undefined> {
  return (await dbPromise).get(KEYVAL_STORE_NAME, key);
}

export async function set(key: string, value: any): Promise<void> {
  await (await dbPromise).put(KEYVAL_STORE_NAME, value, key);
}

// Project-Specific Store Functions
export async function getProject(id: string): Promise<Project | undefined> {
    return (await dbPromise).get(PROJECTS_STORE_NAME, id);
}

export async function setProject(project: Project): Promise<string> {
    return (await dbPromise).put(PROJECTS_STORE_NAME, project);
}

export async function deleteProject(id: string): Promise<void> {
    await (await dbPromise).delete(PROJECTS_STORE_NAME, id);
}

export async function getAllProjects(): Promise<Project[]> {
    return (await dbPromise).getAll(PROJECTS_STORE_NAME);
}

// Clear All Data Function
export async function clear(): Promise<void> {
  await (await dbPromise).clear(KEYVAL_STORE_NAME);
  await (await dbPromise).clear(PROJECTS_STORE_NAME);
}