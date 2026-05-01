import { http } from './http';

export const createTeamApi = (payload) => http.post('/teams', payload);

export const getTeamsApi = () => http.get('/teams');

export const getTeamMembersApi = (teamId) => http.get(`/teams/${teamId}/members`);

export const addTeamMemberApi = (teamId, payload) => http.post(`/teams/${teamId}/members`, payload);

export const deleteTeamMemberApi = (teamId, userId) => http.delete(`/teams/${teamId}/members/${userId}`);

export const updateTeamMemberRoleApi = (teamId, userId, payload) => http.put(`/teams/${teamId}/members/${userId}`, payload);

export const getTeamTasksApi = (teamId) => http.get('/tasks', { params: { team_id: teamId, page: 1, size: 100 } });

export const updateTeamApi = (teamId, payload) => http.put(`/teams/${teamId}`, payload);

export const dissolveTeamApi = (teamId) => http.delete(`/teams/${teamId}`);

export const getInvitationsApi = () => http.get('/teams/invitations/pending');

export const acceptInvitationApi = (invitationId) => http.post(`/teams/invitations/${invitationId}/accept`);

export const declineInvitationApi = (invitationId) => http.post(`/teams/invitations/${invitationId}/decline`);
