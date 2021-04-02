package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gojekfarm/darkroom-operator/cmd/version"
	apiserver "github.com/gojekfarm/darkroom-operator/internal/api-server"
	"github.com/gojekfarm/darkroom-operator/internal/runtime"
	pkglog "github.com/gojekfarm/darkroom-operator/pkg/log"
)

var (
	setupLog = pkglog.Log.WithName("api-server").WithName("setup")
)

func newRootCmd(opts rootCmdOpts) *cobra.Command {
	args := struct {
		port int
	}{}
	cmd := &cobra.Command{
		RunE: func(c *cobra.Command, _ []string) error {
			pkglog.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(c.OutOrStderr())))

			mgr, err := opts.NewManager(opts.GetConfigOrDie(), apiserver.Options{
				Scheme: runtime.Scheme(),
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
	SetupSignalHandler func() context.Context
	NewManager         apiserver.NewManagerFunc
	GetConfigOrDie     func() *rest.Config
}

func NewRootCmd() *cobra.Command {
	cmd := newRootCmd(rootCmdOpts{
		SetupSignalHandler: ctrl.SetupSignalHandler,
		NewManager: apiserver.NewManager(apiserver.NewManagerFuncOptions{
			NewDynamicRESTMapper: apiutil.NewDynamicRESTMapper,
			NewCache:             cache.New,
			NewClientBuilder:     manager.NewClientBuilder(),
		}),
		GetConfigOrDie: ctrl.GetConfigOrDie,
	})
	cmd.AddCommand(version.New())
	return cmd
}
