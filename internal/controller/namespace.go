// namespace.go
package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CreateNamespace creates a new namespace if it doesn't already exist
func CreateNamespace(ctx context.Context, r *TeamReconciler, namespaceName string) error {
	log := log.FromContext(ctx)

	// Check if a namespace with the same name as the team already exists
	var existingNamespace corev1.Namespace
	if err := r.Get(ctx, types.NamespacedName{Name: namespaceName}, &existingNamespace); err != nil {
		if errors.IsNotFound(err) {
			// The namespace doesn't exist, so create it
			newNamespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespaceName,
					Labels: map[string]string{
						"harmonizer.io/team-owner": namespaceName,
					},
				},
			}
			if err := r.Create(ctx, newNamespace); err != nil {
				log.Error(err, "failed to create Namespace")
				return err
			}
			log.Info("created Namespace", "name", namespaceName)
		} else {
			// Error reading the object - requeue the request.
			log.Error(err, "failed to get Namespace")
			return err
		}
	}

	return nil
}
