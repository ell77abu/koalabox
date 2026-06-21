<template>
  <section class="trash-panel">
    <header class="trash-heading">
      <div>
        <h2>{{ title }}</h2>
        <p>{{ subtitle }}</p>
      </div>
      <button class="btn btn-outline btn-small" type="button" :disabled="loading" @click="loadTrash">
        {{ loading ? '刷新中...' : '刷新' }}
      </button>
    </header>

    <p v-if="loadError" class="error-banner">{{ loadError }}</p>

    <div v-if="loading" class="loading-state">加载回收站中...</div>

    <div v-else-if="tasks.length === 0" class="empty-state">
      当前没有已删除任务
    </div>

    <div v-else class="trash-list">
      <article v-for="task in tasks" :key="task.id" class="trash-item">
        <div class="trash-main">
          <div class="task-title-section">
            <h3>{{ task.title || `任务 #${task.id}` }}</h3>
            <span class="status-badge" :class="`status-${String(task.status || 'TODO').toLowerCase()}`">
              {{ statusLabel(task.status) }}
            </span>
            <span class="priority-badge" :class="`priority-${task.priority || 2}`">
              {{ priorityLabel(task.priority) }}
            </span>
            <span class="deleted-badge">已删除</span>
          </div>

          <p v-if="task.description" class="task-description">{{ task.description }}</p>

          <div class="trash-meta">
            <span>删除时间：{{ formatDate(task.deleted_at) }}</span>
            <span v-if="task.deadline">截止时间：{{ formatDate(task.deadline) }}</span>
            <span>#{{ task.id }}</span>
          </div>
        </div>

        <div class="trash-actions">
          <RouterLink class="btn btn-small btn-dependency" :to="buildLogRoute(task)">
            查看日志
          </RouterLink>
          <button
            v-if="canRestore"
            class="btn btn-small btn-complete"
            type="button"
            :disabled="restoringId === task.id"
            @click="confirmRestore(task)"
          >
            {{ restoringId === task.id ? '恢复中...' : '恢复' }}
          </button>
          <span v-else class="permission-note">仅团队管理员可恢复</span>
        </div>
      </article>

      <div v-if="totalPages > 1" class="pagination">
        <button class="btn btn-outline btn-small" type="button" :disabled="page === 1" @click="page--">上一页</button>
        <span class="page-info">第 {{ page }} 页 / 共 {{ totalPages }} 页</span>
        <button class="btn btn-outline btn-small" type="button" :disabled="page === totalPages" @click="page++">下一页</button>
      </div>
    </div>

    <MessageModal v-model="messageOpen" :title="messageTitle" :message="messageText" />
    <ConfirmModal
      v-model="confirmOpen"
      :title="confirmTitle"
      :message="confirmText"
      @confirm="runConfirmAction"
    />
  </section>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import { RouterLink } from 'vue-router';
import ConfirmModal from './ConfirmModal.vue';
import MessageModal from './MessageModal.vue';
import { getTaskTrashApi, restoreTaskApi } from '../services/task';
import { getApiErrorMessage, getApiFailureMessage } from '../utils/apiError';

const props = defineProps({
  teamId: {
    type: [Number, String],
    default: null,
  },
  canRestore: {
    type: Boolean,
    default: true,
  },
  title: {
    type: String,
    default: '回收站',
  },
  subtitle: {
    type: String,
    default: '查看已软删除的任务，并在需要时恢复。',
  },
});

const emit = defineEmits(['restored']);

const tasks = ref([]);
const page = ref(1);
const size = ref(10);
const total = ref(0);
const loading = ref(false);
const loadError = ref('');
const restoringId = ref(null);
const messageOpen = ref(false);
const messageTitle = ref('提示');
const messageText = ref('');
const confirmOpen = ref(false);
const confirmTitle = ref('确认恢复');
const confirmText = ref('');
let confirmAction = null;

const normalizedTeamId = computed(() => {
  const parsed = Number(props.teamId);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
});

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / size.value)));

const statusLabel = (status) => ({ TODO: '待做', DOING: '进行中', DONE: '已完成' }[status] || status || '未知');
const priorityLabel = (priority) => ({ 1: '低', 2: '中', 3: '高' }[Number(priority)] || priority || '-');

const formatDate = (value) => {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return String(value);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
};

const showMessage = (message, title = '提示') => {
  messageTitle.value = title;
  messageText.value = message;
  messageOpen.value = true;
};

const showConfirm = (message, action, title = '确认恢复') => {
  confirmTitle.value = title;
  confirmText.value = message;
  confirmAction = action;
  confirmOpen.value = true;
};

const runConfirmAction = async () => {
  const action = confirmAction;
  confirmAction = null;
  if (typeof action === 'function') {
    await action();
  }
};

const buildLogRoute = (task) => {
  const query = { deleted: '1' };
  if (task.title) query.title = task.title;
  const teamId = Number(task.team_id) || normalizedTeamId.value;
  if (teamId) query.team_id = teamId;
  return { path: `/tasks/${task.id}`, query };
};

const loadTrash = async () => {
  loading.value = true;
  loadError.value = '';
  try {
    const res = await getTaskTrashApi({
      page: page.value,
      size: size.value,
      team_id: normalizedTeamId.value,
    });
    if (res.data?.code === 0) {
      const data = res.data.data || {};
      tasks.value = Array.isArray(data.list) ? data.list : [];
      total.value = Number(data.total) || tasks.value.length;
      const fixedPage = Number(data.page);
      if (Number.isFinite(fixedPage) && fixedPage > 0) page.value = fixedPage;
    } else {
      tasks.value = [];
      total.value = 0;
      loadError.value = getApiFailureMessage(res, '加载回收站失败');
    }
  } catch (error) {
    tasks.value = [];
    total.value = 0;
    loadError.value = getApiErrorMessage(error, '加载回收站失败');
  } finally {
    loading.value = false;
  }
};

const confirmRestore = (task) => {
  showConfirm(`确定要恢复任务「${task.title || `#${task.id}`}」吗？`, async () => {
    restoringId.value = task.id;
    try {
      const res = await restoreTaskApi(task.id);
      if (res.status === 200 && (res.data?.code === undefined || res.data.code === 0)) {
        showMessage('任务已恢复');
        await loadTrash();
        emit('restored', task);
      } else {
        showMessage(getApiFailureMessage(res, '恢复任务失败', '恢复失败'), '操作失败');
      }
    } catch (error) {
      showMessage(getApiErrorMessage(error, '恢复任务失败'), '操作失败');
    } finally {
      restoringId.value = null;
    }
  });
};

watch(page, loadTrash);
watch(normalizedTeamId, () => {
  page.value = 1;
  loadTrash();
});

onMounted(loadTrash);
</script>

<style scoped>
.trash-panel {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.trash-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.trash-heading h2 {
  margin: 0;
  color: var(--text);
  font-size: 1.1rem;
  font-weight: 700;
}

.trash-heading p {
  margin: 0.35rem 0 0;
  color: var(--muted);
  font-size: 0.85rem;
}

.trash-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.trash-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.2rem;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #fff;
  transition: transform 0.2s, box-shadow 0.2s;
}

.trash-item:hover {
  transform: translateY(-2px);
  border-color: #cbd5e1;
  box-shadow: 0 10px 20px rgba(15, 23, 42, 0.08);
}

.trash-main {
  flex: 1;
  min-width: 0;
}

.task-title-section {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.6rem;
  margin-bottom: 0.45rem;
}

.task-title-section h3 {
  margin: 0;
  color: #0f172a;
  font-size: 1.05rem;
  font-weight: 700;
}

.task-description {
  margin: 0 0 0.55rem;
  color: #64748b;
  font-size: 0.9rem;
  line-height: 1.55;
}

.trash-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  color: #64748b;
  font-size: 0.8rem;
}

.trash-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.6rem;
  flex-wrap: wrap;
  min-width: 180px;
}

.permission-note {
  color: #94a3b8;
  font-size: 0.78rem;
  font-weight: 700;
}

.deleted-badge {
  padding: 0.2rem 0.5rem;
  border: 1px solid #fecaca;
  border-radius: 6px;
  background: #fef2f2;
  color: #b91c1c;
  font-size: 0.75rem;
  font-weight: 700;
}

.priority-badge,
.status-badge {
  padding: 0.2rem 0.5rem;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 700;
}

.priority-1 { background: #e0e7ff; color: #4338ca; }
.priority-2 { background: #fef08a; color: #92400e; }
.priority-3 { background: #fee2e2; color: #b91c1c; }
.status-todo { background: #dbeafe; color: #1e40af; }
.status-doing { background: #fef08a; color: #92400e; }
.status-done { background: #dcfce7; color: #166534; }

.btn {
  min-height: 34px;
  padding: 0.45rem 0.8rem;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.82rem;
  font-weight: 700;
  text-decoration: none;
}

.btn:disabled {
  cursor: not-allowed;
  opacity: 0.65;
}

.btn-small {
  min-height: 34px;
}

.btn-outline {
  background: #f1f5f9;
  color: #475569;
  border: 1px solid #cbd5e1;
}

.btn-outline:hover:not(:disabled) {
  background: #e2e8f0;
}

.btn-complete {
  background: linear-gradient(135deg, #22c55e, #16a34a);
  color: #fff;
  box-shadow: 0 4px 10px rgba(34, 197, 94, 0.2);
}

.btn-dependency {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  color: #1e40af;
  border: 1px solid #93c5fd;
}

.btn-dependency:hover {
  background: #eff6ff;
  border-color: #60a5fa;
}

.empty-state,
.loading-state {
  padding: 3rem;
  border: 1px dashed #cbd5e1;
  border-radius: 8px;
  background: #f8fafc;
  color: #6b7280;
  text-align: center;
  font-weight: 600;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  margin-top: 0.5rem;
  padding-top: 1.2rem;
  border-top: 1px solid #e5e7eb;
}

.page-info {
  color: #6b7280;
  font-size: 0.875rem;
  font-weight: 600;
}

:global([data-theme="dark"]) .trash-item,
:global([data-theme="dark"]) .btn-dependency {
  background: #1e293b;
  border-color: #334155;
}

:global([data-theme="dark"]) .task-title-section h3 {
  color: #f8fafc;
}

:global([data-theme="dark"]) .task-description,
:global([data-theme="dark"]) .trash-meta,
:global([data-theme="dark"]) .permission-note {
  color: #94a3b8;
}

:global([data-theme="dark"]) .empty-state,
:global([data-theme="dark"]) .loading-state {
  background: #0f172a;
  color: #94a3b8;
  border-color: #334155;
}

@media (max-width: 720px) {
  .trash-heading,
  .trash-item {
    align-items: stretch;
    flex-direction: column;
  }

  .trash-actions {
    justify-content: flex-start;
    min-width: 0;
  }
}
</style>
