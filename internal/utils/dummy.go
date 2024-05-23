// internal/utils/dummy.go
package utils

import (
	"context"
	harmonizeriov1 "github.com/ibexmonj/harmonizer/api/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateDummyTeam(ctx context.Context, client client.Client) error {
	// Create a "dummy" Team resource
	dummyTeam := &harmonizeriov1.Team{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dummy",
			Namespace: "default", // Modify as needed
		},
		Spec: harmonizeriov1.TeamSpec{
			TeamName: "dummy",
			Members:  []string{},
		},
	}
	// Check if the dummy team already exists
	if err := client.Get(ctx, types.NamespacedName{Name: "dummy", Namespace: "default"}, dummyTeam); err != nil {
		if errors.IsNotFound(err) {
			// The dummy team doesn't exist, so create it
			if err := client.Create(ctx, dummyTeam); err != nil {
				return err
			}
		} else {
			// Other error occurred
			return err
		}
	}
	return nil
}
