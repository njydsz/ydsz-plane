/**
 * sw.js — Ydsz Plane Service Worker
 *
 * 策略：
 *  - 预缓存：manifest / icons / 内联关键资源；
 *  - 导航请求：network-first，失败回退 index.html（SPA 深链离线可达）；
 *  - 静态资源（assets/、icons/）：cache-first，后台更新；
 *  - API 请求：network-only（不缓存，保证数据新鲜）。
 *
 * 版本号变更时自动清理旧缓存。
 */
const CACHE_VERSION = 'ydsz-plane-v1';
const PRECACHE = ['/', '/manifest.webmanifest', '/icons/icon-192.png', '/icons/icon-512.png'];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_VERSION).then((cache) => cache.addAll(PRECACHE)).then(() => self.skipWaiting()),
  );
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(keys.filter((k) => k !== CACHE_VERSION).map((k) => caches.delete(k))),
    ).then(() => self.clients.claim()),
  );
});

self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);

  // 只处理 GET
  if (request.method !== 'GET') return;

  // API：不走缓存
  if (url.pathname.startsWith('/api/')) return;

  // 导航请求：network-first → 回退 index.html
  if (request.mode === 'navigate') {
    event.respondWith(
      fetch(request)
        .then((resp) => {
          const copy = resp.clone();
          caches.open(CACHE_VERSION).then((cache) => cache.put('/', copy));
          return resp;
        })
        .catch(() => caches.match('/').then((r) => r || caches.match('/index.html'))),
    );
    return;
  }

  // 静态资源：cache-first
  if (url.pathname.startsWith('/assets/') || url.pathname.startsWith('/icons/')) {
    event.respondWith(
      caches.match(request).then(
        (cached) =>
          cached ||
          fetch(request).then((resp) => {
            if (resp.ok) {
              const copy = resp.clone();
              caches.open(CACHE_VERSION).then((cache) => cache.put(request, copy));
            }
            return resp;
          }),
      ),
    );
  }
});

// 接收页面发来的 SKIP_WAITING
self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting();
  }
});
