import { ref } from 'vue';

const teams = ref([]); // 当前用户所在的团队列表
const currentTeamMembers = ref([]); // 当前正在查看的团队的成员列表
const loading = ref(false);

const setTeams = (newTeams) => {
  teams.value = Array.isArray(newTeams) ? newTeams : [];
};

const setCurrentTeamMembers = (members) => {
  currentTeamMembers.value = Array.isArray(members) ? members : [];
};

const setLoading = (isLoading) => {
  loading.value = isLoading;
};

const clearTeams = () => {
  teams.value = [];
  currentTeamMembers.value = [];
};

export const teamStore = {
  teams,
  currentTeamMembers,
  loading,
  setTeams,
  setCurrentTeamMembers,
  setLoading,
  clearTeams,
};