import { createApp } from 'vue';
import App from './App.vue';
import router from './router';
import { getSessionApi } from './services/auth';
import { authStore } from './stores/auth';
import { getApiErrorMessage } from './utils/apiError';
import './styles.css';

const bootstrapSession = async () => {
	try {
		const response = await getSessionApi();
		const payload = response.data?.data;
		if (payload?.username) {
			authStore.setAuth(payload.username, payload.level ?? 0);
			return;
		}
	} catch (error) {
		if (error?.response?.status !== 401) {
			console.warn(getApiErrorMessage(error, '恢复登录状态失败'));
		}
	}
	authStore.clearAuth();
};

await bootstrapSession();

createApp(App).use(router).mount('#app');
