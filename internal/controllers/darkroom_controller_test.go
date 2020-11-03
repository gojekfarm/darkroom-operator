/*
MIT License

Copyright (c) 2020 GO-JEK Tech

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package controllers

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	deploymentsv1alpha1 "github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"
	// +kubebuilder:scaffold:imports
)

type DarkroomControllerSuite struct {
	suite.Suite
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
	buf       *bytes.Buffer
}

func TestDarkroomControllerSuite(t *testing.T) {
	suite.Run(t, new(DarkroomControllerSuite))
}

func (s *DarkroomControllerSuite) SetupSuite() {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(s.buf)))
	s.testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "config", "crd", "bases")},
	}

	var err error
	s.cfg, err = s.testEnv.Start()
	s.NoError(err)
	s.NotNil(s.cfg)

	err = deploymentsv1alpha1.AddToScheme(scheme.Scheme)
	s.NoError(err)

	// +kubebuilder:scaffold:scheme

	s.k8sClient, err = client.New(s.cfg, client.Options{Scheme: scheme.Scheme})
	s.NoError(err)
	s.NotNil(s.k8sClient)
}

func (s *DarkroomControllerSuite) TearDownSuite() {
	err := s.testEnv.Stop()
	s.NoError(err)
}
