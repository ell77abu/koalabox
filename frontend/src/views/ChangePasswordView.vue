<template>
  <section class="card password-page">
    <div class="page-head">
      <div>
        <h1>修改密码</h1>
        <p class="hint">请先输入旧密码，再设置新的登录密码。</p>
      </div>
      <RouterLink to="/profile" class="back-link">返回个人中心</RouterLink>
    </div>

    <form class="form" @submit.prevent="onSubmit">
      <label>
        旧密码
        <div class="password-field">
          <input
            v-model="form.oldPassword"
            :type="showOldPassword ? 'text' : 'password'"
            placeholder="请输入当前密码"
          />
          <button
            type="button"
            class="password-toggle"
            :aria-label="showOldPassword ? '隐藏密码' : '显示密码'"
            @click="showOldPassword = !showOldPassword"
          >
            <svg
              v-if="!showOldPassword"
              viewBox="0 0 24 24"
              aria-hidden="true"
              class="password-toggle-icon"
            >
              <path d="M3 3L21 21M10.58 10.58A2 2 0 0013.41 13.41M9.88 5.09A10.94 10.94 0 0112 5c5 0 9.27 3.11 11 7-0.56 1.27-1.37 2.42-2.37 3.38M6.61 6.61C4.62 7.84 3.05 9.73 2 12c1.73 3.89 6 7 10 7a10.7 10.7 0 004.25-.85" />
            </svg>
            <svg
              v-else
              viewBox="0 0 24 24"
              aria-hidden="true"
              class="password-toggle-icon"
            >
              <path d="M2 12S5.64 5 12 5s10 7 10 7-3.64 7-10 7S2 12 2 12z" />
              <circle cx="12" cy="12" r="3" />
            </svg>
          </button>
        </div>
      </label>

      <label>
        新密码
        <div class="password-field">
          <input
            v-model="form.newPassword"
            :type="showNewPassword ? 'text' : 'password'"
            placeholder="至少6位，包含字母和数字"
          />
          <button
            type="button"
            class="password-toggle"
            :aria-label="showNewPassword ? '隐藏密码' : '显示密码'"
            @click="showNewPassword = !showNewPassword"
          >
            <svg
              v-if="!showNewPassword"
              viewBox="0 0 24 24"
              aria-hidden="true"
              class="password-toggle-icon"
            >
              <path d="M3 3L21 21M10.58 10.58A2 2 0 0013.41 13.41M9.88 5.09A10.94 10.94 0 0112 5c5 0 9.27 3.11 11 7-0.56 1.27-1.37 2.42-2.37 3.38M6.61 6.61C4.62 7.84 3.05 9.73 2 12c1.73 3.89 6 7 10 7a10.7 10.7 0 004.25-.85" />
            </svg>
            <svg
              v-else
              viewBox="0 0 24 24"
              aria-hidden="true"
              class="password-toggle-icon"
            >
              <path d="M2 12S5.64 5 12 5s10 7 10 7-3.64 7-10 7S2 12 2 12z" />
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
            placeholder="请再次输入新密码"
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
              <path d="M3 3L21 21M10.58 10.58A2 2 0 0013.41 13.41M9.88 5.09A10.94 10.94 0 0112 5c5 0 9.27 3.11 11 7-0.56 1.27-1.37 2.42-2.37 3.38M6.61 6.61C4.62 7.84 3.05 9.73 2 12c1.73 3.89 6 7 10 7a10.7 10.7 0 004.25-.85" />
            </svg>
            <svg
              v-else
              viewBox="0 0 24 24"
              aria-hidden="true"
              class="password-toggle-icon"
            >
              <path d="M2 12S5.64 5 12 5s10 7 10 7-3.64 7-10 7S2 12 2 12z" />
              <circle cx="12" cy="12" r="3" />
            </svg>
          </button>
        </div>
      </label>

      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
      <p v-if="successMessage" class="success">{{ successMessage }}</p>

      <div class="page-actions">
        <RouterLink to="/profile" class="btn btn-outline cancel-link">取消</RouterLink>
        <button class="btn" type="submit" :disabled="submitting">
          {{ submitting ? '提交中...' : '确认修改' }}
        </button>
      </div>
    </form>
  </section>
</template>

<script setup>
import { reactive, ref } from 'vue';
import { RouterLink, useRouter } from 'vue-router';
import { changePasswordApi } from '../services/auth';
import { validatePassword } from '../utils/validators';

const router = useRouter();

const form = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
});

const submitting = ref(false);
const errorMessage = ref('');
const successMessage = ref('');
const showOldPassword = ref(false);
const showNewPassword = ref(false);
const showConfirmPassword = ref(false);

const onSubmit = async () => {
  errorMessage.value = '';
  successMessage.value = '';

  if (!form.oldPassword || !form.newPassword || !form.confirmPassword) {
    errorMessage.value = '请完整填写旧密码、新密码和确认密码';
    return;
  }

  const passwordError = validatePassword(form.newPassword);
  if (passwordError) {
    errorMessage.value = passwordError;
    return;
  }

  if (form.oldPassword === form.newPassword) {
    errorMessage.value = '新密码不能与旧密码相同';
    return;
  }

  if (form.newPassword !== form.confirmPassword) {
    errorMessage.value = '两次输入的新密码不一致';
    return;
  }

  submitting.value = true;
  try {
    await changePasswordApi({
      old_password: form.oldPassword,
      new_password: form.newPassword,
    });

    successMessage.value = '密码修改成功，正在返回个人中心...';
    form.oldPassword = '';
    form.newPassword = '';
    form.confirmPassword = '';

    setTimeout(() => {
      router.replace('/profile');
    }, 900);
  } catch (error) {
    if (error.response?.status === 401) {
      errorMessage.value = '旧密码错误，请重新输入';
    } else {
      errorMessage.value = error.response?.data?.message || '修改密码失败，请稍后重试';
    }
  } finally {
    submitting.value = false;
  }
};
</script>

<style scoped>
.password-page {
  max-width: 560px;
  margin: 40px auto;
}

.page-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.page-head .hint {
  margin-bottom: 0;
}

.back-link {
  color: #1c7ed6;
  text-decoration: none;
  font-size: 14px;
  white-space: nowrap;
}

.page-actions {
  display: flex;
  gap: 12px;
  margin-top: 8px;
}

.page-actions .btn {
  flex: 1;
}

.cancel-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  text-decoration: none;
}

@media (max-width: 640px) {
  .page-head,
  .page-actions {
    flex-direction: column;
  }

  .page-actions .btn {
    width: 100%;
  }
}
</style>
