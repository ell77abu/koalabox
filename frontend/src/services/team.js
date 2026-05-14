import { http } from './http';

export const createTeamApi = (payload) => http.post('/teams', payload);

export const getTeamsApi = () => http.get('/teams');

export const getTeamMembersApi = (teamId) => http.get(`/teams/${teamId}/members`);

export const addTeamMemberApi = (teamId, payload) => http.post(`/teams/${teamId}/invitations`, payload);

export const deleteTeamMemberApi = (teamId, userId) => http.delete(`/teams/${teamId}/members/${userId}`);

export const updateTeamMemberRoleApi = (teamId, userId, payload) => http.put(`/teams/${teamId}/members/${userId}`, payload);

export const getTeamTasksApi = (teamId) => http.get('/tasks', { params: { team_id: teamId, page: 1, size: 100 } });

export const updateTeamApi = (teamId, payload) => http.put(`/teams/${teamId}`, payload);

export const dissolveTeamApi = (teamId) => http.delete(`/teams/${teamId}`);

// 团队邀请的三个接口
export const getInvitationsApi = () => http.get('/invitations', { params: { invitee_id: 'me' } });

export const acceptInvitationApi = (invitationId) => http.post(`/invitations/${invitationId}/accept`);

export const declineInvitationApi = (invitationId) => http.post(`/invitations/${invitationId}/decline`);

// 任务指派：只暴露获取接口，前端不做同意/拒绝操作
export const getAssignmentRequestsApi = () => http.get('/assignment-requests', { params: { assignee_id: 'me' } });
// 供 静默已读 API
export const acceptAssignmentApi = (requestId) => http.post(`/assignment-requests/${requestId}/accept`);

// ==================== Lab3团队生命周期演化 ====================
export const leaveTeamApi = (teamId) => http.post(`/teams/${teamId}/leave`);

export const transferTeamApi = (teamId, payload) => http.post(`/teams/${teamId}/transfer`, payload);
