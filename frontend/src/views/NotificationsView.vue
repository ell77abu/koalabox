<template>
  <div class="page-shell">
    <section class="card card-wide">
      <header class="task-header">
        <div class="task-title-row">
          <div class="title-text">
            <h1>🔔 系统消息</h1>
            <p class="hint">处理你的团队邀请与任务指派</p>
          </div>
        </div>
      </header>

      <div class="list-section">
        <p v-if="loadError" class="error-banner">{{ loadError }}</p>
        <div v-if="loading" class="loading-state">加载消息中...</div>
        <div v-else-if="notifications.length === 0" class="empty-state">
          你目前没有任何待处理的消息
        </div>

        <div v-else class="task-list">
          <div v-for="item in notifications" :key="item.type + item.id" class="task-item">
            <div class="task-main">
              <div class="task-title-section">
                <!-- 区分渲染：如果是团队邀请 -->
                <template v-if="item.type === 'INVITATION'">
                  <h3>团队邀请：{{ item.team_name || `Team #${item.team_id}` }}</h3>
                  <span class="status-badge" :class="`status-${item.status.toLowerCase()}`">
                    {{ item.status === 'PENDING' ? '待处理' : (item.status === 'ACCEPTED' ? '已同意' : '已拒绝') }}
                  </span>
                </template>
                
                <!-- 区分渲染：如果是任务指派 (纯提示) -->
                <template v-else>
                  <h3>任务指派：{{ item.task_title || `Task #${item.task_id}` }}</h3>
                  <span class="status-badge status-accepted">系统通知</span>
                </template>
              </div>
              
              <p class="task-description">
                <template v-if="item.type === 'INVITATION'">
                  用户 <strong>{{ item.inviter_username || item.inviter_id }}</strong> 邀请你加入该团队。
                </template>
                <template v-else>
                  用户 <strong>{{ item.requester_username || item.requester_id }}</strong> 将一项团队任务指派给了你。
                </template>
                <br>
                <span class="time-text">发送时间: {{ formatDate(item.created_at) }}</span>
              </p>
            </div>

            <!-- 操作区：仅对团队邀请且状态为 PENDING 时展示按钮 (任务指派不展示按钮) -->
            <div class="task-actions" v-if="item.type === 'INVITATION' && item.status === 'PENDING'">
              <button class="btn btn-small btn-complete" @click="handleAccept(item.id)" :disabled="processingId === item.id">
                同意
              </button>
              <button class="btn btn-small btn-danger" @click="handleDecline(item.id)" :disabled="processingId === item.id">
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
import { 
  getInvitationsApi, 
  acceptInvitationApi, 
  declineInvitationApi, 
  getAssignmentRequestsApi, 
  acceptAssignmentApi 
} from '../services/team';
import { getApiErrorMessage, getApiFailureMessage } from '../utils/apiError';

const notifications = ref([]);
const loading = ref(false);
const loadError = ref('');
const processingId = ref(null);

const formatDate = (dateString) => {
  if (!dateString) return '';
  return new Date(dateString).toLocaleString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit'
  });
};

const loadNotifications = async () => {
  loading.value = true;
  loadError.value = '';
  try {
    const [invRes, assignRes] = await Promise.allSettled([
      getInvitationsApi(),
      getAssignmentRequestsApi()
    ]);

    const requestErrors = [];
    let invList = [];
    if (invRes.status === 'fulfilled') {
      // 兼容后端新旧两种格式：[...] 或 { code: 0, data: [...] }
      const resData = invRes.value.data;
      invList = Array.isArray(resData) ? resData : (resData?.data || []);
      invList = invList.map(item => ({ ...item, type: 'INVITATION' }));
    } else {
      requestErrors.push(getApiErrorMessage(invRes.reason, '加载团队邀请失败'));
    }

    let assignList = [];
    if (assignRes.status === 'fulfilled') {
      const resData = assignRes.value.data;
      assignList = Array.isArray(resData) ? resData : (resData?.data || []);
      assignList = assignList.map(item => ({ ...item, type: 'ASSIGNMENT' }));
      
      // 无感静默已读
      const pendingAssignments = assignList.filter(a => a.status === 'PENDING');
      if (pendingAssignments.length > 0) {
        const readResults = await Promise.allSettled(pendingAssignments.map(a => acceptAssignmentApi(a.id)));
        readResults.forEach((result, index) => {
          if (result.status === 'rejected') {
            const assignment = pendingAssignments[index];
            requestErrors.push(getApiErrorMessage(result.reason, `确认任务指派 #${assignment.task_id || assignment.id} 已读失败`));
          }
        });
        assignList.forEach(a => { if (a.status === 'PENDING') a.status = 'ACCEPTED'; });
      }
    } else {
      requestErrors.push(getApiErrorMessage(assignRes.reason, '加载任务指派通知失败'));
    }

    const combinedList = [...invList, ...assignList];
    notifications.value = combinedList.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
    if (requestErrors.length > 0) {
      loadError.value = requestErrors.join('；');
    }
  } catch (error) {
    loadError.value = getApiErrorMessage(error, '加载通知失败');
    console.error(loadError.value, error);
  } finally {
    loading.value = false;
  }
};

const handleAccept = async (id) => {
  if (processingId.value === id) return;
  processingId.value = id;
  try {
    const res = await acceptInvitationApi(id);
    if (res.status === 200 && (res.data?.code === undefined || res.data.code === 0)) {
      await loadNotifications();
    } else {
      alert(getApiFailureMessage(res, '同意团队邀请失败', '操作失败，请重试'));
    }
  } catch (error) {
    alert(getApiErrorMessage(error, '同意团队邀请失败'));
  } finally {
    processingId.value = null;
  }
};

const handleDecline = async (id) => {
  if (processingId.value === id) return;  
  processingId.value = id;
  try {
    const res = await declineInvitationApi(id);
    if (res.status === 200 && (res.data?.code === undefined || res.data.code === 0)) {
      await loadNotifications();
    } else {
      alert(getApiFailureMessage(res, '拒绝团队邀请失败', '操作失败，请重试'));
    }
  } catch (error) {
    alert(getApiErrorMessage(error, '拒绝团队邀请失败'));
  } finally {
    processingId.value = null;
  }
};

onMounted(() => {
  loadNotifications();
});
</script>

<style scoped>
/* 样式原封不动 */
.task-header { margin-bottom: 1.5rem; padding-bottom: 1rem; border-bottom: 1px solid #e5e7eb; }
.title-text h1 { margin: 0; font-size: 1.5rem; color: var(--text); }
.hint { margin: 4px 0 0; color: var(--muted); font-size: 0.85rem; }
.list-section { margin-top: 1rem; }
.empty-state, .loading-state { text-align: center; padding: 3rem; color: #6b7280; background: #f8fafc; border-radius: 0.5rem; border: 1px dashed #cbd5e1; font-weight: 500; }
.task-list { display: flex; flex-direction: column; gap: 1rem; }
.task-item { display: flex; justify-content: space-between; align-items: center; padding: 1.2rem; background: #fff; border: 1px solid #e2e8f0; border-radius: 12px; transition: transform 0.2s, box-shadow 0.2s; }
.task-item:hover { transform: translateY(-2px); box-shadow: 0 10px 20px rgba(15, 23, 42, 0.08); border-color: #cbd5e1; }
.task-main { flex: 1; }
.task-title-section { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 0.4rem; }
.task-title-section h3 { margin: 0; font-size: 1.1rem; color: #0f172a; font-weight: 600; }
.task-description { margin: 0; font-size: 0.9rem; color: #475569; line-height: 1.6; }
.time-text { font-size: 12px; color: #94a3b8; }
.status-badge { padding: 0.25rem 0.6rem; border-radius: 6px; font-size: 0.75rem; font-weight: 600; letter-spacing: 0.02em; }
.status-pending { background: #fef08a; color: #92400e; }
.status-accepted { background: #dcfce7; color: #166534; }
.status-declined { background: #fee2e2; color: #b91c1c; }
.task-actions { display: flex; gap: 0.75rem; }
.btn { padding: 0.5rem 1rem; font-size: 0.85rem; font-weight: 600; border-radius: 8px; cursor: pointer; border: none; transition: all 0.2s ease; }
.btn:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-small { padding: 0.45rem 0.9rem; font-size: 0.8rem; }
.btn-complete { background: linear-gradient(135deg, #22c55e, #16a34a); color: white; box-shadow: 0 4px 10px rgba(34,197,94,0.2); }
.btn-complete:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 6px 14px rgba(34,197,94,0.3); }
.btn-danger { background: #fee2e2; color: #b91c1c; border: 1px solid #fecaca; }
.btn-danger:hover:not(:disabled) { background: #fecaca; }
:global([data-theme="dark"]) .task-header { border-bottom-color: #334155; }
:global([data-theme="dark"]) .task-item { background: #1e293b; border-color: #334155; }
:global([data-theme="dark"]) .task-title-section h3 { color: #f8fafc; }
:global([data-theme="dark"]) .task-description { color: #94a3b8; }
:global([data-theme="dark"]) .empty-state,
:global([data-theme="dark"]) .loading-state { background: #0f172a; color: #94a3b8; border-color: #334155; }
</style>
