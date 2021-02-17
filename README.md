# terraform-registry-address

This package helps with representation, comparison and parsing of
Terraform Registry addresses, such as
`registry.terraform.io/grafana/grafana` or `hashicorp/aws`.

The most common source of these addresses outside of Terraform Core
is JSON representation of state, plan, or schemas as obtained
via [`hashicorp/terraform-exec`](https://github.com/hashicorp/terraform-exec).

## Example

```go
p, err := ParseProviderSourceString("hashicorp/aws")
if err != nil {
	// deal with error
}

// p == Provider{
//   Type:      "aws",
//   Namespace: "hashicorp",
//   Hostname:  svchost.Hostname("registry.terraform.io"),
// }
```

Please note that the `ParseProviderSourceString` is **NOT** equipped
to deal with legacy addresses such as `aws`. Such address will be parsed
as if provider belongs to `hashicorp` namespace in the public Registry.

## Legacy address

A legacy address is by itself (without more context) ambiguous.
For example `aws` may represent either the official `hashicorp/aws`
or just any custom-built provider called `aws`.

If you know the address was produced by Terraform <=0.12 and/or that you're
dealing with a legacy address, the following sequence of steps should be taken.

(optional) Parse such legacy address by `NewLegacyProvider(name)`.

Ask the Registry API whether and where the provider was moved to

(`-` represents the legacy, basically unknown namespace)

```sh
# grafana (redirected to its own namespace)
$ curl -s https://registry.terraform.io/v1/providers/-/grafana/versions | jq .moved_to
"grafana/grafana"

# aws (provider without redirection)
$ curl -s https://registry.terraform.io/v1/providers/-/aws/versions | jq .moved_to
null
```

Then:

 - Use `ParseProviderSourceString` for the _new_ (`moved_to`) address of any _moved_ provider (e.g. `grafana/grafana`).
 - Use `ImpliedProviderForUnqualifiedType` for any other provider (e.g. `aws`)
   - Depending on context `terraform` may also be parsed by `ParseProviderSourceString`,
   	 which assumes `hashicorp/terraform` provider. Read more about this provider below.

### `terraform` provider

Like any other legacy address `terraform` is also ambiguous. Such address may
(most unlikely) represent a custom-built provider called `terraform`,
or the now archived [`hashicorp/terraform` provider in the registry](https://registry.terraform.io/providers/hashicorp/terraform/latest),
or (most likely) the `terraform` provider built into 0.12+, which is
represented via a dedicated FQN of `terraform.io/builtin/terraform` in 0.13+.

You may be able to differentiate between these different providers if you
know the version of Terraform.

Alternatively you may just treat the address as the builtin provider,
i.e. assume all of its logic including schema is contained within
Terraform Core.

In such case you should use `ImpliedProviderForUnqualifiedType(typeName)`,
as the function makes such assumption.
