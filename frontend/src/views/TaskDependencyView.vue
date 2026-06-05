<template>
  <div class="page-shell">
    <section class="dependency-page">
      <header class="dependency-header">
        <button class="back-btn" @click="router.back()">← 返回</button>
        <div class="header-main">
          <p class="eyebrow">任务依赖详情</p>
          <h1>{{ taskTitle }}</h1>
          <div class="header-meta">
            <span class="status-badge" :class="statusClass(task?.status || currentTask?.status)">
              {{ statusLabel(task?.status || currentTask?.status) }}
            </span>
            <span
              v-if="task?.priority"
              class="priority-badge"
              :class="`priority-${task?.priority}`"
            >
              优先级 {{ priorityLabel(task?.priority) }}
            </span>
            <span class="task-id">#{{ taskId }}</span>
          </div>
        </div>
        <RouterLink class="task-list-btn" to="/tasks">任务列表</RouterLink>
      </header>

      <div v-if="loading" class="state-panel">依赖数据加载中...</div>

      <div v-else-if="error" class="state-panel error">
        <p>{{ error }}</p>
        <button class="btn btn-primary" type="button" @click="loadDependencyData">重试</button>
      </div>

      <template v-else>
        <section class="detail-grid">
          <div class="detail-item">
            <span class="detail-label">任务描述</span>
            <p class="detail-value">{{ task?.description || '暂无描述' }}</p>
          </div>
          <div class="detail-item">
            <span class="detail-label">任务范围</span>
            <p class="detail-value">{{ scopeLabel }}</p>
          </div>
          <div class="detail-item">
            <span class="detail-label">截止时间</span>
            <p class="detail-value">{{ formatDate(task?.deadline) }}</p>
          </div>
          <div class="detail-item">
            <span class="detail-label">创建时间</span>
            <p class="detail-value">{{ formatDate(task?.created_at) }}</p>
          </div>
        </section>

        <section class="dependency-columns">
          <div class="dependency-panel">
            <div class="panel-heading">
              <h2>直接前置任务</h2>
              <span>{{ predecessors.length }}</span>
            </div>
            <div v-if="predecessors.length === 0" class="empty-box">当前任务没有前置依赖。</div>
            <div v-else class="task-node-list">
              <RouterLink
                v-for="task in predecessors"
                :key="task.id"
                class="task-node-row"
                :class="{ unfinished: task.status !== 'DONE' }"
                :to="buildDependencyRoute(task.id)"
              >
                <span class="node-title">{{ task.title }}</span>
                <span class="status-badge" :class="statusClass(task.status)">
                  {{ statusLabel(task.status) }}
                </span>
              </RouterLink>
            </div>
          </div>

          <div class="dependency-panel">
            <div class="panel-heading">
              <h2>直接后继任务</h2>
              <span>{{ successors.length }}</span>
            </div>
            <div v-if="successors.length === 0" class="empty-box">暂无任务依赖当前任务。</div>
            <div v-else class="task-node-list">
              <RouterLink
                v-for="task in successors"
                :key="task.id"
                class="task-node-row"
                :to="buildDependencyRoute(task.id)"
              >
                <span class="node-title">{{ task.title }}</span>
                <span class="status-badge" :class="statusClass(task.status)">
                  {{ statusLabel(task.status) }}
                </span>
              </RouterLink>
            </div>
          </div>
        </section>

        <section class="graph-panel" :class="{ expanded: showGraph }">
          <div class="panel-heading graph-heading">
            <div>
              <h2>全量依赖图</h2>
              <span>{{ graphLinks.length }} 条依赖边</span>
            </div>
            <button class="graph-toggle-btn" type="button" @click="showGraph = !showGraph">
              {{ showGraph ? '收起依赖图' : '查看依赖图' }}
            </button>
          </div>

          <div v-if="!showGraph" class="graph-collapsed">
            <span>当前图包含 {{ graphNodes.length }} 个节点、{{ graphLinks.length }} 条依赖边。</span>
          </div>

          <template v-else>
            <div v-if="graphNodes.length <= 1 && graphLinks.length === 0" class="empty-box">
              当前任务尚未形成依赖链。
            </div>

            <div v-else class="graph-scroll">
              <div
                class="graph-canvas"
                :style="{ width: `${graphLayout.width}px`, height: `${graphLayout.height}px` }"
              >
                <svg class="graph-svg" :viewBox="`0 0 ${graphLayout.width} ${graphLayout.height}`">
                  <defs>
                    <marker
                      id="arrow"
                      viewBox="0 0 10 10"
                      refX="9"
                      refY="5"
                      markerWidth="7"
                      markerHeight="7"
                      orient="auto-start-reverse"
                    >
                      <path d="M 0 0 L 10 5 L 0 10 z" fill="#94a3b8" />
                    </marker>
                  </defs>
                  <path
                    v-for="link in graphLayout.links"
                    :key="`${link.source}-${link.target}`"
                    class="graph-link"
                    :d="link.path"
                    marker-end="url(#arrow)"
                  />
                </svg>

                <RouterLink
                  v-for="node in graphLayout.nodes"
                  :key="node.id"
                  class="graph-node"
                  :class="{ current: node.id === taskId, done: node.status === 'DONE' }"
                  :style="{ left: `${node.x}px`, top: `${node.y}px` }"
                  :to="buildDependencyRoute(node.id)"
                >
                  <span class="graph-node-title">{{ node.title }}</span>
                  <span class="status-badge" :class="statusClass(node.status)">
                    {{ statusLabel(node.status) }}
                  </span>
                </RouterLink>
              </div>
            </div>
          </template>
        </section>
      </template>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import { taskStore } from '../stores/task';
import { getTasksApi, getTaskDependenciesApi, getTaskDependencyGraphApi } from '../services/task';
import { getTeamTasksApi } from '../services/team';
import { getApiErrorMessage, getApiFailureMessage } from '../utils/apiError';

const route = useRoute();
const router = useRouter();

const loading = ref(false);
const error = ref('');
const task = ref(null);
const predecessors = ref([]);
const successors = ref([]);
const graphNodes = ref([]);
const graphLinks = ref([]);
const showGraph = ref(false);

const taskId = computed(() => Number(route.params.id));
const teamIdFromQuery = computed(() => {
  const raw = route.query.team_id;
  if (raw === undefined || raw === null || raw === '') return null;
  const parsed = Number(raw);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
});

const currentTask = computed(() =>
  graphNodes.value.find((node) => Number(node.id) === taskId.value) || null
);

const taskTitle = computed(() => task.value?.title || currentTask.value?.title || `任务 #${taskId.value}`);

const isPersonalTask = (item) => {
  const teamId = item?.team_id;
  return teamId === null || teamId === undefined || teamId === '' || teamId === 0 || teamId === '0';
};

const scopeLabel = computed(() => {
  if (!task.value) {
    return teamIdFromQuery.value ? '团队任务' : '-';
  }
  return isPersonalTask(task.value) ? '个人任务' : '团队任务';
});

const priorityLabel = (priority) => {
  const map = { 1: '低', 2: '中', 3: '高' };
  return map[priority] || priority || '-';
};

const formatDate = (dateStr) => {
  if (!dateStr) return '-';
  const date = new Date(dateStr);
  if (Number.isNaN(date.getTime())) return String(dateStr);
  const year = date.getUTCFullYear();
  const month = String(date.getUTCMonth() + 1).padStart(2, '0');
  const day = String(date.getUTCDate()).padStart(2, '0');
  const hour = String(date.getUTCHours()).padStart(2, '0');
  const minute = String(date.getUTCMinutes()).padStart(2, '0');
  return `${year}年${month}月${day}日 ${hour}:${minute}`;
};

const normalizeTaskNode = (node) => ({
  id: Number(node.id),
  title: node.title || `任务 #${node.id}`,
  status: node.status || 'TODO',
});

const resolveTeamId = () => {
  if (teamIdFromQuery.value) return teamIdFromQuery.value;
  if (task.value && !isPersonalTask(task.value)) {
    const parsed = Number(task.value.team_id);
    return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
  }
  return null;
};

const fallbackTaskFromGraph = (teamId) => {
  const graphNode = graphNodes.value.find((node) => Number(node.id) === taskId.value);
  if (!graphNode) return null;
  return {
    id: taskId.value,
    title: graphNode.title,
    status: graphNode.status,
    team_id: teamId,
    description: '',
  };
};

const loadTaskInfo = async () => {
  task.value = null;
  const cached = taskStore.tasks.value.find((item) => Number(item.id) === taskId.value);
  if (cached) {
    task.value = cached;
    return;
  }

  const teamId = resolveTeamId();
  if (teamId) {
    const teamRes = await getTeamTasksApi(teamId);
    if (teamRes.data?.code !== 0) {
      throw new Error(getApiFailureMessage(teamRes, '加载任务失败', '团队任务查询失败'));
    }
    const teamList = Array.isArray(teamRes.data?.data?.list) ? teamRes.data.data.list : [];
    const teamFound = teamList.find((item) => Number(item.id) === taskId.value);
    if (teamFound) {
      task.value = teamFound;
      return;
    }
  }

  const response = await getTasksApi(1, 1000, 'default');
  if (response.data?.code !== 0) {
    throw new Error(getApiFailureMessage(response, '加载任务失败', '任务查询失败'));
  }

  const list = Array.isArray(response.data?.data?.list) ? response.data.data.list : [];
  const found = list.find((item) => Number(item.id) === taskId.value);
  if (found) {
    task.value = found;
    return;
  }

  // 依赖接口已按团队成员权限返回数据；列表接口可能不含「分配给别人的」团队任务
  const fallback = fallbackTaskFromGraph(teamId);
  if (fallback) {
    task.value = fallback;
    return;
  }

  task.value = {
    id: taskId.value,
    title: `任务 #${taskId.value}`,
    status: 'TODO',
    team_id: teamId,
    description: '',
  };
};

const buildDependencyRoute = (id) => {
  const teamId = resolveTeamId();
  const path = `/tasks/${id}/dependencies`;
  return teamId ? { path, query: { team_id: teamId } } : path;
};

const loadDependencyData = async () => {
  if (!Number.isFinite(taskId.value) || taskId.value <= 0) {
    error.value = '任务 ID 无效';
    return;
  }

  showGraph.value = false;
  loading.value = true;
  error.value = '';
  try {
    const [depsRes, graphRes] = await Promise.all([
      getTaskDependenciesApi(taskId.value),
      getTaskDependencyGraphApi(taskId.value),
    ]);

    if (depsRes.data?.code !== 0) {
      throw new Error(getApiFailureMessage(depsRes, '加载直接依赖失败', '依赖数据查询失败'));
    }
    if (graphRes.data?.code !== 0) {
      throw new Error(getApiFailureMessage(graphRes, '加载依赖图失败', '依赖数据查询失败'));
    }

    const deps = depsRes.data.data || {};
    const graph = graphRes.data.data || {};
    predecessors.value = Array.isArray(deps.predecessors)
      ? deps.predecessors.map(normalizeTaskNode)
      : [];
    successors.value = Array.isArray(deps.successors)
      ? deps.successors.map(normalizeTaskNode)
      : [];
    graphNodes.value = Array.isArray(graph.nodes)
      ? graph.nodes.map(normalizeTaskNode)
      : [];
    graphLinks.value = Array.isArray(graph.links)
      ? graph.links.map((link) => ({
          source: Number(link.source),
          target: Number(link.target),
        }))
      : [];

    if (!graphNodes.value.some((node) => node.id === taskId.value)) {
      graphNodes.value = [
        ...graphNodes.value,
        { id: taskId.value, title: `任务 #${taskId.value}`, status: 'TODO' },
      ];
    }

    await loadTaskInfo();
  } catch (err) {
    error.value = getApiErrorMessage(err, '加载依赖详情失败', '依赖数据查询失败');
  } finally {
    loading.value = false;
  }
};

const graphLayout = computed(() => {
  const nodeWidth = 150;                                                                                                                                                                                                 
  const nodeHeight = 56;                                                                                                                                                                                                 
  const columnGap = 260;                                                                                                                                                                                                 
  const rowGap = 110;   
  const margin = 48;
  const minCanvasWidth = 1000;

  const nodes = graphNodes.value.map(normalizeTaskNode);
  const nodeIds = new Set(nodes.map((node) => node.id));
  const links = graphLinks.value.filter((link) => nodeIds.has(link.source) && nodeIds.has(link.target));
  const rank = new Map(nodes.map((node) => [node.id, 0]));
  const indegree = new Map(nodes.map((node) => [node.id, 0]));
  const outgoing = new Map(nodes.map((node) => [node.id, []]));

  links.forEach((link) => {
    indegree.set(link.target, (indegree.get(link.target) || 0) + 1);
    outgoing.get(link.source)?.push(link.target);
  });

  const queue = nodes
    .filter((node) => (indegree.get(node.id) || 0) === 0)
    .map((node) => node.id);
  const visited = new Set();

  while (queue.length > 0) {
    const current = queue.shift();
    if (visited.has(current)) continue;
    visited.add(current);

    for (const next of outgoing.get(current) || []) {
      rank.set(next, Math.max(rank.get(next) || 0, (rank.get(current) || 0) + 1));
      indegree.set(next, (indegree.get(next) || 0) - 1);
      if ((indegree.get(next) || 0) === 0) queue.push(next);
    }
  }

  const groups = new Map();
  nodes.forEach((node) => {
    const group = rank.get(node.id) || 0;
    if (!groups.has(group)) groups.set(group, []);
    groups.get(group).push(node);
  });

  const positioned = [];
  const sortedGroups = [...groups.entries()].sort(([left], [right]) => left - right);
  sortedGroups.forEach(([group, groupNodes]) => {
    groupNodes
      .sort((left, right) => left.id - right.id)
      .forEach((node, index) => {
        positioned.push({
          ...node,
          x: margin + group * columnGap,
          y: margin + index * rowGap,
          width: nodeWidth,
          height: nodeHeight,
        });
      });
  });

  const byID = new Map(positioned.map((node) => [node.id, node]));
  const maxRank = Math.max(0, ...sortedGroups.map(([group]) => group));
  const width = Math.max(minCanvasWidth, margin * 2 + (maxRank + 1) * columnGap);
  const maxRows = Math.max(1, ...sortedGroups.map(([, groupNodes]) => groupNodes.length));
  const height = Math.max(320, margin * 2 + maxRows * rowGap);
  const layoutLinks = links.map((link) => {
    const source = byID.get(link.source);
    const target = byID.get(link.target);
    const startX = source.x + source.width;
    const startY = source.y + source.height / 2;
    const endX = target.x;
    const endY = target.y + target.height / 2;
    const midX = startX + Math.max(50, (endX - startX) / 2);
    return {
      ...link,
      path: `M ${startX} ${startY} C ${midX} ${startY}, ${midX} ${endY}, ${endX - 8} ${endY}`,
    };
  });

  return { nodes: positioned, links: layoutLinks, width, height };
});

const statusLabel = (status) => {
  const map = {
    TODO: '待做',
    DOING: '进行中',
    DONE: '已完成',
    ERROR: '异常',
  };
  return map[status] || status || '未知';
};

const statusClass = (status) => `status-${String(status || 'todo').toLowerCase()}`;

onMounted(loadDependencyData);
watch(() => route.params.id, loadDependencyData);
</script>

<style scoped>
* {
  box-sizing: border-box;
}

.dependency-page {
  width: 750px;
  max-width: calc(100vw - 32px);
  margin: 0 auto;
  min-height: 760px;
}

.dependency-header {
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

.dependency-header h1 {
  margin: 0;
  color: #0f172a;
  font-size: 1.45rem;
  line-height: 1.25;
}

.header-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.75rem;
}

.task-id {
  color: #64748b;
  font-size: 0.85rem;
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


.btn-primary {
  border-color: #2563eb;
  background: #2563eb;
  color: #ffffff;
}

.task-list-btn {
  min-height: 38px;
  padding: 0 14px;
  border: 1px solid #c9ddff;
  border-radius: 8px;
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

.state-panel,
.empty-box {
  padding: 1.25rem;
  border: 1px dashed #cbd5e1;
  border-radius: 10px;
  background: #f8fafc;
  color: #64748b;
  text-align: center;
}

.state-panel {
  min-height: 520px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.state-panel.error {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  border-color: #fecaca;
  background: #fef2f2;
  color: #991b1b;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
  margin-bottom: 1rem;
}

.detail-item {
  padding: 1rem;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  background: #f8fafc;
}

.detail-label {
  display: block;
  font-size: 0.8rem;
  color: #64748b;
  margin-bottom: 0.35rem;
  font-weight: 600;
}

.detail-value {
  margin: 0;
  color: #0f172a;
  font-size: 0.92rem;
  line-height: 1.4;
}


.dependency-columns {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
  margin-bottom: 1rem;
}

.dependency-panel,
.graph-panel {
  display: flex;
  flex-direction: column;
  padding: 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #ffffff;
}

.dependency-panel {
  min-height: 246px;
}

.graph-panel {
  min-height: 132px;
}

.graph-panel.expanded {
  min-height: 404px;
}

.panel-heading {
  flex-shrink: 0;
  display: flex;
  align-items: center;
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
  color: #64748b;
  font-size: 0.82rem;
  font-weight: 700;
}

.graph-heading {
  align-items: flex-start;
}

.graph-heading > div {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.graph-toggle-btn {
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

.graph-toggle-btn:hover {
  background: #eff6ff;
  border-color: #60a5fa;
}

.graph-collapsed {
  min-height: 52px;
  padding: 0.85rem 1rem;
  border: 1px dashed #cbd5e1;
  border-radius: 10px;
  background: #f8fafc;
  color: #64748b;
  display: flex;
  align-items: center;
}

.task-node-list {
  min-height: 148px;
  max-height: 220px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

.dependency-panel > .empty-box {
  min-height: 148px;
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.task-node-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.75rem;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  color: #0f172a;
  text-decoration: none;
}

.task-node-row:hover {
  border-color: #93c5fd;
  background: #f8fbff;
}

.task-node-row.unfinished {
  border-color: #fed7aa;
}

.node-title,
.graph-node-title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-badge {
  flex-shrink: 0;
  padding: 0.24rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: 700;
}

.priority-badge {
  flex-shrink: 0;
  padding: 0.22rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: 700;
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

.status-error {
  background: #fee2e2;
  color: #991b1b;
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

.graph-scroll {
  height: 320px;
  overflow: auto;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #f8fafc;
}

.graph-canvas {
  position: relative;
  min-width: 100%;
  min-height: 320px;
}

.graph-panel > .empty-box {
  min-height: 320px;
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.graph-svg {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.graph-link {
  fill: none;
  stroke: #94a3b8;
  stroke-width: 2;
}

.graph-node {
  position: absolute;
  width: 150px;                                                                                                                                                                                                        
  min-height: 56px;   
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 0.45rem;
  padding: 0.65rem 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 10px;
  background: #ffffff;
  color: #0f172a;
  text-decoration: none;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.08);
}

.graph-node.current {
  border-color: #2563eb;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.14), 0 8px 18px rgba(15, 23, 42, 0.08);
}

.graph-node.done {
  opacity: 0.86;
}

@media (max-width: 860px) {
  .dependency-header {
    flex-direction: column;
    align-items: stretch;
  }

  .dependency-columns {
    grid-template-columns: 1fr;
  }
}
</style>
