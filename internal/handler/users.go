package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/doverlof/avito_help/internal/convert"
	"net/http"

	"github.com/doverlof/avito_help/api"
	"github.com/doverlof/avito_help/internal/model"
	userUseCase "github.com/doverlof/avito_help/internal/usecase/user"
)

func (h *handler) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println(err)
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.userUseCase.SetIsActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		fmt.Println(err)
		if errors.Is(err, userUseCase.ErrUserNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(map[string]interface{}{
		"user": convertUserToApi(user),
	}); err != nil {
		fmt.Println(err)
		http.Error(w, "failed to write response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *handler) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params api.GetUsersGetReviewParams) {
	prs, err := h.pullRequestUseCase.GetByReviewer(r.Context(), params.UserId)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":       params.UserId,
		"pull_requests": convert.Many(convertPRToShort, prs),
	}); err != nil {
		fmt.Println(err)
		http.Error(w, "failed to write response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func convertPRToShort(pr model.PullRequest) api.PullRequestShort {
	return api.PullRequestShort{
		PullRequestId:   pr.PullRequestID,
		PullRequestName: pr.PullRequestName,
		AuthorId:        pr.AuthorID,
		Status:          api.PullRequestShortStatus(pr.Status),
	}
}

func convertUserToApi(user model.User) api.User {
	return api.User{
		UserId:   user.ID,
		Username: user.Name,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}
}
