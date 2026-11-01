/**
 * Global API and Server Configuration for Mim
 * Automatically handles:
 * - Web Browser Dev: relative URLs (handled by Vite proxy)
 * - Web Browser Prod: relative URLs (handled by Nginx/Reverse Proxy)
 * - Tauri Desktop (Dev & Release): http://127.0.0.1:8080 or custom configured host
 */

export const isTauriEnv = (): boolean => {
  if (typeof window === 'undefined') return false;
  return (
    '__TAURI_INTERNALS__' in window ||
    '__TAURI__' in window ||
    '__TAURI_METADATA__' in window ||
    window.location.hostname === 'tauri.localhost' ||
    window.location.protocol === 'tauri:' ||
    ((window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') && !window.location.port)
  );
};

export const getServerHost = (): string => {
  if (typeof window !== 'undefined') {
    const saved = localStorage.getItem('mim_server_host');
    if (saved) return saved.trim();
  }
  // In Tauri Desktop, default to backend 127.0.0.1:8080
  if (isTauriEnv() || (typeof window !== 'undefined' && (window.location.hostname === 'tauri.localhost' || window.location.protocol === 'tauri:'))) {
    return '127.0.0.1:8080';
  }
  // In Browser
  return typeof window !== 'undefined' ? window.location.host : '127.0.0.1:8080';
};

export const setServerHost = (host: string): void => {
  if (typeof window === 'undefined') return;
  const clean = host.replace(/^https?:\/\//, '').replace(/\/$/, '').trim();
  if (!clean || clean === '127.0.0.1:8080' || clean === 'localhost:8080') {
    localStorage.removeItem('mim_server_host');
  } else {
    localStorage.setItem('mim_server_host', clean);
  }
};

export const getApiBaseUrl = (): string => {
  if (typeof window === 'undefined') return '';

  // In Tauri Desktop, default to backend 127.0.0.1:8080 or custom host
  if (isTauriEnv()) {
    const host = getServerHost();
    return `http://${host}`;
  }

  // In Browser (Direct backend serving or reverse proxy): relative URL always works perfectly!
  return '';
};

export const getWsUrl = (userId: number | string, token: string, deviceId: string): string => {
  let host = getServerHost();
  if (isTauriEnv() || (typeof window !== 'undefined' && window.location.hostname === 'tauri.localhost')) {
    host = '127.0.0.1:8080';
  } else if (typeof window !== 'undefined' && window.location.host) {
    host = window.location.host;
  }

  const proto = (typeof window !== 'undefined' && window.location.protocol === 'https:') ? 'wss:' : 'ws:';
  const platform = isTauriEnv() ? 'desktop' : 'web';
  return `${proto}//${host}/ws?token=${encodeURIComponent(token)}&platform=${platform}&device_id=${encodeURIComponent(deviceId)}`;
};

/**
 * Resolve any relative API/static url (e.g. '/api/...' or '/uploads/...') to full URL in Tauri
 */
export const resolveApiUrl = (path: string): string => {
  if (!path) return '';

  // 1. Strip scheme and host if it was expanded against tauri.localhost or localhost without port
  if (path.startsWith('http://tauri.localhost') || path.startsWith('https://tauri.localhost')) {
    path = path.replace(/^https?:\/\/tauri\.localhost(:[0-9]+)?/, '');
  }
  if (path.startsWith('http://localhost/') || path.startsWith('http://127.0.0.1/')) {
    path = path.replace(/^http:\/\/(localhost|127\.0\.0\.1)/, '');
  }
  if (path.startsWith('tauri://localhost')) {
    path = path.replace(/^tauri:\/\/localhost/, '');
  }

  // 2. If it's already an external / absolute URL (including 127.0.0.1:8080, blob:, data:), preserve it
  if (
    path.startsWith('http://') ||
    path.startsWith('https://') ||
    path.startsWith('blob:') ||
    path.startsWith('data:')
  ) {
    return path;
  }

  // 3. Resolve relative URL against API base
  const base = getApiBaseUrl();
  if (!base) return path;
  return `${base}${path.startsWith('/') ? '' : '/'}${path}`;
};
