package controller

import (
	"context"
	hyperdriveharmonizeriov1 "github.com/ibexmonj/hyperdriveharmonizer/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	err := FetchAndCreateTeams(ctx, r, req)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

func (r *TeamReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hyperdriveharmonizeriov1.Team{}).
		Complete(r)
}
