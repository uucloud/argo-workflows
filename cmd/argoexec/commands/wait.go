package commands

import (
	"time"

	"github.com/argoproj/pkg/stats"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewWaitCommand() *cobra.Command {
	var command = cobra.Command{
		Use:   "wait",
		Short: "wait for main container to finish and save artifacts",
		Run: func(cmd *cobra.Command, args []string) {
			err := waitContainer()
			if err != nil {
				log.Fatalf("%+v", err)
			}
		},
	}
	return &command
}

func waitContainer() error {
	wfExecutor := initExecutor()
	defer wfExecutor.HandleError() // Must be placed at the bottom of defers stack.
	defer stats.LogStats()
	stats.StartStatsTicker(5 * time.Minute)

	defer func() {
		// Killing sidecar containers
		err := wfExecutor.KillSidecars()
		if err != nil {
			wfExecutor.AddError(err)
		}
	}()

	// Wait for main container to complete
	err := wfExecutor.Wait(ctx)
	if err != nil {
		wfExecutor.AddError(err)
	}
	// Capture output script result
	err = wfExecutor.CaptureScriptResult()
	if err != nil {
		wfExecutor.AddError(err)
	}
	// Capture output script exit code
	err = wfExecutor.CaptureScriptExitCode()
	if err != nil {
		wfExecutor.AddError(err)
	}
	// Saving logs
	logArt, err := wfExecutor.SaveLogs()
	if err != nil {
		wfExecutor.AddError(err)
	}
	// Saving output parameters
	err = wfExecutor.SaveParameters()
	if err != nil {
		wfExecutor.AddError(err)
	}
	// Saving output artifacts
	err = wfExecutor.SaveArtifacts()
	if err != nil {
		wfExecutor.AddError(err)
	}
	// Annotating pod with output
	err = wfExecutor.AnnotateOutputs(ctx, logArt)
	if err != nil {
		wfExecutor.AddError(err)
	}

	return wfExecutor.HasError()
}
