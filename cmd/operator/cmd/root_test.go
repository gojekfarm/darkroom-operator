package cmd

import (
	"bytes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
	"time"
)

type RootCmdSuite struct {
	suite.Suite
	rootCmd *cobra.Command
	buf     *bytes.Buffer
	stopCh  chan struct{}
}

func TestRootCmd(t *testing.T) {
	suite.Run(t, new(RootCmdSuite))
}

func (s *RootCmdSuite) SetupTest() {
	s.stopCh = make(chan struct{})
	s.buf = &bytes.Buffer{}
	s.rootCmd = newRootCmd(rootCmdOpts{
		SetupSignalHandler: func() <-chan struct{} {
			return s.stopCh
		},
	})
	s.rootCmd.SetOut(s.buf)
}

func (s *RootCmdSuite) TestControllerStartup() {
	errCh := make(chan error)

	go func() {
		defer close(errCh)
		errCh <- s.rootCmd.Execute()
	}()

	s.True(s.Eventually(func() bool {
		return strings.Contains(s.buf.String(), "starting manager")
	}, 5*time.Second, 100*time.Millisecond), "failed to start controller")

	close(s.stopCh)
	s.NoError(<-errCh)
}
