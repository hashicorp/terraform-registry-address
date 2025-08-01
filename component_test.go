// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfaddr

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	svchost "github.com/hashicorp/terraform-svchost"
)

func TestParseComponentSource_simple(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    Component
		wantErr string
	}{
		"main registry implied": {
			input: "hashicorp/k8-cluster",
			want: Component{
				Package: ComponentPackage{
					Host:      svchost.Hostname("registry.terraform.io"),
					Namespace: "hashicorp",
					Name:      "k8-cluster",
				},
				Subdir: "",
			},
		},
		"main registry implied, subdir": {
			input: "hashicorp/k8-cluster//path/to/dir",
			want: Component{
				Package: ComponentPackage{
					Host:      svchost.Hostname("registry.terraform.io"),
					Namespace: "hashicorp",
					Name:      "k8-cluster",
				},
				Subdir: "path/to/dir",
			},
		},
		"custom registry": {
			input: "terraform.registry.io/hashicorp/k8-cluster",
			want: Component{
				Package: ComponentPackage{
					Host:      svchost.Hostname("terraform.registry.io"),
					Namespace: "hashicorp",
					Name:      "k8-cluster",
				},
				Subdir: "",
			},
		},
		"custom registry, subdir": {
			input: "app.terraform.io/hashicorp/k8-cluster//examples/foo",
			want: Component{
				Package: ComponentPackage{
					Host:      svchost.Hostname("app.terraform.io"),
					Namespace: "hashicorp",
					Name:      "k8-cluster",
				},
				Subdir: "examples/foo",
			},
		},
		"private registry": {
			input: "example.com/awesomecorp/network",
			want: Component{
				Package: ComponentPackage{
					Host:      svchost.Hostname("example.com"),
					Namespace: "awesomecorp",
					Name:      "network",
				},
				Subdir: "",
			},
		},
		"private registry, subdir": {
			input: "example.com/awesomecorp/network//configs",
			want: Component{
				Package: ComponentPackage{
					Host:      svchost.Hostname("example.com"),
					Namespace: "awesomecorp",
					Name:      "network",
				},
				Subdir: "configs",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			addr, err := ParseComponentSource(test.input)

			if test.wantErr != "" {
				switch {
				case err == nil:
					t.Errorf("unexpected success\nwant error: %s", test.wantErr)
				case err.Error() != test.wantErr:
					t.Errorf("wrong error messages\ngot:  %s\nwant: %s", err.Error(), test.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			if diff := cmp.Diff(addr, test.want); diff != "" {
				t.Errorf("wrong result\n%s", diff)
			}
		})
	}
}

func TestParseComponentSource(t *testing.T) {
	tests := map[string]struct {
		input           string
		wantString      string
		wantForDisplay  string
		wantForProtocol string
		wantErr         string
	}{
		"public registry": {
			input:           `hashicorp/k8-cluster`,
			wantString:      `registry.terraform.io/hashicorp/k8-cluster`,
			wantForDisplay:  `hashicorp/k8-cluster`,
			wantForProtocol: `hashicorp/k8-cluster`,
		},
		"public registry with subdir": {
			input:           `hashicorp/k8-cluster//configs`,
			wantString:      `registry.terraform.io/hashicorp/k8-cluster//configs`,
			wantForDisplay:  `hashicorp/k8-cluster//configs`,
			wantForProtocol: `hashicorp/k8-cluster`,
		},
		"public registry using explicit hostname": {
			input:           `registry.terraform.io/hashicorp/k8-cluster`,
			wantString:      `registry.terraform.io/hashicorp/k8-cluster`,
			wantForDisplay:  `hashicorp/k8-cluster`,
			wantForProtocol: `hashicorp/k8-cluster`,
		},
		"terraform.registry.io example": {
			input:           `terraform.registry.io/hashicorp/k8-cluster`,
			wantString:      `terraform.registry.io/hashicorp/k8-cluster`,
			wantForDisplay:  `terraform.registry.io/hashicorp/k8-cluster`,
			wantForProtocol: `hashicorp/k8-cluster`,
		},
		"app.terraform.io example": {
			input:           `app.terraform.io/hashicorp/k8-cluster`,
			wantString:      `app.terraform.io/hashicorp/k8-cluster`,
			wantForDisplay:  `app.terraform.io/hashicorp/k8-cluster`,
			wantForProtocol: `hashicorp/k8-cluster`,
		},
		"app.terraform.io with subdir": {
			input:           `app.terraform.io/hashicorp/k8-cluster//path/to/dir`,
			wantString:      `app.terraform.io/hashicorp/k8-cluster//path/to/dir`,
			wantForDisplay:  `app.terraform.io/hashicorp/k8-cluster//path/to/dir`,
			wantForProtocol: `hashicorp/k8-cluster`,
		},
		"public registry with mixed case names": {
			input:           `HashiCorp/K8-Cluster`,
			wantString:      `registry.terraform.io/HashiCorp/K8-Cluster`,
			wantForDisplay:  `HashiCorp/K8-Cluster`,
			wantForProtocol: `HashiCorp/K8-Cluster`,
		},
		"private registry with non-standard port": {
			input:           `Example.com:1234/HashiCorp/K8-Cluster`,
			wantString:      `example.com:1234/HashiCorp/K8-Cluster`,
			wantForDisplay:  `example.com:1234/HashiCorp/K8-Cluster`,
			wantForProtocol: `HashiCorp/K8-Cluster`,
		},
		"private registry with IDN hostname": {
			input:           `Испытание.com/HashiCorp/K8-Cluster`,
			wantString:      `испытание.com/HashiCorp/K8-Cluster`,
			wantForDisplay:  `испытание.com/HashiCorp/K8-Cluster`,
			wantForProtocol: `HashiCorp/K8-Cluster`,
		},
		"private registry with IDN hostname and non-standard port": {
			input:           `Испытание.com:1234/HashiCorp/K8-Cluster//Foo`,
			wantString:      `испытание.com:1234/HashiCorp/K8-Cluster//Foo`,
			wantForDisplay:  `испытание.com:1234/HashiCorp/K8-Cluster//Foo`,
			wantForProtocol: `HashiCorp/K8-Cluster`,
		},
		"invalid hostname": {
			input:   `---.com/HashiCorp/K8-Cluster`,
			wantErr: `invalid component registry hostname "---.com"; internationalized domain names must be given as direct unicode characters, not in punycode`,
		},
		"hostname with only one label": {
			// This was historically forbidden in our initial implementation,
			// so we keep it forbidden to avoid newly interpreting such
			// addresses as registry addresses rather than remote source
			// addresses.
			input:   `foo/var/baz`,
			wantErr: `invalid component registry hostname: must contain at least one dot`,
		},
		"invalid namespace": {
			input:   `boop!/var`,
			wantErr: `invalid namespace "boop!": must be between one and 64 characters, including ASCII letters, digits, dashes, and underscores, where dashes and underscores may not be the prefix or suffix`,
		},
		"invalid component name": {
			input:   `hashicorp/no-no-no!`,
			wantErr: `invalid component name "no-no-no!": must be between one and 64 characters, including ASCII letters, digits, dashes, and underscores, where dashes and underscores may not be the prefix or suffix`,
		},
		"missing part with explicit hostname": {
			input:   `foo.com/var`,
			wantErr: `source address must have two more components after the hostname: the namespace and the name`,
		},
		"errant query string": {
			input:   `foo/var?otherthing`,
			wantErr: `component registry addresses may not include a query string portion`,
		},
		"github.com": {
			// We don't allow using github.com like a component registry because
			// that conflicts with the historically-supported shorthand for
			// installing directly from GitHub-hosted git repositories.
			input:   `github.com/HashiCorp/K8-Cluster`,
			wantErr: `can't use "github.com" as a component registry host, because it's reserved for installing directly from version control repositories`,
		},
		"bitbucket.org": {
			// We don't allow using bitbucket.org like a component registry because
			// that conflicts with the historically-supported shorthand for
			// installing directly from BitBucket-hosted git repositories.
			input:   `bitbucket.org/HashiCorp/K8-Cluster`,
			wantErr: `can't use "bitbucket.org" as a component registry host, because it's reserved for installing directly from version control repositories`,
		},
		"gitlab.com": {
			// We don't allow using gitlab.com like a component registry because
			// that conflicts with the historically-supported shorthand for
			// installing directly from GitLab-hosted git repositories.
			input:   `gitlab.com/HashiCorp/K8-Cluster`,
			wantErr: `can't use "gitlab.com" as a component registry host, because it's reserved for installing directly from version control repositories`,
		},
		"local path from current dir": {
			// Can't use a local path when we're specifically trying to parse
			// a _registry_ source address.
			input:   `./boop`,
			wantErr: `source address must have two more components after the hostname: the namespace and the name`,
		},
		"local path from parent dir": {
			// Can't use a local path when we're specifically trying to parse
			// a _registry_ source address.
			input:   `../boop`,
			wantErr: `source address must have two more components after the hostname: the namespace and the name`,
		},
		"main registry implied, escaping subdir": {
			input:   "hashicorp/k8-cluster//../nope",
			wantErr: `subdirectory path "../nope" leads outside of the component package`,
		},
		"too few segments": {
			input:   "boop",
			wantErr: "a component registry source address must have either two or three slash-separated segments",
		},
		"too many segments": {
			input:   "registry.terraform.io/hashicorp/k8-cluster/extra/segment",
			wantErr: "a component registry source address must have either two or three slash-separated segments",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			addr, err := ParseComponentSource(test.input)

			if test.wantErr != "" {
				switch {
				case err == nil:
					t.Errorf("unexpected success\nwant error: %s", test.wantErr)
				case err.Error() != test.wantErr:
					t.Errorf("wrong error messages\ngot:  %s\nwant: %s", err.Error(), test.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			if got, want := addr.String(), test.wantString; got != want {
				t.Errorf("wrong String() result\ngot:  %s\nwant: %s", got, want)
			}
			if got, want := addr.ForDisplay(), test.wantForDisplay; got != want {
				t.Errorf("wrong ForDisplay() result\ngot:  %s\nwant: %s", got, want)
			}
			if got, want := addr.Package.ForRegistryProtocol(), test.wantForProtocol; got != want {
				t.Errorf("wrong ForRegistryProtocol() result\ngot:  %s\nwant: %s", got, want)
			}
		})
	}
}
