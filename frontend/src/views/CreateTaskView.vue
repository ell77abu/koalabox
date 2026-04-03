<template>
  <section class="card create-task-page">
    <header class="header">
      <button class="back-btn" @click="router.back()">← 返回</button>
      <h1>新建任务</h1>
    </header>

    <form class="form" @submit.prevent="onCreateTask">
      <label>
        任务标题
        <input v-model.trim="newTask.title" type="text" placeholder="想要做什么？" required />
      </label>

      <label>
        任务描述
        <textarea v-model.trim="newTask.description" placeholder="添加更多细节..." rows="4"></textarea>
      </label>

      <div class="form-row">
        <label class="priority-field">
          优先级
          <div ref="priorityDropdownRef" class="priority-dropdown">
            <button
              type="button"
              class="priority-trigger"
              :aria-expanded="priorityOpen"
              @click="togglePriorityMenu"
            >
              <span class="priority-left">
                <span class="priority-dot" :class="`level-${newTask.priority}`"></span>
                <span>{{ getPriorityLabel(newTask.priority) }}</span>
              </span>
              <span class="priority-caret" :class="{ open: priorityOpen }"></span>
            </button>

            <transition name="priority-menu-fade">
              <ul v-if="priorityOpen" class="priority-menu">
                <li v-for="option in priorityOptions" :key="option.value">
                  <button
                    type="button"
                    class="priority-option"
                    :class="{ active: newTask.priority === option.value }"
                    @click="selectPriority(option.value)"
                  >
                    <span class="priority-left">
                      <span class="priority-dot" :class="`level-${option.value}`"></span>
                      <span>{{ option.label }}优先级</span>
                    </span>
                    <span class="priority-tag" :class="`tag-${option.value}`">{{ option.tag }}</span>
                  </button>
                </li>
              </ul>
            </transition>
          </div>
        </label>

        <label>
          截止日期
          <input
            v-model.trim="newTask.deadline"
            type="datetime-local"
            class="date-input"
          />
        </label>
      </div>

      <div class="actions">
        <button class="btn cancel-btn" type="button" @click="router.back()">取消</button>
        <button class="btn" type="submit" :disabled="submitting">
          {{ submitting ? '创建中...' : '确认添加' }}
        </button>
      </div>
    </form>
  </section>
</template>

<script setup>
import { onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { createTaskApi } from '../services/task';

const router = useRouter();
const submitting = ref(false);

const newTask = reactive({
  title: '',
  description: '',
  priority: 2,
  deadline: '',
});

const priorityOpen = ref(false);
const priorityDropdownRef = ref(null);

const priorityOptions = [
  { value: 1, label: '低', tag: '轻松' },
  { value: 2, label: '中', tag: '平衡' },
  { value: 3, label: '高', tag: '紧急' },
];

const getPriorityLabel = (value) => {
  const option = priorityOptions.find((item) => item.value === value);
  return option ? `${option.label}优先级` : '请选择优先级';
};

const togglePriorityMenu = () => {
  priorityOpen.value = !priorityOpen.value;
};

const selectPriority = (value) => {
  newTask.priority = value;
  priorityOpen.value = false;
};

const handleOutsideClick = (event) => {
  if (priorityDropdownRef.value && !priorityDropdownRef.value.contains(event.target)) {
    priorityOpen.value = false;
  }
};

const parseDateInput = (value) => {
  if (!value) return null;
  const normalized = String(value).trim().replace(/\//g, '-');
  const match = normalized.match(/^(\d{4})-(\d{2})-(\d{2})(?:T(\d{2}):(\d{2}))?$/);
  if (!match) return null;
  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3]);
  const hour = match[4] ? Number(match[4]) : 0;
  const minute = match[5] ? Number(match[5]) : 0;
  if (!year || !month || !day) return null;
  return { year, month, day, hour, minute };
};

const toBackendDeadline = (value) => {
  const parts = parseDateInput(value);
  if (!parts) return null;
  const date = new Date(Date.UTC(parts.year, parts.month - 1, parts.day, parts.hour, parts.minute, 0));
  return date.toISOString();
};

onMounted(() => {
  document.addEventListener('click', handleOutsideClick);
});

onBeforeUnmount(() => {
  document.removeEventListener('click', handleOutsideClick);
});

const onCreateTask = async () => {
  if (!newTask.title) return;
  
  submitting.value = true;
  try {
    // 转换日期格式为后端需要的字符串（如果后端需要 ISO 格式）
    const payload = {
      ...newTask,
      deadline: toBackendDeadline(newTask.deadline),
      status: 'TODO'
    };
    
    await createTaskApi(payload);
    router.push('/tasks');
  } catch (error) {
    alert(error.response?.data?.message || '创建任务失败');
  } finally {
    submitting.value = false;
  }
};
</script>

<style scoped>
.create-task-page {
  position: relative;
  max-width: 900px;
  margin: 26px auto;
  padding: 36px 34px;
  border-radius: 20px;
  border: 1px solid rgba(148, 163, 184, 0.25);
  background: linear-gradient(155deg, rgba(255, 255, 255, 0.97), rgba(196, 221, 226, 0.874));
  box-shadow: 0 18px 42px rgba(16, 25, 45, 0.1);
  overflow: hidden;
}

.create-task-page::before,
.create-task-page::after {
  content: '';
  position: absolute;
  border-radius: 999px;
  pointer-events: none;
  z-index: 0;
}

.create-task-page::before {
  width: 240px;
  height: 240px;
  top: -120px;
  right: -80px;
  background: radial-gradient(circle at center, rgba(56, 189, 248, 0.18), rgba(56, 189, 248, 0));
}

.create-task-page::after {
  width: 280px;
  height: 280px;
  bottom: -170px;
  left: -120px;
  background: radial-gradient(circle at center, rgba(59, 130, 246, 0.14), rgba(59, 130, 246, 0));
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 28px;
  gap: 16px;
  position: relative;
  z-index: 1;
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

.back-btn:hover {
  background: #eaf3ff;
  border-color: #accbff;
  transform: translateY(-1px);
}

.header h1 {
  margin: 0;
  font-size: 30px;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: #0f172a;
}

.form {
  display: grid;
  gap: 20px;
  position: relative;
  z-index: 1;
}

.form label {
  display: grid;
  gap: 8px;
  color: #334155;
  font-size: 14px;
  font-weight: 600;
}

.form input,
.form textarea,
.form select {
  width: 100%;
  border: 1px solid #d7e2ef;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.92);
  color: #0f172a;
  font-size: 14px;
  transition: border-color 0.22s ease, box-shadow 0.22s ease, background-color 0.22s ease;
}

.form input,
.form select {
  height: 44px;
  padding: 0 12px;
}

.form textarea {
  min-height: 108px;
  padding: 12px;
  resize: vertical;
}

.form input:focus,
.form textarea:focus,
.form select:focus {
  outline: none;
  border-color: #60a5fa;
  background: #fff;
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.13);
}

.form-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: clamp(14px, 2vw, 20px);
  align-items: start;
}

.form-row > label {
  padding: 14px 14px 12px;
  border-radius: 14px;
  border: 1px solid rgba(203, 213, 225, 0.75);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.92), rgba(248, 251, 255, 0.88));
  box-shadow: 0 6px 16px rgba(15, 23, 42, 0.06);
  transition: border-color 0.22s ease, box-shadow 0.22s ease, transform 0.22s ease;
}

.form-row > label:focus-within {
  border-color: #93c5fd;
  box-shadow: 0 10px 24px rgba(37, 99, 235, 0.14);
  transform: translateY(-1px);
}

.priority-field {
  position: relative;
}

.priority-dropdown {
  position: relative;
}

.priority-trigger {
  width: 100%;
  min-height: 44px;
  padding: 0 12px;
  border: 1px solid #d7e2ef;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.94);
  color: #0f172a;
  font-size: 14px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  transition: border-color 0.22s ease, box-shadow 0.22s ease, background-color 0.22s ease;
}

.priority-trigger:hover {
  border-color: #bfd2ec;
  background: #fff;
}

.priority-trigger:focus-visible {
  outline: none;
  border-color: #60a5fa;
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.13);
}

.priority-left {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.priority-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.priority-dot.level-1 {
  background: #22c55e;
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.16);
}

.priority-dot.level-2 {
  background: #f59e0b;
  box-shadow: 0 0 0 3px rgba(245, 158, 11, 0.16);
}

.priority-dot.level-3 {
  background: #ef4444;
  box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.16);
}

.priority-caret {
  width: 9px;
  height: 9px;
  border-right: 2px solid #64748b;
  border-bottom: 2px solid #64748b;
  transform: rotate(45deg);
  transition: transform 0.2s ease;
}

.priority-caret.open {
  transform: rotate(225deg);
}

.priority-menu {
  list-style: none;
  margin: 8px 0 0;
  padding: 8px;
  position: absolute;
  left: 0;
  right: 0;
  z-index: 20;
  border: 1px solid #d5e1f0;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.98);
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.14);
  backdrop-filter: blur(6px);
}

.priority-option {
  width: 100%;
  min-height: 42px;
  border: 1px solid transparent;
  border-radius: 10px;
  background: transparent;
  color: #1e293b;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 10px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.18s ease;
}

.priority-option:hover {
  background: #f3f8ff;
  border-color: #dbeafe;
}

.priority-option.active {
  background: #eaf2ff;
  border-color: #c7ddff;
}

.priority-tag {
  min-width: 44px;
  padding: 2px 8px;
  border-radius: 999px;
  text-align: center;
  font-size: 12px;
  font-weight: 700;
}

.priority-tag.tag-1 {
  color: #15803d;
  background: rgba(34, 197, 94, 0.14);
}

.priority-tag.tag-2 {
  color: #b45309;
  background: rgba(245, 158, 11, 0.18);
}

.priority-tag.tag-3 {
  color: #b91c1c;
  background: rgba(239, 68, 68, 0.16);
}

.priority-menu-fade-enter-active,
.priority-menu-fade-leave-active {
  transition: opacity 0.16s ease, transform 0.16s ease;
}

.priority-menu-fade-enter-from,
.priority-menu-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid #e5ecf4;
}

.actions .btn {
  min-width: 116px;
  height: 42px;
  border-radius: 12px;
  border: none;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.25s ease;
}

.actions .btn:not(.cancel-btn) {
  color: #fff;
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  box-shadow: 0 8px 16px rgba(37, 99, 235, 0.28);
}

.actions .btn:not(.cancel-btn):hover {
  transform: translateY(-1px);
  box-shadow: 0 10px 20px rgba(37, 99, 235, 0.34);
}

.actions .btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  box-shadow: none;
}

.cancel-btn:hover {
  background: #edf2f8;
}

.cancel-btn {
  background: #f4f7fb;
  color: #475569;
  border: 1px solid #d9e3ef;
}

@media (max-width: 860px) {
  .create-task-page {
    margin: 18px auto;
    padding: 30px 24px;
  }

  .form-row {
    grid-template-columns: 1fr;
    gap: 16px;
  }
}

@media (max-width: 640px) {
  .create-task-page {
    padding: 24px 18px;
    border-radius: 16px;
  }

  .header {
    flex-direction: column;
    align-items: flex-start;
    margin-bottom: 22px;
  }

  .header h1 {
    font-size: 24px;
  }

  .actions {
    justify-content: stretch;
  }

  .actions .btn {
    flex: 1;
  }
}
</style>
