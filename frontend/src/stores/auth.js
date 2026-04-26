import { computed, ref } from 'vue';

const username = ref('');
const level = ref(0);

const isLoggedIn = computed(() => Boolean(username.value));

const setAuth = (newUsername, newLevel = 0) => {
  username.value = newUsername;
  level.value = Number(newLevel || 0);
};

const clearAuth = () => {
  username.value = '';
  level.value = 0;
};

export const authStore = {
  username,
  level,
  isLoggedIn,
  setAuth,
  clearAuth,
};
