package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"task-backend/internal/middleware"
	"task-backend/internal/model"
	"task-backend/internal/repository"
	"task-backend/internal/service"
	"task-backend/pkg/response"
)

type mockTeamService struct {
	createFn    func(ctx context.Context, ownerID uint64, req service.CreateTeamRequest) (*model.Team, error)
	updateFn    func(ctx context.Context, operatorID, teamID uint64, req service.UpdateTeamRequest) error
	addMemberFn func(ctx context.Context, operatorID, teamID uint64, req service.AddTeamMemberRequest) (*model.TeamMember, error)
	leaveFn     func(ctx context.Context, userID, teamID uint64) error
}

func (m *mockTeamService) Create(ctx context.Context, ownerID uint64, req service.CreateTeamRequest) (*model.Team, error) {
	if m.createFn == nil {
		panic("createFn not set")
	}
	return m.createFn(ctx, ownerID, req)
}

func (m *mockTeamService) ListMine(ctx context.Context, userID uint64, search string) ([]model.TeamWithRole, error) {
	panic("not used")
}

func (m *mockTeamService) ListMinePage(ctx context.Context, userID uint64, query service.TeamListQuery) ([]model.TeamWithRole, int64, int, int, error) {
	panic("not used")
}

func (m *mockTeamService) Update(ctx context.Context, operatorID, teamID uint64, req service.UpdateTeamRequest) error {
	if m.updateFn == nil {
		panic("updateFn not set")
	}
	return m.updateFn(ctx, operatorID, teamID, req)
}

func (m *mockTeamService) AddMember(ctx context.Context, operatorID, teamID uint64, req service.AddTeamMemberRequest) (*model.TeamMember, error) {
	if m.addMemberFn == nil {
		panic("addMemberFn not set")
	}
	return m.addMemberFn(ctx, operatorID, teamID, req)
}

func (m *mockTeamService) RemoveMember(ctx context.Context, operatorID, teamID, targetUserID uint64) error {
	panic("not used")
}

func (m *mockTeamService) UpdateMemberRole(ctx context.Context, operatorID, teamID, targetUserID uint64, req service.UpdateTeamMemberRoleRequest) (*model.TeamMember, error) {
	panic("not used")
}

func (m *mockTeamService) ListMembers(ctx context.Context, operatorID, teamID uint64) ([]model.TeamMemberDetail, error) {
	panic("not used")
}

func (m *mockTeamService) ListPendingInvitations(ctx context.Context, userID uint64) ([]model.TeamInvitationDetail, error) {
	panic("not used")
}

func (m *mockTeamService) AcceptInvitation(ctx context.Context, userID, invitationID uint64) error {
	panic("not used")
}

func (m *mockTeamService) DeclineInvitation(ctx context.Context, userID, invitationID uint64) error {
	panic("not used")
}

func (m *mockTeamService) Leave(ctx context.Context, userID, teamID uint64) error {
	if m.leaveFn == nil {
		panic("leaveFn not set")
	}
	return m.leaveFn(ctx, userID, teamID)
}

func (m *mockTeamService) TransferOwnership(ctx context.Context, operatorID, teamID, newOwnerID uint64) error {
	panic("not used")
}

func TestTeamHandler_CreateDuplicateName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ownerID := uint64(9)
	mock := &mockTeamService{
		createFn: func(ctx context.Context, gotOwnerID uint64, req service.CreateTeamRequest) (*model.Team, error) {
			assert.Equal(t, ownerID, gotOwnerID)
			assert.Equal(t, "Alpha", req.Name)
			return nil, repository.ErrConflict
		},
	}

	h := NewTeamHandler(mock)
	r := gin.New()
	r.POST("/teams", func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, ownerID)
		h.Create(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/teams", strings.NewReader(`{"name":"Alpha"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)

	var resp response.Common
	assert.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, 409, resp.Code)
	assert.Equal(t, "team name already exists", resp.Message)
}

func TestTeamHandler_UpdateDuplicateName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	operatorID := uint64(9)
	teamID := uint64(123)
	mock := &mockTeamService{
		updateFn: func(ctx context.Context, gotOperatorID, gotTeamID uint64, req service.UpdateTeamRequest) error {
			assert.Equal(t, operatorID, gotOperatorID)
			assert.Equal(t, teamID, gotTeamID)
			assert.Equal(t, "Beta", req.Name)
			return repository.ErrConflict
		},
	}

	h := NewTeamHandler(mock)
	r := gin.New()
	r.PUT("/teams/:team_id", func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, operatorID)
		h.Update(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/teams/123", strings.NewReader(`{"name":"Beta"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)

	var resp response.Common
	assert.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, 409, resp.Code)
	assert.Equal(t, "team name already exists", resp.Message)
}

func TestTeamHandler_AddMember_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	operatorID := uint64(9)
	teamID := uint64(123)
	assigneeID := uint64(22)

	mock := &mockTeamService{
		addMemberFn: func(ctx context.Context, gotOperatorID, gotTeamID uint64, req service.AddTeamMemberRequest) (*model.TeamMember, error) {
			assert.Equal(t, operatorID, gotOperatorID)
			assert.Equal(t, teamID, gotTeamID)
			assert.Equal(t, assigneeID, req.UserID)
			return &model.TeamMember{TeamID: teamID, UserID: assigneeID, Role: model.TeamRoleMember}, nil
		},
	}

	h := NewTeamHandler(mock)
	r := gin.New()
	r.POST("/teams/:team_id/members", func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, operatorID)
		h.AddMember(c)
	})

	body := `{"user_id": 22}`
	req := httptest.NewRequest(http.MethodPost, "/teams/123/members", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp response.Common
	assert.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "invitation sent", resp.Message)

	data, ok := resp.Data.(map[string]interface{})
	if assert.True(t, ok) {
		assert.EqualValues(t, teamID, data["team_id"])
		assert.EqualValues(t, assigneeID, data["user_id"])
		assert.Equal(t, string(model.TeamRoleMember), data["role"])
		assert.Equal(t, string(model.InvitationStatusPending), data["status"])
	}
}

func TestTeamHandler_AddMember_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock := &mockTeamService{}
	h := NewTeamHandler(mock)
	r := gin.New()
	r.POST("/teams/:team_id/members", func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, uint64(1))
		h.AddMember(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/teams/1/members", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp response.Common
	assert.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "invalid request body", resp.Message)
}

func TestTeamHandler_AddMember_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock := &mockTeamService{
		addMemberFn: func(ctx context.Context, operatorID, teamID uint64, req service.AddTeamMemberRequest) (*model.TeamMember, error) {
			return nil, service.ErrTeamOwnerOnly
		},
	}

	h := NewTeamHandler(mock)
	r := gin.New()
	r.POST("/teams/:team_id/members", func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, uint64(1))
		h.AddMember(c)
	})

	body := `{"user_id": 2}`
	req := httptest.NewRequest(http.MethodPost, "/teams/1/members", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var resp response.Common
	assert.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, 403, resp.Code)
	assert.Equal(t, "forbidden", resp.Message)
}

func TestTeamHandler_Leave_OwnerForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uint64(9)
	teamID := uint64(123)

	mock := &mockTeamService{
		leaveFn: func(ctx context.Context, gotUserID, gotTeamID uint64) error {
			assert.Equal(t, userID, gotUserID)
			assert.Equal(t, teamID, gotTeamID)
			return service.ErrOwnerCannotLeave
		},
	}

	h := NewTeamHandler(mock)
	r := gin.New()
	r.POST("/teams/:team_id/leave", func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, userID)
		h.Leave(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/teams/123/leave", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var resp response.Common
	assert.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, 403, resp.Code)
	assert.Equal(t, "forbidden", resp.Message)
}

func TestTeamHandler_Leave_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uint64(11)
	teamID := uint64(321)

	mock := &mockTeamService{
		leaveFn: func(ctx context.Context, gotUserID, gotTeamID uint64) error {
			assert.Equal(t, userID, gotUserID)
			assert.Equal(t, teamID, gotTeamID)
			return nil
		},
	}

	h := NewTeamHandler(mock)
	r := gin.New()
	r.POST("/teams/:team_id/leave", func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, userID)
		h.Leave(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/teams/321/leave", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp response.Common
	assert.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已成功离开团队", resp.Message)
}
