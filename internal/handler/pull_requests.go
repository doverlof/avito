package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/doverlof/avito_help/api"
	"github.com/doverlof/avito_help/internal/model"
)

func (h *handler) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	var req api.PostPullRequestCreateJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println(err)
		writeError(w, http.StatusBadRequest, api.NOTFOUND, "invalid request body")
		return
	}
	err := h.pullRequestUseCase.Create(r.Context(), convertFromApi(req))
	if err != nil {
		fmt.Println(err)
		status, code, msg := mapErrorToAPI(err)
		writeError(w, status, code, msg)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(map[string]interface{}{
		"team": req,
	}); err != nil {
		fmt.Println(err)
		writeError(w, http.StatusInternalServerError, api.NOTFOUND, "failed to write response")
		return
	}
}

func convertFromApi(create api.PostPullRequestCreateJSONRequestBody) model.CreatePullRequest {
	return model.CreatePullRequest{
		AuthorID:        create.AuthorId,
		PullRequestID:   create.PullRequestId,
		PullRequestName: create.PullRequestName,
	}
}

func (h *handler) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	var req api.PostPullRequestMergeJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println(err)
		writeError(w, http.StatusBadRequest, api.NOTFOUND, "invalid request body")
		return
	}

	pullRequest, err := h.pullRequestUseCase.Merge(r.Context(), req.PullRequestId)
	if err != nil {
		fmt.Println(err)
		status, code, msg := mapErrorToAPI(err)
		writeError(w, status, code, msg)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(map[string]interface{}{
		"pr": api.PullRequest{
			AuthorId:          pullRequest.AuthorID,
			AssignedReviewers: pullRequest.ReviewerIDs,
			MergedAt:          &pullRequest.MergedAt,
			PullRequestId:     pullRequest.PullRequestID,
			PullRequestName:   pullRequest.PullRequestName,
			Status:            convertPRStatus(pullRequest.Status),
		},
	}); err != nil {
		fmt.Println(err)
		writeError(w, http.StatusInternalServerError, api.NOTFOUND, "failed to write response")
		return
	}
}

func convertPRStatus(status model.PullRequestStatus) api.PullRequestStatus {
	switch status {
	case model.StatusOpen:
		return api.PullRequestStatusOPEN
	default:
		return api.PullRequestStatusMERGED
	}
}

func (h *handler) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	var req api.PostPullRequestReassignJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println(err)
		writeError(w, http.StatusBadRequest, api.NOTFOUND, "invalid request body")
		return
	}

	pullRequest, newRewieverID, err := h.pullRequestUseCase.Reassign(r.Context(), req.PullRequestId, req.OldUserId)
	if err != nil {
		fmt.Println(err)
		status, code, msg := mapErrorToAPI(err)
		writeError(w, status, code, msg)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(map[string]interface{}{
		"pr": api.PullRequest{
			AuthorId:          pullRequest.AuthorID,
			AssignedReviewers: pullRequest.ReviewerIDs,
			MergedAt:          &pullRequest.MergedAt,
			PullRequestId:     pullRequest.PullRequestID,
			PullRequestName:   pullRequest.PullRequestName,
			Status:            convertPRStatus(pullRequest.Status),
		},
		"replaced_by": newRewieverID,
	}); err != nil {
		fmt.Println(err)
		writeError(w, http.StatusInternalServerError, api.NOTFOUND, "failed to write response")
		return
	}
}
