

// src/lib/mermaid-loader.ts
let loadPromise: Promise<any> | null = null;

export function preloadMermaidCDN() {
  if (document.querySelector('link[data-mermaid-preload]')) return;
  const l = document.createElement('link');
  l.rel = 'preload';
  l.as = 'script';
  l.href = 'https://cdn.jsdelivr.net/npm/mermaid@10.9.1/dist/mermaid.min.js';
  l.setAttribute('data-mermaid-preload', '1');
  document.head.appendChild(l);
}

export function loadMermaidFromCDN(): Promise<any> {
  const mmd = (window as any).mermaid;

  if (mmd) return Promise.resolve(mmd);
  if (loadPromise) return loadPromise;

  loadPromise = new Promise((resolve, reject) => {
    const s = document.createElement('script');
    s.src = 'https://cdn.jsdelivr.net/npm/mermaid@10.9.1/dist/mermaid.min.js';
    s.async = true;
    s.defer = true;
    s.onload = () => {
      try {
        mmd?.initialize({
          startOnLoad: false,
          theme: 'dark',
          securityLevel: 'loose',
          fontFamily: 'Inter, system-ui, sans-serif'
        });
        resolve(mmd);
      } catch (e) { reject(e); }
    };
    s.onerror = () => reject(new Error('CDN mermaid load failed'));
    document.head.appendChild(s);
  });

  return loadPromise;
}
