<template>
  <BaseModal v-model="visible" :title="title">
    <p class="message-text">{{ message }}</p>

    <template #footer>
      <button type="button" class="message-ok" @click="visible = false">确定</button>
    </template>
  </BaseModal>
</template>

<script setup>
import { computed } from 'vue';
import BaseModal from './BaseModal.vue';

const props = defineProps({
  modelValue: {
    type: Boolean,
    required: true,
  },
  title: {
    type: String,
    default: '提示',
  },
  message: {
    type: String,
    default: '',
  },
});

const emit = defineEmits(['update:modelValue']);

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
});
</script>

<style scoped>
.message-text {
  margin: 0;
  color: #334155;
  font-size: 0.95rem;
  line-height: 1.7;
  white-space: pre-wrap;
}

.message-ok {
  min-width: 96px;
  padding: 0.65rem 1rem;
  border: none;
  border-radius: 8px;
  background: #2563eb;
  color: #fff;
  font-weight: 600;
  cursor: pointer;
}

.message-ok:hover {
  background: #1d4ed8;
}

:global([data-theme="dark"]) .message-text {
  color: #e2e8f0;
}
</style>
