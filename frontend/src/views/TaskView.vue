<template>
  <div class="page-shell">
    <section class="card card-wide">
      <header class="task-header">
        <div class="task-title">
          <h1>任务管理</h1>
          <p class="hint">欢迎你，{{ authStore.username.value }}</p>
        </div>
        <div class="header-actions">
          <div class="search-box">
            <span class="search-fab" aria-hidden="true">
              <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
                <circle cx="11" cy="11" r="7" fill="none" stroke="currentColor" stroke-width="4" />
                <line x1="16.65" y1="16.65" x2="22" y2="22" stroke="currentColor" stroke-width="4" stroke-linecap="round" />
              </svg>
            </span>
            <input
              v-model.trim="searchQuery"
              type="search"
              placeholder="搜索任务"
              class="search-input"
            />
            <button
              v-if="isSearching"
              type="button"
              class="search-clear"
              @click="searchQuery = ''"
            >
              清除
            </button>
          </div>
          <RouterLink to="/tasks/create" class="fab"  data-hint="创建任务" title="创建任务">
            <span class="plus-icon">+</span>
          </RouterLink>
        </div>
      </header>

      <p v-if="teamLoadError" class="error-banner">{{ teamLoadError }}</p>

      <div class="list-section">
        <div class="list-header">
          <h2>任务列表</h2>
          <p v-if="!loading && displayedTasks.length > 0" class="hint">
            <template v-if="isSearching">
              搜索结果 {{ displayedTasks.length }} 项
            </template>
            <template v-else>
              共 {{ taskStore.total.value }} 个任务
            </template>
          </p>
        </div>

        <div v-if="loading" class="loading-state">
          <p>加载中...</p>
        </div>

        <div v-else-if="loadError" class="empty-state error-state">
          <p>{{ loadError }}</p>
          <div class="error-actions">
            <button class="btn btn-small btn-primary" type="button" @click="loadTasks">重试</button>
          </div>
        </div>

        <!-- 只有当账户里【彻底一个任务都没有】时，才显示全局大空状态 -->
        <div v-else-if="taskStore.tasks.value.length === 0" class="empty-state">
          <p>暂无任务，创建一个新任务开始吧！</p>
        </div>

        <div v-else class="task-list">
          <section class="task-group">
            <div class="task-group-header">
              <div>
                <h3>待办任务</h3>
              </div>
              <div class="task-group-actions">
                <!-- 筛选下拉控件 -->
                <div class="sort-control">
                  <span class="sort-fab" aria-hidden="true" style="background: #10b981;"> <!-- 用绿色区分筛选 -->
                    <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
                      <!-- 漏斗图标 -->
                      <path d="M3 4c0-.55.45-1 1-1h16c.55 0 1 .45 1 1v1.59c0 .27-.11.52-.29.71L15 13v6c0 .55-.45 1-1 1h-4c-.55 0-1-.45-1-1v-6L3.29 6.3A.996.996 0 0 1 3 5.59V4z" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                  </span>
                  <div ref="filterDropdownRef" class="sort-dropdown">
                    <button
                      type="button"
                      class="sort-trigger"
                      :aria-expanded="filterOpen"
                      @click="toggleFilterMenu"
                    >
                      <span>{{ currentFilterLabel }}</span>
                      <span class="sort-caret" :class="{ open: filterOpen }"></span>
                    </button>

                    <transition name="sort-menu-fade">
                      <ul v-if="filterOpen" class="sort-menu" role="listbox" aria-label="任务筛选">
                        <li v-for="option in filterOptions" :key="option.value">
                          <button
                            type="button"
                            class="sort-option"
                            :class="{ active: filterOption === option.value }"
                            @click="selectFilterOption(option.value)"
                          >
                            {{ option.label }}
                          </button>
                        </li>
                      </ul>
                    </transition>
                  </div>
                </div>

                <div class="sort-control">
                  <span class="sort-fab" aria-hidden="true">
                    <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
                      <path
                        d="M22 3H2l8 9.46V19l4 2v-8.54L22 3z"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                      />
                    </svg>
                  </span>
                  <div ref="sortDropdownRef" class="sort-dropdown">
                    <button
                      type="button"
                      class="sort-trigger"
                      :aria-expanded="sortOpen"
                      @click="toggleSortMenu"
                    >
                      <span>{{ currentSortLabel }}</span>
                      <span class="sort-caret" :class="{ open: sortOpen }"></span>
                    </button>

                    <transition name="sort-menu-fade">
                      <ul v-if="sortOpen" class="sort-menu" role="listbox" aria-label="任务排序">
                        <li v-for="option in sortOptions" :key="option.value">
                          <button
                            type="button"
                            class="sort-option"
                            :class="{ active: sortOption === option.value }"
                            @click="selectSortOption(option.value)"
                          >
                            {{ option.label }}
                          </button>
                        </li>
                      </ul>
                    </transition>
                  </div>
                </div>
                <span class="task-group-count">{{ incompleteTasks.length }}</span>
              </div>
            </div>

            <!-- 当前分类/搜索下，一个任务都没 -->
            <div v-if="displayedTasks.length === 0" class="sub-empty-state">
              <p v-if="isSearching">未找到包含关键词的匹配任务。</p>
              <p v-else>当前分类下暂无任务。</p>
            </div>

            <!-- 有任务，且有【未完成】的任务 -->
            <div v-else-if="incompleteTasks.length > 0" class="task-group-list">
              <div
                v-for="task in incompleteTasks"
                :key="task.id"
                class="task-item"
                :class="{
                  'task-item-celebrating': celebrationTaskId === task.id,
                  'task-item-completing': completingTaskId === task.id,
                }"
              >
                <div v-if="celebrationTaskId === task.id" class="confetti-burst" aria-hidden="true">
                  <span
                    v-for="piece in confettiPieces"
                    :key="`${task.id}-${piece}`"
                    class="confetti-piece"
                    :style="{ '--piece-index': piece }"
                  ></span>
                </div>
                <div class="task-main">
                  <div class="task-title-section">
                    <h3>{{ task.title }}</h3>
                    <RouterLink
                      v-if="!isPersonalTask(task)"
                      class="priority-badge team-link-badge"
                      :to="`/teams/${task.team_id}`"
                    >
                      👥 {{ getTeamName(task) }}
                    </RouterLink>
                    <span class="status-badge" :class="`status-${task.status.toLowerCase()}`">
                      {{ statusLabel(task.status) }}
                    </span>
                    <span
                      class="priority-badge"
                      :class="`priority-${task.priority}`"
                    >
                      {{ priorityLabel(task.priority) }}
                    </span>
                  </div>

                  <p v-if="task.description" class="task-description">{{ task.description }}</p>

                  <div class="task-meta">
                    <span v-if="task.deadline" class="meta-item">
                      <span class="meta-label">截止时间：</span>{{ formatDate(task.deadline) }}
                    </span>
                    <span class="meta-item">
                      <span class="meta-label">创建时间：</span>{{ formatDate(task.created_at) }}
                    </span>
                  </div>
                </div>

                <div class="task-actions">
                  <button
                    v-if="canUpdateTaskStatus(task)"
                    class="btn btn-small btn-complete"
                    @click="markTaskDone(task)"
                    :disabled="updatingId === task.id"
                  >
                    {{ updatingId === task.id ? '完成中...' : '完成任务' }}
                  </button>
                  <RouterLink
                    class="btn btn-small btn-dependency"
                    :to="`/tasks/${task.id}/dependencies`"
                  >
                    依赖详情
                  </RouterLink>
                  <button
                    v-if="canUpdateTaskStatus(task)"
                    class="btn btn-small btn-edit"
                    @click="startEdit(task)"
                    :disabled="updatingId === task.id"
                  >
                    编辑
                  </button>
                  <button
                    v-if="canDeleteTask(task)"
                    class="btn btn-small btn-danger"
                    @click="onDeleteTask(task)"
                    :disabled="deletingId === task.id"
                  >
                    {{ deletingId === task.id ? '删除中...' : '删除' }}
                  </button>
                </div>
              </div>
            </div>

            <!-- 有任务，全都是【已完成】的任务 -->
            <div v-else class="sub-empty-state">
              <p>当前没有未完成任务。</p>
            </div>
          </section>

          <section class="task-group task-group-history">
            <button
              type="button"
              class="history-toggle"
              @click="showCompletedHistory = !showCompletedHistory"
            >
              <div>
                <h3>已完成历史</h3>
                <p>点击 {{ showCompletedHistory ? '收起' : '展开' }}</p>
              </div>
              <div class="history-toggle-meta">
                <span class="task-group-count done">{{ completedTasks.length }}</span>
              </div>
            </button>

            <div v-if="showCompletedHistory">
              <div v-if="completedTasks.length" class="task-group-list">
                <div
                  v-for="task in completedTasks"
                  :key="task.id"
                  class="task-item task-item-done"
                  :class="{
                    'task-item-celebrating': celebrationTaskId === task.id,
                    'task-item-arrived': arrivedTaskId === task.id,
                  }"
                >
                  <div v-if="celebrationTaskId === task.id" class="confetti-burst" aria-hidden="true">
                    <span
                      v-for="piece in confettiPieces"
                      :key="`${task.id}-${piece}`"
                      class="confetti-piece"
                      :style="{ '--piece-index': piece }"
                    ></span>
                  </div>
                  <div class="task-main">
                    <div class="task-title-section">
                      <h3>{{ task.title }}</h3>
                      <RouterLink
                        v-if="!isPersonalTask(task)"
                        class="priority-badge team-link-badge"
                        :to="`/teams/${task.team_id}`"
                      >
                        👥 {{ getTeamName(task) }}
                      </RouterLink>
                      <span class="status-badge" :class="`status-${task.status.toLowerCase()}`">
                        {{ statusLabel(task.status) }}
                      </span>
                      <span
                        class="priority-badge"
                        :class="`priority-${task.priority}`"
                      >
                        {{ priorityLabel(task.priority) }}
                      </span>
                    </div>

                    <p v-if="task.description" class="task-description">{{ task.description }}</p>

                    <div class="task-meta">
                      <span v-if="task.deadline" class="meta-item">
                        <span class="meta-label">截止时间：</span>{{ formatDate(task.deadline) }}
                      </span>
                      <span class="meta-item">
                        <span class="meta-label">创建时间：</span>{{ formatDate(task.created_at) }}
                      </span>
                    </div>
                  </div>

                  <div class="task-actions">
                    <RouterLink
                      class="btn btn-small btn-dependency"
                      :to="`/tasks/${task.id}/dependencies`"
                    >
                      依赖详情
                    </RouterLink>
                    <button
                      v-if="canUpdateTaskStatus(task)"
                      class="btn btn-small btn-edit"
                      @click="startEdit(task)"
                      :disabled="updatingId === task.id"
                    >
                      编辑
                    </button>
                    <button
                      v-if="canDeleteTask(task)"
                      class="btn btn-small btn-danger"
                      @click="onDeleteTask(task)"
                      :disabled="deletingId === task.id"
                    >
                      {{ deletingId === task.id ? '删除中...' : '删除' }}
                    </button>
                  </div>
                </div>
              </div>

              <div v-else class="sub-empty-state">
                <p>还没有已完成的历史任务。</p>
              </div>
            </div>
          </section>
        </div>
      </div>

      <div v-if="editingTask.id" class="modal-overlay" @click="cancelEdit">
        <div class="modal" @click.stop>
          <h2>编辑任务</h2>
          <form @submit.prevent="onUpdateTask" class="task-form">
            <input
              v-model.trim="editingTask.title"
              type="text"
              placeholder="任务标题"
              required
              class="form-input"
              :disabled="!canEditTaskDetails(editingTask)"
            />

            <textarea
              v-model.trim="editingTask.description"
              placeholder="任务描述"
              class="form-input form-textarea"
              :disabled="!canEditTaskDetails(editingTask)"
            ></textarea>
            
            <p v-if="!canEditTaskDetails(editingTask)" style="color: #ef4444; font-size: 13px; margin: 0; font-weight: 600;">
              * 作为该团队成员，您仅有权限修改此任务的执行状态。
            </p>

            <select v-model="editingTask.status" class="form-input">
              <option value="TODO">状态：待做</option>
              <option value="DOING">状态：进行中</option>
              <option value="DONE">状态：已完成</option>
            </select>

            <select v-model="editingTask.priority" class="form-input" :disabled="!canEditTaskDetails(editingTask)">
              <option value="1">优先级：低</option>
              <option value="2">优先级：中</option>
              <option value="3">优先级：高</option>
            </select>

            <input
              v-model.trim="editingTask.deadline"
              type="datetime-local"
              class="form-input date-input"
              :disabled="!canEditTaskDetails(editingTask)"
            />

            <div v-if="isPersonalTask(editingTask)" class="dependency-section">
              <label class="section-label" style="font-size: 0.9rem; font-weight: 600; color: #475569; display: block; margin-bottom: 0.5rem;">
                设置前置依赖 (仅限个人任务)
              </label>
              <div class="checkbox-list" style="display: flex; flex-direction: column; gap: 0.5rem; max-height: 160px; overflow-y: auto; padding: 0.75rem; background: #f8fafc; border: 1px solid #cbd5e1; border-radius: 0.5rem;">
                <label v-for="t in availablePersonalTasks" :key="t.id" class="checkbox-item" style="display: flex; align-items: center; gap: 0.5rem; font-size: 0.85rem; color: #1e293b; cursor: pointer;">
                  <input type="checkbox" :value="t.id" v-model="editingTask.predecessor_ids" style="accent-color: #3b82f6; width: 16px; height: 16px; cursor: pointer;" />
                  {{ t.title }}
                  <span class="status-badge" :class="`status-${t.status.toLowerCase()}`" style="transform: scale(0.85); transform-origin: left;">
                    {{ statusLabel(t.status) }}
                  </span>
                </label>
                <span v-if="availablePersonalTasks.length === 0" style="font-size: 0.85rem; color: #94a3b8;">暂无可依赖的其他个人任务</span>
              </div>
            </div>

            <div class="modal-actions">
              <button type="button" class="btn btn-outline" @click="cancelEdit">取消</button>
              <button type="submit" class="btn btn-primary" :disabled="updatingId === editingTask.id">
                {{ updatingId === editingTask.id ? '保存中...' : '保存' }}
              </button>
            </div>

            <p v-if="updateError" class="error-msg">{{ updateError }}</p>
          </form>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch } from 'vue';
import { useRouter, RouterLink } from 'vue-router';
import { authStore } from '../stores/auth';
import { taskStore } from '../stores/task';
import {
  getTasksApi,
  createTaskApi,
  updateTaskApi,
  deleteTaskApi,
  getTaskDependenciesApi,
  addTaskDependencyApi,
  removeTaskDependencyApi
} from '../services/task';
import {
  getApiErrorMessage,
  getApiFailureMessage,
  getUnresolvedDependenciesMessage,
} from '../utils/apiError';

// 团队数据与接口
import { teamStore } from '../stores/team';
import { getTeamsApi } from '../services/team';

const router = useRouter();
const isSystemAdmin = computed(() => Number(authStore.level.value) === 1);

const isPersonalTask = (task) => {
  const teamId = task?.team_id;
  return teamId === null || teamId === undefined || teamId === '' || teamId === 0 || teamId === '0';
};

// === Lab2 权限控制逻辑 ===
// 进页面时拉取用户的团队列表
const fetchUserTeams = async () => {
  try {
    const res = await getTeamsApi();
    if (res.data.code === 0) {
      teamStore.setTeams(res.data.data.list || []);
      teamLoadError.value = '';
    } else {
      teamLoadError.value = getApiFailureMessage(res, '加载团队列表失败');
    }
  } catch (err) {
    teamLoadError.value = getApiErrorMessage(err, '加载团队列表失败');
    console.error(teamLoadError.value, err);
  }
};

// 判断当前用户在这个任务中，是否有高级编辑权限（修改标题、删除等）
const canEditTaskDetails = (task) => {
  if (isPersonalTask(task)) return !isSystemAdmin.value;
  
  const teamId = Number(task.team_id);
  const team = teamStore.teams.value.find(t => Number(t.team_id) === teamId);
  const role = team ? team.role : 'Member'; 
  
  return role === 'Owner' || role === 'Admin';
};

const canDeleteTask = (task) => {
  return isSystemAdmin.value || canEditTaskDetails(task);
};

const canUpdateTaskStatus = (task) => {
  if (canEditTaskDetails(task)) return true;
  if (isSystemAdmin.value) return false;
  return !isPersonalTask(task);
};

const getTeamName = (task) => {
  const teamId = Number(task?.team_id);
  const team = teamStore.teams.value.find((item) => Number(item.team_id) === teamId);
  return team?.name || '团队任务';
};
// === Lab2 逻辑结束 ===

// 新任务表单
const newTask = reactive({
  title: '',
  description: '',
  priority: '2',
  deadline: '',
});

// 编辑任务
const editingTask = reactive({
  id: null,
  title: '',
  description: '',
  status: 'TODO',
  priority: '2',
  deadline: '',
  team_id: null,
  predecessor_ids: [] // 用于绑定弹窗里的多选框（用于依赖设置）
});

//  记录打开弹窗时的原始依赖，用于比对增删
const editingDependencies = ref([]); 

// 过滤出“自己的其他个人任务”作为可选依赖池
const availablePersonalTasks = computed(() => {
  return taskStore.tasks.value.filter(t => isPersonalTask(t) && t.id !== editingTask.id);
});

// 状态管理
const findTaskTitleById = (id) => {
  const task = taskStore.tasks.value.find((item) => Number(item.id) === Number(id));
  return task?.title || '';
};

const loading = ref(false);
const loadError = ref('');
const teamLoadError = ref('');
const creatingTask = ref(false);
const updatingId = ref(null);
const deletingId = ref(null);
const createError = ref('');
const updateError = ref('');
const searchQuery = ref('');
const sortOption = ref('default');
const sortOpen = ref(false);
const sortDropdownRef = ref(null);
// 筛选专用的状态变量
const filterOption = ref('all');
const filterOpen = ref(false);
const filterDropdownRef = ref(null);
const filterOptions = [
  { value: 'all', label: '全部任务' },
  { value: 'personal', label: '个人任务' },
  { value: 'team', label: '团队任务' },
];

const currentFilterLabel = computed(() => {
  const option = filterOptions.find((item) => item.value === filterOption.value);
  return option ? option.label : '全部任务';
});

const toggleFilterMenu = () => {
  filterOpen.value = !filterOpen.value;
  if (filterOpen.value) sortOpen.value = false; // 打开筛选时，自动关闭排序菜单
};

const selectFilterOption = (value) => {
  filterOption.value = value;
  filterOpen.value = false;
};

const showCompletedHistory = ref(false);
const celebrationTaskId = ref(null);
const completingTaskId = ref(null);
const arrivedTaskId = ref(null);
const confettiPieces = Array.from({ length: 14 }, (_, index) => index);
let celebrationTimer = null;
let completeMoveTimer = null;
let arrivalTimer = null;

const sortOptions = [
  { value: 'default', label: '按默认顺序' },
  { value: 'deadline', label: '按截止时间' },
  { value: 'priority', label: '按优先级' },
];

const currentSortLabel = computed(() => {
  const option = sortOptions.find((item) => item.value === sortOption.value);
  return option ? option.label : '默认顺序';
});

// 让两个菜单不会同时重叠打开
const toggleSortMenu = () => {
  sortOpen.value = !sortOpen.value;
  if (sortOpen.value) filterOpen.value = false;
};

const selectSortOption = (value) => {
  sortOption.value = value;
  sortOpen.value = false;
};

// 兼容两个下拉菜单的外部点击关闭事件
const handleOutsideClick = (event) => {
  if (sortDropdownRef.value && !sortDropdownRef.value.contains(event.target)) {
    sortOpen.value = false;
  }
  if (filterDropdownRef.value && !filterDropdownRef.value.contains(event.target)) {
    filterOpen.value = false;
  }
};

// 核心计算属性 (添加筛选管道)
const displayedTasks = computed(() => {
  let list = [...taskStore.tasks.value];

  // 拦截层：按个人/团队筛选
  if (filterOption.value === 'personal') {
    list = list.filter(isPersonalTask); // 没有 team_id 的是个人任务
  } else if (filterOption.value === 'team') {
    list = list.filter(task => !isPersonalTask(task)); // 有 team_id 的是团队任务
  }

  // 拦截层：按关键词搜索
  const keyword = searchQuery.value.trim().toLowerCase();
  let filtered = keyword
    ? list.filter((task) => String(task.title || '').toLowerCase().includes(keyword))
    : list;

  // 拦截层：排序逻辑
  if (sortOption.value === 'deadline') {
    return filtered.sort((a, b) => {
      const aTime = a.deadline ? new Date(a.deadline).getTime() : Number.POSITIVE_INFINITY;
      const bTime = b.deadline ? new Date(b.deadline).getTime() : Number.POSITIVE_INFINITY;
      const aValid = Number.isFinite(aTime);
      const bValid = Number.isFinite(bTime);
      if (!aValid && !bValid) return 0;
      if (!aValid) return 1;
      if (!bValid) return -1;
      return aTime - bTime;
    });
  }
  if (sortOption.value === 'priority') {
    return filtered.sort((a, b) => {
      const priorityDiff = Number(b.priority || 0) - Number(a.priority || 0);
      if (priorityDiff !== 0) return priorityDiff;
      return new Date(b.created_at || 0).getTime() - new Date(a.created_at || 0).getTime();
    });
  }

  return filtered;
});

const isSearching = computed(() => searchQuery.value.trim().length > 0);

const incompleteTasks = computed(() =>
  displayedTasks.value.filter((task) => task.status !== 'DONE')
);

const completedTasks = computed(() =>
  displayedTasks.value.filter((task) => task.status === 'DONE')
);

// 退出登录
const logout = async () => {
  taskStore.clearTasks();
  authStore.clearAuth();
  await router.replace('/login');
};

// 加载任务列表 - 一次加载所有任务
const loadTasks = async () => {
  loading.value = true;
  loadError.value = '';
  taskStore.setLoading(true);
  try {
    // 一次加载所有任务，使用大的 pageSize
    const response = await getTasksApi(1, 1000, sortOption.value);
    if (response.data.code === 0) {
      const pageData = response.data.data || {};
      const list = Array.isArray(pageData.list) ? pageData.list : [];
      taskStore.setTasks(list);
      const total = Number.isFinite(pageData.total) ? pageData.total : list.length;
      taskStore.setTotal(total);
    } else {
      loadError.value = getApiFailureMessage(response, '加载任务失败');
      taskStore.clearTasks();
    }
  } catch (error) {
    loadError.value = getApiErrorMessage(error, '加载任务失败');
    console.error(loadError.value, error);
  } finally {
    loading.value = false;
    taskStore.setLoading(false);
  }
};

// 创建任务
const onCreateTask = async () => {
  if (!newTask.title.trim()) {
    createError.value = '任务标题不能为空';
    return;
  }

  creatingTask.value = true;
  createError.value = '';
  try {
    const payload = {
      title: newTask.title,
      description: newTask.description || null,
      priority: parseInt(newTask.priority),
      deadline: toBackendDeadline(newTask.deadline),
      status: 'TODO',
    };

    const response = await createTaskApi(payload);
    if (response.data.code === 0) {
      // 清空表单
      newTask.title = '';
      newTask.description = '';
      newTask.priority = '2';
      newTask.deadline = '';
      // 刷新列表
      await loadTasks();
    } else {
      createError.value = getApiFailureMessage(response, '创建任务失败', '创建失败');
    }
  } catch (error) {
    createError.value = getApiErrorMessage(error, '创建任务失败');
  } finally {
    creatingTask.value = false;
  }
};

// 更新任务
const onUpdateTask = async () => {
  if (!canUpdateTaskStatus(editingTask)) {
    updateError.value = '无权限更新该任务';
    return;
  }

  updatingId.value = editingTask.id;
  updateError.value = '';
  try {
    const payload = canEditTaskDetails(editingTask)
      ? {
          title: editingTask.title,
          description: editingTask.description || null,
          status: editingTask.status,
          priority: parseInt(editingTask.priority),
          deadline: toBackendDeadline(editingTask.deadline),
        }
      : { status: editingTask.status };

    const response = await updateTaskApi(editingTask.id, payload);
    if (response.data.code === 0) {
      if (isPersonalTask(editingTask)) {
        // 算出用户勾选了哪些新任务，取消了哪些老任务
        const toAdd = editingTask.predecessor_ids.filter(id => !editingDependencies.value.includes(id));
        const toRemove = editingDependencies.value.filter(id => !editingTask.predecessor_ids.includes(id));
        const dependencyErrors = [];

        for (const id of toAdd) {
          try {
            await addTaskDependencyApi(editingTask.id, id);
          } catch (err) {
            const title = findTaskTitleById(id) || `ID ${id}`;
            dependencyErrors.push(getApiErrorMessage(err, `添加前置依赖「${title}」失败`));
          }
        }
        for (const id of toRemove) {
          try {
            await removeTaskDependencyApi(editingTask.id, id);
          } catch (err) {
            const title = findTaskTitleById(id) || `ID ${id}`;
            dependencyErrors.push(getApiErrorMessage(err, `移除前置依赖「${title}」失败`));
          }
        }

        if (dependencyErrors.length > 0) {
          updateError.value = dependencyErrors.join('；');
          return;
        }
      }
      cancelEdit();
      await loadTasks();
    } else {
      updateError.value = getApiFailureMessage(response, '更新任务失败', '更新失败');
    }
  } catch (error) {
    if (error.response?.status === 409) {
      updateError.value = getUnresolvedDependenciesMessage(error, findTaskTitleById, '更新任务失败');
    } else {
      updateError.value = getApiErrorMessage(error, '更新任务失败');
    }
  } finally {
    updatingId.value = null;
  }
};

const triggerCelebration = (taskId) => {
  celebrationTaskId.value = taskId;
  if (celebrationTimer) {
    clearTimeout(celebrationTimer);
  }
  celebrationTimer = window.setTimeout(() => {
    celebrationTaskId.value = null;
    celebrationTimer = null;
  }, 1600);
};

const moveTaskToCompleted = (taskId) => {
  taskStore.setTasks(
    taskStore.tasks.value.map((item) =>
      item.id === taskId
        ? { ...item, status: 'DONE' }
        : item
    )
  );
};

const markTaskDone = async (task) => {
  if (task.status === 'DONE') {
    return;
  }
  if (!canUpdateTaskStatus(task)) {
    alert('无权限更新该任务');
    return;
  }

  updatingId.value = task.id;
  try {
    const payload = canEditTaskDetails(task)
      ? {
          title: task.title,
          description: task.description || null,
          status: 'DONE',
          priority: parseInt(task.priority),
          deadline: task.deadline || null,
        }
      : { status: 'DONE' };

    const response = await updateTaskApi(task.id, payload);
    if (response.data.code === 0) {
      completingTaskId.value = task.id;
      triggerCelebration(task.id);
      showCompletedHistory.value = true;

      if (completeMoveTimer) {
        clearTimeout(completeMoveTimer);
      }
      if (arrivalTimer) {
        clearTimeout(arrivalTimer);
      }

      completeMoveTimer = window.setTimeout(() => {
        moveTaskToCompleted(task.id);
        completingTaskId.value = null;
        arrivedTaskId.value = task.id;
        completeMoveTimer = null;

        arrivalTimer = window.setTimeout(() => {
          arrivedTaskId.value = null;
          arrivalTimer = null;
        }, 850);
      }, 950);
    } else {
      alert(getApiFailureMessage(response, '完成任务失败', '更新失败'));
    }
  } catch (error) {
    // 新增：拦截 409 状态码，提示前置任务未完成
    if (error.response?.status === 409) {
      alert(getUnresolvedDependenciesMessage(error, findTaskTitleById, '完成任务失败'));
    } else {
      alert(getApiErrorMessage(error, '完成任务失败'));
    }
  } finally {
    updatingId.value = null;
  }
};

// 删除任务
const onDeleteTask = async (task) => {
  if (!canDeleteTask(task)) {
    alert('无权限删除该任务');
    return;
  }

  const id = task.id;
  if (!window.confirm('确定要删除这个任务吗？')) {
    return;
  }

  deletingId.value = id;
  try {
    const response = await deleteTaskApi(id);
    if (response.data.code === 0) {
      await loadTasks();
    } else {
      alert(getApiFailureMessage(response, '删除任务失败', '删除失败'));
    }
  } catch (error) {
    alert(getApiErrorMessage(error, '删除任务失败'));
  } finally {
    deletingId.value = null;
  }
};

// 开始编辑任务
const startEdit = async (task) => {
  if (!canUpdateTaskStatus(task)) {
    updateError.value = '无权限编辑该任务';
    return;
  }

  updateError.value = '';
  editingTask.id = task.id;
  editingTask.title = task.title;
  editingTask.description = task.description || '';
  editingTask.status = task.status;
  editingTask.priority = String(task.priority);
  editingTask.deadline = task.deadline ? formatDateForInput(task.deadline) : '';
  editingTask.team_id = isPersonalTask(task) ? null : task.team_id;
  editingTask.predecessor_ids = []; // 清空重置
  editingDependencies.value = [];   // 清空重置

  // 如果是个人任务，从后端拉取现有前置依赖用于回显
  if (isPersonalTask(task)) {
    try {
      const res = await getTaskDependenciesApi(task.id);
      if (res.data && res.data.code === 0) {
        const preds = res.data.data.predecessors || [];
        editingDependencies.value = preds.map(p => p.id);
        editingTask.predecessor_ids = [...editingDependencies.value]; // 绑定给多选框
      }
    } catch (err) {
      updateError.value = getApiErrorMessage(err, '获取任务依赖失败');
      console.error(updateError.value, err);
    }
  }
};

// 取消编辑
const cancelEdit = () => {
  editingTask.id = null;
  editingTask.title = '';
  editingTask.description = '';
  editingTask.status = 'TODO';
  editingTask.priority = '2';
  editingTask.deadline = '';
  editingTask.team_id = null;
  editingTask.predecessor_ids = [];
  editingDependencies.value = [];
};



// 格式化日期显示
const parseDateInput = (value) => {
  if (!value) return null;
  const normalized = String(value).trim().replace(/\//g, '-');
  const match = normalized.match(/^(\d{4})-(\d{2})-(\d{2})(?:T(\d{2}):(\d{2}))?$/);
  if (!match) return null;
  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3]);
  const hour = match[4] ? Number(match[4]) : 0;
  const minute = match[5] ? Number(match[5]) : 0;
  if (!year || !month || !day) return null;
  return { year, month, day, hour, minute };
};

const toBackendDeadline = (value) => {
  const parts = parseDateInput(value);
  if (!parts) return null;
  const date = new Date(Date.UTC(parts.year, parts.month - 1, parts.day, parts.hour, parts.minute, 0));
  return date.toISOString();
};

// 格式化日期显示为 YYYY/MM/DD
const formatDate = (dateStr) => {
  if (!dateStr) return '-';
  const date = new Date(dateStr);
  if (Number.isNaN(date.getTime())) return '-';
  const year = date.getUTCFullYear();
  const month = String(date.getUTCMonth() + 1).padStart(2, '0');
  const day = String(date.getUTCDate()).padStart(2, '0');
  const hour = String(date.getUTCHours()).padStart(2, '0');
  const minute = String(date.getUTCMinutes()).padStart(2, '0');
  return `${year}年${month}月${day}日 ${hour}:${minute}`;
};

// 格式化日期为输入框显示
const formatDateForInput = (dateStr) => {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  if (Number.isNaN(date.getTime())) return '';
  const year = date.getUTCFullYear();
  const month = String(date.getUTCMonth() + 1).padStart(2, '0');
  const day = String(date.getUTCDate()).padStart(2, '0');
  const hour = String(date.getUTCHours()).padStart(2, '0');
  const minute = String(date.getUTCMinutes()).padStart(2, '0');
  return `${year}-${month}-${day}T${hour}:${minute}`;
};

// 状态标签
const statusLabel = (status) => {
  const map = {
    TODO: '待做',
    DOING: '进行中',
    DONE: '已完成',
  };
  return map[status] || status;
};

// 优先级标签
const priorityLabel = (priority) => {
  const map = { 1: '低', 2: '中', 3: '高' };
  return map[priority] || priority;
};

// 页面加载时获取任务列表
onMounted(() => {
  document.addEventListener('click', handleOutsideClick);
  fetchUserTeams(); // 👇 新增：挂载时加载团队
  loadTasks();
});

watch(sortOption, async () => {
  await loadTasks();
});

onBeforeUnmount(() => {
  document.removeEventListener('click', handleOutsideClick);
  if (celebrationTimer) {
    clearTimeout(celebrationTimer);
  }
  if (completeMoveTimer) {
    clearTimeout(completeMoveTimer);
  }
  if (arrivalTimer) {
    clearTimeout(arrivalTimer);
  }
});
</script>

<style scoped>
/* 头部 */
.task-header {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 1rem;
  margin-bottom: 2rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #e5e7eb;
}

.task-title {
  text-align: center;
}

.header-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.75rem;
  width: 100%;
}

.task-group-actions {
  display: inline-flex;
  align-items: center;
  gap: 0.75rem;
}

.fab {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: var(--primary);
  color: #fff;
  text-decoration: none;
  
  border: none;
  padding: 0;
  outline: none;
  cursor: pointer;
  
  box-shadow: 0 4px 10px rgba(28, 126, 214, 0.28);
  transition: transform 0.2s, box-shadow 0.2s;
  flex-shrink: 0;
}

.fab:hover {
  transform: scale(1.05);
  box-shadow: 0 6px 14px rgba(28, 126, 214, 0.35);
}

.plus-icon {
  font-size: 30px;
  font-weight: 600 ;
  line-height: 1;
  margin-top: -6px;
}

.sort-control {
  display: inline-flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.25rem 0.75rem 0.25rem 0.4rem;
  border-radius: 999px;
  border: 1px solid #e2e8f0;
  background: #f8fafc;
}

.sort-fab {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: 50%;
  background: var(--primary);
  color: #fff;
  box-shadow: 0 4px 10px rgba(28, 126, 214, 0.28);
  flex-shrink: 0;
}

.sort-fab svg {
  width: 16px;
  height: 16px;
}

.sort-dropdown {
  position: relative;
  min-width: 120px;
}

.sort-trigger {
  width: 100%;
  min-height: 36px;
  border: 1px solid #d7e2ef;
  border-radius: 999px;
  background: #fff;
  font-size: 0.85rem;
  color: #1f2937;
  cursor: pointer;
  padding: 6px 10px 6px 12px;
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.sort-trigger:hover {
  border-color: #bfd2ec;
}

.sort-trigger:focus-visible {
  outline: none;
  border-color: #60a5fa;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.12);
}

.sort-caret {
  width: 8px;
  height: 8px;
  border-right: 2px solid #64748b;
  border-bottom: 2px solid #64748b;
  transform: rotate(45deg);
  transition: transform 0.18s ease;
}

.sort-caret.open {
  transform: rotate(225deg);
}

.sort-menu {
  list-style: none;
  margin: 6px 0 0;
  padding: 6px;
  position: absolute;
  left: 0;
  right: 0;
  z-index: 25;
  border: 1px solid #d5e1f0;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.98);
  box-shadow: 0 12px 24px rgba(15, 23, 42, 0.14);
  backdrop-filter: blur(6px);
}

.sort-option {
  width: 100%;
  min-height: 34px;
  border: 1px solid transparent;
  border-radius: 8px;
  background: transparent;
  color: #1f2937;
  font-size: 0.82rem;
  font-weight: 600;
  text-align: left;
  cursor: pointer;
  padding: 0 8px;
  transition: background-color 0.18s ease, border-color 0.18s ease;
}

.sort-option:hover {
  background: #f3f8ff;
  border-color: #dbeafe;
}

.sort-option.active {
  background: #eaf2ff;
  border-color: #c7ddff;
  color: #1d4ed8;
}

.sort-menu-fade-enter-active,
.sort-menu-fade-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.sort-menu-fade-enter-from,
.sort-menu-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

.search-box {
  display: inline-flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.25rem 0.75rem 0.25rem 0.4rem;
  border-radius: 999px;
  border: 1px solid #e2e8f0;
  background: #f8fafc;
  flex: 1;
}

.search-fab {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: var(--primary);
  color: #fff;
  box-shadow: 0 4px 10px rgba(28, 126, 214, 0.28);
  flex-shrink: 0;
}

.search-fab svg {
  width: 18px;
  height: 18px;
}

.search-input {
  border: none;
  background: transparent;
  width: 100%;
  font-size: 0.85rem;
  color: #1f2937;
}

.search-input:focus {
  outline: none;
}

.search-clear {
  border: none;
  background: transparent;
  color: #64748b;
  font-size: 0.75rem;
  cursor: pointer;
  padding: 0.2rem 0.4rem;
}

.search-clear:hover {
  color: #0f172a;
}

.task-header h1 {
  margin: 0 0 0.5rem 0;
  font-size: 1.5rem;
  color: var(--text);
}

.task-header .hint {
  margin: 0;
  color: var(--muted);
  font-size: 0.875rem;
}

/* 表单部分 */
.form-section {
  background: #f9fafb;
  border-radius: 0.5rem;
  padding: 1.5rem;
  margin-bottom: 2rem;
}

.form-section h2 {
  margin: 0 0 1rem 0;
  font-size: 1rem;
  color: #374151;
  font-weight: 600;
}

.task-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-input {
  padding: 0.5rem 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  font-family: inherit;
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-textarea {
  resize: vertical;
  min-height: 80px;
  font-family: inherit;
}

/* 列表部分 */
.list-section {
  margin-bottom: 2rem;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.list-header h2 {
  margin: 0;
  font-size: 1rem;
  color: #374151;
  font-weight: 600;
}

/* 空/加载状态 */
.empty-state,
.loading-state {
  text-align: center;
  padding: 2rem;
  color: #6b7280;
  background: #f9fafb;
  border-radius: 0.5rem;
}

/* 任务列表 */
.task-list {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.task-group {
  padding: 1rem;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
}

.task-group-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 1rem;
}

.task-group-header h3,
.history-toggle h3 {
  margin: 0;
  font-size: 1rem;
  color: #1f2937;
}

.task-group-header p,
.history-toggle p {
  margin: 0.35rem 0 0;
  color: #64748b;
  font-size: 0.85rem;
}

.task-group-count {
  min-width: 2.25rem;
  height: 2.25rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0 0.75rem;
  border-radius: 999px;
  background: #dbeafe;
  color: #1d4ed8;
  font-weight: 700;
  font-size: 0.85rem;
}

.task-group-count.done {
  background: #dcfce7;
  color: #15803d;
}

.task-group-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.history-toggle {
  width: 100%;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 0;
  margin-bottom: 1rem;
  border: none;
  background: transparent;
  cursor: pointer;
  text-align: left;
}

.history-toggle-meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.history-toggle-text {
  color: #0f766e;
  font-size: 0.85rem;
  font-weight: 600;
}

.task-item-done {
  opacity: 0.92;
}

.sub-empty-state {
  padding: 1rem;
  text-align: center;
  color: #6b7280;
  background: #ffffff;
  border: 1px dashed #dbe3f0;
  border-radius: 0.5rem;
}

.task-item {
  position: relative;
  overflow: hidden;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 1rem;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
  transition: transform 0.22s ease, box-shadow 0.22s ease, border-color 0.22s ease;
}

.task-item-celebrating {
  border-color: #86efac;
  box-shadow: 0 16px 36px rgba(34, 197, 94, 0.22);
}

.task-item-completing {
  animation: task-complete-exit 0.95s cubic-bezier(0.22, 1, 0.36, 1) forwards;
}

.task-item-arrived {
  animation: task-complete-arrive 0.7s cubic-bezier(0.22, 1, 0.36, 1);
}

.task-item:hover {
  transform: translateY(-4px);
  border-color: #bfdbfe;
  box-shadow: 0 12px 24px rgba(15, 23, 42, 0.12);
}

.confetti-burst {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
}

.confetti-piece {
  --x-start: calc(8% + var(--piece-index) * 6%);
  position: absolute;
  left: var(--x-start);
  top: 52%;
  width: 10px;
  height: 18px;
  border-radius: 999px;
  opacity: 0;
  transform: translate3d(0, 0, 0) rotate(0deg) scale(0.8);
  animation: confetti-pop 1.2s ease-out forwards;
  animation-delay: calc(var(--piece-index) * 0.03s);
}

.confetti-piece:nth-child(4n + 1) {
  background: #f97316;
}

.confetti-piece:nth-child(4n + 2) {
  background: #22c55e;
}

.confetti-piece:nth-child(4n + 3) {
  background: #38bdf8;
}

.confetti-piece:nth-child(4n) {
  background: #facc15;
}

.task-main {
  flex: 1;
  min-width: 0;
}

.task-title-section {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
  flex-wrap: wrap;
}

.task-title-section h3 {
  margin: 0;
  font-size: 0.95rem;
  color: #1f2937;
  font-weight: 600;
}

.status-badge,
.priority-badge {
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: 500;
}

.status-todo {
  background: #dbeafe;
  color: #1e40af;
}

.status-doing {
  background: #fef08a;
  color: #92400e;
}

.status-done {
  background: #dcfce7;
  color: #166534;
}

.priority-1 {
  background: #e0e7ff;
  color: #4338ca;
}

.priority-2 {
  background: #fce7f3;
  color: #be456f;
}

.priority-3 {
  background: #fee2e2;
  color: #b91c1c;
}

.team-link-badge {
  background: #f1f5f9;
  color: #475569;
  border: 1px solid #cbd5e1;
  text-decoration: none;
  cursor: pointer;
  transition: background-color 0.18s ease, border-color 0.18s ease, color 0.18s ease;
}

.team-link-badge:hover {
  background: #eff6ff;
  border-color: #93c5fd;
  color: #1d4ed8;
}

.task-description {
  margin: 0.5rem 0;
  color: #6b7280;
  font-size: 0.875rem;
  line-height: 1.4;
}

.task-meta {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
  margin-top: 0.75rem;
  font-size: 0.8rem;
  color: #9ca3af;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.meta-label {
  color: #64748b;
  font-size: 0.85rem;
  min-width: 72px;
  display: inline-block;
}

/* 任务操作 */
.task-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-left: 1rem;
  flex-shrink: 0;
}

.btn-small {
  padding: 0.375rem 0.75rem;
  font-size: 0.75rem;
  white-space: nowrap;
}

.btn-edit {
  background: #dbeafe;
  color: #1e40af;
  border: 1px solid #93c5fd;
}

.btn-edit:hover:not(:disabled) {
  background: #bfdbfe;
}

.btn-complete {
  position: relative;
  overflow: hidden;
  background: linear-gradient(135deg, #22c55e, #16a34a);
  color: #fff;
  border: 1px solid rgba(34, 197, 94, 0.5);
  box-shadow: 0 10px 18px rgba(34, 197, 94, 0.2);
  transition: transform 0.22s ease, box-shadow 0.22s ease, filter 0.22s ease;
}

.btn-complete::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(120deg, transparent 0%, rgba(255, 255, 255, 0.28) 45%, transparent 100%);
  transform: translateX(-120%);
  transition: transform 0.45s ease;
}

.btn-complete:hover:not(:disabled) {
  transform: translateY(-2px) scale(1.03);
  filter: saturate(1.08);
  box-shadow: 0 14px 24px rgba(34, 197, 94, 0.28);
}

.btn-complete:hover:not(:disabled)::before {
  transform: translateX(120%);
}

.btn-dependency {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 42px;
  background: #fff;
  color: #1e40af;
  border: 1px solid #93c5fd;
  text-align: center;
  text-decoration: none;
  font-weight: 600;
  line-height: 1;
  transition: background-color 0.22s ease, border-color 0.22s ease, color 0.22s ease;
}

.btn-dependency:hover {
  background: #eff6ff;
  border-color: #60a5fa;
  color: #1d4ed8;
}

.btn-danger {
  background: #fee2e2;
  color: #b91c1c;
  border: 1px solid #fecaca;
}

.btn-danger:hover:not(:disabled) {
  background: #fecaca;
}

/* 分页 */
.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid #e5e7eb;
}

.page-info {
  font-size: 0.875rem;
  color: #6b7280;
}

/* 模态框 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal {
  background: white;
  border-radius: 0.5rem;
  padding: 2rem;
  max-width: 500px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
}

.modal h2 {
  margin: 0 0 1.5rem 0;
  font-size: 1.25rem;
  color: #1f2937;
}

.modal-actions {
  display: flex;
  gap: 1rem;
  margin-top: 1.5rem;
}

.modal-actions button {
  flex: 1;
}

/* 错误消息 */
.error-msg {
  margin: 0.5rem 0 0;
  padding: 0.5rem;
  background: #fee2e2;
  color: #991b1b;
  border-radius: 0.375rem;
  font-size: 0.875rem;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .header-actions {
    width: 100%;
    flex-wrap: wrap;
    justify-content: flex-end;
  }

  .search-box {
    flex: 1;
  }

  .search-input {
    width: 100%;
  }

  .task-group-header,
  .history-toggle {
    flex-direction: column;
    align-items: flex-start;
  }

  .task-group-actions {
    width: 100%;
    justify-content: space-between;
  }

  .history-toggle-meta {
    width: 100%;
    justify-content: space-between;
  }

  .form-row {
    grid-template-columns: 1fr;
  }

  .task-item {
    flex-direction: column;
    gap: 1rem;
  }

  .task-actions {
    margin-left: 0;
    width: 100%;
  }

  .task-actions button {
    flex: 1;
  }

  .pagination {
    flex-wrap: wrap;
  }
}

@keyframes confetti-pop {
  0% {
    opacity: 0;
    transform: translate3d(0, 0, 0) rotate(0deg) scale(0.8);
  }

  15% {
    opacity: 1;
  }

  100% {
    opacity: 0;
    transform:
      translate3d(
        calc((var(--piece-index) - 6) * 12px),
        calc(-90px - var(--piece-index) * 6px),
        0
      )
      rotate(calc((var(--piece-index) - 4) * 28deg))
      scale(1);
  }
}

@keyframes task-complete-exit {
  0% {
    opacity: 1;
    transform: translate3d(0, 0, 0) scale(1);
  }

  45% {
    opacity: 1;
    transform: translate3d(10px, -10px, 0) scale(1.02);
  }

  100% {
    opacity: 0;
    transform: translate3d(30px, -18px, 0) scale(0.96);
  }
}

@keyframes task-complete-arrive {
  0% {
    opacity: 0;
    transform: translate3d(0, 18px, 0) scale(0.96);
  }

  100% {
    opacity: 0.92;
    transform: translate3d(0, 0, 0) scale(1);
  }
}
</style>
