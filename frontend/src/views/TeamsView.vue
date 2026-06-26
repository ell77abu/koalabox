<template>
  <div class="page-shell">
    <section class="card card-wide">
      <header class="task-header">
        <div class="task-title">
          <h1>我的团队</h1>
          <p class="hint">和伙伴们高效协同，管理你的团队空间</p>
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
              placeholder="搜索团队"
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
          <RouterLink to="/teams/create" class="fab" data-hint="创建新团队" title="创建新团队">
            <span class="plus-icon">+</span>
          </RouterLink>
        </div>
      </header>

      <div class="list-section">
        <div class="list-header">
          <h2>团队列表</h2>
          <p v-if="!loading" class="hint">
            <template v-if="isSearching">
              搜索结果 {{ teamTotal }} 项
            </template>
            <template v-else>
              共 {{ teamTotal }} 个团队
            </template>
          </p>
        </div>

        <div v-if="loading" class="loading-state">
          <p>加载中...</p>
        </div>

        <div v-else-if="loadError" class="empty-state error-state">
          <p>{{ loadError }}</p>
          <div class="error-actions">
            <button class="btn btn-small btn-primary-style" type="button" @click="loadTeams">重试</button>
          </div>
        </div>

        <div v-else-if="teamTotal === 0" class="empty-state">
          <p v-if="isSearching">未找到匹配的团队。</p>
          <p v-else>你还没有加入任何团队，快去创建一个吧！</p>
        </div>

        <div v-else class="task-list">
          <section class="task-group">
            <div class="task-group-list">
              <div
                v-for="team in paginatedTeams"
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

            <div v-if="teamTotal > 0" class="pagination">
              <button class="btn btn-outline btn-small" :disabled="currentPage === 1" @click="currentPage--">上一页</button>
              <span class="page-info">第 {{ currentPage }} 页 / 共 {{ totalPages }} 页</span>
              <button class="btn btn-outline btn-small" :disabled="currentPage === totalPages" @click="currentPage++">下一页</button>
            </div>
          </section>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, computed, nextTick, onMounted, watch } from 'vue'; // 新增导入 watch
import { useRouter, RouterLink } from 'vue-router';
import { teamStore } from '../stores/team';
import { getTeamsApi } from '../services/team';
import { getApiErrorMessage, getApiFailureMessage } from '../utils/apiError';

const router = useRouter();
const loading = ref(false);
const loadError = ref('');
const searchQuery = ref('');
const teamTotal = ref(0);

const isSearching = computed(() => searchQuery.value.trim().length > 0);

const displayedTeams = computed(() => teamStore.teams.value);

// ================== 分页逻辑 ==================
const currentPage = ref(1);
const pageSize = ref(8); // 每页显示 8 个团队

const totalPages = computed(() => Math.ceil(teamTotal.value / pageSize.value) || 1);

// 当前页显示的数据
const paginatedTeams = computed(() => displayedTeams.value);

let suppressTeamPageWatch = false;

const reloadTeamsFromFirstPage = async () => {
  suppressTeamPageWatch = true;
  currentPage.value = 1;
  await nextTick();
  suppressTeamPageWatch = false;
  await loadTeams();
};

// 搜索条件变化时重置页码
watch(searchQuery, reloadTeamsFromFirstPage);

watch(currentPage, () => {
  if (!suppressTeamPageWatch) {
    loadTeams();
  }
});

// 自动退页（例如：解散或退出了最后一页的最后一个团队时）
watch(paginatedTeams, (newVal) => {
  if (newVal.length === 0 && currentPage.value > 1) {
    currentPage.value--;
  }
});
// ==============================================

const roleBadgeClass = (role) => {
  if (role === 'Owner') return 'priority-3';
  if (role === 'Admin') return 'priority-2';
  return 'priority-1';
};

const extractTeamPage = (response) => {
  const pageData = response?.data?.data || {};
  const list = Array.isArray(pageData.list) ? pageData.list : [];
  const total = Number.isFinite(Number(pageData.total)) ? Number(pageData.total) : list.length;
  const page = Number.isFinite(Number(pageData.page)) ? Number(pageData.page) : 1;
  const size = Number.isFinite(Number(pageData.size)) ? Number(pageData.size) : 10;
  return { list, total, page, size };
};

const loadTeams = async () => {
  loading.value = true;
  loadError.value = '';
  try {
    const keyword = searchQuery.value.trim();
    const res = await getTeamsApi({
      page: currentPage.value,
      size: pageSize.value,
      search: keyword || undefined,
    });
    if (res.data.code === 0) {
      const pageData = extractTeamPage(res);
      teamStore.setTeams(pageData.list);
      teamTotal.value = pageData.total;
      currentPage.value = pageData.page || currentPage.value;
      pageSize.value = pageData.size || pageSize.value;
    } else {
      teamStore.setTeams([]);
      teamTotal.value = 0;
      loadError.value = getApiFailureMessage(res, '加载团队失败');
    }
  } catch (error) {
    teamStore.setTeams([]);
    teamTotal.value = 0;
    loadError.value = getApiErrorMessage(error, '加载团队失败');
    console.error(loadError.value, error);
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

/*分页控件及其配套按钮样式*/
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
  font-weight: 500;
}
.btn-outline { background: #f1f5f9; color: #475569; border: 1px solid #cbd5e1; }
.btn-outline:hover:not(:disabled) { background: #e2e8f0; }
.btn-outline:disabled { opacity: 0.5; cursor: not-allowed; }

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

/* 深色模式：分页控件 */
:global([data-theme="dark"]) .pagination { border-top-color: #334155; }
:global([data-theme="dark"]) .btn-outline { background: #0f172a; color: #cbd5e1; border-color: #334155; }
:global([data-theme="dark"]) .page-info { color: #94a3b8; }
</style>
