package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/leefowlercu/go-contextforge/contextforge"
)

func getTeamNoSlash(ctx context.Context, client *contextforge.Client, teamID string) (*contextforge.Team, *contextforge.Response, error) {
	u := fmt.Sprintf("teams/%s", url.PathEscape(teamID))

	req, err := client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var team *contextforge.Team
	resp, err := client.Do(ctx, req, &team)
	if err != nil {
		return nil, resp, err
	}

	return team, resp, nil
}

func updateTeamNoSlash(ctx context.Context, client *contextforge.Client, teamID string, team *contextforge.TeamUpdate) (*contextforge.Team, *contextforge.Response, error) {
	u := fmt.Sprintf("teams/%s", url.PathEscape(teamID))

	req, err := client.NewRequest(http.MethodPut, u, team)
	if err != nil {
		return nil, nil, err
	}

	var updated *contextforge.Team
	resp, err := client.Do(ctx, req, &updated)
	if err != nil {
		return nil, resp, err
	}

	return updated, resp, nil
}

func deleteTeamNoSlash(ctx context.Context, client *contextforge.Client, teamID string) (*contextforge.Response, error) {
	u := fmt.Sprintf("teams/%s", url.PathEscape(teamID))

	req, err := client.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(ctx, req, nil)
	return resp, err
}

func findTeamByIDViaList(ctx context.Context, client *contextforge.Client, teamID string) (*contextforge.Team, error) {
	const pageSize = 100
	skip := 0

	for {
		opts := &contextforge.TeamListOptions{
			Skip:  skip,
			Limit: pageSize,
		}

		teams, _, err := client.Teams.List(ctx, opts)
		if err != nil {
			return nil, err
		}

		for _, team := range teams {
			if team.ID == teamID {
				return team, nil
			}
		}

		if len(teams) < pageSize {
			break
		}

		skip += len(teams)
	}

	return nil, nil
}
