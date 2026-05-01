<template>
  <div class="page-shell">
    <section class="card card-wide">
      <!-- 页头 -->
      <header class="task-header">
        <div class="task-title-row">
          <div class="title-text">
            <h1>🔔 系统消息</h1>
            <p class="hint">处理你的团队邀请与系统通知</p>
          </div>
        </div>
      </header>

      <!-- 消息列表区 -->
      <div class="list-section">
        <div v-if="loading" class="loading-state">加载消息中...</div>
        <div v-else-if="invitations.length === 0" class="empty-state">
          你目前没有任何待处理的消息
        </div>

        <div v-else class="task-list">
          <div v-for="inv in invitations" :key="inv.id" class="task-item">
            <div class="task-main">
              <div class="task-title-section">
                <!-- 兼容处理：如果有 team_name 则展示，否则回退展示 team_id -->
                <h3>团队邀请：{{ inv.team_name || `Team #${inv.team_id}` }}</h3>
                
                <!-- 状态徽章 -->
                <span class="status-badge" :class="`status-${inv.status.toLowerCase()}`">
                  {{ inv.status === 'PENDING' ? '待处理' : (inv.status === 'ACCEPTED' ? '已同意' : '已拒绝') }}
                </span>
              </div>
              
              <p class="task-description">
                用户 <strong>{{ inv.inviter_id }}</strong> 邀请你加入该团队。
                <br>
                <span class="time-text">发送时间: {{ formatDate(inv.created_at) }}</span>
              </p>
            </div>

            <!-- 操作区：只有处于 PENDING (待处理) 状态才显示操作按钮 -->
            <div class="task-actions" v-if="inv.status === 'PENDING'">
              <button class="btn btn-small btn-complete" @click="handleAccept(inv.id)" :disabled="processingId === inv.id">
                同意
              </button>
              <button class="btn btn-small btn-danger" @click="handleDecline(inv.id)" :disabled="processingId === inv.id">
                拒绝
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { getInvitationsApi, acceptInvitationApi, declineInvitationApi } from '../services/team';

// 页面状态
const invitations = ref([]);
const loading = ref(false);
const processingId = ref(null); // 防止用户连击按钮

// 格式化时间辅助函数
const formatDate = (dateString) => {
  if (!dateString) return '';
  return new Date(dateString).toLocaleString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit'
  });
};

// 加载邀请列表
const loadInvitations = async () => {
  loading.value = true;
  try {
    const res = await getInvitationsApi();
    // 确保有数据才进行处理
    if (res.data && Array.isArray(res.data)) {
      // 按照时间倒序排列，让最新的消息排在最上面
      invitations.value = res.data.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
    }
  } catch (error) {
    console.error('加载邀请失败', error);
  } finally {
    loading.value = false;
  }
};

// 接受邀请
const handleAccept = async (id) => {
    // 新增内存锁：如果当前这条消息正在处理，直接拦截
  if (processingId.value === id) return;
  processingId.value = id;
  try {
    const res = await acceptInvitationApi(id);
    if (res.status === 200) {
      await loadInvitations(); // 重新加载列表，更新状态
    }
  } catch (error) {
    alert(error.response?.data?.message || '操作失败，请重试');
  } finally {
    processingId.value = null;
  }
};

// 拒绝邀请
const handleDecline = async (id) => {
  // 内存锁
  if (processingId.value === id) return;  
  processingId.value = id;
  try {
    const res = await declineInvitationApi(id);
    if (res.status === 200) {
      await loadInvitations(); // 重新加载列表，更新状态
    }
  } catch (error) {
    alert(error.response?.data?.message || '操作失败，请重试');
  } finally {
    processingId.value = null;
  }
};

// 组件挂载时自动拉取数据
onMounted(() => {
  loadInvitations();
});
</script>

<style scoped>
/*基础高定 UI 样式*/
.task-header { margin-bottom: 1.5rem; padding-bottom: 1rem; border-bottom: 1px solid #e5e7eb; }
.title-text h1 { margin: 0; font-size: 1.5rem; color: var(--text); }
.hint { margin: 4px 0 0; color: var(--muted); font-size: 0.85rem; }

.list-section { margin-top: 1rem; }
.empty-state, .loading-state { text-align: center; padding: 3rem; color: #6b7280; background: #f8fafc; border-radius: 0.5rem; border: 1px dashed #cbd5e1; font-weight: 500; }

.task-list { display: flex; flex-direction: column; gap: 1rem; }
.task-item { 
  display: flex; justify-content: space-between; align-items: center; 
  padding: 1.2rem; background: #fff; border: 1px solid #e2e8f0; border-radius: 12px; 
  transition: transform 0.2s, box-shadow 0.2s; 
}
.task-item:hover { transform: translateY(-2px); box-shadow: 0 10px 20px rgba(15, 23, 42, 0.08); border-color: #cbd5e1; }

.task-main { flex: 1; }
.task-title-section { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 0.4rem; }
.task-title-section h3 { margin: 0; font-size: 1.1rem; color: #0f172a; font-weight: 600; }

.task-description { margin: 0; font-size: 0.9rem; color: #475569; line-height: 1.6; }
.time-text { font-size: 12px; color: #94a3b8; }

/* 状态徽章 */
.status-badge { padding: 0.25rem 0.6rem; border-radius: 6px; font-size: 0.75rem; font-weight: 600; letter-spacing: 0.02em; }
.status-pending { background: #fef08a; color: #92400e; }
.status-accepted { background: #dcfce7; color: #166534; }
.status-declined { background: #fee2e2; color: #b91c1c; }

/* 按钮样式 */
.task-actions { display: flex; gap: 0.75rem; }
.btn { padding: 0.5rem 1rem; font-size: 0.85rem; font-weight: 600; border-radius: 8px; cursor: pointer; border: none; transition: all 0.2s ease; }
.btn:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-small { padding: 0.45rem 0.9rem; font-size: 0.8rem; }
.btn-complete { background: linear-gradient(135deg, #22c55e, #16a34a); color: white; box-shadow: 0 4px 10px rgba(34,197,94,0.2); }
.btn-complete:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 6px 14px rgba(34,197,94,0.3); }
.btn-danger { background: #fee2e2; color: #b91c1c; border: 1px solid #fecaca; }
.btn-danger:hover:not(:disabled) { background: #fecaca; }

/*深色模式适配*/
:global([data-theme="dark"]) .task-header { border-bottom-color: #334155; }
:global([data-theme="dark"]) .task-item { background: #1e293b; border-color: #334155; }
:global([data-theme="dark"]) .task-title-section h3 { color: #f8fafc; }
:global([data-theme="dark"]) .task-description { color: #94a3b8; }
:global([data-theme="dark"]) .empty-state,
:global([data-theme="dark"]) .loading-state { background: #0f172a; color: #94a3b8; border-color: #334155; }
</style>