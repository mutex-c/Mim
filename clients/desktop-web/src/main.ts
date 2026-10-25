import { createApp } from 'vue';
import { createPinia } from 'pinia';
import { router } from './router';
import App from './App.vue';
import './style.css';
import { resolveApiUrl } from './utils/api';

// Intercept window.fetch to automatically resolve relative /api/... and /uploads/... in Tauri desktop release mode
const originalFetch = window.fetch;
window.fetch = (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
  if (typeof input === 'string') {
    return originalFetch(resolveApiUrl(input), init);
  }
  if (input instanceof URL) {
    return originalFetch(new URL(resolveApiUrl(input.toString())), init);
  }
  if (input instanceof Request) {
    const resolved = resolveApiUrl(input.url);
    if (resolved !== input.url) {
      const newReq = new Request(resolved, input);
      return originalFetch(newReq, init);
    }
  }
  return originalFetch(input, init);
};

const app = createApp(App);
app.use(createPinia());
app.use(router);
app.mount('#app');
