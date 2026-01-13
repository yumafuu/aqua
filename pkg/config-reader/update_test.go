package reader_test

import (
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	reader "github.com/aquaproj/aqua/v2/pkg/config-reader"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
	"github.com/google/go-cmp/cmp"
)

func Test_configReader_ReadToUpdate(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name           string
		cfg            *aqua.Config
		cfgs           map[string]*aqua.Config
		isErr          bool
		errContains    []string // expected substrings in error message
		files          map[string]string
		configFilePath string
		homeDir        string
	}{
		{
			name:  "file isn't found",
			isErr: true,
		},
		{
			name: "normal",
			files: map[string]string{
				"/home/workspace/foo/aqua.yaml": `registries:
- type: standard
  ref: v2.5.0
- type: local
  name: local
  path: registry.yaml
packages:`,
			},
			configFilePath: "/home/workspace/foo/aqua.yaml",
			cfg: &aqua.Config{
				Registries: aqua.Registries{
					"standard": {
						Type:      "github_content",
						Name:      "standard",
						Ref:       "v2.5.0",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Path:      "registry.yaml",
					},
					"local": {
						Type: "local",
						Name: "local",
						Path: "/home/workspace/foo/registry.yaml",
					},
				},
				Packages: []*aqua.Package{},
			},
			cfgs: map[string]*aqua.Config{},
		},
		{
			name: "import package",
			files: map[string]string{
				"/home/workspace/foo/aqua.yaml": `registries:
- type: standard
  ref: v2.5.0
packages:
- name: suzuki-shunsuke/ci-info@v1.0.0
- import: aqua-installer.yaml
`,
				"/home/workspace/foo/aqua-installer.yaml": `packages:
- name: aquaproj/aqua-installer@v1.0.0
`,
			},
			configFilePath: "/home/workspace/foo/aqua.yaml",
			cfg: &aqua.Config{
				Registries: aqua.Registries{
					"standard": {
						Type:      "github_content",
						Name:      "standard",
						Ref:       "v2.5.0",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Path:      "registry.yaml",
					},
				},
				Packages: []*aqua.Package{
					{
						Name:     "suzuki-shunsuke/ci-info",
						Registry: "standard",
						Version:  "v1.0.0",
					},
				},
			},
			cfgs: map[string]*aqua.Config{
				"/home/workspace/foo/aqua-installer.yaml": {
					Packages: []*aqua.Package{
						{
							Name:     "aquaproj/aqua-installer",
							Registry: "standard",
							Version:  "v1.0.0",
						},
					},
					Registries: aqua.Registries{
						"standard": {
							Type:      "github_content",
							Name:      "standard",
							Ref:       "v2.5.0",
							RepoOwner: "aquaproj",
							RepoName:  "aqua-registry",
							Path:      "registry.yaml",
						},
					},
				},
			},
		},
		{
			name: "circular import - self reference",
			files: map[string]string{
				"/home/workspace/foo/aqua.yaml": `registries:
- type: standard
  ref: v2.5.0
packages:
- import: aqua.yaml
`,
			},
			configFilePath: "/home/workspace/foo/aqua.yaml",
			isErr:          true,
			errContains:    []string{"circular import detected", "aqua.yaml -> aqua.yaml"},
		},
		{
			name: "circular import - A imports B, B imports A",
			files: map[string]string{
				"/home/workspace/foo/aqua.yaml": `registries:
- type: standard
  ref: v2.5.0
packages:
- import: b.yaml
`,
				"/home/workspace/foo/b.yaml": `packages:
- import: aqua.yaml
`,
			},
			configFilePath: "/home/workspace/foo/aqua.yaml",
			isErr:          true,
			errContains:    []string{"circular import detected", "aqua.yaml -> b.yaml -> aqua.yaml"},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs, err := testutil.NewFs(d.files)
			if err != nil {
				t.Fatal(err)
			}
			reader := reader.New(fs, &config.Param{
				HomeDir: d.homeDir,
			})
			cfg := &aqua.Config{}
			cfgs, err := reader.ReadToUpdate(d.configFilePath, cfg)
			if err != nil {
				if d.isErr {
					// Verify error message contains expected substrings
					for _, substr := range d.errContains {
						if !strings.Contains(err.Error(), substr) {
							t.Errorf("error message should contain %q, got: %s", substr, err.Error())
						}
					}
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if diff := cmp.Diff(d.cfg, cfg); diff != "" {
				t.Fatal("cfg:", diff)
			}
			if diff := cmp.Diff(d.cfgs, cfgs); diff != "" {
				t.Fatal("cfgs:", diff)
			}
		})
	}
}
