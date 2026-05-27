<template>
  <section class="card profile-page">
    <div class="profile-header">
      <img :src="avatarUrl" alt="用户头像" class="profile-avatar" />
      <div>
        <h1>个人中心</h1>
        <p class="hint">在这里查看完整个人资料，并维护账号安全设置。</p>
      </div>
    </div>

    <div class="profile-info">
      <div class="info-item">
        <span class="label">用户名：</span>
        <span class="value">{{ authStore.username.value }}</span>
      </div>
      <div class="info-item">
        <span class="label">用户身份：</span>
        <span class="value tag">{{ levelLabel }}</span>
      </div>
      <div class="info-item">
        <span class="label">身份等级：</span>
        <span class="value">Level {{ authStore.level.value }}</span>
      </div>
      <div class="info-item">
        <span class="label">账号头像：</span>
        <img :src="avatarUrl" alt="用户头像" class="profile-inline-avatar" />
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
import { computed } from 'vue';
import { RouterLink, useRouter } from 'vue-router';
import { logoutApi } from '../services/auth';
import { authStore } from '../stores/auth';
import { getApiErrorMessage } from '../utils/apiError';

const router = useRouter();

const avatarUrl = computed(
  () => `https://api.dicebear.com/7.x/adventurer/svg?seed=${encodeURIComponent(authStore.username.value || 'user')}`
);

const levelLabel = computed(() => (Number(authStore.level.value) === 1 ? '管理员' : '普通用户'));

const onLogout = async () => {
  if (confirm('确定要退出登录吗？')) {
    try {
      await logoutApi();
    } catch (error) {
      // Ignore API logout errors and clear local state anyway.
      console.warn(getApiErrorMessage(error, '退出登录接口失败，本地登录状态已清除'));
    }
    authStore.clearAuth();
    router.replace('/login');
  }
};
</script>

<style scoped>
.profile-page {
  max-width: 720px;
  margin: 40px auto;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 18px;
}

.profile-header .hint {
  margin-bottom: 0;
}

.profile-avatar {
  width: 88px;
  height: 88px;
  border-radius: 50%;
  display: block;
  border: 4px solid #dbeafe;
  background: #fff;
  box-shadow: 0 12px 26px rgba(37, 99, 235, 0.14);
}

.profile-info {
  margin: 30px 0;
  padding: 24px;
  background: linear-gradient(180deg, #f8fbff 0%, #f8fafc 100%);
  border-radius: 16px;
  border: 1px solid #dbe3f0;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
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

.profile-inline-avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  display: block;
  border: 2px solid #dbeafe;
  background: #fff;
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

@media (max-width: 640px) {
  .profile-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .info-item {
    align-items: flex-start;
    flex-direction: column;
    gap: 8px;
  }
}
</style>
