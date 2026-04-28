<template>
  <div class="page-shell">
    <section class="card card-wide">
      
      <header class="task-header">
        <div class="task-title-row">
          <button class="back-btn" @click="router.back()">← 返回</button>
          <div class="title-text">
            <h1>{{ currentTeam ? currentTeam.name : '加载中...' }}</h1>
            <p class="hint">团队空间 ID: #{{ teamId }} | 你的角色: <span :class="roleTextClass(myRole)">{{ myRole }}</span></p>
          </div>
        </div>
      </header>

      <div class="tabs-container">
        <button 
          class="tab-btn" 
          :class="{ active: activeTab === 'tasks' }" 
          @click="activeTab = 'tasks'"
        >
          📋 团队任务
        </button>
        <button 
          class="tab-btn" 
          :class="{ active: activeTab === 'members' }" 
          @click="activeTab = 'members'"
        >
          👥 成员管理
        </button>
      </div>

      <div v-if="activeTab === 'tasks'" class="tab-content">
        <div class="list-header">
          <h2>任务大厅</h2>
          <button 
            v-if="myRole === 'Owner' || myRole === 'Admin'" 
            class="btn btn-primary btn-small" 
            @click="showCreateTaskModal = true"
          >
            + 发布团队任务
          </button>
        </div>

        <div v-if="loadingTasks" class="loading-state">加载任务中...</div>
        <div v-else-if="tasks.length === 0" class="empty-state">当前团队暂无任务。</div>
        
        <div v-else class="task-list">
          <div v-for="task in tasks" :key="task.id" class="task-item">
            <div class="task-main">
              <div class="task-title-section">
                <h3>{{ task.title }}</h3>
                <span class="status-badge" :class="`status-${task.status.toLowerCase()}`">
                  {{ statusLabel(task.status) }}
                </span>
                <span class="priority-badge" :class="`priority-${task.priority}`">
                  优先级: {{ priorityLabel(task.priority) }}
                </span>
                <span v-if="task.assignee_id" class="priority-badge" style="background: #eef2ff; color: #4f46e5;">
                  执行者: {{ getAssigneeName(task.assignee_id) }}
                </span>
                <span v-else class="priority-badge" style="background: #f1f5f9; color: #64748b;">
                  待分配
                </span>
              </div>
              <p class="task-description">{{ task.description || '暂无描述' }}</p>
            </div>

            <div class="task-actions" v-if="canEditTask(task)">
              <button class="btn btn-small btn-edit" @click="startEdit(task)">
                编辑
              </button>
              <button 
                v-if="myRole === 'Owner' || myRole === 'Admin'" 
                class="btn btn-small btn-danger" 
                @click="handleDeleteTask(task.id)"
              >
                删除
              </button>
            </div>

          </div>
        </div>
      </div>

      <div v-if="activeTab === 'members'" class="tab-content">
        <div class="list-header">
          <h2>成员列表</h2>
          <button 
            v-if="myRole === 'Owner'" 
            class="btn btn-complete btn-small" 
            @click="showAddMemberModal = true"
          >
            + 邀请新成员
          </button>
        </div>

        <div v-if="loadingMembers" class="loading-state">加载成员中...</div>
        
        <div v-else class="task-list">
          <div v-for="member in members" :key="member.user_id" class="task-item">
            <div class="task-main">
              <div class="task-title-section">
                <h3>{{ member.username }}</h3>
                <span class="priority-badge" :class="roleBadgeClass(member.role)">
                  {{ member.role }}
                </span>
              </div>
            </div>

            <div class="task-actions" v-if="myRole === 'Owner' && member.role !== 'Owner'">
              <button 
                class="btn btn-small btn-outline" 
                @click="toggleAdminRole(member)"
              >
                {{ member.role === 'Admin' ? '撤销管理员' : '设为管理员' }}
              </button>
              <button class="btn btn-small btn-danger" @click="removeMember(member.user_id)">
                移除
              </button>
            </div>
          </div>
        </div>
      </div>

    </section>

    <div v-if="showCreateTaskModal" class="modal-overlay" @click="showCreateTaskModal = false">
      <div class="modal" @click.stop>
        <h2>发布团队任务</h2>
        <form @submit.prevent="handleCreateTask" class="task-form">
          <input v-model.trim="newTask.title" type="text" placeholder="任务标题" required class="form-input" />
          <textarea v-model.trim="newTask.description" placeholder="任务描述 (选填)" class="form-input form-textarea"></textarea>
          
          <div class="form-row">
            <select v-model="newTask.priority" class="form-input">
              <option value="1">优先级：低</option>
              <option value="2">优先级：中</option>
              <option value="3">优先级：高</option>
            </select>
            <input v-model.trim="newTask.deadline" type="datetime-local" class="form-input" />
          </div>

          <label style="font-size: 14px; font-weight: 600; color: #475569; margin-top: 10px;">指派给成员 (选填):</label>
          <select v-model="newTask.assignee_id" class="form-input">
            <option :value="null">-- 暂不分配，放入团队池 --</option>
            <option v-for="m in members" :key="m.user_id" :value="m.user_id">
              {{ m.username }} ({{ m.role }})
            </option>
          </select>

          <div class="modal-actions">
            <button type="button" class="btn btn-outline" @click="showCreateTaskModal = false">取消</button>
            <button type="submit" class="btn btn-primary" :disabled="submittingTask">发布</button>
          </div>
        </form>
      </div>
    </div>

    <div v-if="showEditTaskModal" class="modal-overlay" @click="showEditTaskModal = false">
      <div class="modal" @click.stop>
        <h2>编辑任务</h2>
        <form @submit.prevent="handleUpdateTask" class="task-form">
          
          <input 
            v-model.trim="editingTask.title" 
            type="text" 
            placeholder="任务标题" 
            required 
            class="form-input" 
            :disabled="myRole === 'Member'"
          />
          
          <textarea 
            v-model.trim="editingTask.description" 
            placeholder="任务描述" 
            class="form-input form-textarea"
            :disabled="myRole === 'Member'"
          ></textarea>
          
          <p v-if="myRole === 'Member'" style="color: #ef4444; font-size: 13px; margin: 0; font-weight: 600;">
            * 您的角色为 Member，仅能修改分派给自己的任务状态。
          </p>

          <div class="form-row">
            <select v-model="editingTask.status" class="form-input">
              <option value="TODO">状态：待做</option>
              <option value="DOING">状态：进行中</option>
              <option value="DONE">状态：已完成</option>
            </select>

            <select v-model="editingTask.priority" class="form-input" :disabled="myRole === 'Member'">
              <option value="1">优先级：低</option>
              <option value="2">优先级：中</option>
              <option value="3">优先级：高</option>
            </select>
          </div>

          <div class="form-row">
            <input 
              v-model.trim="editingTask.deadline" 
              type="datetime-local" 
              class="form-input date-input"
              :disabled="myRole === 'Member'"
            />
            
            <select v-model="editingTask.assignee_id" class="form-input" :disabled="myRole === 'Member'">
              <option :value="null">-- 暂不分配，放入团队池 --</option>
              <option v-for="m in members" :key="m.user_id" :value="m.user_id">
                {{ m.username }} ({{ m.role }})
              </option>
            </select>
          </div>

          <div class="modal-actions">
            <button type="button" class="btn btn-outline" @click="showEditTaskModal = false">取消</button>
            <button type="submit" class="btn btn-primary" :disabled="submittingTask">保存</button>
          </div>
        </form>
      </div>
    </div>

    <div v-if="showAddMemberModal" class="modal-overlay" @click="showAddMemberModal = false">
      <div class="modal" @click.stop>
        <h2>邀请新成员</h2>
        <form @submit.prevent="handleAddMember" class="task-form">
          <label style="font-size: 14px; color: #475569; margin-bottom: 5px;">请输入你要邀请的【用户名】：</label>
          <input v-model.trim="newMemberName" type="text" placeholder="例如: Alice123" required class="form-input" />

          <div class="modal-actions">
            <button type="button" class="btn btn-outline" @click="showAddMemberModal = false">取消</button>
            <button type="submit" class="btn btn-complete" :disabled="submittingMember">确认邀请</button>
          </div>
        </form>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
// 👇 新增引入：当前登录用户信息 (用于判断该任务是否分配给自己)
import { authStore } from '../stores/auth'; 
import { teamStore } from '../stores/team';
import { 
  getTeamsApi,
  getTeamMembersApi, 
  addTeamMemberApi, 
  deleteTeamMemberApi, 
  updateTeamMemberRoleApi,
  getTeamTasksApi
} from '../services/team';
// 👇 新增引入：修改和删除任务的 API
import { createTaskApi, updateTaskApi, deleteTaskApi } from '../services/task';

const route = useRoute();
const router = useRouter();
const teamId = Number(route.params.team_id);

const activeTab = ref('tasks');

// 权限与基础数据
const currentTeam = computed(() => teamStore.teams.value.find(t => t.team_id === teamId));
const myRole = computed(() => currentTeam.value ? currentTeam.value.role : 'Member');

// 团队任务状态
const tasks = ref([]);
const loadingTasks = ref(false);
const showCreateTaskModal = ref(false);
const showEditTaskModal = ref(false);
const submittingTask = ref(false);

const newTask = ref({ title: '', description: '', priority: '2', deadline: '', assignee_id: null });
const editingTask = ref({ id: null, title: '', description: '', status: 'TODO', priority: '2', deadline: '', assignee_id: null });

// 成员管理状态
const members = ref([]);
const loadingMembers = ref(false);
const showAddMemberModal = ref(false);
const submittingMember = ref(false);
const newMemberName = ref(''); 

/* ================== 工具函数 ================== */
const statusLabel = (status) => ({ TODO: '待做', DOING: '进行中', DONE: '已完成' }[status] || status);
const priorityLabel = (p) => ({ 1: '低', 2: '中', 3: '高' }[p] || p);

const roleBadgeClass = (role) => {
  if (role === 'Owner') return 'priority-3';
  if (role === 'Admin') return 'priority-2';
  return 'priority-1';
};

// 通过成员列表中的 user_id 查找对应的 username
const getAssigneeName = (assigneeId) => {
  if (!assigneeId) return '';
  const member = members.value.find(m => m.user_id === assigneeId);
  return member ? member.username : assigneeId;
};

const roleTextClass = (role) => {
  if (role === 'Owner') return 'text-red-600 font-bold';
  if (role === 'Admin') return 'text-orange-500 font-bold';
  return 'text-blue-600 font-bold';
};

// 👇 新增：权限判断函数：是否有资格点击“编辑”按钮
const canEditTask = (task) => {
  if (myRole.value === 'Owner' || myRole.value === 'Admin') return true;
  // 普通 Member 只能操作分派给自己的任务
  return task.assignee_id === authStore.username.value;
};

// 日期解析转换
const parseDateInput = (value) => {
  if (!value) return null;
  const match = String(value).trim().replace(/\//g, '-').match(/^(\d{4})-(\d{2})-(\d{2})(?:T(\d{2}):(\d{2}))?$/);
  if (!match) return null;
  return new Date(Date.UTC(Number(match[1]), Number(match[2]) - 1, Number(match[3]), match[4] ? Number(match[4]) : 0, match[5] ? Number(match[5]) : 0)).toISOString();
};

const formatDateForInput = (dateStr) => {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  if (Number.isNaN(date.getTime())) return '';
  const year = date.getUTCFullYear();
  const month = String(date.getUTCMonth() + 1).padStart(2, '0');
  const day = String(date.getUTCDate()).padStart(2, '0');
  const hour = String(date.getUTCHours()).padStart(2, '0');
  const minute = String(date.getUTCMinutes()).padStart(2, '0');
  return `${year}-${month}-${day}T${hour}:${minute}`;
};

/* ================== 数据加载 ================== */

const loadTeams = async () => {
  try {
    const res = await getTeamsApi();
    if (res.data.code === 0) {
      teamStore.setTeams(res.data.data.list || []);
    }
  } catch (error) {
    console.error('加载团队列表失败', error);
  }
};

const loadTeamTasks = async () => {
  loadingTasks.value = true;
  try {
    const res = await getTeamTasksApi(teamId);
    if (res.data.code === 0) tasks.value = res.data.data.list || [];
  } catch (error) {
    console.error('加载任务失败', error);
  } finally {
    loadingTasks.value = false;
  }
};

const loadTeamMembers = async () => {
  loadingMembers.value = true;
  try {
    const res = await getTeamMembersApi(teamId);
    if (res.data.code === 0) members.value = res.data.data || [];
  } catch (error) {
    console.error('加载成员失败', error);
  } finally {
    loadingMembers.value = false;
  }
};

/* ================== 任务操作 ================== */
const handleCreateTask = async () => {
  submittingTask.value = true;
  try {
    const payload = {
      ...newTask.value,
      priority: Number(newTask.value.priority),
      deadline: parseDateInput(newTask.value.deadline),
      team_id: teamId, 
      status: 'TODO'
    };
    const res = await createTaskApi(payload);
    if (res.data.code === 0) {
      showCreateTaskModal.value = false;
      newTask.value = { title: '', description: '', priority: '2', deadline: '', assignee_id: null };
      await loadTeamTasks();
    } else alert(res.data.message || '发布失败');
  } catch (error) {
    alert(error.response?.data?.message || '请求失败');
  } finally {
    submittingTask.value = false;
  }
};

// 👇 新增：打开编辑任务面板
const startEdit = (task) => {
  editingTask.value = {
    id: task.id,
    title: task.title,
    description: task.description || '',
    status: task.status,
    priority: String(task.priority),
    deadline: task.deadline ? formatDateForInput(task.deadline) : '',
    assignee_id: task.assignee_id || null
  };
  showEditTaskModal.value = true;
};

// 👇 新增：保存更新任务
const handleUpdateTask = async () => {
  submittingTask.value = true;
  try {
    const payload = {
      title: editingTask.value.title,
      description: editingTask.value.description || null,
      status: editingTask.value.status,
      priority: Number(editingTask.value.priority),
      deadline: parseDateInput(editingTask.value.deadline),
      assignee_id: editingTask.value.assignee_id || null
    };

    const res = await updateTaskApi(editingTask.value.id, payload);
    if (res.data.code === 0) {
      showEditTaskModal.value = false;
      await loadTeamTasks();
    } else {
      alert(res.data.message || '更新失败');
    }
  } catch (error) {
    alert(error.response?.data?.message || '更新失败，可能是权限不足');
  } finally {
    submittingTask.value = false;
  }
};

// 👇 新增：删除团队任务
const handleDeleteTask = async (taskId) => {
  if (!confirm('确定要删除这个团队任务吗？此操作不可撤销。')) return;
  try {
    const res = await deleteTaskApi(taskId);
    if (res.data.code === 0) {
      await loadTeamTasks();
    } else {
      alert(res.data.message || '删除失败');
    }
  } catch (error) {
    alert(error.response?.data?.message || '无权删除');
  }
};

/* ================== 成员操作 ================== */
const handleAddMember = async () => {
  if (!newMemberName.value) return;
  submittingMember.value = true;
  try {
    const res = await addTeamMemberApi(teamId, { username: newMemberName.value });
    if (res.data.code === 0) {
      showAddMemberModal.value = false;
      newMemberName.value = ''; 
      await loadTeamMembers();
    } else alert(res.data.message || '邀请失败');
  } catch (error) {
    alert(error.response?.data?.message || '请求失败，可能用户不存在或已在团队中');
  } finally {
    submittingMember.value = false;
  }
};

const removeMember = async (userId) => {
  if (!confirm('确定要将该成员移出团队吗？')) return;
  try {
    const res = await deleteTeamMemberApi(teamId, userId);
    if (res.data.code === 0) await loadTeamMembers();
    else alert(res.data.message || '移除失败');
  } catch (error) {
    alert(error.response?.data?.message || '无权操作或请求失败');
  }
};

const toggleAdminRole = async (member) => {
  const newRole = member.role === 'Admin' ? 'Member' : 'Admin';
  try {
    const res = await updateTeamMemberRoleApi(teamId, member.user_id, { role: newRole });
    if (res.data.code === 0) await loadTeamMembers();
    else alert(res.data.message || '权限修改失败');
  } catch (error) {
    alert(error.response?.data?.message || '无权操作');
  }
};

onMounted(() => {
  loadTeams();
  loadTeamTasks();
  loadTeamMembers();
});
</script>

<style scoped>
.task-header { margin-bottom: 1.5rem; padding-bottom: 1rem; }
.task-title-row { display: flex; align-items: center; gap: 1rem; }
.back-btn { min-height: 38px; padding: 0 14px; border: 1px solid #c9ddff; border-radius: 999px; background: #f8fbff; color: #1d4ed8; cursor: pointer; font-size: 14px; font-weight: 600; }
.back-btn:hover { background: #eaf3ff; border-color: #accbff; }
.title-text h1 { margin: 0; font-size: 1.5rem; color: var(--text); }
.hint { margin: 4px 0 0; color: var(--muted); font-size: 0.85rem; }

/* 选项卡样式 */
.tabs-container { display: flex; gap: 1.5rem; border-bottom: 2px solid #e5e7eb; margin-bottom: 1.5rem; padding-left: 10px; }
.tab-btn { padding: 0.5rem 0.5rem 0.75rem; border: none; background: transparent; color: #64748b; font-size: 1rem; font-weight: 600; cursor: pointer; border-bottom: 3px solid transparent; transition: all 0.2s; margin-bottom: -2px; }
.tab-btn:hover { color: #3b82f6; }
.tab-btn.active { color: #2563eb; border-bottom-color: #2563eb; }

/* 列表容器 */
.list-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; }
.list-header h2 { font-size: 1.1rem; color: var(--text); font-weight: 700; margin: 0; }
.empty-state, .loading-state { text-align: center; padding: 3rem; color: #6b7280; background: #f8fafc; border-radius: 0.5rem; border: 1px dashed #cbd5e1; }

/* 卡片 */
.task-list { display: flex; flex-direction: column; gap: 1rem; }
.task-item { display: flex; justify-content: space-between; align-items: center; padding: 1.2rem; background: #fff; border: 1px solid #e2e8f0; border-radius: 12px; transition: transform 0.2s, box-shadow 0.2s; }
.task-item:hover { transform: translateY(-2px); box-shadow: 0 10px 20px rgba(15, 23, 42, 0.08); border-color: #cbd5e1; }
.task-main { flex: 1; }
.task-title-section { display: flex; align-items: center; gap: 0.75rem; flex-wrap: wrap; margin-bottom: 0.4rem; }
.task-title-section h3 { margin: 0; font-size: 1.05rem; color: #0f172a; font-weight: 600; }
.task-description { margin: 0; font-size: 0.85rem; color: #64748b; }

/* 徽章 */
.priority-badge { padding: 0.2rem 0.5rem; border-radius: 6px; font-size: 0.75rem; font-weight: 600; }
.priority-1 { background: #e0e7ff; color: #4338ca; }
.priority-2 { background: #fef08a; color: #92400e; }
.priority-3 { background: #fee2e2; color: #b91c1c; }
.status-badge { padding: 0.2rem 0.5rem; border-radius: 6px; font-size: 0.75rem; font-weight: 600; }
.status-todo { background: #dbeafe; color: #1e40af; }
.status-doing { background: #fef08a; color: #92400e; }
.status-done { background: #dcfce7; color: #166534; }

/* 按钮 */
.btn { padding: 0.5rem 1rem; font-size: 0.85rem; font-weight: 600; border-radius: 8px; cursor: pointer; border: none; transition: all 0.2s ease; }
.btn-small { padding: 0.4rem 0.8rem; font-size: 0.8rem; }
.btn-primary { background: #3b82f6; color: white; }
.btn-primary:hover { background: #2563eb; }
.btn-complete { background: linear-gradient(135deg, #22c55e, #16a34a); color: white; box-shadow: 0 4px 10px rgba(34,197,94,0.2); }
.btn-complete:hover { transform: translateY(-1px); box-shadow: 0 6px 14px rgba(34,197,94,0.3); }
.btn-outline { background: #f1f5f9; color: #475569; border: 1px solid #cbd5e1; }
.btn-outline:hover { background: #e2e8f0; }
.btn-edit { background: #dbeafe; color: #1e40af; border: 1px solid #93c5fd; }
.btn-edit:hover:not(:disabled) { background: #bfdbfe; }
.btn-danger { background: #fee2e2; color: #b91c1c; border: 1px solid #fecaca; }
.btn-danger:hover { background: #fecaca; }

/* 模态框 */
/*弹窗 (Modal) 核心控制层*/
.modal-overlay { 
  position: fixed; 
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.45); /* 半透明黑色遮罩 */
  backdrop-filter: blur(3px); /* 背景模糊效果 */
  display: flex; 
  justify-content: center; 
  align-items: center; 
  z-index: 9999; /* 强制置于页面最顶层 */
}

.modal { 
  background: #fff; 
  padding: 2.5rem; 
  border-radius: 16px; 
  width: 90%; 
  max-width: 650px; /* 宽敞的表单尺寸 */
  max-height: 90vh; /* 防止屏幕较小导致弹窗撑爆，允许内部滚动 */
  overflow-y: auto;
  box-shadow: 0 24px 50px rgba(0, 0, 0, 0.2); 
  animation: modal-pop 0.3s cubic-bezier(0.16, 1, 0.3, 1); /* 增加一点弹出的动效 */
}

.modal h2 { 
  margin-top: 0; 
  color: var(--text); 
  text-align: center; 
  margin-bottom: 1.5rem; 
  font-size: 1.4rem; 
}

/* ================== 弹窗内部表单 ================== */
.task-form { display: flex; flex-direction: column; gap: 1.25rem; }
.form-input { padding: 0.8rem; border: 1px solid #cbd5e1; border-radius: 8px; font-family: inherit; font-size: 0.95rem; }
.form-input:focus { outline: none; border-color: #3b82f6; box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15); }
.form-textarea { min-height: 120px; resize: vertical; line-height: 1.5; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.modal-actions { display: flex; gap: 1rem; margin-top: 1.5rem; }
.modal-actions button { flex: 1; padding: 0.8rem; border-radius: 8px; font-weight: 600; cursor: pointer; border: none; transition: all 0.2s ease; }

/* 弹窗弹出的丝滑小动效 */
@keyframes modal-pop {
  0% { opacity: 0; transform: scale(0.95) translateY(10px); }
  100% { opacity: 1; transform: scale(1) translateY(0); }
}

.text-red-600 { color: #dc2626; }
.text-orange-500 { color: #f97316; }
.text-blue-600 { color: #2563eb; }
.font-bold { font-weight: bold; }

/* === 深色模式适配补丁 === */
:global([data-theme="dark"]) .task-item { background: #1e293b; border-color: #334155; }
:global([data-theme="dark"]) .task-title-section h3 { color: #f8fafc; }
:global([data-theme="dark"]) .task-description { color: #94a3b8; }
:global([data-theme="dark"]) .empty-state,
:global([data-theme="dark"]) .loading-state { background: #0f172a; color: #94a3b8; border-color: #334155; }
:global([data-theme="dark"]) .modal { background: #1e293b; }
:global([data-theme="dark"]) .form-input { background: #0f172a; color: #f8fafc; border-color: #334155; }
:global([data-theme="dark"]) .tab-btn { color: #94a3b8; }
:global([data-theme="dark"]) .tab-btn:hover { color: #60a5fa; }
:global([data-theme="dark"]) .tab-btn.active { color: #3b82f6; border-bottom-color: #3b82f6; }
:global([data-theme="dark"]) .tabs-container { border-bottom-color: #334155; }
:global([data-theme="dark"]) .btn-outline { background: #0f172a; color: #cbd5e1; border-color: #334155; }
</style>