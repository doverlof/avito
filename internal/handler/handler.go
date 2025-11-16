package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/doverlof/avito_help/api"
	pullRequestUseCase "github.com/doverlof/avito_help/internal/usecase/pull-request"
	statsUseCase "github.com/doverlof/avito_help/internal/usecase/stats"
	teamUseCase "github.com/doverlof/avito_help/internal/usecase/team"
	userUseCase "github.com/doverlof/avito_help/internal/usecase/user"
)

type handler struct {
	teamUseCase        teamUseCase.UseCase
	userUseCase        userUseCase.UseCase
	statsUseCase       statsUseCase.UseCase
	pullRequestUseCase pullRequestUseCase.UseCase
}

func New(
	teamUseCase teamUseCase.UseCase,
	userUseCase userUseCase.UseCase,
	statsUseCase statsUseCase.UseCase,
	pullRequestUseCase pullRequestUseCase.UseCase,
) api.ServerInterface {
	return &handler{
		teamUseCase:        teamUseCase,
		userUseCase:        userUseCase,
		statsUseCase:       statsUseCase,
		pullRequestUseCase: pullRequestUseCase,
	}
}

func mapErrorToAPI(err error) (int, api.ErrorResponseErrorCode, string) {
	switch {

	case errors.Is(err, pullRequestUseCase.ErrPRNotFound):
		return http.StatusNotFound, api.NOTFOUND, "pull request not found"

	case errors.Is(err, pullRequestUseCase.ErrTeamOrAuthorNotFound):
		return http.StatusNotFound, api.NOTFOUND, "author or team not found"

	case errors.Is(err, pullRequestUseCase.ErrPRAlreadyMerged):
		return http.StatusConflict, api.PRMERGED, "cannot reassign on merged PR"

	case errors.Is(err, pullRequestUseCase.ErrDontHaveReviewers):
		return http.StatusConflict, api.NOCANDIDATE, "no active replacement candidate in team"

	case errors.Is(err, pullRequestUseCase.ErrNotAssigned):
		return http.StatusConflict, api.NOTASSIGNED, "reviewer is not assigned to this PR"

	case errors.Is(err, teamUseCase.ErrTeamExists):
		return http.StatusConflict, api.TEAMEXISTS, "team already exists"

	case errors.Is(err, teamUseCase.ErrTeamNotFound):
		return http.StatusNotFound, api.NOTFOUND, "team not found"

	default:
		return http.StatusInternalServerError, api.NOTFOUND, "internal server error"
	}
}

func writeError(w http.ResponseWriter, status int, code api.ErrorResponseErrorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(api.ErrorResponse{
		Error: struct {
			Code    api.ErrorResponseErrorCode `json:"code"`
			Message string                     `json:"message"`
		}{
			Code:    code,
			Message: message,
		},
	})
}
