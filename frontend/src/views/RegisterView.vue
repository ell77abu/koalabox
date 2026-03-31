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
          <div class="password-field">
            <input
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="至少6位，包含字母和数字"
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

        <label>
          确认密码
          <div class="password-field">
            <input
              v-model="form.confirmPassword"
              :type="showConfirmPassword ? 'text' : 'password'"
              placeholder="请再次输入密码"
            />
            <button
              type="button"
              class="password-toggle"
              :aria-label="showConfirmPassword ? '隐藏密码' : '显示密码'"
              @click="showConfirmPassword = !showConfirmPassword"
            >
              <svg
                v-if="!showConfirmPassword"
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
const showPassword = ref(false);
const showConfirmPassword = ref(false);

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
