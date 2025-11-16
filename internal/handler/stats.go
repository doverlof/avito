package handler

import (
	"encoding/json"
	"fmt"
	"github.com/doverlof/avito_help/api"
	"github.com/doverlof/avito_help/internal/convert"
	"github.com/doverlof/avito_help/internal/model"
	"net/http"
)

func (h *handler) GetStatsUsers(w http.ResponseWriter, r *http.Request) {
	stats, err := h.statsUseCase.GetUserStatistics(r.Context())
	if err != nil {
		fmt.Println(err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(map[string]interface{}{
		"statistics": convert.Many(convertUserStatsToApi, stats),
	}); err != nil {
		fmt.Println(err)
		http.Error(w, "failed to write response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func convertUserStatsToApi(stats model.UserStatistics) api.UserStatistics {
	return api.UserStatistics{
		UserId:   stats.UserId,
		Username: stats.Username,
		TeamName: stats.TeamName,
		IsActive: stats.IsActive,

		TotalReviewAssignments:  stats.TotalReviewAssignments,
		OpenReviewAssignments:   stats.OpenReviewAssignments,
		MergedReviewAssignments: stats.MergedReviewAssignments,

		TotalAuthoredPrs:  stats.TotalAuthoredPrs,
		OpenAuthoredPrs:   stats.OpenAuthoredPrs,
		MergedAuthoredPrs: stats.MergedAuthoredPrs,
	}
}
