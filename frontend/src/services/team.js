import { http } from './http';

export const createTeamApi = (payload) => http.post('/teams', payload);

export const getTeamsApi = (params = undefined) => http.get('/teams', { params });

export const getTeamMembersApi = (teamId) => http.get(`/teams/${teamId}/members`);

export const addTeamMemberApi = (teamId, payload) => http.post(`/teams/${teamId}/invitations`, payload);

export const deleteTeamMemberApi = (teamId, userId) => http.delete(`/teams/${teamId}/members/${userId}`);

export const updateTeamMemberRoleApi = (teamId, userId, payload) => http.put(`/teams/${teamId}/members/${userId}`, payload);

export const getTeamTasksApi = (teamId, params = {}) => http.get('/tasks', {
  params: { team_id: teamId, page: 1, size: 100, sort: 'priority', ...params },
});

export const updateTeamApi = (teamId, payload) => http.put(`/teams/${teamId}`, payload);

export const dissolveTeamApi = (teamId) => http.delete(`/teams/${teamId}`);

// 团队邀请的三个接口
export const getInvitationsApi = () => http.get('/teams/invitations/pending');

export const acceptInvitationApi = (invitationId) => http.post(`/teams/invitations/${invitationId}/accept`);

export const declineInvitationApi = (invitationId) => http.post(`/teams/invitations/${invitationId}/decline`);


// 任务指派消息接口：
// - 获取任务指派消息；
// - 按“已读即接受”规则将 PENDING 指派更新为 ACCEPTED；
// - 支持删除（忽略）消息，后端写入 dismiss_at。
export const getAssignmentRequestsApi = () => http.get('/assignment-requests', { params: { assignee_id: 'me' } });
// 供消息中心自动“已读即接受”使用
export const acceptAssignmentApi = (requestId) => http.post(`/assignment-requests/${requestId}/accept`);
// 忽略任务指派消息：后端写入 dismiss_at，仅从消息列表隐藏，不改变任务分配业务状态
export const dismissAssignmentApi = (requestId) => http.post(`/assignment-requests/${requestId}/dismiss`);

// ==================== Lab3团队生命周期演化 ====================
export const leaveTeamApi = (teamId) => http.post(`/teams/${teamId}/leave`);

export const transferTeamApi = (teamId, payload) => http.post(`/teams/${teamId}/transfer`, payload);
