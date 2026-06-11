<template>
  <section class="log-panel">
    <div class="panel-heading">
      <div>
        <h2>操作日志</h2>
        <span>最近 {{ logs.length }} 条记录</span>
      </div>
      <button class="refresh-btn" type="button" :disabled="loading" @click="loadLogs">
        {{ loading ? '加载中...' : '刷新' }}
      </button>
    </div>

    <div v-if="loading" class="log-state">日志加载中...</div>

    <div v-else-if="error" class="log-state log-error">
      {{ error }}
    </div>

    <div v-else-if="logs.length === 0" class="log-state">
      暂无操作记录
    </div>

    <ol v-else class="log-list">
      <li v-for="log in logs" :key="log.id" class="log-item">
        <div class="log-dot" aria-hidden="true"></div>
        <div class="log-content">
          <p>
            <strong>{{ log.actor_name || '未知用户' }}</strong>
            在 {{ formatLogTime(log.created_at) }}
            {{ log.content || actionLabel(log.action_type) }}
          </p>
          <span>{{ actionLabel(log.action_type) }}</span>
        </div>
      </li>
    </ol>
  </section>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue';
import { getTaskLogsApi } from '../services/task';
import { getApiErrorMessage, getApiFailureMessage } from '../utils/apiError';

const props = defineProps({
  taskId: {
    type: Number,
    required: true,
  },
});

const logs = ref([]);
const loading = ref(false);
const error = ref('');

const actionLabel = (type) => {
  const labels = {
    CREATE_TASK: '创建任务',
    UPDATE_TASK_INFO: '修改任务信息',
    DELETE_TASK: '删除任务',
    TASK_ASSIGNEE_CHANGED_BY_MEMBER_LEAVE: '成员离开导致任务归属变化',
  };
  return labels[type] || type || '操作记录';
};

const formatLogTime = (value) => {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return String(value);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  const hour = String(date.getHours()).padStart(2, '0');
  const minute = String(date.getMinutes()).padStart(2, '0');
  return `${year}-${month}-${day} ${hour}:${minute}`;
};

const loadLogs = async () => {
  if (!Number.isFinite(props.taskId) || props.taskId <= 0) return;

  loading.value = true;
  error.value = '';
  try {
    const res = await getTaskLogsApi(props.taskId, { page: 1, size: 10 });
    if (res.data.code === 0) {
      logs.value = res.data.data?.list || [];
    } else {
      logs.value = [];
      error.value = getApiFailureMessage(res, '获取任务日志失败');
    }
  } catch (err) {
    logs.value = [];
    error.value = getApiErrorMessage(err, '获取任务日志失败');
  } finally {
    loading.value = false;
  }
};

onMounted(loadLogs);
watch(() => props.taskId, loadLogs);
</script>

<style scoped>
.log-panel {
  display: flex;
  flex-direction: column;
  min-height: 180px;
  padding: 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #ffffff;
}

.panel-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 0.9rem;
}

.panel-heading h2 {
  margin: 0;
  color: #0f172a;
  font-size: 1rem;
}

.panel-heading span {
  display: inline-block;
  margin-top: 0.35rem;
  color: #64748b;
  font-size: 0.82rem;
  font-weight: 700;
}

.refresh-btn {
  min-height: 34px;
  padding: 0 12px;
  border: 1px solid #93c5fd;
  border-radius: 8px;
  background: #fff;
  color: #1d4ed8;
  cursor: pointer;
  font-size: 0.82rem;
  font-weight: 700;
}

.refresh-btn:hover:not(:disabled) {
  background: #eff6ff;
  border-color: #60a5fa;
}

.refresh-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.log-state {
  padding: 2rem 1rem;
  text-align: center;
  color: #64748b;
  background: #f8fafc;
  border: 1px dashed #cbd5e1;
  border-radius: 8px;
}

.log-error {
  color: #991b1b;
  background: #fef2f2;
  border-color: #fecaca;
}

.log-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.log-item {
  display: grid;
  grid-template-columns: 14px 1fr;
  gap: 0.75rem;
  align-items: start;
}

.log-dot {
  width: 10px;
  height: 10px;
  margin-top: 0.45rem;
  border-radius: 999px;
  background: #2563eb;
  box-shadow: 0 0 0 4px #dbeafe;
}

.log-content {
  padding: 0.85rem 1rem;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.log-content p {
  margin: 0;
  color: #0f172a;
  font-size: 0.9rem;
  line-height: 1.6;
}

.log-content span {
  display: inline-block;
  margin-top: 0.35rem;
  color: #64748b;
  font-size: 0.76rem;
}

:global([data-theme="dark"]) .log-panel,
:global([data-theme="dark"]) .log-content,
:global([data-theme="dark"]) .log-state {
  background: #0f172a;
  border-color: #334155;
}

:global([data-theme="dark"]) .panel-heading h2,
:global([data-theme="dark"]) .log-content p {
  color: #f8fafc;
}

:global([data-theme="dark"]) .panel-heading span,
:global([data-theme="dark"]) .log-content span,
:global([data-theme="dark"]) .log-state {
  color: #94a3b8;
}
</style>
