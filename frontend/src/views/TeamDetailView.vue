<template>
  <div class="page-shell">
    <section class="card card-wide">
      
      <header class="task-header">
        <div class="task-title-row" style="justify-content: space-between; width: 100%;">
          <div style="display: flex; align-items: center; gap: 1rem;">
            <button class="back-btn" @click="router.back()">← 返回</button>
            <div class="title-text">
              <h1>{{ currentTeam ? currentTeam.name : '加载中...' }}</h1>
              <p class="hint">团队空间 ID: #{{ teamId }} | 你的角色: <span :class="roleTextClass(myRole)">{{ myRole }}</span></p>
            </div>
          </div>
          
          <button 
            v-if="myRole === 'Member' || myRole === 'Admin'" 
            class="btn btn-outline btn-small" 
            style="color: #dc2626; border-color: #fca5a5; background: #fef2f2;"
            @click="handleLeaveTeam"
          >
            退出团队
          </button>
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
        <button
          v-if="myRole === 'Owner'"
          class="tab-btn"
          :class="{ active: activeTab === 'manage' }"
          @click="activeTab = 'manage'"
        >
          ⚙️ 团队管理
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

            <div class="task-actions">
              <RouterLink
                class="btn btn-small btn-dependency"
                :to="`/tasks/${task.id}/dependencies`"
              >
                依赖详情
              </RouterLink>
              <button v-if="canEditTask(task)" class="btn btn-small btn-edit" @click="startEdit(task)">编辑</button>
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

            <div class="task-actions" v-if="canRemoveMember(member) || (myRole === 'Owner' && member.role !== 'Owner')">
              
              <button 
                v-if="myRole === 'Owner' && member.role !== 'Owner'"
                class="btn btn-small btn-outline" 
                @click="toggleAdminRole(member)"
              >
                {{ member.role === 'Admin' ? '撤销管理员' : '设为管理员' }}
              </button>
              
              <button 
                v-if="canRemoveMember(member)"
                class="btn btn-small btn-danger" 
                @click="removeMember(member.user_id)"
              >
                移除
              </button>
            </div>
          </div>
        </div>
      </div>

      <div v-if="activeTab === 'manage'" class="tab-content">
        <div class="list-header">
          <h2>团队管理</h2>
        </div>

        <div class="manage-section">
          <div class="manage-card">
            <h3>修改团队名称</h3>
            <form @submit.prevent="handleRenameTeam" class="rename-form">
              <input
                v-model.trim="newTeamName"
                type="text"
                :placeholder="currentTeam ? currentTeam.name : '团队名称'"
                required
                class="form-input"
              />
              <button type="submit" class="btn btn-primary" :disabled="renamingTeam">
                {{ renamingTeam ? '保存中...' : '保存修改' }}
              </button>
            </form>
          </div>

          <div class="manage-card">
            <h3>转让团队</h3>
            <p>将团队所有权转让给其他成员。转让后，您将降级为管理员并可以退出团队。</p>
            <form @submit.prevent="handleTransferTeam" class="rename-form">
              <select v-model="newOwnerId" required class="form-input">
                <option :value="null" disabled>请选择要转让的成员...</option>
                <option v-for="m in members.filter(m => m.role !== 'Owner')" :key="m.user_id" :value="m.user_id">
                  {{ m.username }} ({{ m.role }})
                </option>
              </select>
              <button type="submit" class="btn btn-primary" :disabled="transferringTeam || !newOwnerId" style="background: #eab308;">
                {{ transferringTeam ? '转让中...' : '确认转让' }}
              </button>
            </form>
          </div>

          <div class="manage-card danger-zone">
            <h3>危险操作</h3>
            <p>解散团队将删除所有相关任务数据，此操作不可撤销。</p>
            <button class="btn btn-danger" :disabled="dissolvingTeam" @click="handleDissolveTeam">
              {{ dissolvingTeam ? '处理中...' : '解散团队' }}
            </button>
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

          <select v-model="newTask.assignee_id" class="form-input">
            <option :value="null">任务指派：暂不指派 (留入团队池)</option>
            <option v-for="m in members" :key="m.user_id" :value="m.user_id">
              指派给: {{ m.username }} ({{ m.role }})
            </option>
          </select>

          <div class="dependency-section" style="margin-top: 8px;">
            <label style="font-size: 14px; color: #475569; font-weight: 600; display: block; margin-bottom: 5px;">📌 设置前置依赖 (仅限本团队任务)</label>
            <div style="display: flex; flex-direction: column; gap: 6px; max-height: 140px; overflow-y: auto; padding: 10px; border: 1px solid #cbd5e1; border-radius: 8px; background: #f8fafc;">
              <label v-for="t in tasks" :key="t.id" style="display: flex; align-items: center; gap: 8px; font-size: 13px; color: #1e293b; cursor: pointer;">
                <input type="checkbox" :value="t.id" v-model="newTask.predecessor_ids" style="accent-color: #3b82f6; width: 14px; height: 14px; cursor: pointer;" />
                {{ t.title }}
                <span class="status-badge" :class="`status-${t.status.toLowerCase()}`" style="transform: scale(0.8); transform-origin: left;">{{ statusLabel(t.status) }}</span>
              </label>
              <span v-if="tasks.length === 0" style="font-size: 13px; color: #94a3b8;">团队暂无其他任务</span>
            </div>
          </div>

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
          <input v-model.trim="editingTask.title" type="text" placeholder="任务标题" required class="form-input" :disabled="myRole === 'Member'" />
          <textarea v-model.trim="editingTask.description" placeholder="任务描述" class="form-input form-textarea" :disabled="myRole === 'Member'"></textarea>
          
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
            <input v-model.trim="editingTask.deadline" type="datetime-local" class="form-input date-input" :disabled="myRole === 'Member'" />
            <select v-model="editingTask.assignee_id" class="form-input" :disabled="myRole === 'Member'">
              <option :value="null">任务指派：暂不指派 (留入团队池)</option>
              <option v-for="m in members" :key="m.user_id" :value="m.user_id">
                指派给: {{ m.username }} ({{ m.role }})
              </option>
            </select>
          </div>

          <div v-if="myRole === 'Owner' || myRole === 'Admin'" class="dependency-section" style="margin-top: 8px;">
            <label style="font-size: 14px; color: #475569; font-weight: 600; display: block; margin-bottom: 5px;">📌 设置前置依赖 (仅限本团队任务)</label>
            <div style="display: flex; flex-direction: column; gap: 6px; max-height: 140px; overflow-y: auto; padding: 10px; border: 1px solid #cbd5e1; border-radius: 8px; background: #f8fafc;">
              <label v-for="t in tasks.filter(task => task.id !== editingTask.id)" :key="t.id" style="display: flex; align-items: center; gap: 8px; font-size: 13px; color: #1e293b; cursor: pointer;">
                <input type="checkbox" :value="t.id" v-model="editingTask.predecessor_ids" style="accent-color: #3b82f6; width: 14px; height: 14px; cursor: pointer;" />
                {{ t.title }}
                <span class="status-badge" :class="`status-${t.status.toLowerCase()}`" style="transform: scale(0.8); transform-origin: left;">{{ statusLabel(t.status) }}</span>
              </label>
              <span v-if="tasks.length <= 1" style="font-size: 13px; color: #94a3b8;">团队暂无可依赖的其他任务</span>
            </div>
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
import { RouterLink, useRoute, useRouter } from 'vue-router';
import { authStore } from '../stores/auth'; 
import { teamStore } from '../stores/team';
import {
  getTeamsApi, getTeamMembersApi, addTeamMemberApi, deleteTeamMemberApi, updateTeamMemberRoleApi,
  getTeamTasksApi, updateTeamApi, dissolveTeamApi,
  // 👇 Lab3 引入：退出与转让 API
  leaveTeamApi, transferTeamApi
} from '../services/team';
import { 
  createTaskApi, updateTaskApi, deleteTaskApi,
  getTaskDependenciesApi, addTaskDependencyApi, removeTaskDependencyApi
} from '../services/task';

const route = useRoute();
const router = useRouter();
const teamId = Number(route.params.team_id);

const activeTab = ref('tasks');

const currentTeam = computed(() => teamStore.teams.value.find(t => t.team_id === teamId));
const myRole = computed(() => currentTeam.value ? currentTeam.value.role : 'Member');

const tasks = ref([]);
const loadingTasks = ref(false);
const showCreateTaskModal = ref(false);
const showEditTaskModal = ref(false);
const submittingTask = ref(false);

const newTask = ref({ title: '', description: '', priority: '2', deadline: '', assignee_id: null, predecessor_ids: [] });
const editingTask = ref({ id: null, title: '', description: '', status: 'TODO', priority: '2', deadline: '', assignee_id: null, predecessor_ids: [] });
const editingDependencies = ref([]); 

const members = ref([]);
const loadingMembers = ref(false);
const showAddMemberModal = ref(false);
const submittingMember = ref(false);
const newMemberName = ref('');

const newTeamName = ref('');
const renamingTeam = ref(false);
const dissolvingTeam = ref(false);

// 👇 Lab3 新增：转让团队状态
const newOwnerId = ref(null);
const transferringTeam = ref(false);

/* ================== 工具函数 ================== */
const statusLabel = (status) => ({ TODO: '待做', DOING: '进行中', DONE: '已完成' }[status] || status);
const priorityLabel = (p) => ({ 1: '低', 2: '中', 3: '高' }[p] || p);

const roleBadgeClass = (role) => {
  if (role === 'Owner') return 'priority-3';
  if (role === 'Admin') return 'priority-2';
  return 'priority-1';
};

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

const canEditTask = (task) => {
  if (myRole.value === 'Owner' || myRole.value === 'Admin') return true;
  const currentMember = members.value.find(m => m.username === authStore.username.value);
  return currentMember ? task.assignee_id === currentMember.user_id : false;
};

// 👇 Lab3 新增：细化移除成员权限校验
const canRemoveMember = (member) => {
  if (myRole.value === 'Owner' && member.role !== 'Owner') return true;
  if (myRole.value === 'Admin' && member.role === 'Member') return true;
  return false;
};

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
    if (res.data.code === 0) teamStore.setTeams(res.data.data.list || []);
  } catch (error) { console.error('加载团队列表失败', error); }
};

const loadTeamTasks = async () => {
  loadingTasks.value = true;
  try {
    const res = await getTeamTasksApi(teamId);
    if (res.data.code === 0) tasks.value = res.data.data.list || [];
  } catch (error) { console.error('加载任务失败', error); } 
  finally { loadingTasks.value = false; }
};

const loadTeamMembers = async () => {
  loadingMembers.value = true;
  try {
    const res = await getTeamMembersApi(teamId);
    if (res.data.code === 0) members.value = res.data.data || [];
  } catch (error) { console.error('加载成员失败', error); } 
  finally { loadingMembers.value = false; }
};

/* ================== 任务操作 (含依赖) ================== */
const handleCreateTask = async () => {
  if (submittingTask.value) return; 
  submittingTask.value = true;
  try {
    const payload = {
      title: newTask.value.title, description: newTask.value.description || null,
      priority: Number(newTask.value.priority), deadline: parseDateInput(newTask.value.deadline),
      team_id: teamId, status: 'TODO', assignee_id: newTask.value.assignee_id || null
    };
    
    const res = await createTaskApi(payload);
    if (res.data.code === 0) {
      const newTaskId = res.data.data?.id;
      if (newTaskId && newTask.value.predecessor_ids.length > 0) {
        for (const predId of newTask.value.predecessor_ids) {
          try { await addTaskDependencyApi(newTaskId, predId); } catch (err) { console.error('添加依赖失败', err); }
        }
      }
      showCreateTaskModal.value = false;
      newTask.value = { title: '', description: '', priority: '2', deadline: '', assignee_id: null, predecessor_ids: [] }; 
      await loadTeamTasks();
    } else alert(res.data.message || '发布失败');
  } catch (error) { alert(error.response?.data?.message || '请求失败'); } 
  finally { submittingTask.value = false; }
};

const startEdit = async (task) => {
  editingTask.value = {
    id: task.id, title: task.title, description: task.description || '',
    status: task.status, priority: String(task.priority), deadline: task.deadline ? formatDateForInput(task.deadline) : '',
    assignee_id: task.assignee_id || null, predecessor_ids: []
  };
  editingDependencies.value = [];
  try {
    const res = await getTaskDependenciesApi(task.id);
    if (res.data && res.data.code === 0) {
      editingDependencies.value = (res.data.data.predecessors || []).map(p => p.id);
      editingTask.value.predecessor_ids = [...editingDependencies.value];
    }
  } catch (err) { console.error('获取依赖失败', err); }
  showEditTaskModal.value = true;
};

const handleUpdateTask = async () => {
  if (submittingTask.value) return;
  submittingTask.value = true;
  try {
    const payload = myRole.value === 'Member'
      ? { status: editingTask.value.status }
      : {
          title: editingTask.value.title, description: editingTask.value.description || null,
          status: editingTask.value.status, priority: Number(editingTask.value.priority),
          deadline: parseDateInput(editingTask.value.deadline), assignee_id: editingTask.value.assignee_id || null
        };

    const res = await updateTaskApi(editingTask.value.id, payload);
    if (res.data.code === 0) {
      if (myRole.value === 'Owner' || myRole.value === 'Admin') {
        const toAdd = editingTask.value.predecessor_ids.filter(id => !editingDependencies.value.includes(id));
        const toRemove = editingDependencies.value.filter(id => !editingTask.value.predecessor_ids.includes(id));
        for (const id of toAdd) {
          try { await addTaskDependencyApi(editingTask.value.id, id); } 
          catch (err) { if (err.response?.status === 400) alert(`⚠️ 无法添加依赖: ${err.response.data.message}`); }
        }
        for (const id of toRemove) { try { await removeTaskDependencyApi(editingTask.value.id, id); } catch (e) {} }
      }
      showEditTaskModal.value = false;
      await loadTeamTasks();
    } else alert(res.data.message || '更新失败');
  } catch (error) {
    if (error.response?.status === 409) alert(`⚠️ 无法保存：${error.response.data.message || '尚有前置任务未完成'}`);
    else alert(error.response?.data?.message || '更新失败，可能是权限不足');
  } finally { submittingTask.value = false; }
};

const handleDeleteTask = async (taskId) => {
  if (!confirm('确定要删除这个团队任务吗？此操作不可撤销。')) return;
  try {
    const res = await deleteTaskApi(taskId);
    if (res.data.code === 0) await loadTeamTasks();
    else alert(res.data.message || '删除失败');
  } catch (error) { alert(error.response?.data?.message || '无权删除'); }
};

/* ================== 团队管理生命周期操作 ================== */

// 成员主动退出团队
const handleLeaveTeam = async () => {
  if (!confirm('确定要退出该团队吗？退出后您负责的未完成任务将退回团队池。')) return;
  try {
    const res = await leaveTeamApi(teamId);
    if (res.data.code === 0) {
      alert('已成功退出团队');
      await loadTeams();
      router.push('/teams');
    } else alert(res.data.message || '退出失败');
  } catch (error) {
    alert(error.response?.data?.message || '退出请求失败');
  }
};

// 转让团队所有权
const handleTransferTeam = async () => {
  if (!newOwnerId.value) return;
  if (!confirm('确定要将团队转让给该成员吗？转让后您将失去 Owner 权限。')) return;
  transferringTeam.value = true;
  try {
    const res = await transferTeamApi(teamId, { new_owner_id: newOwnerId.value });
    if (res.data.code === 0) {
      alert('团队转让成功！');
      newOwnerId.value = null;
      await loadTeams(); 
      await loadTeamMembers();
      activeTab.value = 'tasks'; // 失去 Owner 权限后自动切回任务大厅
    } else alert(res.data.message || '转让失败');
  } catch (error) {
    alert(error.response?.data?.message || '转让请求失败');
  } finally {
    transferringTeam.value = false;
  }
};

const handleRenameTeam = async () => {
  if (!newTeamName.value) return;
  renamingTeam.value = true;
  try {
    const res = await updateTeamApi(teamId, { name: newTeamName.value });
    if (res.data.code === 0) {
      alert('团队名称修改成功');
      newTeamName.value = '';
      await loadTeams();
    } else alert(res.data.message || '修改失败');
  } catch (error) { alert(error.response?.data?.message || '请求失败'); } 
  finally { renamingTeam.value = false; }
};

const handleDissolveTeam = async () => {
  if (!confirm('确定要解散该团队吗？此操作将删除所有相关任务数据且不可撤销！')) return;
  dissolvingTeam.value = true;
  try {
    const res = await dissolveTeamApi(teamId);
    if (res.data.code === 0) {
      alert('团队已解散');
      await loadTeams();
      router.push('/teams');
    } else alert(res.data.message || '解散失败');
  } catch (error) { alert(error.response?.data?.message || '请求失败'); } 
  finally { dissolvingTeam.value = false; }
};

/* ================== 成员邀请与移除操作 ================== */
const handleAddMember = async () => {
  if (!newMemberName.value || submittingMember.value) return;
  submittingMember.value = true;
  try {
    const res = await addTeamMemberApi(teamId, { username: newMemberName.value });
    if (res.data.code === 0) {
      showAddMemberModal.value = false;
      newMemberName.value = ''; 
      await loadTeamMembers();
      alert('邀请已发送！');
    } else alert(res.data.message || '邀请失败');
  } catch (error) { alert(error.response?.data?.message || '请求失败，可能用户不存在或已在团队中'); } 
  finally { submittingMember.value = false; }
};

const removeMember = async (userId) => {
  if (!confirm('确定要将该成员移出团队吗？其未完成的任务将被退回团队池。')) return;
  try {
    const res = await deleteTeamMemberApi(teamId, userId);
    if (res.data.code === 0) await loadTeamMembers();
    else alert(res.data.message || '移除失败');
  } catch (error) { alert(error.response?.data?.message || '无权操作或请求失败'); }
};

const toggleAdminRole = async (member) => {
  const newRole = member.role === 'Admin' ? 'Member' : 'Admin';
  try {
    const res = await updateTeamMemberRoleApi(teamId, member.user_id, { role: newRole });
    if (res.data.code === 0) await loadTeamMembers();
    else alert(res.data.message || '权限修改失败');
  } catch (error) { alert(error.response?.data?.message || '无权操作'); }
};

onMounted(() => {
  loadTeams();
  loadTeamTasks();
  loadTeamMembers();
});
</script>

<style scoped>
/* 样式依然原封不动保留！ */
.task-header { margin-bottom: 1.5rem; padding-bottom: 1rem; }
.task-title-row { display: flex; align-items: center; gap: 1rem; }
.back-btn { min-height: 38px; padding: 0 14px; border: 1px solid #c9ddff; border-radius: 999px; background: #f8fbff; color: #1d4ed8; cursor: pointer; font-size: 14px; font-weight: 600; }
.back-btn:hover { background: #eaf3ff; border-color: #accbff; }
.title-text h1 { margin: 0; font-size: 1.5rem; color: var(--text); }
.hint { margin: 4px 0 0; color: var(--muted); font-size: 0.85rem; }

.tabs-container { display: flex; gap: 1.5rem; border-bottom: 2px solid #e5e7eb; margin-bottom: 1.5rem; padding-left: 10px; }
.tab-btn { padding: 0.5rem 0.5rem 0.75rem; border: none; background: transparent; color: #64748b; font-size: 1rem; font-weight: 600; cursor: pointer; border-bottom: 3px solid transparent; transition: all 0.2s; margin-bottom: -2px; }
.tab-btn:hover { color: #3b82f6; }
.tab-btn.active { color: #2563eb; border-bottom-color: #2563eb; }

.list-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; }
.list-header h2 { font-size: 1.1rem; color: var(--text); font-weight: 700; margin: 0; }
.empty-state, .loading-state { text-align: center; padding: 3rem; color: #6b7280; background: #f8fafc; border-radius: 0.5rem; border: 1px dashed #cbd5e1; }

.task-list { display: flex; flex-direction: column; gap: 1rem; }
.task-item { display: flex; justify-content: space-between; align-items: center; padding: 1.2rem; background: #fff; border: 1px solid #e2e8f0; border-radius: 12px; transition: transform 0.2s, box-shadow 0.2s; }
.task-item:hover { transform: translateY(-2px); box-shadow: 0 10px 20px rgba(15, 23, 42, 0.08); border-color: #cbd5e1; }
.task-main { flex: 1; }
.task-title-section { display: flex; align-items: center; gap: 0.75rem; flex-wrap: wrap; margin-bottom: 0.4rem; }
.task-title-section h3 { margin: 0; font-size: 1.05rem; color: #0f172a; font-weight: 600; }
.task-description { margin: 0; font-size: 0.85rem; color: #64748b; }

.priority-badge { padding: 0.2rem 0.5rem; border-radius: 6px; font-size: 0.75rem; font-weight: 600; }
.priority-1 { background: #e0e7ff; color: #4338ca; }
.priority-2 { background: #fef08a; color: #92400e; }
.priority-3 { background: #fee2e2; color: #b91c1c; }
.status-badge { padding: 0.2rem 0.5rem; border-radius: 6px; font-size: 0.75rem; font-weight: 600; }
.status-todo { background: #dbeafe; color: #1e40af; }
.status-doing { background: #fef08a; color: #92400e; }
.status-done { background: #dcfce7; color: #166534; }

.btn { padding: 0.5rem 1rem; font-size: 0.85rem; font-weight: 600; border-radius: 8px; cursor: pointer; border: none; transition: all 0.2s ease; }
.btn-small { padding: 0.4rem 0.8rem; font-size: 0.8rem; }
.btn-primary { background: #3b82f6; color: white; }
.btn-primary:hover { background: #2563eb; }
.btn-complete { background: linear-gradient(135deg, #22c55e, #16a34a); color: white; box-shadow: 0 4px 10px rgba(34,197,94,0.2); }
.btn-complete:hover { transform: translateY(-1px); box-shadow: 0 6px 14px rgba(34,197,94,0.3); }
.btn-outline { background: #f1f5f9; color: #475569; border: 1px solid #cbd5e1; }
.btn-outline:hover { background: #e2e8f0; }
.btn-dependency {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  color: #1e40af;
  border: 1px solid #93c5fd;
  text-align: center;
  text-decoration: none;
  line-height: 1;
}
.btn-dependency:hover { background: #eff6ff; border-color: #60a5fa; color: #1d4ed8; }
.btn-edit { background: #dbeafe; color: #1e40af; border: 1px solid #93c5fd; }
.btn-edit:hover:not(:disabled) { background: #bfdbfe; }
.btn-danger { background: #fee2e2; color: #b91c1c; border: 1px solid #fecaca; }
.btn-danger:hover { background: #fecaca; }

.modal-overlay { 
  position: fixed; top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.45); backdrop-filter: blur(3px); 
  display: flex; justify-content: center; align-items: center; z-index: 9999; 
}
.modal { 
  background: #fff; padding: 2.5rem; border-radius: 16px; width: 90%; max-width: 650px; 
  max-height: 90vh; overflow-y: auto; box-shadow: 0 24px 50px rgba(0, 0, 0, 0.2); 
  animation: modal-pop 0.3s cubic-bezier(0.16, 1, 0.3, 1); 
}
.modal h2 { margin-top: 0; color: var(--text); text-align: center; margin-bottom: 1.5rem; font-size: 1.4rem; }

.task-form { display: flex; flex-direction: column; gap: 1.25rem; }
.form-input { padding: 0.8rem; border: 1px solid #cbd5e1; border-radius: 8px; font-family: inherit; font-size: 0.95rem; }
.form-input:focus { outline: none; border-color: #3b82f6; box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15); }
.form-textarea { min-height: 120px; resize: vertical; line-height: 1.5; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.modal-actions { display: flex; gap: 1rem; margin-top: 1.5rem; }
.modal-actions button { flex: 1; padding: 0.8rem; border-radius: 8px; font-weight: 600; cursor: pointer; border: none; transition: all 0.2s ease; }

@keyframes modal-pop {
  0% { opacity: 0; transform: scale(0.95) translateY(10px); }
  100% { opacity: 1; transform: scale(1) translateY(0); }
}

.text-red-600 { color: #dc2626; }
.text-orange-500 { color: #f97316; }
.text-blue-600 { color: #2563eb; }
.font-bold { font-weight: bold; }

.manage-section { display: flex; flex-direction: column; gap: 1.5rem; }
.manage-card { background: #fff; border: 1px solid #e2e8f0; border-radius: 12px; padding: 1.5rem; }
.manage-card h3 { margin: 0 0 0.75rem; font-size: 1.05rem; color: #0f172a; }
.manage-card p { color: #64748b; font-size: 0.85rem; margin: 0 0 1rem; }
.rename-form { display: flex; gap: 0.75rem; align-items: center; }
.rename-form .form-input { flex: 1; }
.danger-zone { border-color: #fecaca; background: #fef2f2; }
.danger-zone h3 { color: #b91c1c; }

:global([data-theme="dark"]) .manage-card { background: #1e293b; border-color: #334155; }
:global([data-theme="dark"]) .manage-card h3 { color: #f8fafc; }
:global([data-theme="dark"]) .manage-card p { color: #94a3b8; }
:global([data-theme="dark"]) .danger-zone { border-color: #7f1d1d; background: #1a0f0f; }
:global([data-theme="dark"]) .danger-zone h3 { color: #fca5a5; }

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
