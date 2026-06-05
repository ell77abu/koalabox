<template>
  <div class="page-shell">
    <section class="detail-card">
      <header class="detail-header">
        <button class="back-btn" @click="router.back()">← 返回</button>
        <div class="header-main">
          <p class="eyebrow">任务日志</p>
          <h1>{{ taskTitleDisplay }}</h1>
        </div>
        <RouterLink class="task-list-btn" to="/tasks">任务列表</RouterLink>
      </header>

      <section class="log-panel">
        <div class="panel-heading">
          <h2>操作日志</h2>
          <div class="panel-actions">
            <span class="log-count">共 {{ logTotal }} 条</span>
            <button
              class="btn btn-outline btn-small"
              type="button"
              @click="loadLogs"
              :disabled="logsLoading"
            >
              {{ logsLoading ? '刷新中...' : '刷新' }}
            </button>
          </div>
        </div>

        <div v-if="logsLoading" class="state-panel">日志加载中...</div>

        <div v-else-if="logsError" class="state-panel error">
          <p>{{ logsError }}</p>
          <button class="btn btn-primary" type="button" @click="loadLogs">重试</button>
        </div>

        <div v-else-if="logs.length === 0" class="state-panel empty">
          暂无操作记录
        </div>

        <ul v-else class="log-list">
          <li v-for="log in logs" :key="log.id" class="log-item">
            <div class="log-line">{{ formatLogLine(log) }}</div>
            <div class="log-meta">
              <span class="log-type">{{ log.action_type || 'UNKNOWN' }}</span>
            </div>
          </li>
        </ul>

        <div v-if="logTotalPages > 1" class="pagination">
          <button
            class="btn btn-outline btn-small"
            :disabled="logPage === 1 || logsLoading"
            @click="logPage--"
          >
            上一页
          </button>
          <span class="page-info">第 {{ logPage }} 页 / 共 {{ logTotalPages }} 页</span>
          <button
            class="btn btn-outline btn-small"
            :disabled="logPage === logTotalPages || logsLoading"
            @click="logPage++"
          >
            下一页
          </button>
        </div>
      </section>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import { taskStore } from '../stores/task';
import { getTasksApi, getTaskLogsApi } from '../services/task';
import { getTeamTasksApi } from '../services/team';
import { getApiErrorMessage, getApiFailureMessage } from '../utils/apiError';

const route = useRoute();
const router = useRouter();

const logs = ref([]);
const logsLoading = ref(false);
const logsError = ref('');
const logPage = ref(1);
const logSize = ref(10);
const logTotal = ref(0);
const taskTitle = ref('');

const taskId = computed(() => Number(route.params.id));
const teamIdFromQuery = computed(() => {
  const raw = route.query.team_id;
  if (raw === undefined || raw === null || raw === '') return null;
  const parsed = Number(raw);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
});
const taskTitleDisplay = computed(() => taskTitle.value || `任务 #${taskId.value}`);

const logTotalPages = computed(() => Math.ceil(logTotal.value / logSize.value) || 1);

const formatLogTime = (time) => {
  if (!time) return '-';
  const date = new Date(time.replace(' ', 'T') + 'Z');
  if (Number.isNaN(date.getTime())) return String(time);
  const year = date.getUTCFullYear();
  const month = String(date.getUTCMonth() + 1).padStart(2, '0');
  const day = String(date.getUTCDate()).padStart(2, '0');
  const hour = String(date.getUTCHours()).padStart(2, '0');
  const minute = String(date.getUTCMinutes()).padStart(2, '0');
  return `${year}-${month}-${day} ${hour}:${minute}`;
};

const formatLogLine = (log) => {
  const actor = log?.actor_name || '未知用户';
  const time = formatLogTime(log?.created_at);
  const content = log?.content || log?.action_type || '进行了操作';
  return `${actor} 在 ${time} ${content}`;
};

const loadTaskTitle = async () => {
  taskTitle.value = '';
  if (!Number.isFinite(taskId.value) || taskId.value <= 0) {
    taskTitle.value = `任务 #${taskId.value}`;
    return;
  }

  const cached = taskStore.tasks.value.find((item) => Number(item.id) === taskId.value);
  if (cached?.title) {
    taskTitle.value = cached.title;
    return;
  }

  try {
    const teamId = teamIdFromQuery.value;
    if (teamId) {
      const teamRes = await getTeamTasksApi(teamId);
      if (teamRes.data?.code !== 0) {
        throw new Error(getApiFailureMessage(teamRes, '加载任务标题失败', '团队任务查询失败'));
      }
      const teamList = Array.isArray(teamRes.data?.data?.list) ? teamRes.data.data.list : [];
      const teamFound = teamList.find((item) => Number(item.id) === taskId.value);
      if (teamFound?.title) {
        taskTitle.value = teamFound.title;
        return;
      }
    }

    const response = await getTasksApi(1, 1000, 'default');
    if (response.data?.code !== 0) {
      throw new Error(getApiFailureMessage(response, '加载任务标题失败', '任务查询失败'));
    }

    const list = Array.isArray(response.data?.data?.list) ? response.data.data.list : [];
    const found = list.find((item) => Number(item.id) === taskId.value);
    if (found?.title) {
      taskTitle.value = found.title;
    }
  } catch (err) {
    taskTitle.value = `任务 #${taskId.value}`;
  }
};


const loadLogs = async () => {
  if (!Number.isFinite(taskId.value) || taskId.value <= 0) {
    logsError.value = '任务 ID 无效';
    return;
  }

  logsLoading.value = true;
  logsError.value = '';
  try {
    const response = await getTaskLogsApi(taskId.value, logPage.value, logSize.value);
    if (response.data?.code !== 0) {
      throw new Error(getApiFailureMessage(response, '加载日志失败', '日志查询失败'));
    }
    const data = response.data?.data || {};
    logs.value = Array.isArray(data.list) ? data.list : [];
    logTotal.value = Number.isFinite(data.total) ? data.total : logs.value.length;
  } catch (err) {
    logsError.value = getApiErrorMessage(err, '加载日志失败', '日志查询失败');
  } finally {
    logsLoading.value = false;
  }
};

onMounted(async () => {
  await loadTaskTitle();
  await loadLogs();
});

watch(() => route.params.id, async () => {
  logPage.value = 1;
  await loadTaskTitle();
  await loadLogs();
});

watch(logPage, async () => {
  await loadLogs();
});
</script>

<style scoped>
* {
  box-sizing: border-box;
}

.detail-card {
  width: 750px;
  max-width: calc(100vw - 32px);
  margin: 0 auto;
  min-height: 760px;
}

.detail-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  min-height: 128px;
  padding: 1.5rem;
  margin-bottom: 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #ffffff;
}

.header-main h1 {
  margin: 0 0 0.35rem;
  font-size: 1.5rem;
  color: #0f172a;
}

.eyebrow {
  margin: 0 0 0.35rem;
  color: #64748b;
  font-size: 0.82rem;
  font-weight: 700;
}

.header-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
}

.back-btn {
  min-height: 38px;
  padding: 0 14px;
  border: 1px solid #c9ddff;
  border-radius: 999px;
  background: #f8fbff;
  color: #1d4ed8;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.25s ease;
}

.task-list-btn {
  min-height: 38px;
  padding: 0 14px;
  border: 1px solid #c9ddff;
  border-radius: 999px;
  background: #f8fbff;
  color: #1d4ed8;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.25s ease;
}

.task-list-btn:hover {
  background: #eaf3ff;
  border-color: #accbff;
}

.back-btn:hover {
  border-color: #93c5fd;
  background: #eff6ff;
}

.task-name {
  font-size: 0.85rem;
  color: #64748b;
}

.log-panel {
  padding: 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #ffffff;
}

.panel-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}

.panel-heading h2 {
  margin: 0;
  font-size: 1.05rem;
  color: #0f172a;
}

.panel-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.log-count {
  font-size: 0.8rem;
  color: #64748b;
}

.state-panel {
  padding: 1rem;
  border-radius: 10px;
  background: #f8fafc;
  color: #64748b;
  text-align: center;
  border: 1px solid #e2e8f0;
}

.state-panel.error {
  background: #fee2e2;
  color: #991b1b;
  border-color: #fecaca;
}

.state-panel.empty {
  background: #ffffff;
  border-style: dashed;
}

.log-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.log-item {
  padding: 0.85rem 1rem;
  border-radius: 10px;
  background: #ffffff;
  border: 1px solid #e2e8f0;
}

.log-line {
  color: #0f172a;
  font-size: 0.9rem;
}

.log-meta {
  margin-top: 0.5rem;
  font-size: 0.75rem;
  color: #64748b;
}

.log-type {
  padding: 0.2rem 0.45rem;
  background: #e2e8f0;
  border-radius: 999px;
  font-weight: 600;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  margin-top: 1.5rem;
}

.page-info {
  font-size: 0.85rem;
  color: #6b7280;
}

@media (max-width: 768px) {
  .detail-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .panel-heading {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
