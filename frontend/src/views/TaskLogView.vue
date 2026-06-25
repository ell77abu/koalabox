<template>
  <div class="page-shell">
    <section class="log-page">
      <header class="log-header">
        <button class="back-btn" type="button" @click="router.back()">← 返回</button>
        <div class="header-main">
          <p class="eyebrow">任务日志</p>
          <h1>{{ title }}</h1>
          <div class="header-meta">
            <span v-if="isDeleted" class="deleted-badge">已删除</span>
            <span v-if="teamId" class="scope-badge">团队任务</span>
            <span v-else class="scope-badge">个人任务</span>
          </div>
        </div>
        <span class="task-id">#{{ taskId }}</span>
      </header>

      <TaskActivityLog :task-id="taskId" />
    </section>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import TaskActivityLog from '../components/TaskActivityLog.vue';

const route = useRoute();
const router = useRouter();

const taskId = computed(() => Number(route.params.id));
const teamId = computed(() => {
  const parsed = Number(route.query.team_id);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
});
const isDeleted = computed(() => route.query.deleted === '1' || route.query.deleted === 'true');
const title = computed(() => route.query.title || `任务 #${taskId.value}`);
</script>

<style scoped>
.log-page {
  width: 750px;
  max-width: calc(100vw - 32px);
  margin: 0 auto;
  min-height: 520px;
}

.log-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  min-height: 118px;
  padding: 1.5rem;
  margin-bottom: 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #ffffff;
}

.header-main {
  flex: 1;
  min-width: 0;
}

.eyebrow {
  margin: 0 0 0.35rem;
  color: #64748b;
  font-size: 0.82rem;
  font-weight: 700;
}

.log-header h1 {
  margin: 0;
  color: #0f172a;
  font-size: 1.45rem;
  line-height: 1.25;
}

.header-meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.75rem;
}

.task-id {
  color: #1d4ed8;
  font-size: 1.35rem;
  font-weight: 800;
  line-height: 1;
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
}

.back-btn:hover {
  border-color: #93c5fd;
  background: #eff6ff;
}

.deleted-badge,
.scope-badge {
  padding: 0.22rem 0.55rem;
  border-radius: 6px;
  font-size: 0.76rem;
  font-weight: 700;
}

.deleted-badge {
  border: 1px solid #fecaca;
  background: #fef2f2;
  color: #b91c1c;
}

.scope-badge {
  border: 1px solid #bfdbfe;
  background: #eff6ff;
  color: #1d4ed8;
}

:global([data-theme="dark"]) .log-header {
  background: #1e293b;
  border-color: #334155;
}

:global([data-theme="dark"]) .log-header h1 {
  color: #f8fafc;
}

:global([data-theme="dark"]) .eyebrow {
  color: #94a3b8;
}

@media (max-width: 640px) {
  .log-header {
    flex-direction: column;
  }
}
</style>
