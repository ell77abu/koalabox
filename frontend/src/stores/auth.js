import { computed, ref } from 'vue';

const TOKEN_KEY = 'lab1_token';
const USERNAME_KEY = 'lab1_username';
const LEVEL_KEY = 'lab1_level';

const token = ref(localStorage.getItem(TOKEN_KEY) || '');
const username = ref(localStorage.getItem(USERNAME_KEY) || '');
const level = ref(Number(localStorage.getItem(LEVEL_KEY) || '0'));

const isLoggedIn = computed(() => Boolean(token.value));

const setAuth = (newToken, newUsername, newLevel = 0) => {
  token.value = newToken;
  username.value = newUsername;
  level.value = Number(newLevel || 0);
  localStorage.setItem(TOKEN_KEY, newToken);
  localStorage.setItem(USERNAME_KEY, newUsername);
  localStorage.setItem(LEVEL_KEY, String(level.value));
};

const clearAuth = () => {
  token.value = '';
  username.value = '';
  level.value = 0;
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USERNAME_KEY);
  localStorage.removeItem(LEVEL_KEY);
};

export const authStore = {
  token,
  username,
  level,
  isLoggedIn,
  setAuth,
  clearAuth,
};
