<template>
  <div class="page-shell">
    <section class="card">
      <h1>注册</h1>
      <p class="hint">创建账号后可登录系统</p>

      <form class="form" @submit.prevent="onSubmit">
        <label>
          用户名
          <input
            v-model.trim="form.username"
            type="text"
            placeholder="4~20位，仅字母数字下划线"
          />
        </label>

        <label>
          密码
          <input
            v-model="form.password"
            type="password"
            placeholder="至少6位，包含字母和数字"
          />
        </label>

        <label>
          确认密码
          <input v-model="form.confirmPassword" type="password" placeholder="请再次输入密码" />
        </label>

        <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
        <p v-if="successMessage" class="success">{{ successMessage }}</p>

        <button class="btn" type="submit" :disabled="submitting">
          {{ submitting ? '注册中...' : '注册' }}
        </button>
      </form>

      <p class="foot">
        已有账号？
        <RouterLink to="/login">去登录</RouterLink>
      </p>
    </section>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue';
import { RouterLink, useRouter } from 'vue-router';
import { registerApi } from '../services/auth';
import { validatePassword, validateUsername } from '../utils/validators';

const router = useRouter();

const form = reactive({
  username: '',
  password: '',
  confirmPassword: '',
});

const submitting = ref(false);
const errorMessage = ref('');
const successMessage = ref('');

const onSubmit = async () => {
  errorMessage.value = '';
  successMessage.value = '';

  const usernameError = validateUsername(form.username);
  if (usernameError) {
    errorMessage.value = usernameError;
    return;
  }

  const passwordError = validatePassword(form.password);
  if (passwordError) {
    errorMessage.value = passwordError;
    return;
  }

  if (form.password !== form.confirmPassword) {
    errorMessage.value = '两次输入密码不一致';
    return;
  }

  submitting.value = true;
  try {
    await registerApi({
      username: form.username,
      password: form.password,
    });

    successMessage.value = '注册成功，正在跳转到登录页...';
    setTimeout(() => {
      router.replace('/login');
    }, 800);
  } catch (error) {
    if (error.response?.status === 409) {
      errorMessage.value = '用户名已存在，请更换用户名';
    } else {
      errorMessage.value = error.response?.data?.message || '注册失败，请稍后重试';
    }
  } finally {
    submitting.value = false;
  }
};
</script>
