package test

import (
	"context"
	"testing"

	harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
	"github.com/kudobuilder/kuttl/pkg/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/darkowlzz/kubecover/declarative"
)

func TestExample(t *testing.T) {
	// NOTE: Generate a common namespace for all the harnesses and delete it at
	// the very end. The deferred code for namespace delete need not be a
	// harness.

	testNamespace := "test11"

	// 1. Set up an application.
	// 2. Create a new configmap, separate from the one in the app setup.
	// 3. Create using kubectl.
	// 4. Assert the existence of application and configmap exists.
	// 5. Assert using go.
	// 6. Delete everything.
	// TODO:
	// - kuttl error assertion example after delete.

	h1 := test.Harness{
		TestSuite: harness.TestSuite{
			TestDirs:   []string{"testdata/app-setup-kuttl"},
			Namespace:  testNamespace,
			SkipDelete: true,
		},
		T: t,
	}
	// Can't run RunTests() directly because Harness.report is a private
	// attribute and can only be initialized by calling Harness.Setup(), which
	// Harness.Run() does.
	// h1.RunTests()
	h1.Run()

	// Creates a new configmap with patched data using kustomize.
	h2, err := declarative.NewKustomizedHarness("testdata/separate-cm", testNamespace, t)
	assert.Nil(t, err)
	defer h2.Cleanup()
	h2.Run()

	// Create a new configmap using kubectl create command.
	h3, err := declarative.NewKustomizedHarness("testdata/kubectl-create", testNamespace, t)
	assert.Nil(t, err)
	defer h3.Cleanup()
	h3.Run()

	// Runs kuttl assertion and checks if a configmap exists.
	h4, err := declarative.NewKustomizedHarness("testdata/configmap-exists", testNamespace, t)
	assert.Nil(t, err)
	defer h4.Cleanup()
	h4.Run()

	// Check if configmap exists using go.

	// Create a k8s client.
	cfg, err := config.GetConfig()
	assert.Nil(t, err)
	cl, err := client.New(cfg, client.Options{})
	assert.Nil(t, err)

	// Get the target configmap.
	cm := &corev1.ConfigMap{}
	nsn := client.ObjectKey{Name: "game-demo0", Namespace: testNamespace}
	assert.Nil(t, cl.Get(context.TODO(), nsn, cm))

	// Delete everything.
	h9 := test.Harness{
		TestSuite: harness.TestSuite{
			TestDirs:   []string{"testdata/uninstall-kuttl"},
			Namespace:  testNamespace,
			SkipDelete: true,
		},
		T: t,
	}
	h9.Run()
}
