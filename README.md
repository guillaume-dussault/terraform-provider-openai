![nventive](https://nventive-public-assets.s3.amazonaws.com/nventive_logo_github.svg?v=2)

# Terraform Provider OpenAI

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](LICENSE) [![Latest Release](https://img.shields.io/github/release/guillaume-dussault/terraform-provider-openai.svg?style=flat-square)](https://github.com/guillaume-dussault/terraform-provider-openai/releases/latest)

## Requirements

* [Terraform](https://www.terraform.io/downloads) (>= 0.12)
* [Go](https://go.dev/doc/install) (1.22)
* [GNU Make](https://www.gnu.org/software/make/)
* Ideally a OpenAI ChatGPT Plus account to test

## Development

### Build provider

Run the following command to build the provider

```shell
go build -o terraform-provider-openai
```

### Test sample configuration

First, build and install the provider.

```shell
make install
```

Then, navigate to the `examples` directory.

```shell
cd examples
```

Run the following command to initialize the workspace and apply the sample configuration.

```shell
terraform init && terraform apply
```

### Generating documentation

This provider uses [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs/)
to generate documentation and store it in the `docs/` directory.
Once a release is cut, the Terraform Registry will download the documentation from `docs/`
and associate it with the release version. Read more about how this works on the
[official page](https://www.terraform.io/registry/providers/docs).

Use `make generate` to ensure the documentation is regenerated with any changes.

## Breaking Changes

Please consult [BREAKING_CHANGES.md](BREAKING_CHANGES.md) for more information about version
history and compatibility.

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on the process for
contributing to this project.

Be mindful of our [Code of Conduct](CODE_OF_CONDUCT.md).

## We're hiring

Look for current openings on BambooHR https://nventive.bamboohr.com/careers/

## Stay in touch

[nventive.com](https://nventive.com/) | [Linkedin](https://www.linkedin.com/company/nventive/) | [Instagram](https://www.instagram.com/hellonventive/) | [YouTube](https://www.youtube.com/channel/UCFQyvGEKMO10hEyvCqprp5w) | [Spotify](https://open.spotify.com/show/0lsxfIb6Ttm76jB4wgutob)
