package declarative

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
	"github.com/kudobuilder/kuttl/pkg/test"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/krusty"
)

// Resource kinds from kuttl.
const (
	TestAssertKind = "TestAssert"
	TestStepKind   = "TestStep"
)

// Harness embeds kuttl Harness and adds support for running the test suite
// with kustomize.
type Harness struct {
	test.Harness
	fs         filesys.FileSystem
	harnessDir string
}

// Cleanup deletes all the files generated by the test harness.
func (h *Harness) Cleanup() {
	h.T.Log("Cleaning up", h.harnessDir)
	h.fs.RemoveAll(h.harnessDir)

	// TODO: Support cleanup of all the created resources separately that can
	// be called as a deferred function at the end of the caller function.
}

// NewKustomizedHarness runs kustomization on a given directory to generate
// manifests to be used with kuttl test suite.
func NewKustomizedHarness(path string, namespace string, t *testing.T) (*Harness, error) {
	// Default test kind.
	testKind := TestStepKind

	testName := filepath.Base(path)

	// Set up krusty with local filesystem and run kustomize at the given path.
	fs := filesys.MakeFsOnDisk()
	opt := krusty.MakeDefaultOptions()
	k := krusty.MakeKustomizer(opt)
	resmap, err := k.Run(fs, path)
	if err != nil {
		return nil, err
	}

	// Analyze the ResMap to find kuttl resources and determine the type of
	// kuttl resource to create.
	for _, res := range resmap.Resources() {
		if res.GetKind() == TestAssertKind {
			testKind = TestAssertKind
		}
	}

	t.Log("Test type:", testKind)

	r, err := resmap.AsYaml()
	if err != nil {
		return nil, err
	}

	// Create a temporary directory for the test suite.
	dir, err := ioutil.TempDir("", fmt.Sprintf("test-harness-%s", testName))
	if err != nil {
		return nil, err
	}
	t.Log("Test harness dir:", dir)
	// Create a test directory.
	if err := fs.MkdirAll(filepath.Join(dir, testName)); err != nil {
		return nil, err
	}

	// Set the resource file name based on the resource kind.
	var filename string
	if testKind == TestStepKind {
		filename = "01-step.yaml"
	} else if testKind == TestAssertKind {
		filename = "01-assert.yaml"
	}
	// TODO: Add support for error assertion.

	// Write the kustomized result.
	if err := fs.WriteFile(filepath.Join(dir, testName, filename), r); err != nil {
		return nil, err
	}

	// TODO: Write other files as well that aren't part of kustomization. They
	// may be used by a kuttl step.

	// Create and return a kuttl harness with the above test data as test
	// suite.
	return &Harness{
		Harness: test.Harness{
			TestSuite: harness.TestSuite{
				TestDirs:   []string{dir},
				Namespace:  namespace,
				SkipDelete: true, // Skip delete to persist the objects to be used in other harnesses.
			},
			T: t,
		},
		fs:         fs,
		harnessDir: dir,
	}, nil
}
