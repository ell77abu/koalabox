<template>
  <div v-if="authStore.isLoggedIn.value" class="app-layout">
    <!-- 侧边栏导航 -->
    <aside class="sidebar">
      <div class="logo">任务管理系统</div>
      <nav class="nav-menu">
        <RouterLink to="/tasks" class="nav-item" active-class="active">
          <span class="icon">📋</span> 任务中心
        </RouterLink>
      </nav>
      <RouterLink to="/profile" class="user-info" active-class="active-user-info">
        <span class="username">{{ authStore.username.value }}</span>
      </RouterLink>
    </aside>

    <!-- 右侧内容区 -->
    <main class="main-container">
      <div class="content-wrapper">
        <RouterView />
      </div>
    </main>
  </div>

  <!-- 登录注册页不显示侧边栏 -->
  <RouterView v-else />
</template>

<script setup>
import { RouterView, RouterLink } from 'vue-router';
import { authStore } from './stores/auth';
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

.user-info {
  display: block;
  padding: 16px 24px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  font-size: 14px;
  color: rgba(255, 255, 255, 0.85);
  text-decoration: none;
  transition: all 0.3s;
}

.user-info:hover,
.active-user-info {
  color: #fff;
  background: rgba(255, 255, 255, 0.08);
}

.username {
  display: inline-block;
  font-weight: 500;
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
</style>
