# Changelog

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
