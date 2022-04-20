package tfaddr

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	svchost "github.com/hashicorp/terraform-svchost"
)

func TestParseModuleSource(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    ModuleSource
		wantErr string
	}{
		// Local paths
		"local in subdirectory": {
			input: "./child",
			want:  ModuleSourceLocal("./child"),
		},
		"local in subdirectory non-normalized": {
			input: "./nope/../child",
			want:  ModuleSourceLocal("./child"),
		},
		"local in sibling directory": {
			input: "../sibling",
			want:  ModuleSourceLocal("../sibling"),
		},
		"local in sibling directory non-normalized": {
			input: "./nope/../../sibling",
			want:  ModuleSourceLocal("../sibling"),
		},
		"Windows-style local in subdirectory": {
			input: `.\child`,
			want:  ModuleSourceLocal("./child"),
		},
		"Windows-style local in subdirectory non-normalized": {
			input: `.\nope\..\child`,
			want:  ModuleSourceLocal("./child"),
		},
		"Windows-style local in sibling directory": {
			input: `..\sibling`,
			want:  ModuleSourceLocal("../sibling"),
		},
		"Windows-style local in sibling directory non-normalized": {
			input: `.\nope\..\..\sibling`,
			want:  ModuleSourceLocal("../sibling"),
		},
		"an abominable mix of different slashes": {
			input: `./nope\nope/why\./please\don't`,
			want:  ModuleSourceLocal("./nope/nope/why/please/don't"),
		},

		// Registry addresses
		// (NOTE: There is another test function TestParseModuleSourceRegistry
		// which tests this situation more exhaustively, so this is just a
		// token set of cases to see that we are indeed calling into the
		// registry address parser when appropriate.)
		"main registry implied": {
			input: "hashicorp/subnets/cidr",
			want: ModuleSourceRegistry{
				PackageAddr: ModuleRegistryPackage{
					Host:         svchost.Hostname("registry.terraform.io"),
					Namespace:    "hashicorp",
					Name:         "subnets",
					TargetSystem: "cidr",
				},
				Subdir: "",
			},
		},
		"main registry implied, subdir": {
			input: "hashicorp/subnets/cidr//examples/foo",
			want: ModuleSourceRegistry{
				PackageAddr: ModuleRegistryPackage{
					Host:         svchost.Hostname("registry.terraform.io"),
					Namespace:    "hashicorp",
					Name:         "subnets",
					TargetSystem: "cidr",
				},
				Subdir: "examples/foo",
			},
		},
		"main registry implied, escaping subdir": {
			input:   "hashicorp/subnets/cidr//../nope",
			wantErr: `unsupported module source "hashicorp/subnets/cidr//../nope"`,
		},
		"custom registry": {
			input: "example.com/awesomecorp/network/happycloud",
			want: ModuleSourceRegistry{
				PackageAddr: ModuleRegistryPackage{
					Host:         svchost.Hostname("example.com"),
					Namespace:    "awesomecorp",
					Name:         "network",
					TargetSystem: "happycloud",
				},
				Subdir: "",
			},
		},
		"custom registry, subdir": {
			input: "example.com/awesomecorp/network/happycloud//examples/foo",
			want: ModuleSourceRegistry{
				PackageAddr: ModuleRegistryPackage{
					Host:         svchost.Hostname("example.com"),
					Namespace:    "awesomecorp",
					Name:         "network",
					TargetSystem: "happycloud",
				},
				Subdir: "examples/foo",
			},
		},

		"relative path without the needed prefix": {
			input: "boop/bloop",
			// For this case we return a generic error message from the addrs
			// layer, but using a specialized error type which our module
			// installer checks for and produces an extra hint for users who
			// were intending to write a local path which then got
			// misinterpreted as a remote source due to the missing prefix.
			// However, the main message is generic here because this is really
			// just a general "this string doesn't match any of our source
			// address patterns" situation, not _necessarily_ about relative
			// local paths.
			wantErr: `unsupported module source "boop/bloop"`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			addr, err := ParseRawModuleSource(test.input)

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

func TestParseModuleSourceRegistry(t *testing.T) {
	// We test parseModuleSourceRegistry alone here, in addition to testing
	// it indirectly as part of TestParseModuleSource, because general
	// module parsing unfortunately eats all of the error situations from
	// registry passing by falling back to trying for a direct remote package
	// address.

	// Historical note: These test cases were originally derived from the
	// ones in the old internal/registry/regsrc package that the
	// ModuleSourceRegistry type is replacing. That package had the notion
	// of "normalized" addresses as separate from the original user input,
	// but this new implementation doesn't try to preserve the original
	// user input at all, and so the main string output is always normalized.
	//
	// That package also had some behaviors to turn the namespace, name, and
	// remote system portions into lowercase, but apparently we didn't
	// actually make use of that in the end and were preserving the case
	// the user provided in the input, and so for backward compatibility
	// we're continuing to do that here, at the expense of now making the
	// "ForDisplay" output case-preserving where its predecessor in the
	// old package wasn't. The main Terraform Registry at registry.terraform.io
	// is itself case-insensitive anyway, so our case-preserving here is
	// entirely for the benefit of existing third-party registry
	// implementations that might be case-sensitive, which we must remain
	// compatible with now.

	tests := map[string]struct {
		input           string
		wantString      string
		wantForDisplay  string
		wantForProtocol string
		wantErr         string
	}{
		"public registry": {
			input:           `hashicorp/consul/aws`,
			wantString:      `registry.terraform.io/hashicorp/consul/aws`,
			wantForDisplay:  `hashicorp/consul/aws`,
			wantForProtocol: `hashicorp/consul/aws`,
		},
		"public registry with subdir": {
			input:           `hashicorp/consul/aws//foo`,
			wantString:      `registry.terraform.io/hashicorp/consul/aws//foo`,
			wantForDisplay:  `hashicorp/consul/aws//foo`,
			wantForProtocol: `hashicorp/consul/aws`,
		},
		"public registry using explicit hostname": {
			input:           `registry.terraform.io/hashicorp/consul/aws`,
			wantString:      `registry.terraform.io/hashicorp/consul/aws`,
			wantForDisplay:  `hashicorp/consul/aws`,
			wantForProtocol: `hashicorp/consul/aws`,
		},
		"public registry with mixed case names": {
			input:           `HashiCorp/Consul/aws`,
			wantString:      `registry.terraform.io/HashiCorp/Consul/aws`,
			wantForDisplay:  `HashiCorp/Consul/aws`,
			wantForProtocol: `HashiCorp/Consul/aws`,
		},
		"private registry with non-standard port": {
			input:           `Example.com:1234/HashiCorp/Consul/aws`,
			wantString:      `example.com:1234/HashiCorp/Consul/aws`,
			wantForDisplay:  `example.com:1234/HashiCorp/Consul/aws`,
			wantForProtocol: `HashiCorp/Consul/aws`,
		},
		"private registry with IDN hostname": {
			input:           `Испытание.com/HashiCorp/Consul/aws`,
			wantString:      `испытание.com/HashiCorp/Consul/aws`,
			wantForDisplay:  `испытание.com/HashiCorp/Consul/aws`,
			wantForProtocol: `HashiCorp/Consul/aws`,
		},
		"private registry with IDN hostname and non-standard port": {
			input:           `Испытание.com:1234/HashiCorp/Consul/aws//Foo`,
			wantString:      `испытание.com:1234/HashiCorp/Consul/aws//Foo`,
			wantForDisplay:  `испытание.com:1234/HashiCorp/Consul/aws//Foo`,
			wantForProtocol: `HashiCorp/Consul/aws`,
		},
		"invalid hostname": {
			input:   `---.com/HashiCorp/Consul/aws`,
			wantErr: `invalid module registry hostname "---.com"; internationalized domain names must be given as direct unicode characters, not in punycode`,
		},
		"hostname with only one label": {
			// This was historically forbidden in our initial implementation,
			// so we keep it forbidden to avoid newly interpreting such
			// addresses as registry addresses rather than remote source
			// addresses.
			input:   `foo/var/baz/qux`,
			wantErr: `invalid module registry hostname: must contain at least one dot`,
		},
		"invalid target system characters": {
			input:   `foo/var/no-no-no`,
			wantErr: `invalid target system "no-no-no": must be between one and 64 ASCII letters or digits`,
		},
		"invalid target system length": {
			input:   `foo/var/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaah`,
			wantErr: `invalid target system "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaah": must be between one and 64 ASCII letters or digits`,
		},
		"invalid namespace": {
			input:   `boop!/var/baz`,
			wantErr: `invalid namespace "boop!": must be between one and 64 characters, including ASCII letters, digits, dashes, and underscores, where dashes and underscores may not be the prefix or suffix`,
		},
		"missing part with explicit hostname": {
			input:   `foo.com/var/baz`,
			wantErr: `source address must have three more components after the hostname: the namespace, the name, and the target system`,
		},
		"errant query string": {
			input:   `foo/var/baz?otherthing`,
			wantErr: `module registry addresses may not include a query string portion`,
		},
		"github.com": {
			// We don't allow using github.com like a module registry because
			// that conflicts with the historically-supported shorthand for
			// installing directly from GitHub-hosted git repositories.
			input:   `github.com/HashiCorp/Consul/aws`,
			wantErr: `can't use "github.com" as a module registry host, because it's reserved for installing directly from version control repositories`,
		},
		"bitbucket.org": {
			// We don't allow using bitbucket.org like a module registry because
			// that conflicts with the historically-supported shorthand for
			// installing directly from BitBucket-hosted git repositories.
			input:   `bitbucket.org/HashiCorp/Consul/aws`,
			wantErr: `can't use "bitbucket.org" as a module registry host, because it's reserved for installing directly from version control repositories`,
		},
		"local path from current dir": {
			// Can't use a local path when we're specifically trying to parse
			// a _registry_ source address.
			input:   `./boop`,
			wantErr: `can't use local directory "./boop" as a module registry address`,
		},
		"local path from parent dir": {
			// Can't use a local path when we're specifically trying to parse
			// a _registry_ source address.
			input:   `../boop`,
			wantErr: `can't use local directory "../boop" as a module registry address`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			addrI, err := ParseRawModuleSourceRegistry(test.input)

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

			addr, ok := addrI.(ModuleSourceRegistry)
			if !ok {
				t.Fatalf("wrong address type %T; want %T", addrI, addr)
			}

			if got, want := addr.String(), test.wantString; got != want {
				t.Errorf("wrong String() result\ngot:  %s\nwant: %s", got, want)
			}
			if got, want := addr.ForDisplay(), test.wantForDisplay; got != want {
				t.Errorf("wrong ForDisplay() result\ngot:  %s\nwant: %s", got, want)
			}
			if got, want := addr.PackageAddr.ForRegistryProtocol(), test.wantForProtocol; got != want {
				t.Errorf("wrong ForRegistryProtocol() result\ngot:  %s\nwant: %s", got, want)
			}
		})
	}
}
