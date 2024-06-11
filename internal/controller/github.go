package controller

import (
	"context"
	"fmt"
	"github.com/google/go-github/v33/github"
	harmonizeriov1 "github.com/ibexmonj/harmonizer/api/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type GitHubClient interface {
	ListTeams(ctx context.Context, org string, opt *github.ListOptions) ([]*github.Team, *github.Response, error)
	ListTeamMembersBySlug(ctx context.Context, org string, slug string, opt *github.TeamListTeamMembersOptions) ([]*github.User, *github.Response, error)
}

func FetchAndCreateTeams(ctx context.Context, r *TeamReconciler, req ctrl.Request, ghClient GitHubClient) error {
	log := log.FromContext(ctx)

	teams, _, err := ghClient.ListTeams(ctx, "PistonPioneers", nil)
	if err != nil {
		log.Error(err, "failed to get teams from GitHub")
		return err
	}

	fmt.Println("teams found:", teams)

	// For each team, ensure there's a corresponding Kubernetes custom resource
	for _, team := range teams {
		teamName := team.GetName()

		// Check if the custom resource already exists
		namespacedName := types.NamespacedName{
			Name:      teamName,
			Namespace: "default", // TODO: revisit
		}
		var existingTeam harmonizeriov1.Team
		if err := r.Get(ctx, namespacedName, &existingTeam); err != nil {
			if errors.IsNotFound(err) {

				members, _, err := ghClient.ListTeamMembersBySlug(ctx, "PistonPioneers", *team.Slug, nil)
				if err != nil {
					log.Error(err, "failed to get team members from GitHub")
					return err
				}
				fmt.Println("members found:", members)

				// Convert the members to a slice of strings
				memberNames := make([]string, 0, len(members))
				for _, member := range members {
					if member.Login != nil {
						memberNames = append(memberNames, *member.Login)
					}
				}

				// The custom resource doesn't exist, so create it
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
					log.Error(err, "failed to create Team resource")
					return err
				}
				log.Info("created Team resource", "name", teamName)

			} else {
				// Error reading the object - requeue the request.
				log.Error(err, "failed to get Team resource")
				return err
			}
		}

		// Create a namespace for the team
		if err := CreateNamespace(ctx, r, teamName); err != nil {
			return err
		}
	}

	return nil
}
