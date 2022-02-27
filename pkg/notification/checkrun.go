package notification

import (
	"context"
	"fmt"
	"github.com/argoproj/gitops-engine/pkg/health"
	synccommon "github.com/argoproj/gitops-engine/pkg/sync/common"
	"github.com/go-logr/logr"
	"github.com/int128/argocd-commenter/pkg/github"
	"strings"
)

func (c client) CheckRun(ctx context.Context, e Event) error {
	logger := logr.FromContextOrDiscard(ctx)

	checkRunURL := e.Application.Annotations["argocd-commenter.int128.github.io/check-run-url"]
	checkRunID := github.ParseCheckRunURL(checkRunURL)
	if checkRunID == nil {
		return nil
	}

	cr := generateCheckRun(e)
	if cr == nil {
		logger.Info("nothing to update the check run", "event", e)
		return nil
	}

	logger.Info("updating the check run", "checkRun", checkRunURL)
	if err := c.ghc.UpdateCheckRun(ctx, *checkRunID, *cr); err != nil {
		return fmt.Errorf("unable to create a deployment status: %w", err)
	}
	return nil
}

func generateCheckRun(e Event) *github.CheckRun {
	applicationURL := fmt.Sprintf("%s/applications/%s", e.ArgoCDURL, e.Application.Name)
	externalURLs := strings.Join(e.Application.Status.Summary.ExternalURLs, "\n")
	summary := fmt.Sprintf(`
## Argo CD
%s

## External URL
%s
`, applicationURL, externalURLs)

	if e.PhaseIsChanged {
		if e.Application.Status.OperationState == nil {
			return nil
		}
		switch e.Application.Status.OperationState.Phase {
		case synccommon.OperationRunning:
			return &github.CheckRun{
				Status:  "in_progress",
				Title:   fmt.Sprintf("Syncing: %s", e.Application.Status.OperationState.Message),
				Summary: summary,
			}
		case synccommon.OperationSucceeded:
			return &github.CheckRun{
				Status:     "completed",
				Conclusion: "success",
				Title:      fmt.Sprintf("Synced: %s", e.Application.Status.OperationState.Message),
				Summary:    summary,
			}
		case synccommon.OperationFailed:
			return &github.CheckRun{
				Status:     "completed",
				Conclusion: "failure",
				Title:      fmt.Sprintf("Sync Failed: %s", e.Application.Status.OperationState.Message),
				Summary:    summary,
			}
		case synccommon.OperationError:
			return &github.CheckRun{
				Status:     "completed",
				Conclusion: "failure",
				Title:      fmt.Sprintf("Sync Error: %s", e.Application.Status.OperationState.Message),
				Summary:    summary,
			}
		}
	}

	if e.HealthIsChanged {
		switch e.Application.Status.Health.Status {
		case health.HealthStatusProgressing:
			return &github.CheckRun{
				Status:  "in_progress",
				Title:   fmt.Sprintf("Progressing: %s", e.Application.Status.Health.Message),
				Summary: summary,
			}
		case health.HealthStatusHealthy:
			return &github.CheckRun{
				Status:     "completed",
				Conclusion: "success",
				Title:      fmt.Sprintf("Healthy: %s", e.Application.Status.Health.Message),
				Summary:    summary,
			}
		case health.HealthStatusDegraded:
			return &github.CheckRun{
				Status:     "completed",
				Conclusion: "failure",
				Title:      fmt.Sprintf("Degraded: %s", e.Application.Status.Health.Message),
				Summary:    summary,
			}
		}
	}

	return nil
}
