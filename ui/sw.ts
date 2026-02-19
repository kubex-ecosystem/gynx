/// <reference lib="webworker" />

const CACHE_NAME = 'gemx-analyzer-cache-v1';

// Pre-cache essential assets for the app shell to work offline.
const APP_SHELL_URLS = [
  '/',
  '/index.html',
];

const swOriginIgnore = [
  '//ai.studio',
  'scf.usercontent.goog',
  'generativelanguage.googleapis.com'
]

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      console.log('Opened cache and caching app shell');
      return cache.addAll(APP_SHELL_URLS);
    })
  );
});

self.addEventListener('activate', (event) => {
  const cacheWhitelist = [CACHE_NAME];
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheWhitelist.indexOf(cacheName) === -1) {
            console.log('Deleting old cache:', cacheName);
            return caches.delete(cacheName);
          }
        })
      );
    })
  );
});

self.addEventListener('fetch', (event) => {
  // Use a "network falling back to cache" strategy for navigation requests
  if (event.request.mode === 'navigate') {
    event.respondWith(
      fetch(event.request).catch(() => caches.match('/index.html') as Promise<Response>)
    );
    return;
  }

  // Use a "stale-while-revalidate" strategy for other assets (CSS, JS, images, fonts, etc.)
  // Skip non-GET requests and Gemini API calls
  if (event.request.method !== 'GET' || swOriginIgnore.filter((v, i) => event.request.url.includes(v)).length > 0) {
    return;
  }
  
  event.respondWith(
    caches.open(CACHE_NAME).then(async (cache) => {
      const cachedResponse = await cache.match(event.request);
      
      const fetchPromise = fetch(event.request).then((networkResponse) => {
        if (networkResponse.ok) {
          cache.put(event.request, networkResponse.clone());
        }
        return networkResponse;
      }).catch(err => {
        console.warn(`Fetch failed for ${event.request.url}; returning cached response if available.`, err);
        // If fetch fails and we have a cached response, the cachedResponse will be returned.
        // If not, the promise will reject, leading to a network error page.
        if (cachedResponse) {
          return cachedResponse;
        }
        throw err;
      });

      return cachedResponse || fetchPromise;
    })
  );
});