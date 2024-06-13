package controller

import (
	"context"
	"fmt"
	"github.com/google/go-github/v33/github"
	harmonizeriov1 "github.com/ibexmonj/harmonizer/api/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TeamReconciler reconciles a Team object
type TeamReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	GitHubClient GitHubClient
}

type GitHubClient interface {
	ListTeams(ctx context.Context, org string, opt *github.ListOptions) ([]*github.Team, *github.Response, error)
	ListTeamMembersBySlug(ctx context.Context, org string, slug string, opt *github.TeamListTeamMembersOptions) ([]*github.User, *github.Response, error)
}

// FetchTeams fetches teams from GitHub.
func FetchTeams(ctx context.Context, ghClient GitHubClient) ([]*github.Team, error) {
	teams, _, err := ghClient.ListTeams(ctx, "PistonPioneers", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams from GitHub: %w", err)
	}
	return teams, nil
}

// CreateTeamResource creates a new Team resource in Kubernetes if it doesn't already exist.
func CreateTeamResource(ctx context.Context, r *TeamReconciler, team *github.Team, ghClient GitHubClient) error {
	if r == nil {
		return fmt.Errorf("client is nil")
	}
	if team == nil {
		return fmt.Errorf("team is nil")
	}
	if ghClient == nil {
		return fmt.Errorf("GitHub client is nil")
	}
	teamName := team.GetName()
	namespacedName := types.NamespacedName{
		Name:      teamName,
		Namespace: "default", // TODO: revisit
	}
	var existingTeam harmonizeriov1.Team
	if err := r.Get(ctx, namespacedName, &existingTeam); err != nil {
		if errors.IsNotFound(err) {
			members, _, err := ghClient.ListTeamMembersBySlug(ctx, "PistonPioneers", *team.Slug, nil)
			if err != nil {
				return fmt.Errorf("failed to get team members from GitHub: %w", err)
			}
			memberNames := make([]string, 0, len(members))
			for _, member := range members {
				if member.Login != nil {
					memberNames = append(memberNames, *member.Login)
				}
			}
			newTeam := &harmonizeriov1.Team{
				ObjectMeta: metav1.ObjectMeta{
					Name:      teamName,
					Namespace: "default", // TODO: revisit
				},
				Spec: harmonizeriov1.TeamSpec{
					TeamName: teamName,
					Members:  memberNames,
				},
			}
			if err := r.Create(ctx, newTeam); err != nil {
				return fmt.Errorf("failed to create Team resource: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get Team resource: %w", err)
		}
	}
	return nil
}

// FetchAndCreateTeams fetches teams from GitHub and creates corresponding resources in Kubernetes.
func FetchAndCreateTeams(ctx context.Context, r *TeamReconciler, req ctrl.Request, ghClient GitHubClient) error {
	teams, err := FetchTeams(ctx, ghClient)
	if err != nil {
		return err
	}

	for _, team := range teams {
		if err := CreateTeamResource(ctx, r, team, ghClient); err != nil {
			return err
		}

		if err := CreateNamespace(ctx, r, team.GetName()); err != nil {
			return err
		}
	}

	return nil
}
