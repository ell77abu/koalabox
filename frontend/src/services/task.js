import { http } from './http';

/**
 * 获取任务列表
 * @param {number} page - 页码
 * @param {number} size - 每页数量
 * @returns {Promise}
 */
export const getTasksApi = (page = 1, size = 10) =>
  http.get('/tasks', { params: { page, size } });

/**
 * 创建任务
 * @param {Object} payload - 任务数据
 * @returns {Promise}
 */
export const createTaskApi = (payload) => http.post('/tasks', payload);

/**
 * 更新任务
 * @param {number} id - 任务 ID
 * @param {Object} payload - 更新的任务数据
 * @returns {Promise}
 */
export const updateTaskApi = (id, payload) => http.put(`/tasks/${id}`, payload);

/**
 * 删除任务
 * @param {number} id - 任务 ID
 * @returns {Promise}
 */
export const deleteTaskApi = (id) => http.delete(`/tasks/${id}`);
