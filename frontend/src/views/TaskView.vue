<template>
  <div class="page-shell">
    <section class="card card-wide">
      <!-- 页头 -->
      <header class="task-header">
        <div>
          <h1>任务管理</h1>
          <p class="hint">欢迎你，{{ authStore.username.value }}</p>
        </div>
        <div class="header-actions">
          <RouterLink to="/tasks/create" class="add-task-fab" title="创建任务">
            <span class="plus-icon">+</span>
          </RouterLink>
        </div>
      </header>

      <!-- 任务列表 -->
      <div class="list-section">
        <div class="list-header">
          <h2>任务列表</h2>
          <p v-if="!loading && taskStore.tasks.value.length > 0" class="hint">
            共 {{ taskStore.total.value }} 个任务
          </p>
        </div>

        <!-- 加载状态 -->
        <div v-if="loading" class="loading-state">
          <p>加载中...</p>
        </div>

        <!-- 空状态 -->
        <div v-else-if="taskStore.tasks.value.length === 0" class="empty-state">
          <p>暂无任务，创建一个新任务开始吧！</p>
        </div>

        <!-- 任务列表 -->
        <div v-else class="task-list">
          <section class="task-group">
            <div class="task-group-header">
              <div>
                <h3>待办任务</h3>
                <!-- <p>优先展示待做和进行中的任务</p> -->
              </div>
              <span class="task-group-count">{{ incompleteTasks.length }}</span>
            </div>

            <div v-if="incompleteTasks.length" class="task-group-list">
              <div
                v-for="task in incompleteTasks"
                :key="task.id"
                class="task-item"
                :class="{ 'task-item-celebrating': celebrationTaskId === task.id }"
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
                      📅 {{ formatDate(task.deadline) }}
                    </span>
                    <span class="meta-item">
                      🕐 {{ formatDate(task.created_at) }}
                    </span>
                  </div>
                </div>

                <div class="task-actions">
                  <button
                    class="btn btn-small btn-complete"
                    @click="markTaskDone(task)"
                    :disabled="updatingId === task.id"
                  >
                    {{ updatingId === task.id ? '完成中...' : '完成任务' }}
                  </button>
                  <button
                    class="btn btn-small btn-edit"
                    @click="startEdit(task)"
                    :disabled="updatingId === task.id"
                  >
                    编辑
                  </button>
                  <button
                    class="btn btn-small btn-danger"
                    @click="onDeleteTask(task.id)"
                    :disabled="deletingId === task.id"
                  >
                    {{ deletingId === task.id ? '删除中...' : '删除' }}
                  </button>
                </div>
              </div>
            </div>

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
                <!-- <span class="history-toggle-text">{{ showCompletedHistory ? '收起' : '展开' }}</span> -->
              </div>
            </button>

            <div v-if="showCompletedHistory">
              <div v-if="completedTasks.length" class="task-group-list">
                <div
                  v-for="task in completedTasks"
                  :key="task.id"
                  class="task-item task-item-done"
                  :class="{ 'task-item-celebrating': celebrationTaskId === task.id }"
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
                        📅 {{ formatDate(task.deadline) }}
                      </span>
                      <span class="meta-item">
                        🕐 {{ formatDate(task.created_at) }}
                      </span>
                    </div>
                  </div>

                  <div class="task-actions">
                    <button
                      class="btn btn-small btn-edit"
                      @click="startEdit(task)"
                      :disabled="updatingId === task.id"
                    >
                      编辑
                    </button>
                    <button
                      class="btn btn-small btn-danger"
                      @click="onDeleteTask(task.id)"
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

        <!-- 分页 -->
        <div v-if="taskStore.tasks.value.length > 0" class="pagination">
          <button
            :disabled="taskStore.currentPage.value === 1 || loading"
            @click="onPrevPage"
            class="btn btn-small"
          >
            ← 上一页
          </button>
          <span class="page-info">
            第 {{ taskStore.currentPage.value }} 页 / 共
            {{ Math.ceil(taskStore.total.value / taskStore.pageSize.value) }} 页
          </span>
          <button
            :disabled="!taskStore.hasMore.value || loading"
            @click="onNextPage"
            class="btn btn-small"
          >
            下一页 →
          </button>
        </div>
      </div>

      <!-- 编辑模态框 -->
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
            />

            <textarea
              v-model.trim="editingTask.description"
              placeholder="任务描述"
              class="form-input form-textarea"
            ></textarea>

            <select v-model="editingTask.status" class="form-input">
              <option value="TODO">状态：待做</option>
              <option value="DOING">状态：进行中</option>
              <option value="DONE">状态：已完成</option>
            </select>

            <select v-model="editingTask.priority" class="form-input">
              <option value="1">优先级：低</option>
              <option value="2">优先级：中</option>
              <option value="3">优先级：高</option>
            </select>

            <input v-model="editingTask.deadline" type="datetime-local" class="form-input" />

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
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue';
import { useRouter, RouterLink } from 'vue-router';
import { authStore } from '../stores/auth';
import { taskStore } from '../stores/task';
import {
  getTasksApi,
  createTaskApi,
  updateTaskApi,
  deleteTaskApi,
} from '../services/task';

const router = useRouter();

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
});

// 状态管理
const loading = ref(false);
const creatingTask = ref(false);
const updatingId = ref(null);
const deletingId = ref(null);
const createError = ref('');
const updateError = ref('');
const showCompletedHistory = ref(false);
const celebrationTaskId = ref(null);
const confettiPieces = Array.from({ length: 14 }, (_, index) => index);
let celebrationTimer = null;

const incompleteTasks = computed(() =>
  taskStore.tasks.value.filter((task) => task.status !== 'DONE')
);

const completedTasks = computed(() =>
  taskStore.tasks.value.filter((task) => task.status === 'DONE')
);

// 退出登录
const logout = async () => {
  taskStore.clearTasks();
  authStore.clearAuth();
  await router.replace('/login');
};

// 加载任务列表
const loadTasks = async () => {
  loading.value = true;
  taskStore.setLoading(true);
  try {
    const response = await getTasksApi(taskStore.currentPage.value, taskStore.pageSize.value);
    if (response.data.code === 0) {
      taskStore.setTasks(response.data.data || []);
      // 假设后端返回 total 字段，若无则使用返回的数据长度
      const total = response.data.total || (response.data.data?.length || 0);
      taskStore.setPageInfo(
        taskStore.currentPage.value,
        taskStore.pageSize.value,
        total
      );
    }
  } catch (error) {
    console.error('加载任务失败:', error);
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
      deadline: newTask.deadline || null,
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
      createError.value = response.data.message || '创建失败';
    }
  } catch (error) {
    createError.value = error.response?.data?.message || '创建任务失败，请重试';
  } finally {
    creatingTask.value = false;
  }
};

// 更新任务
const onUpdateTask = async () => {
  updatingId.value = editingTask.id;
  updateError.value = '';
  try {
    const payload = {
      title: editingTask.title,
      description: editingTask.description || null,
      status: editingTask.status,
      priority: parseInt(editingTask.priority),
      deadline: editingTask.deadline || null,
    };

    const response = await updateTaskApi(editingTask.id, payload);
    if (response.data.code === 0) {
      cancelEdit();
      await loadTasks();
    } else {
      updateError.value = response.data.message || '更新失败';
    }
  } catch (error) {
    updateError.value = error.response?.data?.message || '更新任务失败，请重试';
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

const markTaskDone = async (task) => {
  if (task.status === 'DONE') {
    return;
  }

  updatingId.value = task.id;
  try {
    const payload = {
      title: task.title,
      description: task.description || null,
      status: 'DONE',
      priority: parseInt(task.priority),
      deadline: task.deadline || null,
    };

    const response = await updateTaskApi(task.id, payload);
    if (response.data.code === 0) {
      triggerCelebration(task.id);
      showCompletedHistory.value = true;
      await loadTasks();
    } else {
      alert(response.data.message || '更新失败');
    }
  } catch (error) {
    alert(error.response?.data?.message || '完成任务失败，请重试');
  } finally {
    updatingId.value = null;
  }
};

// 删除任务
const onDeleteTask = async (id) => {
  if (!window.confirm('确定要删除这个任务吗？')) {
    return;
  }

  deletingId.value = id;
  try {
    const response = await deleteTaskApi(id);
    if (response.data.code === 0) {
      await loadTasks();
    } else {
      alert(response.data.message || '删除失败');
    }
  } catch (error) {
    alert(error.response?.data?.message || '删除任务失败，请重试');
  } finally {
    deletingId.value = null;
  }
};

// 开始编辑任务
const startEdit = (task) => {
  editingTask.id = task.id;
  editingTask.title = task.title;
  editingTask.description = task.description || '';
  editingTask.status = task.status;
  editingTask.priority = String(task.priority);
  editingTask.deadline = task.deadline ? formatDateForInput(task.deadline) : '';
};

// 取消编辑
const cancelEdit = () => {
  editingTask.id = null;
  editingTask.title = '';
  editingTask.description = '';
  editingTask.status = 'TODO';
  editingTask.priority = '2';
  editingTask.deadline = '';
};

// 分页
const onPrevPage = async () => {
  if (taskStore.currentPage.value > 1) {
    taskStore.currentPage.value--;
    await loadTasks();
  }
};

const onNextPage = async () => {
  if (taskStore.hasMore.value) {
    taskStore.currentPage.value++;
    await loadTasks();
  }
};

// 格式化日期显示
const formatDate = (dateStr) => {
  if (!dateStr) return '-';
  const date = new Date(dateStr);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
};

// 格式化日期为 input[type="datetime-local"] 的格式
const formatDateForInput = (dateStr) => {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  return `${year}-${month}-${day}T${hours}:${minutes}`;
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
  loadTasks();
});

onBeforeUnmount(() => {
  if (celebrationTimer) {
    clearTimeout(celebrationTimer);
  }
});
</script>

<style scoped>
/* 头部 */
.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #e5e7eb;
}

.task-header h1 {
  margin: 0 0 0.5rem 0;
  font-size: 1.5rem;
  color: #1f2937;
}

.task-header .hint {
  margin: 0;
  color: #6b7280;
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
  .task-group-header,
  .history-toggle {
    flex-direction: column;
    align-items: flex-start;
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
</style>
