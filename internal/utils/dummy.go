// internal/utils/dummy.go
package utils

import (
	"context"
	harmonizeriov1 "github.com/ibexmonj/harmonizer/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	if err := client.Create(ctx, dummyTeam); err != nil {
		return err
	}
	return nil
}
