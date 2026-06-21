<template>
  <div class="page-shell">
    <section class="card card-wide">
      <header class="trash-page-header">
        <button class="back-btn" type="button" @click="router.back()">← 返回</button>
        <div>
          <h1>回收站</h1>
          <p class="hint">查看个人已删除任务，并按需恢复。</p>
        </div>
      </header>

      <TaskTrashPanel
        title="个人任务回收站"
        subtitle="这里展示你的个人已删除任务，恢复后会回到任务列表。"
        :can-restore="true"
      />
    </section>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router';
import TaskTrashPanel from '../components/TaskTrashPanel.vue';

const router = useRouter();
</script>

<style scoped>
.trash-page-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #e5e7eb;
}

.trash-page-header h1 {
  margin: 0;
  color: var(--text);
  font-size: 1.5rem;
}

.hint {
  margin: 4px 0 0;
  color: var(--muted);
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
}

.back-btn:hover {
  background: #eaf3ff;
  border-color: #accbff;
}
</style>
