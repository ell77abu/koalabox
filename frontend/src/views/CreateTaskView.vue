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
        <label>
          优先级
          <select v-model.number="newTask.priority">
            <option :value="1">低</option>
            <option :value="2">中</option>
            <option :value="3">高</option>
          </select>
        </label>

        <label>
          截止日期
          <input v-model="newTask.deadline" type="datetime-local" />
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
import { reactive, ref } from 'vue';
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

const onCreateTask = async () => {
  if (!newTask.title) return;
  
  submitting.value = true;
  try {
    // 转换日期格式为后端需要的字符串（如果后端需要 ISO 格式）
    const payload = {
      ...newTask,
      deadline: newTask.deadline ? new Date(newTask.deadline).toISOString() : null,
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
  max-width: 800px;
  margin: 20px auto;
  padding: 30px;
}

.header {
  display: flex;
  align-items: center;
  margin-bottom: 30px;
  gap: 20px;
}

.back-btn {
  background: none;
  border: none;
  color: #1890ff;
  cursor: pointer;
  font-size: 16px;
  padding: 5px 10px;
}

.header h1 {
  margin: 0;
  font-size: 24px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 15px;
  margin-top: 40px;
  padding-top: 20px;
  border-top: 1px solid #eee;
}

.cancel-btn {
  background-color: #f5f5f5;
  color: #666;
}

.cancel-btn:hover {
  background-color: #e8e8e8;
}

textarea {
  width: 100%;
  padding: 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-family: inherit;
  resize: vertical;
}
</style>
