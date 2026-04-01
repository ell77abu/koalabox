<template>
  <div v-if="authStore.isLoggedIn.value" class="app-layout">
    <aside class="sidebar">
      <div class="logo">Collabox</div>
      <nav class="nav-menu">
        <RouterLink to="/tasks" class="nav-item" active-class="active">
          <span class="icon">📋</span> 任务中心
        </RouterLink>
      </nav>
    </aside>

    <main class="main-container">
      <div class="content-wrapper">
        <header class="page-topbar">
          <div class="profile-entry">
            <div class="profile-trigger">
              <img :src="avatarUrl" alt="用户头像" class="profile-avatar" />
              <span class="profile-name">{{ authStore.username.value }}</span>
            </div>

            <div class="profile-popover">
              <img :src="avatarUrl" alt="用户头像" class="profile-popover-avatar" />
              <div class="profile-popover-name">{{ authStore.username.value }}</div>
              <div class="profile-popover-level">{{ levelLabel }}</div>
              <RouterLink to="/profile" class="profile-popover-link">
                个人中心
              </RouterLink>
            </div>
          </div>
        </header>
        <RouterView />
      </div>
    </main>
  </div>

  <!-- 登录注册页不显示侧边栏 -->
  <RouterView v-else />
</template>

<script setup>
import { computed } from 'vue';
import { RouterView, RouterLink } from 'vue-router';
import { authStore } from './stores/auth';

const avatarUrl = computed(
  () => `https://api.dicebear.com/7.x/adventurer/svg?seed=${encodeURIComponent(authStore.username.value || 'user')}`
);

const levelLabel = computed(() => (Number(authStore.level.value) === 1 ? '管理员' : '普通用户'));
</script>

<style scoped>
.app-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  background-color: #f0f2f5;
}

.sidebar {
  width: 240px;
  background-color: #001529;
  color: #fff;
  display: flex;
  flex-direction: column;
}

.logo {
  height: 64px;
  line-height: 64px;
  text-align: center;
  font-size: 20px;
  font-weight: bold;
  background: rgba(255, 255, 255, 0.1);
}

.nav-menu {
  flex: 1;
  padding: 16px 0;
}

.nav-item {
  display: flex;
  align-items: center;
  padding: 12px 24px;
  color: rgba(255, 255, 255, 0.65);
  text-decoration: none;
  transition: all 0.3s;
}

.nav-item:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.05);
}

.nav-item.active {
  color: #fff;
  background: #1890ff;
}

.icon {
  margin-right: 12px;
  font-size: 18px;
}

.main-container {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.content-wrapper {
  max-width: 1200px;
  margin: 0 auto;
}

.page-topbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 18px;
}

.profile-entry {
  position: relative;
}

.profile-trigger {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.78);
  border: 1px solid rgba(203, 213, 225, 0.9);
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.08);
  transition: transform 0.24s ease, box-shadow 0.24s ease;
  backdrop-filter: blur(10px);
}

.profile-entry:hover .profile-trigger {
  transform: translateY(-4px);
  box-shadow: 0 14px 28px rgba(15, 23, 42, 0.14);
}

.profile-avatar {
  width: 42px;
  height: 42px;
  border-radius: 50%;
  display: block;
  border: 2px solid #dbeafe;
  background: #fff;
}

.profile-name {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
}

.profile-popover {
  position: absolute;
  top: calc(100% + 12px);
  right: 0;
  width: 220px;
  padding: 18px 16px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.98);
  border: 1px solid #dbe3f0;
  box-shadow: 0 20px 40px rgba(15, 23, 42, 0.16);
  opacity: 0;
  visibility: hidden;
  transform: translateY(-8px) scale(0.98);
  transform-origin: top right;
  transition: opacity 0.22s ease, transform 0.22s ease, visibility 0.22s ease;
  text-align: center;
  z-index: 10;
}

.profile-entry:hover .profile-popover {
  opacity: 1;
  visibility: visible;
  transform: translateY(0) scale(1);
}

.profile-popover-avatar {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  display: block;
  margin: 0 auto 12px;
  border: 3px solid #dbeafe;
  background: #fff;
}

.profile-popover-name {
  font-size: 16px;
  font-weight: 700;
  color: #111827;
}

.profile-popover-level {
  margin-top: 6px;
  color: #64748b;
  font-size: 13px;
}

.profile-popover-link {
  margin-top: 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 108px;
  min-height: 36px;
  padding: 0 14px;
  border-radius: 999px;
  background: #eff6ff;
  color: #2563eb;
  text-decoration: none;
  font-size: 13px;
  font-weight: 600;
}

.profile-popover-link:hover {
  background: #dbeafe;
}
</style>
