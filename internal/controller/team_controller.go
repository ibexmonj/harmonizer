package controller

import (
	"context"
	harmonizeriov1 "github.com/ibexmonj/harmonizer/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

//+kubebuilder:rbac:groups=harmonizer.io,resources=teams,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=harmonizer.io,resources=teams/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=harmonizer.io,resources=teams/finalizers,verbs=update

func (r *TeamReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	err := FetchAndCreateTeams(ctx, r, req, r.GitHubClient)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

func (r *TeamReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&harmonizeriov1.Team{}).
		Complete(r)
}
