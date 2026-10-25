import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';
import ChatView from '../views/ChatView.vue';
import ContactStage from '../components/ContactStage.vue';
import CallStage from '../components/CallStage.vue';
import AuthView from '../views/AuthView.vue';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/chats',
  },
  {
    path: '/login',
    name: 'Login',
    component: AuthView,
    meta: { public: true },
  },
  {
    path: '/register',
    name: 'Register',
    component: AuthView,
    meta: { public: true },
  },
  {
    path: '/chats',
    name: 'Chats',
    component: ChatView,
  },
  {
    path: '/chats/:id',
    name: 'ChatConversation',
    component: ChatView,
  },
  {
    path: '/contacts',
    name: 'Contacts',
    component: ContactStage,
  },
  {
    path: '/calls',
    name: 'Calls',
    component: CallStage,
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/chats',
  },
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to, _from, next) => {
  const token = typeof localStorage !== 'undefined' ? localStorage.getItem('mim_token') : null;
  const uid = typeof localStorage !== 'undefined' ? localStorage.getItem('mim_active_uid') : null;
  const isExplicitlyLoggedOut = typeof localStorage !== 'undefined' ? localStorage.getItem('mim_logged_out') === 'true' : false;
  const isLoggedIn = !isExplicitlyLoggedOut && !!token && Number(uid) > 0;

  const isPublic = to.meta?.public === true || to.path === '/login' || to.path === '/register';

  if (!isLoggedIn && !isPublic) {
    next({ path: '/login', query: to.fullPath !== '/chats' && to.fullPath !== '/' ? { redirect: to.fullPath } : undefined });
  } else if (isLoggedIn && isPublic) {
    next('/chats');
  } else {
    next();
  }
});
