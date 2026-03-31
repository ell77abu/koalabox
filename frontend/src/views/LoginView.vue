<template>
  <div class="page-shell">
    <section class="card">
      <h1>登录</h1>
      <p class="hint">登录后进入任务管理界面</p>

      <form class="form" @submit.prevent="onSubmit">
        <label>
          用户名
          <input v-model.trim="form.username" type="text" placeholder="请输入用户名" />
        </label>

        <label>
          密码
          <div class="password-field">
            <input
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="请输入密码"
            />
            <button
              type="button"
              class="password-toggle"
              :aria-label="showPassword ? '隐藏密码' : '显示密码'"
              @click="showPassword = !showPassword"
            >
              <svg
                v-if="!showPassword"
                viewBox="0 0 24 24"
                aria-hidden="true"
                class="password-toggle-icon"
              >
                <path
                  d="M3 3L21 21M10.58 10.58A2 2 0 0013.41 13.41M9.88 5.09A10.94 10.94 0 0112 5c5 0 9.27 3.11 11 7-0.56 1.27-1.37 2.42-2.37 3.38M6.61 6.61C4.62 7.84 3.05 9.73 2 12c1.73 3.89 6 7 10 7a10.7 10.7 0 004.25-.85"
                />
              </svg>
              <svg
                v-else
                viewBox="0 0 24 24"
                aria-hidden="true"
                class="password-toggle-icon"
              >
                <path
                  d="M2 12S5.64 5 12 5s10 7 10 7-3.64 7-10 7S2 12 2 12z"
                />
                <circle cx="12" cy="12" r="3" />
              </svg>
            </button>
          </div>
        </label>

        <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

        <button class="btn" type="submit" :disabled="submitting">
          {{ submitting ? '登录中...' : '登录' }}
        </button>
      </form>

      <p class="foot">
        还没有账号？
        <RouterLink to="/register">去注册</RouterLink>
      </p>
    </section>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import { loginApi } from '../services/auth';
import { authStore } from '../stores/auth';

const router = useRouter();
const route = useRoute();

const form = reactive({
  username: '',
  password: '',
});

const submitting = ref(false);
const errorMessage = ref('');
const showPassword = ref(false);

const onSubmit = async () => {
  errorMessage.value = '';
  if (!form.username || !form.password) {
    errorMessage.value = '请输入用户名和密码';
    return;
  }

  submitting.value = true;
  try {
    const response = await loginApi({
      username: form.username,
      password: form.password,
    });

    const payload = response.data?.data;
    if (!payload?.token || !payload?.username) {
      throw new Error('登录响应格式不正确');
    }

    authStore.setAuth(payload.token, payload.username);
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/tasks';
    await router.replace(redirect);
  } catch (error) {
    if (error.response?.status === 401) {
      errorMessage.value = '用户名或密码错误';
    } else {
      errorMessage.value = error.response?.data?.message || '登录失败，请稍后重试';
    }
  } finally {
    submitting.value = false;
  }
};
</script>
