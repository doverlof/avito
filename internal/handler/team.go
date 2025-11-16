package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/doverlof/avito_help/api"
	"github.com/doverlof/avito_help/internal/convert"
	"github.com/doverlof/avito_help/internal/model"
)

func (h *handler) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	var req api.PostTeamAddJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println(err)
		writeError(w, http.StatusBadRequest, api.NOTFOUND, "invalid request body")
		return
	}

	err := h.teamUseCase.Add(r.Context(), model.Team{
		Name:    req.TeamName,
		Members: convert.Many(convertMemberFromApi, req.Members),
	})
	if err != nil {
		fmt.Println(err)
		status, code, msg := mapErrorToAPI(err)
		writeError(w, status, code, msg)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(req); err != nil {
		fmt.Println(err)
		writeError(w, http.StatusInternalServerError, api.NOTFOUND, "failed to write response")
		return
	}
}

func convertMemberFromApi(apiMember api.TeamMember) model.Member {
	return model.Member{
		ID:       apiMember.UserId,
		IsActive: apiMember.IsActive,
		Name:     apiMember.Username,
	}
}

func (h *handler) GetTeamGet(w http.ResponseWriter, r *http.Request, params api.GetTeamGetParams) {
	team, err := h.teamUseCase.Get(r.Context(), params.TeamName)
	if err != nil {
		fmt.Println(err)
		status, code, msg := mapErrorToAPI(err)
		writeError(w, status, code, msg)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(convertTeam(team)); err != nil {
		fmt.Println(err)
		writeError(w, http.StatusInternalServerError, api.NOTFOUND, "failed to write response")
		return
	}
}

func convertTeam(team model.Team) api.Team {
	return api.Team{
		TeamName: team.Name,
		Members:  convert.Many(convertMemberFromModel, team.Members),
	}
}

func convertMemberFromModel(m model.Member) api.TeamMember {
	return api.TeamMember{
		IsActive: m.IsActive,
		Username: m.Name,
		UserId:   m.ID,
	}
}
