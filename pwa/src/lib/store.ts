/**
 * IndexedDB storage for encrypted API keys
 */

import { openDB, DBSchema, IDBPDatabase } from 'idb';
import { generateKey, importKey, exportKey, encryptJSON, decryptJSON } from './crypto';

const DB_NAME = 'agentpay-db';
const DB_VERSION = 1;
const STORE_NAME = 'secrets';
const KEY_SLOT = 'credentials';
const SYMKEY_STORAGE = 'agentpay_symkey';

interface SecretBundle {
  apiKey: string;
  budgetKey: string;
  auth?: string;
}

interface AgentPayDB extends DBSchema {
  secrets: {
    key: string;
    value: {
      iv: number[];
      ciphertext: number[];
    };
  };
}

async function getDB(): Promise<IDBPDatabase<AgentPayDB>> {
  return openDB<AgentPayDB>(DB_NAME, DB_VERSION, {
    upgrade(db) {
      if (!db.objectStoreNames.contains(STORE_NAME)) {
        db.createObjectStore(STORE_NAME);
      }
    },
  });
}

/**
 * Get or create symmetric encryption key (stored in localStorage)
 */
async function getOrCreateSymKey(): Promise<CryptoKey> {
  const hexKey = localStorage.getItem(SYMKEY_STORAGE);

  if (hexKey) {
    // Import existing key
    const bytes = new Uint8Array(
      hexKey.match(/.{1,2}/g)!.map(byte => parseInt(byte, 16))
    );
    return importKey(bytes.buffer);
  }

  // Generate new key
  const key = await generateKey();
  const rawKey = await exportKey(key);
  const hexOut = Array.from(new Uint8Array(rawKey))
    .map(b => b.toString(16).padStart(2, '0'))
    .join('');

  localStorage.setItem(SYMKEY_STORAGE, hexOut);
  return key;
}

/**
 * Save encrypted secrets to IndexedDB
 */
export async function saveSecrets(bundle: SecretBundle): Promise<void> {
  const key = await getOrCreateSymKey();
  const { iv, ciphertext } = await encryptJSON(key, bundle);

  const db = await getDB();
  await db.put(STORE_NAME, {
    iv: Array.from(iv),
    ciphertext: Array.from(new Uint8Array(ciphertext))
  }, KEY_SLOT);
}

/**
 * Load and decrypt secrets from IndexedDB
 */
export async function loadSecrets(): Promise<SecretBundle | null> {
  const db = await getDB();
  const record = await db.get(STORE_NAME, KEY_SLOT);

  if (!record) {
    return null;
  }

  const key = await getOrCreateSymKey();
  const iv = new Uint8Array(record.iv);
  const ciphertext = new Uint8Array(record.ciphertext).buffer;

  return decryptJSON<SecretBundle>(key, iv, ciphertext);
}

/**
 * Clear all stored secrets
 */
export async function clearSecrets(): Promise<void> {
  const db = await getDB();
  await db.delete(STORE_NAME, KEY_SLOT);
  localStorage.removeItem(SYMKEY_STORAGE);
}
