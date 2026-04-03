import { ref } from 'vue';

const tasks = ref([]);
const loading = ref(false);
const total = ref(0);

const setTasks = (newTasks) => {
  tasks.value = Array.isArray(newTasks) ? newTasks : [];
};

const setLoading = (isLoading) => {
  loading.value = isLoading;
};

const setTotal = (totalCount) => {
  total.value = totalCount;
};

const clearTasks = () => {
  tasks.value = [];
  total.value = 0;
};

export const taskStore = {
  tasks,
  loading,
  total,
  setTasks,
  setLoading,
  setTotal,
  clearTasks,
};
