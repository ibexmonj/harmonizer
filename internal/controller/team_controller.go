package controller

import (
	"context"
	"fmt"
	"github.com/google/go-github/v33/github"
	hyperdriveharmonizeriov1 "github.com/ibexmonj/hyperdriveharmonizer/api/v1"
	"golang.org/x/oauth2"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// TeamReconciler reconciles a Team object
type TeamReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=hyperdriveharmonizer.io,resources=teams,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hyperdriveharmonizer.io,resources=teams/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hyperdriveharmonizer.io,resources=teams/finalizers,verbs=update

func (r *TeamReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Set up the GitHub client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	ghClient := github.NewClient(tc)
	// Get the list of teams from GitHub
	orgs, _, err := ghClient.Organizations.List(ctx, "", nil)
	if err != nil {
		log.Error(err, "failed to get Org from GitHub")
		return ctrl.Result{}, err
	}
	fmt.Println(orgs, ": organization found")
	teams, _, err := ghClient.Teams.ListTeams(ctx, "PistonPioneers", nil)
	if err != nil {
		log.Error(err, "failed to get teams from GitHub")
		return ctrl.Result{}, err
	}

	fmt.Println("teams found:", teams)

	// For each team, ensure there's a corresponding Kubernetes custom resource
	for _, team := range teams {
		teamName := team.GetName()

		// Check if the custom resource already exists
		namespacedName := types.NamespacedName{
			Name:      teamName,
			Namespace: "default", // TODO: Modify as needed
		}
		var existingTeam hyperdriveharmonizeriov1.Team
		if err := r.Get(ctx, namespacedName, &existingTeam); err != nil {
			if errors.IsNotFound(err) {

				members, _, err := ghClient.Teams.ListTeamMembersBySlug(ctx, "PistonPioneers", *team.Slug, nil)
				if err != nil {
					log.Error(err, "failed to get team members from GitHub")
					return ctrl.Result{}, err
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
				newTeam := &hyperdriveharmonizeriov1.Team{
					ObjectMeta: metav1.ObjectMeta{
						Name:      teamName,
						Namespace: "default", // TODO: Modify as needed
					},
					Spec: hyperdriveharmonizeriov1.TeamSpec{
						TeamName: teamName,
						Members:  memberNames,
					},
				}
				if err := r.Create(ctx, newTeam); err != nil {
					log.Error(err, "failed to create Team resource")
					return ctrl.Result{}, err
				}
				log.Info("created Team resource", "name", teamName)
			} else {
				// Error reading the object - requeue the request.
				log.Error(err, "failed to get Team resource")
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

func (r *TeamReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hyperdriveharmonizeriov1.Team{}).
		Complete(r)
}
