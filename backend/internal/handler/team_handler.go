package handler

import (
	"context"
	"errors"
	"log"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"task-backend/internal/model"
	"task-backend/internal/repository"
	"task-backend/internal/service"
	"task-backend/pkg/response"
)

type TeamHandler struct {
	teamService service.TeamService
}

type createTeamResponse struct {
	TeamID  uint64 `json:"team_id"`
	Name    string `json:"name"`
	OwnerID uint64 `json:"owner_id"`
}

type addMemberResponse struct {
	TeamID uint64 `json:"team_id"`
	UserID uint64 `json:"user_id"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

type teamListResponse struct {
	List interface{} `json:"list"`
}

type transferTeamRequest struct {
	NewOwnerID uint64 `json:"new_owner_id"`
}

type teamDeleter interface {
	DeleteTeam(ctx context.Context, operatorID, teamID uint64) error
}

func NewTeamHandler(teamService service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

func (h *TeamHandler) ListMine(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	search := strings.TrimSpace(c.Query("search"))

	list, err := h.teamService.ListMine(c.Request.Context(), userID, search)
	if err != nil {
		response.InternalError(c, "query teams failed")
		return
	}

	response.Success(c, "query success", teamListResponse{List: list})
}

func (h *TeamHandler) Create(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	var req service.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	team, err := h.teamService.Create(c.Request.Context(), userID, req)
	if errors.Is(err, service.ErrInvalidTeamName) {
		response.BadRequest(c, err.Error())
		return
	}
	if errors.Is(err, repository.ErrConflict) {
		response.Conflict(c, "team name already exists")
		return
	}
	if err != nil {
		response.InternalError(c, "create team failed")
		return
	}

	response.Created(c, "team created", createTeamResponse{
		TeamID:  team.ID,
		Name:    team.Name,
		OwnerID: team.OwnerID,
	})
}

func (h *TeamHandler) Update(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid team id")
		return
	}

	var req service.UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	err = h.teamService.Update(c.Request.Context(), userID, teamID, req)
	if errors.Is(err, service.ErrInvalidTeamName) {
		response.BadRequest(c, err.Error())
		return
	}
	if errors.Is(err, repository.ErrConflict) {
		response.Conflict(c, "team name already exists")
		return
	}
	if errors.Is(err, service.ErrTeamOwnerOnly) || errors.Is(err, repository.ErrForbidden) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(c, "team not found")
		return
	}
	if err != nil {
		response.InternalError(c, "rename team failed")
		return
	}

	response.Success(c, "rename success", nil)
}

func (h *TeamHandler) AddMember(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid team id")
		return
	}

	var req service.AddTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	// 输入校验：要求至少提供 user_id 或 username 中的一个
	if req.UserID == 0 && strings.TrimSpace(req.Username) == "" {
		response.BadRequest(c, "user_id or username is required")
		return
	}

	member, err := h.teamService.AddMember(c.Request.Context(), userID, teamID, req)
	if errors.Is(err, service.ErrInvalidTeamMember) {
		response.BadRequest(c, err.Error())
		return
	}
	if errors.Is(err, service.ErrInvalidTeamRole) {
		response.BadRequest(c, err.Error())
		return
	}
	if errors.Is(err, service.ErrTeamOwnerOnly) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, repository.ErrForbidden) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(c, "team or user not found")
		return
	}
	if errors.Is(err, repository.ErrConflict) {
		response.Conflict(c, "member already exists or invitation pending")
		return
	}
	if err != nil {
		// 记录详细错误与堆栈到服务器日志，便于排查；对外只返回统一的内部错误信息
		log.Printf("AddMember error: %v\n%s", err, debug.Stack())
		response.InternalError(c, "add member failed")
		return
	}

	response.Created(c, "invitation sent", addMemberResponse{
		TeamID: teamID,
		UserID: member.UserID,
		Role:   string(member.Role),
		Status: string(model.InvitationStatusPending),
	})
}

func (h *TeamHandler) RemoveMember(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid team id")
		return
	}

	targetUserID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	err = h.teamService.RemoveMember(c.Request.Context(), userID, teamID, targetUserID)
	if errors.Is(err, service.ErrTeamOwnerOnly) || errors.Is(err, service.ErrCannotRemoveSelf) || errors.Is(err, service.ErrCannotModifyOwner) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, repository.ErrForbidden) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(c, "team or member not found")
		return
	}
	if err != nil {
		response.InternalError(c, "remove member failed")
		return
	}

	response.Success(c, "remove member success", nil)
}

func (h *TeamHandler) UpdateMemberRole(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid team id")
		return
	}

	targetUserID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req service.UpdateTeamMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	member, err := h.teamService.UpdateMemberRole(c.Request.Context(), userID, teamID, targetUserID, req)
	if errors.Is(err, service.ErrInvalidTeamRole) {
		response.BadRequest(c, err.Error())
		return
	}
	if errors.Is(err, service.ErrTeamOwnerOnly) || errors.Is(err, service.ErrCannotModifyOwner) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, repository.ErrForbidden) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(c, "team or member not found")
		return
	}
	if err != nil {
		response.InternalError(c, "update member role failed")
		return
	}

	response.Success(c, "update member role success", member)
}

func (h *TeamHandler) ListMembers(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid team id")
		return
	}

	list, err := h.teamService.ListMembers(c.Request.Context(), userID, teamID)
	if errors.Is(err, repository.ErrForbidden) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(c, "team not found")
		return
	}
	if err != nil {
		response.InternalError(c, "query team members failed")
		return
	}

	response.Success(c, "query success", list)
}

func (h *TeamHandler) ListPendingInvitations(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	list, err := h.teamService.ListPendingInvitations(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "query invitations failed")
		return
	}

	response.Success(c, "query success", list)
}

func (h *TeamHandler) AcceptInvitation(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	invitationID, err := strconv.ParseUint(c.Param("invitation_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid invitation id")
		return
	}

	err = h.teamService.AcceptInvitation(c.Request.Context(), userID, invitationID)
	if errors.Is(err, service.ErrInvitationNotPending) {
		response.BadRequest(c, "invitation is not pending")
		return
	}
	if errors.Is(err, repository.ErrForbidden) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(c, "invitation or team not found")
		return
	}
	if err != nil {
		response.InternalError(c, "accept invitation failed")
		return
	}

	response.Success(c, "invitation accepted", nil)
}

func (h *TeamHandler) DeclineInvitation(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	invitationID, err := strconv.ParseUint(c.Param("invitation_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid invitation id")
		return
	}

	err = h.teamService.DeclineInvitation(c.Request.Context(), userID, invitationID)
	if errors.Is(err, service.ErrInvitationNotPending) {
		response.BadRequest(c, "invitation is not pending")
		return
	}
	if errors.Is(err, repository.ErrForbidden) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(c, "invitation not found")
		return
	}
	if err != nil {
		response.InternalError(c, "decline invitation failed")
		return
	}

	response.Success(c, "invitation declined", nil)
}

func (h *TeamHandler) Leave(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid team id")
		return
	}

	err = h.teamService.Leave(c.Request.Context(), userID, teamID)
	if errors.Is(err, service.ErrOwnerCannotLeave) || errors.Is(err, repository.ErrForbidden) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(c, "team not found")
		return
	}
	if err != nil {
		response.InternalError(c, "leave team failed")
		return
	}

	response.Success(c, "已成功离开团队", nil)
}

func (h *TeamHandler) TransferOwnership(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid team id")
		return
	}

	var req transferTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	err = h.teamService.TransferOwnership(c.Request.Context(), userID, teamID, req.NewOwnerID)
	if errors.Is(err, service.ErrTeamOwnerOnly) || errors.Is(err, repository.ErrForbidden) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, service.ErrInvalidTransferTarget) {
		response.BadRequest(c, err.Error())
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(c, "team not found")
		return
	}
	if err != nil {
		response.InternalError(c, "transfer owner failed")
		return
	}

	response.Success(c, "团队所有权转让成功", nil)
}

func (h *TeamHandler) Delete(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid team id")
		return
	}

	deleter, ok := h.teamService.(teamDeleter)
	if !ok {
		response.InternalError(c, "team delete not supported")
		return
	}

	err = deleter.DeleteTeam(c.Request.Context(), userID, teamID)
	if errors.Is(err, service.ErrTeamOwnerOnly) || errors.Is(err, repository.ErrForbidden) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(c, "team not found")
		return
	}
	if err != nil {
		response.InternalError(c, "delete team failed")
		return
	}

	response.Success(c, "delete team success", nil)
}
