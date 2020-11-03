package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/gojekfarm/darkroom-operator/internal/controllers"
	deploymentsv1alpha1 "github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(deploymentsv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func newRootCmd(opts rootCmdOpts) *cobra.Command {
	args := struct {
		metricsAddr          string
		enableLeaderElection bool
	}{}
	cmd := &cobra.Command{
		Use:   "darkroom-operator",
		Short: "Darkroom Operator helps deploy Darkroom in a Kubernetes Cluster",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctrl.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(cmd.OutOrStderr())))

			mgr, err := opts.NewManager(opts.GetConfigOrDie(), ctrl.Options{
				Scheme:             scheme,
				MetricsBindAddress: args.metricsAddr,
				Port:               9443,
				LeaderElection:     args.enableLeaderElection,
				LeaderElectionID:   "750f7516.gojek.io",
			})
			if err != nil {
				setupLog.Error(err, "unable to start manager")
				return err
			}

			if err = (&controllers.DarkroomReconciler{
				Client: mgr.GetClient(),
				Log:    ctrl.Log.WithName("controllers").WithName("Darkroom"),
				Scheme: mgr.GetScheme(),
			}).SetupWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create controller", "controller", "Darkroom")
				return err
			}
			// +kubebuilder:scaffold:builder

			setupLog.Info("starting manager")
			if err := mgr.Start(opts.SetupSignalHandler()); err != nil {
				setupLog.Error(err, "problem running manager")
				return err
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(&args.metricsAddr, "metrics-addr", "b", ":8080", "The address the metric endpoint binds to.")
	cmd.PersistentFlags().BoolVar(&args.enableLeaderElection, "enable-leader-election", false, "Enable leader election for controller manager. "+
		"Enabling this will ensure there is only one active controller manager.")
	return cmd
}

type rootCmdOpts struct {
	SetupSignalHandler func() (stopCh <-chan struct{})
	NewManager         func(config *rest.Config, options ctrl.Options) (ctrl.Manager, error)
	GetConfigOrDie     func() *rest.Config
}

func NewRootCmd() *cobra.Command {
	return newRootCmd(rootCmdOpts{
		SetupSignalHandler: ctrl.SetupSignalHandler,
		NewManager:         ctrl.NewManager,
		GetConfigOrDie:     ctrl.GetConfigOrDie,
	})
}
