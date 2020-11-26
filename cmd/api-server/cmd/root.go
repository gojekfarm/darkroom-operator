package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

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
			pkglog.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(c.OutOrStderr())))

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
	NewManager         apiserver.NewManagerFunc
	GetConfigOrDie     func() *rest.Config
}

func NewRootCmd() *cobra.Command {
	cmd := newRootCmd(rootCmdOpts{
		SetupSignalHandler: ctrl.SetupSignalHandler,
		NewManager: apiserver.NewManager(apiserver.NewManagerFuncOptions{
			NewDynamicRESTMapper: apiutil.NewDynamicRESTMapper,
			NewCache:             cache.New,
			NewClient:            client.New,
		}),
		GetConfigOrDie: ctrl.GetConfigOrDie,
	})
	cmd.AddCommand(version.New())
	return cmd
}
