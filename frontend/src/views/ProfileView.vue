<template>
  <section class="card profile-page">
    <h1>个人信息</h1>
    <div class="profile-info">
      <div class="info-item">
        <span class="label">用户名：</span>
        <span class="value">{{ authStore.username.value }}</span>
      </div>
      <div class="info-item">
        <span class="label">当前身份：</span>
        <span class="value tag">普通用户</span>
      </div>
    </div>
    <div class="profile-top-actions">
      <RouterLink to="/profile/password" class="change-password-link">
        修改密码
      </RouterLink>
    </div>
    <div class="profile-bottom-actions">
      <button class="btn logout-btn" @click="onLogout">退出登录</button>
    </div>
  </section>
</template>

<script setup>
import { RouterLink, useRouter } from 'vue-router';
import { authStore } from '../stores/auth';

const router = useRouter();

const onLogout = () => {
  if (confirm('确定要退出登录吗？')) {
    authStore.clearAuth();
    router.replace('/login');
  }
};
</script>

<style scoped>
.profile-page {
  max-width: 600px;
  margin: 40px auto;
}

.profile-info {
  margin: 30px 0;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px border #eee;
}

.info-item:last-child {
  border-bottom: none;
}

.label {
  color: #666;
}

.value {
  font-weight: 500;
  color: #2c3e50;
}

.tag {
  background: #e3f2fd;
  color: #1976d2;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.9em;
}

.profile-top-actions {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  margin-top: 20px;
}

.logout-btn {
  background-color: #ff4d4f;
}

.logout-btn:hover {
  background-color: #ff7875;
}

.change-password-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 30px;
  padding: 0 10px;
  border: 1px solid #bfdbfe;
  border-radius: 7px;
  background: #eff6ff;
  color: #2563eb;
  text-decoration: none;
  font-size: 12px;
  font-weight: 500;
  line-height: 1;
}

.change-password-link:hover {
  background: #dbeafe;
  color: #1d4ed8;
}

.profile-bottom-actions {
  display: flex;
  justify-content: center;
  margin-top: 24px;
}
</style>
