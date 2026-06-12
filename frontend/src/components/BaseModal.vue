<template>
  <Teleport to="body">
    <div
      v-if="modelValue"
      class="base-modal-overlay"
      role="presentation"
      @click="close"
    >
      <section
        class="base-modal"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        @click.stop
      >
        <header class="base-modal-header">
          <h2 :id="titleId">{{ title }}</h2>
          <button
            type="button"
            class="base-modal-close"
            aria-label="关闭弹窗"
            title="关闭"
            @click="close"
          >
            x
          </button>
        </header>

        <div class="base-modal-body">
          <slot />
        </div>

        <footer v-if="$slots.footer" class="base-modal-footer">
          <slot name="footer" />
        </footer>
      </section>
    </div>
  </Teleport>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  modelValue: {
    type: Boolean,
    required: true,
  },
  title: {
    type: String,
    default: '弹窗',
  },
});

const emit = defineEmits(['update:modelValue', 'close']);

const titleId = computed(() => `modal-title-${props.title.replace(/\s+/g, '-').toLowerCase()}`);

const close = () => {
  emit('update:modelValue', false);
  emit('close');
};
</script>

<style scoped>
.base-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  background: rgba(15, 23, 42, 0.46);
  backdrop-filter: blur(3px);
}

.base-modal {
  width: min(100%, 680px);
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 24px 50px rgba(15, 23, 42, 0.22);
}

.base-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid #e2e8f0;
}

.base-modal-header h2 {
  margin: 0;
  color: #0f172a;
  font-size: 1.15rem;
  line-height: 1.3;
}

.base-modal-close {
  width: 34px;
  height: 34px;
  border: 1px solid #cbd5e1;
  border-radius: 999px;
  background: #f8fafc;
  color: #475569;
  cursor: pointer;
  font-size: 1rem;
  line-height: 1;
}

.base-modal-close:hover {
  background: #e2e8f0;
  color: #0f172a;
}

.base-modal-body {
  overflow-y: auto;
  padding: 1.5rem;
}

.base-modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid #e2e8f0;
}

:global([data-theme="dark"]) .base-modal {
  background: #1e293b;
}

:global([data-theme="dark"]) .base-modal-header,
:global([data-theme="dark"]) .base-modal-footer {
  border-color: #334155;
}

:global([data-theme="dark"]) .base-modal-header h2 {
  color: #f8fafc;
}

:global([data-theme="dark"]) .base-modal-close {
  background: #0f172a;
  color: #cbd5e1;
  border-color: #334155;
}
</style>
