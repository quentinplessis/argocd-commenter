package controllers

import (
	"context"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// findArgoCDURL returns the URL of Argo CD if available.
// See https://github.com/argoproj/argo-cd/blob/master/docs/operator-manual/argocd-cm.yaml
func findArgoCDURL(ctx context.Context, c client.Client, namespace string) string {
	logger := log.FromContext(ctx)
	var cm corev1.ConfigMap
	err := c.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "argocd-cm"}, &cm)
	if err != nil {
		logger.Error(err, "unable to determine Argo CD URL")
		return ""
	}
	return cm.Data["url"]
}

func getLastDeployedRevision(a argocdv1alpha1.Application) string {
	if a.Status.OperationState == nil {
		return ""
	}
	if a.Status.OperationState.Operation.Sync == nil {
		return ""
	}
	return a.Status.OperationState.Operation.Sync.Revision
}