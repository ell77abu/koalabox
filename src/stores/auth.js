import { computed, ref } from 'vue';

const TOKEN_KEY = 'lab1_token';
const USERNAME_KEY = 'lab1_username';

const token = ref(localStorage.getItem(TOKEN_KEY) || '');
const username = ref(localStorage.getItem(USERNAME_KEY) || '');

const isLoggedIn = computed(() => Boolean(token.value));

const setAuth = (newToken, newUsername) => {
  token.value = newToken;
  username.value = newUsername;
  localStorage.setItem(TOKEN_KEY, newToken);
  localStorage.setItem(USERNAME_KEY, newUsername);
};

const clearAuth = () => {
  token.value = '';
  username.value = '';
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USERNAME_KEY);
};

export const authStore = {
  token,
  username,
  isLoggedIn,
  setAuth,
  clearAuth,
};
