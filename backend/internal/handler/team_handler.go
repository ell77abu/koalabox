package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"task-backend/internal/repository"
	"task-backend/internal/service"
	"task-backend/pkg/response"
)

type TeamHandler struct {
	teamService service.TeamService
}

func NewTeamHandler(teamService service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

func (h *TeamHandler) ListMine(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	list, err := h.teamService.ListMine(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "query teams failed")
		return
	}

	response.Success(c, "query success", list)
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
	if err != nil {
		response.InternalError(c, "create team failed")
		return
	}

	response.Created(c, "create success", team)
}

func (h *TeamHandler) AddMember(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	teamID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid team id")
		return
	}

	var req service.AddTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	member, err := h.teamService.AddMember(c.Request.Context(), userID, teamID, req)
	if errors.Is(err, service.ErrInvalidTeamRole) {
		response.BadRequest(c, err.Error())
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
		response.Conflict(c, "member already exists")
		return
	}
	if err != nil {
		response.InternalError(c, "add member failed")
		return
	}

	response.Created(c, "add member success", member)
}
