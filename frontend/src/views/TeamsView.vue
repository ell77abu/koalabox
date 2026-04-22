<template>
  <div class="page-shell">
    <section class="card card-wide">
      <header class="task-header">
        <div class="task-title">
          <h1>我的团队</h1>
          <p class="hint">和伙伴们高效协同，管理你的团队空间</p>
        </div>
        <div class="header-actions">
          <RouterLink to="/teams/create" class="fab" data-hint="创建新团队" title="创建新团队">
            <span class="plus-icon">+</span>
          </RouterLink>
        </div>
      </header>

      <div class="list-section">
        <div class="list-header">
          <h2>团队列表</h2>
          <p v-if="!loading" class="hint">共 {{ teamStore.teams.value.length }} 个团队</p>
        </div>

        <div v-if="loading" class="loading-state">
          <p>加载中...</p>
        </div>

        <div v-else-if="teamStore.teams.value.length === 0" class="empty-state">
          <p>你还没有加入任何团队，快去创建一个吧！</p>
        </div>

        <div v-else class="task-list">
          <section class="task-group">
            <div class="task-group-list">
              <div
                v-for="team in teamStore.teams.value"
                :key="team.team_id"
                class="task-item"
              >
                <div class="task-main">
                  <div class="task-title-section">
                    <h3>{{ team.name }}</h3>
                    <span class="priority-badge" :class="roleBadgeClass(team.role)">
                      角色：{{ team.role }}
                    </span>
                  </div>

                  <div class="task-meta">
                    <span class="meta-item">
                      <span class="meta-label">团队 ID：</span>#{{ team.team_id }}
                    </span>
                  </div>
                </div>

                <div class="task-actions">
                  <button
                    class="btn btn-small btn-primary-style"
                    @click="goToTeamDetail(team.team_id)"
                  >
                    查看团队 →
                  </button>
                </div>
              </div>
            </div>
          </section>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRouter, RouterLink } from 'vue-router';
import { teamStore } from '../stores/team';
import { getTeamsApi } from '../services/team';

const router = useRouter();
const loading = ref(false);

const roleBadgeClass = (role) => {
  if (role === 'Owner') return 'priority-3';
  if (role === 'Admin') return 'priority-2';
  return 'priority-1';
};

const loadTeams = async () => {
  loading.value = true;
  try {
    const res = await getTeamsApi();
    if (res.data.code === 0) {
      teamStore.setTeams(res.data.data.list || []);
    }
  } catch (error) {
    console.error('加载团队失败', error);
  } finally {
    loading.value = false;
  }
};

const goToTeamDetail = (teamId) => {
  router.push(`/teams/${teamId}`);
};

onMounted(() => {
  loadTeams();
});
</script>

<style scoped>
/*基础 UI 样式*/
.task-header { display: flex; flex-direction: column; align-items: stretch; gap: 1rem; margin-bottom: 2rem; padding-bottom: 1rem; border-bottom: 1px solid #e5e7eb; }
.task-title { text-align: center; }
.task-header h1 { margin: 0 0 0.5rem 0; font-size: 1.5rem; color: var(--text); }
.task-header .hint { margin: 0; color: var(--muted); font-size: 0.875rem; }
.header-actions { display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; width: 100%; }

.fab {
  display: flex; align-items: center; justify-content: center; width: 48px; height: 48px;
  border-radius: 50%; background: var(--primary); color: #fff; text-decoration: none;
  border: none; padding: 0; outline: none; cursor: pointer;
  box-shadow: 0 4px 10px rgba(28, 126, 214, 0.28); transition: transform 0.2s, box-shadow 0.2s; flex-shrink: 0;
}
.fab:hover { transform: scale(1.05); box-shadow: 0 6px 14px rgba(28, 126, 214, 0.35); }
.plus-icon { font-size: 30px; font-weight: 600; line-height: 1; margin-top: -6px; }

.list-section { margin-bottom: 2rem; }
.list-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }
.list-header h2 { margin: 0; font-size: 1rem; color: var(--text); font-weight: 600; }
.empty-state, .loading-state { text-align: center; padding: 2rem; color: #6b7280; background: #f9fafb; border-radius: 0.5rem; }
.task-group { padding: 1rem; background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 0.75rem; }
.task-group-list { display: flex; flex-direction: column; gap: 1rem; }

.task-item {
  position: relative; overflow: hidden; display: flex; justify-content: space-between; align-items: center;
  padding: 1.25rem 1rem; background: #fff; border: 1px solid #e5e7eb; border-radius: 0.5rem;
  transition: transform 0.22s ease, box-shadow 0.22s ease, border-color 0.22s ease;
}
.task-item:hover { transform: translateY(-4px); border-color: #bfdbfe; box-shadow: 0 12px 24px rgba(15, 23, 42, 0.12); }
.task-main { flex: 1; min-width: 0; }
.task-title-section { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 0.5rem; flex-wrap: wrap; }
.task-title-section h3 { margin: 0; font-size: 1.1rem; color: #1f2937; font-weight: 700; }

.priority-badge { padding: 0.25rem 0.5rem; border-radius: 0.25rem; font-size: 0.75rem; font-weight: 600; }
.priority-1 { background: #e0e7ff; color: #4338ca; }
.priority-2 { background: #fef08a; color: #92400e; }
.priority-3 { background: #fee2e2; color: #b91c1c; }

.task-meta { display: flex; gap: 1rem; flex-wrap: wrap; margin-top: 0.5rem; font-size: 0.8rem; color: #9ca3af; }
.meta-item { display: flex; align-items: center; gap: 0.25rem; }
.meta-label { color: #64748b; font-size: 0.85rem; }

.task-actions { display: flex; align-items: center; gap: 0.5rem; margin-left: 1rem; flex-shrink: 0; }
.btn-small { padding: 0.5rem 1rem; font-size: 0.85rem; font-weight: 600; border-radius: 8px; cursor: pointer; border: none; transition: all 0.2s ease; }
.btn-primary-style { background: #eff6ff; color: #2563eb; border: 1px solid #bfdbfe; }
.btn-primary-style:hover { background: #2563eb; color: #fff; border-color: #2563eb; box-shadow: 0 4px 12px rgba(37, 99, 235, 0.2); }

@media (max-width: 768px) {
  .task-item { flex-direction: column; align-items: flex-start; gap: 1rem; }
  .task-actions { margin-left: 0; width: 100%; }
  .task-actions button { width: 100%; }
}

/* === 深色模式适配补丁 === */
:global([data-theme="dark"]) .task-item { background: #1e293b; border-color: #334155; }
:global([data-theme="dark"]) .task-group { background: #0f172a; border-color: #1e293b; }
:global([data-theme="dark"]) .task-title-section h3 { color: #f8fafc; }
:global([data-theme="dark"]) .empty-state,
:global([data-theme="dark"]) .loading-state { background: #1e293b; color: #94a3b8; }
:global([data-theme="dark"]) .meta-label { color: #94a3b8; }
:global([data-theme="dark"]) .task-meta { color: #cbd5e1; }
:global([data-theme="dark"]) .btn-primary-style { background: #1e3a8a; color: #bfdbfe; border-color: #1e40af; }
:global([data-theme="dark"]) .btn-primary-style:hover { background: #2563eb; color: #fff; }
</style>