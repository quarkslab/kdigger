# Changelog

## `1.4.0` - 2022-09-30

### Added

- New plugin for basic container detection, result of [this
  discussion](https://twitter.com/g3rzi/status/1564594977220562945) on Twitter.
- Add a linting configuration and linting in CI on GitHub.
- Add the nixery.dev docker image build instructions in README.
- Add a demo GIF in the README.

### Changed

- Makefile is up to date with some new targets to setup dev env and the
  default build target runs without the linter.
- Simplify and update the Vagrantfile.
- Updated all dependencies and especially the Go client k8s to `v0.25.2`.
- Made a lot of style modifications and minor fixes thanks to linting.

### Fixes

- Fix the import of all auth providers for k8s Go client thanks to [this user's
  PR](https://github.com/quarkslab/kdigger/pull/7).

## `1.3.0` - 2022-07-12

### Added

- New level one command to generate template of pods with major security features
  disabled. It's mostly something that I needed while doing CTFs to not have some
  canonical YAML in a file somewhere to use, but being able to generate quickly
  those templates with random names, etc.
- New plugin to scan the metadata endpoints in public cloud. I got this idea
  thanks to someone contributing to the [security
  checklist](https://github.com/kubernetes/website/pull/33992) on the Kubernetes
  documentation. It's basically public cloud fingerprinting via network.

## `1.2.1` - 2022-06-21

### Added
- New builds for macOS amd64 and Linux arm64. the macOS build is not really
  useful since kdigger is supposed to be run inside of pods, inside nodes, but
  it can be used to scan the admission control for example, or any remote
  plugins. However, Linux arm64 can be quite useful in case of arm64 node
  pools.
- You can now install kdigger via Nix! Thanks to generous contributor
  @06kellyjac, see the [PR on kdigger repo](https://github.com/quarkslab/kdigger/pull/2)
  and [in nixpkgs](https://github.com/NixOS/nixpkgs/pull/177868).

### Changed
- Fixed minor bugs discovered along running on a different arch.

## `1.2.0` - 2022-06-16
### Added
- A new plugin, apiresources to retrieve all information that can be leaked by
  the discovery API. I had the idea after doing the last CTF challenge at
  KubeCon Europe by ControlPlane, Falco was installed in the cluster and it was
  useful to discover that. It could be discovered via the services plugin
  because Falco exposes one, but CRDs discovery could also be used.

### Changed
- The "active" flag to "side-effects" because it was unclear for some
  person at BlackHat Asia when I presented what "active" meant on the list of
  plugins.
- The API used to register, I grouped all the args in a structure and used the
  new "require client" field to properly load the context or not and fail
  gracefully to run the rest of the plugins in case the context is unavailable.
- Fix a bug when no default namespaced was defined in a kubeconfig, now
  automatically default to the namespace "default".

## `1.1.0` - 2022-04-21
### Added
- Two new plugins, cgroups, and node and checks for NoNewPrivs and Seccomp flag
  in respectively, capabilities and syscall plugins. (Thanks for Andrew Martin
  & Michael Hausenblas for the inspiration from the Appendix 1 "A Pod-Level
  Attack" from the "Hacking Kubernetes" book)
- Documentation about the Wildcard feature removal in CoreDNS.
- New Makefiles rules to quickly start kdigger in a Pod in a kind cluster and
  to make a release.
- Vagrantfile for development on different systems.

### Changed
- Update dependencies and use Go 1.18.
- Fix the `got get` oneliner using `go install`.
- The output mechanism for plugins, now using comments array and flatten
  results that are of length one for better JSON output parsing.

## `1.0.0` - 2021-10-07
- Initial release!
