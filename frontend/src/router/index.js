import { createRouter, createWebHistory } from 'vue-router';
import { authStore } from '../stores/auth';
import LoginView from '../views/LoginView.vue';
import RegisterView from '../views/RegisterView.vue';
import TaskView from '../views/TaskView.vue';
import ProfileView from '../views/ProfileView.vue';
import CreateTaskView from '../views/CreateTaskView.vue';
import ChangePasswordView from '../views/ChangePasswordView.vue';
import TeamsView from '../views/TeamsView.vue';
import CreateTeamView from '../views/CreateTeamView.vue';
import TeamDetailView from '../views/TeamDetailView.vue';

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

export default router;
