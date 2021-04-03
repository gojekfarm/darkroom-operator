package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/gojekfarm/darkroom-operator/cmd/version"
	"github.com/gojekfarm/darkroom-operator/internal/controllers"
	"github.com/gojekfarm/darkroom-operator/internal/runtime"
	pkglog "github.com/gojekfarm/darkroom-operator/pkg/log"
	// +kubebuilder:scaffold:imports
)

var (
	setupLog = pkglog.Log.WithName("operator").WithName("setup")
)

func newRootCmd(opts rootCmdOpts) *cobra.Command {
	args := struct {
		metricsAddr          string
		healthProbeAddr      string
		enableLeaderElection bool
		certDir              string
	}{}
	cmd := &cobra.Command{
		Use:   "darkroom-operator",
		Short: "Darkroom Operator helps deploy Darkroom in a Kubernetes Cluster",
		RunE: func(cmd *cobra.Command, _ []string) error {
			pkglog.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(cmd.OutOrStdout())))

			mgr, err := opts.NewManager(opts.GetConfigOrDie(), ctrl.Options{
				Scheme:                 runtime.Scheme(),
				MetricsBindAddress:     args.metricsAddr,
				Port:                   9443,
				HealthProbeBindAddress: args.healthProbeAddr,
				LeaderElection:         args.enableLeaderElection,
				LeaderElectionID:       "750f7516.gojek.io",
				CertDir:                args.certDir,
			})
			if err != nil {
				setupLog.Error(err, "unable to start manager")
				return err
			}

			r := &controllers.DarkroomReconciler{
				Client: mgr.GetClient(),
				Log:    pkglog.Log.WithName("controllers").WithName("darkroom-reconciler"),
				Scheme: mgr.GetScheme(),
			}

			if err = r.SetupControllerWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create controller", "controller", "Darkroom")
				return err
			}
			_ = r.SetupWebhookWithManager(mgr)

			// +kubebuilder:scaffold:builder

			if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
				setupLog.Error(err, "unable to set up health check")
				return err
			}
			if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
				setupLog.Error(err, "unable to set up ready check")
				return err
			}

			setupLog.Info("starting manager")
			if err := mgr.Start(opts.SetupSignalHandler()); err != nil {
				setupLog.Error(err, "problem running manager")
				return err
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&args.metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	cmd.PersistentFlags().StringVar(&args.healthProbeAddr, "health-probe-bind-address", ":8081", "The address the metric endpoint binds to.")
	cmd.PersistentFlags().StringVar(&args.certDir, "cert-dir", "", "The directory containing server certificate and key.")
	cmd.PersistentFlags().BoolVar(&args.enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager. "+
		"Enabling this will ensure there is only one active controller manager.")
	return cmd
}

type rootCmdOpts struct {
	SetupSignalHandler func() context.Context
	NewManager         func(config *rest.Config, options ctrl.Options) (ctrl.Manager, error)
	GetConfigOrDie     func() *rest.Config
}

func NewRootCmd() *cobra.Command {
	cmd := newRootCmd(rootCmdOpts{
		SetupSignalHandler: ctrl.SetupSignalHandler,
		NewManager:         ctrl.NewManager,
		GetConfigOrDie:     ctrl.GetConfigOrDie,
	})
	cmd.AddCommand(version.New())
	return cmd
}
