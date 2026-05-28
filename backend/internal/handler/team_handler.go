package handler

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"task-backend/internal/apperror"
	"task-backend/internal/model"
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
		apperror.Respond(c, err, "team list mine")
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
		apperror.Respond(c, apperror.BadRequest("请求体格式不正确"), "team create")
		return
	}

	team, err := h.teamService.Create(c.Request.Context(), userID, req)
	if err != nil {
		apperror.Respond(c, err, "team create")
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
		apperror.Respond(c, apperror.BadRequest("团队ID不合法"), "team update")
		return
	}

	var req service.UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("请求体格式不正确"), "team update")
		return
	}

	err = h.teamService.Update(c.Request.Context(), userID, teamID, req)
	if err != nil {
		apperror.Respond(c, err, "team update")
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
		apperror.Respond(c, apperror.BadRequest("团队ID不合法"), "team add member")
		return
	}

	var req service.AddTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("请求体格式不正确"), "team add member")
		return
	}

	// 输入校验：要求至少提供 user_id 或 username 中的一个
	if req.UserID == 0 && strings.TrimSpace(req.Username) == "" {
		apperror.Respond(c, apperror.BadRequest("请提供用户ID或用户名"), "team add member")
		return
	}

	member, err := h.teamService.AddMember(c.Request.Context(), userID, teamID, req)
	if err != nil {
		apperror.Respond(c, err, "team add member")
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
		apperror.Respond(c, apperror.BadRequest("团队ID不合法"), "team remove member")
		return
	}

	targetUserID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("用户ID不合法"), "team remove member")
		return
	}

	err = h.teamService.RemoveMember(c.Request.Context(), userID, teamID, targetUserID)
	if err != nil {
		apperror.Respond(c, err, "team remove member")
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
		apperror.Respond(c, apperror.BadRequest("团队ID不合法"), "team update member role")
		return
	}

	targetUserID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("用户ID不合法"), "team update member role")
		return
	}

	var req service.UpdateTeamMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("请求体格式不正确"), "team update member role")
		return
	}

	member, err := h.teamService.UpdateMemberRole(c.Request.Context(), userID, teamID, targetUserID, req)
	if err != nil {
		apperror.Respond(c, err, "team update member role")
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
		apperror.Respond(c, apperror.BadRequest("团队ID不合法"), "team list members")
		return
	}

	list, err := h.teamService.ListMembers(c.Request.Context(), userID, teamID)
	if err != nil {
		apperror.Respond(c, err, "team list members")
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
		apperror.Respond(c, err, "team list invitations")
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
		apperror.Respond(c, apperror.BadRequest("邀请ID不合法"), "team accept invitation")
		return
	}

	err = h.teamService.AcceptInvitation(c.Request.Context(), userID, invitationID)
	if err != nil {
		apperror.Respond(c, err, "team accept invitation")
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
		apperror.Respond(c, apperror.BadRequest("邀请ID不合法"), "team decline invitation")
		return
	}

	err = h.teamService.DeclineInvitation(c.Request.Context(), userID, invitationID)
	if err != nil {
		apperror.Respond(c, err, "team decline invitation")
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
		apperror.Respond(c, apperror.BadRequest("团队ID不合法"), "team leave")
		return
	}

	err = h.teamService.Leave(c.Request.Context(), userID, teamID)
	if err != nil {
		apperror.Respond(c, err, "team leave")
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
		apperror.Respond(c, apperror.BadRequest("团队ID不合法"), "team transfer ownership")
		return
	}

	var req transferTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("请求体格式不正确"), "team transfer ownership")
		return
	}

	err = h.teamService.TransferOwnership(c.Request.Context(), userID, teamID, req.NewOwnerID)
	if err != nil {
		apperror.Respond(c, err, "team transfer ownership")
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
		apperror.Respond(c, apperror.BadRequest("团队ID不合法"), "team delete")
		return
	}

	deleter, ok := h.teamService.(teamDeleter)
	if !ok {
		apperror.Respond(c, errors.New("team delete not supported"), "team delete")
		return
	}

	err = deleter.DeleteTeam(c.Request.Context(), userID, teamID)
	if err != nil {
		apperror.Respond(c, err, "team delete")
		return
	}

	response.Success(c, "delete team success", nil)
}
