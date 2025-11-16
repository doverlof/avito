package component

import (
	"context"
	"testing"

	"github.com/doverlof/avito_help/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTeam(t *testing.T) {
	ctx := context.Background()
	client, err := api.NewClient("http://localhost:8080")
	require.NoError(t, err)
	tests := []struct {
		name    string
		params  *api.GetTeamGetParams
		wantOk  *api.Team
		wantErr *api.ErrorResponse
	}{
		{
			name: "success",
			params: &api.GetTeamGetParams{
				TeamName: "backend",
			},
			wantOk: &api.Team{
				TeamName: "backend",
				Members: []api.TeamMember{
					{UserId: "u3", Username: "Carol White", IsActive: true},
					{UserId: "u4", Username: "David Brown", IsActive: false},
					{UserId: "u5", Username: "Eve Davis", IsActive: true},
					{UserId: "u6", Username: "Frank Miller", IsActive: true},
					{UserId: "u7", Username: "Grace Wilson", IsActive: true},
					{UserId: "u8", Username: "Henry Moore", IsActive: false},
					{UserId: "u9", Username: "Ivy Taylor", IsActive: true},
					{UserId: "u10", Username: "Jack Anderson", IsActive: true},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := client.GetTeamGet(ctx, test.params)
			require.NoError(t, err)
			resp, err := api.ParseGetTeamGetResponse(got)
			require.NoError(t, err)
			assert.Equal(t, test.wantOk, resp.JSON200)
			assert.Equal(t, test.wantErr, resp.JSON404)
		})
	}
}
