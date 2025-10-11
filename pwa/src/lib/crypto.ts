/**
 * WebCrypto utilities for encrypting API keys at rest
 */

export async function generateKey(): Promise<CryptoKey> {
  return crypto.subtle.generateKey(
    { name: 'AES-GCM', length: 256 },
    true,
    ['encrypt', 'decrypt']
  );
}

export async function importKey(rawKey: ArrayBuffer): Promise<CryptoKey> {
  return crypto.subtle.importKey(
    'raw',
    rawKey,
    'AES-GCM',
    true,
    ['encrypt', 'decrypt']
  );
}

export async function exportKey(key: CryptoKey): Promise<ArrayBuffer> {
  return crypto.subtle.exportKey('raw', key);
}

export async function encryptJSON(
  key: CryptoKey,
  data: unknown
): Promise<{ iv: Uint8Array; ciphertext: ArrayBuffer }> {
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const plaintext = new TextEncoder().encode(JSON.stringify(data));
  const ciphertext = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv },
    key,
    plaintext
  );
  return { iv, ciphertext };
}

export async function decryptJSON<T>(
  key: CryptoKey,
  iv: Uint8Array,
  ciphertext: ArrayBuffer
): Promise<T> {
  const plaintext = await crypto.subtle.decrypt(
    { name: 'AES-GCM', iv },
    key,
    ciphertext
  );
  const json = new TextDecoder().decode(plaintext);
  return JSON.parse(json);
}
