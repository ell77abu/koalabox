<template>
  <div class="page-shell">
    <section class="dependency-page">
      <header class="dependency-header">
        <button class="back-btn" type="button" @click="router.back()">返回</button>
        <div class="header-main">
          <p class="eyebrow">任务依赖详情</p>
          <h1>{{ currentTaskTitle }}</h1>
          <div class="header-meta">
            <span class="status-badge" :class="statusClass(currentTask?.status)">
              {{ statusLabel(currentTask?.status) }}
            </span>
            <span class="task-id">#{{ taskId }}</span>
          </div>
        </div>
        <RouterLink class="btn btn-outline" to="/tasks">任务列表</RouterLink>
      </header>

      <div v-if="loading" class="state-panel">依赖数据加载中...</div>

      <div v-else-if="error" class="state-panel error">
        <p>{{ error }}</p>
        <button class="btn btn-primary" type="button" @click="loadDependencyData">重试</button>
      </div>

      <template v-else>
        <section class="summary-grid">
          <div class="summary-item">
            <span class="summary-label">前置任务</span>
            <strong>{{ predecessors.length }}</strong>
          </div>
          <div class="summary-item" :class="{ warning: incompletePredecessors.length > 0 }">
            <span class="summary-label">未完成前置</span>
            <strong>{{ incompletePredecessors.length }}</strong>
          </div>
          <div class="summary-item">
            <span class="summary-label">后继任务</span>
            <strong>{{ successors.length }}</strong>
          </div>
          <div class="summary-item">
            <span class="summary-label">依赖图节点</span>
            <strong>{{ graphNodes.length }}</strong>
          </div>
        </section>

        <section class="dependency-status" :class="{ blocked: incompletePredecessors.length > 0 }">
          <strong>{{ dependencyStatusTitle }}</strong>
          <span>{{ dependencyStatusText }}</span>
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
                :to="`/tasks/${task.id}/dependencies`"
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
                :to="`/tasks/${task.id}/dependencies`"
              >
                <span class="node-title">{{ task.title }}</span>
                <span class="status-badge" :class="statusClass(task.status)">
                  {{ statusLabel(task.status) }}
                </span>
              </RouterLink>
            </div>
          </div>
        </section>

        <section class="graph-panel">
          <div class="panel-heading">
            <h2>全量依赖图</h2>
            <span>{{ graphLinks.length }} 条依赖边</span>
          </div>

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
                :to="`/tasks/${node.id}/dependencies`"
              >
                <span class="graph-node-title">{{ node.title }}</span>
                <span class="status-badge" :class="statusClass(node.status)">
                  {{ statusLabel(node.status) }}
                </span>
              </RouterLink>
            </div>
          </div>
        </section>
      </template>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import { getTaskDependenciesApi, getTaskDependencyGraphApi } from '../services/task';

const route = useRoute();
const router = useRouter();

const loading = ref(false);
const error = ref('');
const predecessors = ref([]);
const successors = ref([]);
const graphNodes = ref([]);
const graphLinks = ref([]);

const taskId = computed(() => Number(route.params.id));

const currentTask = computed(() =>
  graphNodes.value.find((node) => Number(node.id) === taskId.value) || null
);

const currentTaskTitle = computed(() =>
  currentTask.value?.title || `任务 #${taskId.value}`
);

const incompletePredecessors = computed(() =>
  predecessors.value.filter((task) => task.status !== 'DONE')
);

const dependencyStatusTitle = computed(() =>
  incompletePredecessors.value.length > 0 ? '当前任务仍被前置任务阻塞' : '当前任务前置依赖已就绪'
);

const dependencyStatusText = computed(() => {
  if (predecessors.value.length === 0) return '没有前置任务限制，可以直接推进。';
  if (incompletePredecessors.value.length === 0) return '所有直接前置任务都已完成。';
  return `还有 ${incompletePredecessors.value.length} 个直接前置任务未完成。`;
});

const normalizeTaskNode = (node) => ({
  id: Number(node.id),
  title: node.title || `任务 #${node.id}`,
  status: node.status || 'TODO',
});

const loadDependencyData = async () => {
  if (!Number.isFinite(taskId.value) || taskId.value <= 0) {
    error.value = '任务 ID 无效';
    return;
  }

  loading.value = true;
  error.value = '';
  try {
    const [depsRes, graphRes] = await Promise.all([
      getTaskDependenciesApi(taskId.value),
      getTaskDependencyGraphApi(taskId.value),
    ]);

    if (depsRes.data?.code !== 0 || graphRes.data?.code !== 0) {
      throw new Error(depsRes.data?.message || graphRes.data?.message || '依赖数据查询失败');
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
  } catch (err) {
    error.value = err.response?.data?.message || err.message || '依赖数据查询失败';
  } finally {
    loading.value = false;
  }
};

const graphLayout = computed(() => {
  const nodeWidth = 168;
  const nodeHeight = 62;
  const columnGap = 230;
  const rowGap = 92;
  const margin = 48;

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
  const width = Math.max(760, margin * 2 + (maxRank + 1) * columnGap);
  const maxRows = Math.max(1, ...sortedGroups.map(([, groupNodes]) => groupNodes.length));
  const height = Math.max(280, margin * 2 + maxRows * rowGap);
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
.dependency-page {
  max-width: 1100px;
  margin: 0 auto;
}

.dependency-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
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

.back-btn,
.btn {
  min-height: 38px;
  padding: 0 14px;
  border-radius: 8px;
  border: 1px solid #cbd5e1;
  background: #f8fafc;
  color: #1e293b;
  font-size: 0.85rem;
  font-weight: 700;
  cursor: pointer;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.btn-primary {
  border-color: #2563eb;
  background: #2563eb;
  color: #ffffff;
}

.btn-outline:hover,
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

.state-panel.error {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  border-color: #fecaca;
  background: #fef2f2;
  color: #991b1b;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.summary-item {
  padding: 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #ffffff;
}

.summary-item.warning {
  border-color: #fed7aa;
  background: #fff7ed;
}

.summary-label {
  display: block;
  margin-bottom: 0.45rem;
  color: #64748b;
  font-size: 0.82rem;
}

.summary-item strong {
  color: #0f172a;
  font-size: 1.6rem;
}

.dependency-status {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.9rem 1rem;
  margin-bottom: 1rem;
  border: 1px solid #bbf7d0;
  border-radius: 10px;
  background: #f0fdf4;
  color: #166534;
}

.dependency-status.blocked {
  border-color: #fed7aa;
  background: #fff7ed;
  color: #9a3412;
}

.dependency-columns {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
  margin-bottom: 1rem;
}

.dependency-panel,
.graph-panel {
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

.task-node-list {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
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

.graph-scroll {
  overflow: auto;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #f8fafc;
}

.graph-canvas {
  position: relative;
  min-width: 100%;
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
  width: 168px;
  min-height: 62px;
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
  .dependency-header,
  .dependency-status {
    flex-direction: column;
    align-items: stretch;
  }

  .summary-grid,
  .dependency-columns {
    grid-template-columns: 1fr;
  }
}
</style>
