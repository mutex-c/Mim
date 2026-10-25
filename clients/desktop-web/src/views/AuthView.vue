<template>
  <div class="w-full h-full flex items-center justify-center p-4 md:p-8 bg-mim-light-stage dark:bg-mim-dark-stage overflow-auto">
    <AuthModal
      :is-fullscreen="true"
      :initial-tab="currentTab"
      @tab-change="onTabChange"
      @success="onAuthSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import AuthModal from '../components/AuthModal.vue';

const route = useRoute();
const router = useRouter();

const currentTab = computed<'login' | 'register'>(() => {
  return route.path === '/register' || route.query.tab === 'register' ? 'register' : 'login';
});

const onTabChange = (tab: 'login' | 'register') => {
  if (tab === 'register' && route.path !== '/register') {
    router.replace('/register');
  } else if (tab === 'login' && route.path !== '/login') {
    router.replace('/login');
  }
};

const onAuthSuccess = () => {
  const redirect = (route.query.redirect as string) || '/chats';
  router.push(redirect);
};
</script>
