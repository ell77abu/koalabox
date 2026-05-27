import { createRouter, createWebHistory } from 'vue-router';
import { authStore } from '../stores/auth';
import LoginView from '../views/LoginView.vue';
import RegisterView from '../views/RegisterView.vue';
import TaskView from '../views/TaskView.vue';
import ProfileView from '../views/ProfileView.vue';
import CreateTaskView from '../views/CreateTaskView.vue';
import TaskDependencyView from '../views/TaskDependencyView.vue';
import ChangePasswordView from '../views/ChangePasswordView.vue';
import TeamsView from '../views/TeamsView.vue';
import CreateTeamView from '../views/CreateTeamView.vue';
import TeamDetailView from '../views/TeamDetailView.vue';
import NotificationsView from '../views/NotificationsView.vue';

const routes = [
  {
    path: '/',
    redirect: '/tasks',
  },
  {
    path: '/login',
    name: 'login',
    component: LoginView,
    meta: { guestOnly: true },
  },
  {
    path: '/register',
    name: 'register',
    component: RegisterView,
    meta: { guestOnly: true },
  },
  {
    path: '/tasks',
    name: 'tasks',
    component: TaskView,
    meta: { requiresAuth: true },
  },
  {
    path: '/tasks/create',
    name: 'create-task',
    component: CreateTaskView,
    meta: { requiresAuth: true },
  },
  {
    path: '/tasks/:id/dependencies',
    name: 'task-dependencies',
    component: TaskDependencyView,
    meta: { requiresAuth: true },
  },
  {
    path: '/profile',
    name: 'profile',
    component: ProfileView,
    meta: { requiresAuth: true },
  },
  {
    path: '/profile/password',
    name: 'change-password',
    component: ChangePasswordView,
    meta: { requiresAuth: true },
  },

  // 系统消息列表页路由
  {
    path: '/notifications',
    name: 'notifications',
    component: NotificationsView,
    meta: { requiresAuth: true },
  },

  // 我的团队列表页路由
  {
    path: '/teams',
    name: 'teams',
    component: TeamsView,
    meta: { requiresAuth: true },
  },
  {
    path: '/teams/create',
    name: 'create-team',
    component: CreateTeamView,
    meta: { requiresAuth: true },
  },
  {
    path: '/teams/:team_id',
    name: 'team-detail',
    component: TeamDetailView,
    meta: { requiresAuth: true },
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to) => {
  if (to.meta.requiresAuth && !authStore.isLoggedIn.value) {
    return {
      path: '/login',
      query: { redirect: to.fullPath },
    };
  }

  if (to.meta.guestOnly && authStore.isLoggedIn.value) {
    return '/tasks';
  }

  return true;
});

// 全局前置路由守卫
router.beforeEach((to, from, next) => {
  // 尝试从 localStorage 获取 token (或者通过 authStore)
  const isAuthenticated = !!localStorage.getItem('token'); 

  if (to.meta.requiresAuth && !isAuthenticated) {
    // 试图访问需要登录的页面，但没有 Token，强制踢回登录页
    next({ name: 'login', query: { redirect: to.fullPath } });
  } else if (to.meta.guestOnly && isAuthenticated) {
    // 已经登录了，还试图访问登录/注册页，直接送回任务大厅
    next({ name: 'tasks' });
  } else {
    // 正常放行
    next();
  }
});

export default router;
