import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';
import ChatView from '../views/ChatView.vue';
import ContactStage from '../components/ContactStage.vue';
import CallStage from '../components/CallStage.vue';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/chats',
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
