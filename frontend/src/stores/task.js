import { ref, computed } from 'vue';

const tasks = ref([]);
const loading = ref(false);
const currentPage = ref(1);
const pageSize = ref(10);
const total = ref(0);

const hasMore = computed(() => currentPage.value * pageSize.value < total.value);

const setTasks = (newTasks) => {
  tasks.value = newTasks;
};

const setLoading = (isLoading) => {
  loading.value = isLoading;
};

const setPageInfo = (page, size, totalCount) => {
  currentPage.value = page;
  pageSize.value = size;
  total.value = totalCount;
};

const clearTasks = () => {
  tasks.value = [];
  currentPage.value = 1;
  total.value = 0;
};

export const taskStore = {
  tasks,
  loading,
  currentPage,
  pageSize,
  total,
  hasMore,
  setTasks,
  setLoading,
  setPageInfo,
  clearTasks,
};
