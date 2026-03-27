<template>
  <div class="page-shell">
    <section class="card card-wide">
      <!-- 页头 -->
      <header class="task-header">
        <div>
          <h1>任务管理</h1>
          <p class="hint">欢迎你，{{ authStore.username.value }}</p>
        </div>
        <button class="btn btn-outline" @click="logout">退出登录</button>
      </header>

      <!-- 创建任务表单 -->
      <div class="form-section">
        <h2>创建新任务</h2>
        <form @submit.prevent="onCreateTask" class="task-form">
          <div class="form-row">
            <input
              v-model.trim="newTask.title"
              type="text"
              placeholder="任务标题（必填）"
              required
              maxlength="100"
              class="form-input"
            />
            <select v-model="newTask.priority" class="form-input">
              <option value="1">优先级：低</option>
              <option value="2">优先级：中</option>
              <option value="3">优先级：高</option>
            </select>
          </div>

          <textarea
            v-model.trim="newTask.description"
            placeholder="任务描述（可选）"
            class="form-input form-textarea"
            maxlength="500"
          ></textarea>

          <div class="form-row">
            <input v-model="newTask.deadline" type="datetime-local" class="form-input" />
            <button type="submit" class="btn btn-primary" :disabled="creatingTask">
              {{ creatingTask ? '创建中...' : '创建任务' }}
            </button>
          </div>

          <p v-if="createError" class="error-msg">{{ createError }}</p>
        </form>
      </div>

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

        <!-- 错误状态 -->
        <div v-else-if="listError" class="empty-state">
          <p>{{ listError }}</p>
        </div>

        <!-- 空状态 -->
        <div v-else-if="taskStore.tasks.value.length === 0" class="empty-state">
          <p>暂无任务，创建一个新任务开始吧！</p>
        </div>

        <!-- 任务列表 -->
        <div v-else class="task-list">
          <div v-for="task in taskStore.tasks.value" :key="task.id" class="task-item">
            <div class="task-main">
              <div class="task-title-section">
                <h3>{{ task.title }}</h3>
                <span class="status-badge" :class="`status-${(task.status || 'TODO').toLowerCase()}`">
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
import { ref, reactive, onMounted } from 'vue';
import { useRouter } from 'vue-router';
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
const listError = ref('');

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
  listError.value = '';
  try {
    const response = await getTasksApi(taskStore.currentPage.value, taskStore.pageSize.value);
    if (response.data.code === 0) {
      const pageData = response.data.data || {};
      const list = Array.isArray(pageData.list) ? pageData.list : [];
      const total = Number(pageData.total || 0);
      const currentPage = Number(pageData.page || taskStore.currentPage.value);
      const pageSize = Number(pageData.size || taskStore.pageSize.value);

      taskStore.setTasks(list);
      taskStore.setPageInfo(
        currentPage,
        pageSize,
        total
      );
    } else {
      listError.value = response.data.message || '加载任务失败';
    }
  } catch (error) {
    listError.value = error.response?.data?.message || '加载任务失败，请检查登录状态或后端服务';
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
      deadline: formatDateTimeForBackend(newTask.deadline),
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
      deadline: formatDateTimeForBackend(editingTask.deadline),
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

// 转换 datetime-local 值到后端约定格式：YYYY-MM-DD HH:mm:ss
const formatDateTimeForBackend = (inputValue) => {
  if (!inputValue) return null;
  const normalized = inputValue.replace('T', ' ');
  return normalized.length === 16 ? `${normalized}:00` : normalized;
};

// 状态标签
const statusLabel = (status) => {
  const map = {
    TODO: '待做',
    DOING: '进行中',
    DONE: '已完成',
    ERROR: '异常',
  };
  return map[status] || status || '待做';
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
  gap: 1rem;
}

.task-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 1rem;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
  transition: box-shadow 0.2s;
}

.task-item:hover {
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
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
</style>
