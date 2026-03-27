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
          <input v-model="form.password" type="password" placeholder="请输入密码" />
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
