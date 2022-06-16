# Changelog

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
