const CACHE_NAME = 'agent-payment-v1';
const URLS_TO_CACHE = [
  '/',
  '/index.html',
  '/agent-payment-logo.png'
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => cache.addAll(URLS_TO_CACHE))
      .then(() => self.skipWaiting())
  );
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheName !== CACHE_NAME) {
            return caches.delete(cacheName);
          }
        })
      );
    }).then(() => self.clients.claim())
  );
});

self.addEventListener('fetch', (event) => {
  // Only cache GET requests
  if (event.request.method !== 'GET') {
    return;
  }

  event.respondWith(
    caches.match(event.request)
      .then((response) => {
        // Return cached version or fetch new
        return response || fetch(event.request)
          .then((fetchResponse) => {
            // Cache new responses
            return caches.open(CACHE_NAME).then((cache) => {
              cache.put(event.request, fetchResponse.clone());
              return fetchResponse;
            });
          });
      })
      .catch(() => {
        // Fallback for offline
        return caches.match('/index.html');
      })
  );
});
