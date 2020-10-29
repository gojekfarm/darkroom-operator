package cmd

import (
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var setupLog = ctrl.Log.WithName("setup")

func newRootCmd(opts rootCmdOpts) *cobra.Command {
	return &cobra.Command{
		Use: "darkroom-operator",
		Short: "Darkroom Operator helps deploy Darkroom in a Kubernetes Cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctrl.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(cmd.OutOrStderr())))

			setupLog.Info("starting manager")
			return nil
		},
	}
}

type rootCmdOpts struct {
	SetupSignalHandler func() (stopCh <-chan struct{})
}
