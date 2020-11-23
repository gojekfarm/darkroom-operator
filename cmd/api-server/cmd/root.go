package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/gojekfarm/darkroom-operator/cmd/version"
	apiserver "github.com/gojekfarm/darkroom-operator/internal/api-server"
	deploymentsv1alpha1 "github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"
	pkglog "github.com/gojekfarm/darkroom-operator/pkg/log"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = pkglog.Log.WithName("api-server").WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(deploymentsv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func newRootCmd(opts rootCmdOpts) *cobra.Command {
	args := struct {
		port int
	}{}
	cmd := &cobra.Command{
		RunE: func(c *cobra.Command, _ []string) error {
			mgr, err := opts.NewManager(opts.GetConfigOrDie(), apiserver.Options{
				Scheme: scheme,
				Port:   args.port,
			})
			if err != nil {
				setupLog.Error(err, "unable to create api-server manager")
				return err
			}

			setupLog.Info("starting api-server manager")
			return mgr.Start(opts.SetupSignalHandler())
		},
	}
	cmd.PersistentFlags().IntVarP(&args.port, "port", "p", 5000, "port used by the api-server")
	return cmd
}

type rootCmdOpts struct {
	SetupSignalHandler func() (stopCh <-chan struct{})
	NewManager         func(*rest.Config, apiserver.Options) (apiserver.Manager, error)
	GetConfigOrDie     func() *rest.Config
}

func NewRootCmd() *cobra.Command {
	cmd := newRootCmd(rootCmdOpts{
		SetupSignalHandler: ctrl.SetupSignalHandler,
		NewManager:         apiserver.NewManager,
		GetConfigOrDie:     ctrl.GetConfigOrDie,
	})
	cmd.AddCommand(version.New())
	return cmd
}
