<template>
  <BaseModal v-model="visible" :title="title">
    <p class="confirm-text">{{ message }}</p>

    <template #footer>
      <button type="button" class="confirm-cancel" @click="visible = false">
        {{ cancelText }}
      </button>
      <button type="button" class="confirm-ok" @click="confirm">
        {{ confirmText }}
      </button>
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
    default: '确认操作',
  },
  message: {
    type: String,
    default: '',
  },
  confirmText: {
    type: String,
    default: '确定',
  },
  cancelText: {
    type: String,
    default: '取消',
  },
});

const emit = defineEmits(['update:modelValue', 'confirm']);

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
});

const confirm = () => {
  emit('confirm');
  visible.value = false;
};
</script>

<style scoped>
.confirm-text {
  margin: 0;
  color: #334155;
  font-size: 0.95rem;
  line-height: 1.7;
  white-space: pre-wrap;
}

.confirm-cancel,
.confirm-ok {
  min-width: 96px;
  padding: 0.65rem 1rem;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
}

.confirm-cancel {
  border: 1px solid #cbd5e1;
  background: #f8fafc;
  color: #475569;
}

.confirm-cancel:hover {
  background: #e2e8f0;
}

.confirm-ok {
  border: none;
  background: #dc2626;
  color: #fff;
}

.confirm-ok:hover {
  background: #b91c1c;
}

:global([data-theme="dark"]) .confirm-text {
  color: #e2e8f0;
}

:global([data-theme="dark"]) .confirm-cancel {
  background: #0f172a;
  color: #cbd5e1;
  border-color: #334155;
}
</style>
