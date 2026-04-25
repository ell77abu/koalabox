import { createApp } from 'vue';
import App from './App.vue';
import router from './router';
import { getSessionApi } from './services/auth';
import { authStore } from './stores/auth';
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
		// No valid session found, keep guest state.
	}
	authStore.clearAuth();
};

await bootstrapSession();

createApp(App).use(router).mount('#app');
