package setup

import (
	"context"

	"github.com/solo-io/gloo/pkg/version"

	"github.com/solo-io/reporting-client/pkg/client"

	"github.com/solo-io/gloo/pkg/utils/setuputils"
	"github.com/solo-io/gloo/projects/gloo/pkg/syncer"
)

func Main(customCtx context.Context) error {
	// TODO: Figure out if we want to recreate usage reporter without the envoy metrics service
	return startSetupLoop(customCtx, nil)
}

func StartGlooInTest(customCtx context.Context) error {
	return startSetupLoop(customCtx, nil)
}

func startSetupLoop(ctx context.Context, usageReporter client.UsagePayloadReader) error {
	return setuputils.Main(setuputils.SetupOpts{
		LoggerName:    "gloo",
		Version:       version.Version,
		SetupFunc:     syncer.NewSetupFunc(),
		ExitOnError:   true,
		CustomCtx:     ctx,
		UsageReporter: usageReporter,
	})
}
