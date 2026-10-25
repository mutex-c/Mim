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
    window.location.hostname === 'tauri.localhost' ||
    window.location.protocol === 'tauri:'
  );
};

export const getServerHost = (): string => {
  if (typeof window !== 'undefined') {
    const saved = localStorage.getItem('mim_server_host');
    if (saved) return saved.trim();
  }
  // In Tauri Desktop, default to backend 127.0.0.1:8080
  if (isTauriEnv()) {
    return '127.0.0.1:8080';
  }
  // In Browser
  return window.location.host;
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
  // If in web browser dev or behind reverse proxy, relative path works
  if (!isTauriEnv() && window.location.hostname !== 'tauri.localhost' && window.location.protocol !== 'tauri:') {
    return '';
  }
  const host = getServerHost();
  if (host.startsWith('127.0.0.1') || host.startsWith('localhost') || host.includes(':8080')) {
    return `http://${host}`;
  }
  const proto = window.location.protocol === 'https:' ? 'https:' : 'http:';
  return `${proto}//${host}`;
};

export const getWsUrl = (userId: number | string, token: string, deviceId: string): string => {
  let host = getServerHost();
  if (!isTauriEnv()) {
    host = (window.location.hostname === '127.0.0.1' || window.location.hostname === 'localhost')
      ? `${window.location.hostname}:8080`
      : window.location.host;
  }

  const proto = (host.startsWith('127.0.0.1') || host.startsWith('localhost') || host.includes(':8080'))
    ? 'ws:'
    : (window.location.protocol === 'https:' ? 'wss:' : 'ws:');
  return `${proto}//${host}/ws?user_id=${userId}&token=${encodeURIComponent(token)}&platform=desktop&device_id=${encodeURIComponent(deviceId)}`;
};

/**
 * Resolve any relative API/static url (e.g. '/api/...' or '/uploads/...') to full URL in Tauri
 */
export const resolveApiUrl = (path: string): string => {
  if (!path) return '';
  if (path.startsWith('http://tauri.localhost') || path.startsWith('https://tauri.localhost')) {
    path = path.replace(/^https?:\/\/tauri\.localhost/, '');
  }
  if (
    path.startsWith('http://') ||
    path.startsWith('https://') ||
    path.startsWith('blob:') ||
    path.startsWith('data:')
  ) {
    return path;
  }
  const base = getApiBaseUrl();
  if (!base) return path;
  return `${base}${path.startsWith('/') ? '' : '/'}${path}`;
};
