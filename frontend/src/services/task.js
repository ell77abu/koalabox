import { http } from './http';

/**
 * 获取任务列表
 * @param {number} page - 页码
 * @param {number} size - 每页数量
 * @returns {Promise}
 */
export const getTasksApi = (page = 1, size = 10, sort = 'default') =>
  http.get('/tasks', { params: { page, size, sort } });

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

/**
 * 为任务添加前置依赖
 * @param {number} taskId - 当前任务 ID
 * @param {number} predecessorId - 前置任务 ID
 * @returns {Promise}
 */
export const addTaskDependencyApi = (taskId, predecessorId) => 
  http.post(`/tasks/${taskId}/dependencies`, { predecessor_id: predecessorId });

/**
 * 移除任务的前置依赖
 * @param {number} taskId - 当前任务 ID
 * @param {number} predecessorId - 要移除的前置任务 ID
 * @returns {Promise}
 */
export const removeTaskDependencyApi = (taskId, predecessorId) => 
  http.delete(`/tasks/${taskId}/dependencies/${predecessorId}`);

/**
 * 获取任务的直接依赖树 (直接前置与直接后继)
 * @param {number} taskId - 任务 ID
 * @returns {Promise}
 */
export const getTaskDependenciesApi = (taskId) => 
  http.get(`/tasks/${taskId}/dependencies`);

/**
 * 获取任务的全量依赖图数据 (用于前端绘制无环图 DAG)
 * @param {number} taskId - 任务 ID
 * @returns {Promise}
 */
export const getTaskDependencyGraphApi = (taskId) => 
  http.get(`/tasks/${taskId}/dependency-graph`);