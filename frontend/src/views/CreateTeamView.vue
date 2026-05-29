<template>
  <section class="card create-task-page">
    <header class="header">
      <button class="back-btn" @click="router.back()">← 返回</button>
      <h1>新建团队</h1>
    </header>

    <form class="form" @submit.prevent="onCreateTeam">
      <label>
        团队名称
        <input 
          v-model.trim="newTeamName" 
          type="text" 
          placeholder="此处输入团队名称" 
          required 
          class="form-input"
        />
      </label>

      <div class="actions">
        <button class="btn cancel-btn" type="button" @click="router.back()">取消</button>
        <button class="btn" type="submit" :disabled="submitting">
          {{ submitting ? '创建中...' : '确认创建' }}
        </button>
      </div>

      <p v-if="createError" style="color: #ef4444; font-size: 13px; margin-top: 15px; font-weight: 600;">
        {{ createError }}
      </p>
    </form>
  </section>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { createTeamApi } from '../services/team';
import { getApiErrorMessage, getApiFailureMessage } from '../utils/apiError';

const router = useRouter();
const submitting = ref(false);
const newTeamName = ref('');
const createError = ref('');
const TEAM_NAME_MAX_LENGTH = 100;
const TEAM_NAME_INVALID_MESSAGE = '团队名称不能为空且不能超过 100 个字符';

const onCreateTeam = async () => {
  const teamName = newTeamName.value.trim();
  if (!teamName || teamName.length > TEAM_NAME_MAX_LENGTH) {
    createError.value = TEAM_NAME_INVALID_MESSAGE;
    return;
  }
  
  submitting.value = true;
  createError.value = '';

  try {
    const res = await createTeamApi({ name: teamName });
    if (res.data.code === 0) {
      // 创建成功后自动跳回团队列表
      router.push('/teams');
    } else {
      createError.value = getApiFailureMessage(res, '创建团队失败', '创建失败');
    }
  } catch (error) {
    createError.value = getApiErrorMessage(error, '创建团队失败');
  } finally {
    submitting.value = false;
  }
};
</script>

<style scoped>
/* =========================================
   基础高定样式 (浅色模式)
========================================== */
.create-task-page {
  position: relative; max-width: 700px; margin: 26px auto; padding: 36px 34px;
  border-radius: 20px; border: 1px solid rgba(148, 163, 184, 0.25);
  background: linear-gradient(155deg, rgba(255, 255, 255, 0.97), rgba(196, 221, 226, 0.874));
  box-shadow: 0 18px 42px rgba(16, 25, 45, 0.1); overflow: hidden;
  transition: all 0.3s ease;
}

.create-task-page::before, .create-task-page::after {
  content: ''; position: absolute; border-radius: 999px; pointer-events: none; z-index: 0;
  transition: all 0.3s ease;
}
.create-task-page::before {
  width: 240px; height: 240px; top: -120px; right: -80px;
  background: radial-gradient(circle at center, rgba(56, 189, 248, 0.18), rgba(56, 189, 248, 0));
}
.create-task-page::after {
  width: 280px; height: 280px; bottom: -170px; left: -120px;
  background: radial-gradient(circle at center, rgba(59, 130, 246, 0.14), rgba(59, 130, 246, 0));
}

.header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 28px; gap: 16px; position: relative; z-index: 1; }
.back-btn {
  min-height: 38px; padding: 0 14px; border: 1px solid #c9ddff; border-radius: 999px; background: #f8fbff;
  color: #1d4ed8; cursor: pointer; font-size: 14px; font-weight: 600; transition: all 0.25s ease;
}
.back-btn:hover { background: #eaf3ff; border-color: #accbff; transform: translateY(-1px); }

/* 👇 核心修改：无论浅色深色，“新建团队”四个字被死死锁定为纯黑色 */
.header h1 { 
  margin: 0; font-size: 30px; font-weight: 700; letter-spacing: 0.02em; 
  color: #0f172a !important; 
}

.form { display: grid; gap: 20px; position: relative; z-index: 1; }
.form label { display: grid; gap: 8px; color: #334155; font-size: 14px; font-weight: 600; }
.form-input {
  width: 100%; height: 44px; padding: 0 12px; border: 1px solid #d7e2ef; border-radius: 12px;
  background: rgba(255, 255, 255, 0.92); color: #0f172a; font-size: 14px;
  transition: border-color 0.22s ease, box-shadow 0.22s ease, background-color 0.22s ease;
  box-sizing: border-box;
}
.form-input:focus { outline: none; border-color: #60a5fa; background: #fff; box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.13); }

.actions { display: flex; justify-content: flex-end; gap: 12px; margin-top: 10px; padding-top: 20px; border-top: 1px solid #e5ecf4; }
.actions .btn { min-width: 116px; height: 42px; border-radius: 12px; border: none; font-size: 14px; font-weight: 600; cursor: pointer; transition: all 0.25s ease; }
.actions .btn:not(.cancel-btn) { color: #fff; background: linear-gradient(135deg, #2563eb, #1d4ed8); box-shadow: 0 8px 16px rgba(37, 99, 235, 0.28); }
.actions .btn:not(.cancel-btn):hover { transform: translateY(-1px); box-shadow: 0 10px 20px rgba(37, 99, 235, 0.34); }
.actions .btn:disabled { opacity: 0.7; cursor: not-allowed; box-shadow: none; }
.cancel-btn { background: #f4f7fb; color: #475569; border: 1px solid #d9e3ef; }
.cancel-btn:hover { background: #edf2f8; }

/* =========================================
   深色模式强制接管区 (绝对生效)
========================================== */
:global([data-theme="dark"]) .create-task-page {
  background: linear-gradient(155deg, rgba(30, 41, 59, 1), rgba(15, 23, 42, 0.98)) !important;
  border-color: #334155 !important;
  box-shadow: 0 18px 42px rgba(0, 0, 0, 0.5) !important;
}

:global([data-theme="dark"]) .create-task-page::before {
  background: radial-gradient(circle at center, rgba(56, 189, 248, 0.08), rgba(56, 189, 248, 0)) !important;
}
:global([data-theme="dark"]) .create-task-page::after {
  background: radial-gradient(circle at center, rgba(59, 130, 246, 0.08), rgba(59, 130, 246, 0)) !important;
}

:global([data-theme="dark"]) .form label { 
  color: #94a3b8 !important; 
}

:global([data-theme="dark"]) .form-input { 
  background: rgba(15, 23, 42, 0.8) !important; 
  color: #f8fafc !important; 
  border-color: #334155 !important; 
}
:global([data-theme="dark"]) .form-input:focus {
  border-color: #3b82f6 !important;
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.15) !important;
}

:global([data-theme="dark"]) .back-btn { 
  background: #1e293b !important; color: #93c5fd !important; border-color: #334155 !important; 
}
:global([data-theme="dark"]) .back-btn:hover { 
  background: #0f172a !important; 
}
:global([data-theme="dark"]) .cancel-btn { 
  background: #1e293b !important; color: #cbd5e1 !important; border-color: #334155 !important; 
}
:global([data-theme="dark"]) .cancel-btn:hover { 
  background: #0f172a !important; 
}
</style>
