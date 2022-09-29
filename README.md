# kdigger

`kdigger`, short for "Kubernetes digger", is a context discovery tool for
Kubernetes penetration testing. This tool is a compilation of various plugins
called buckets to facilitate pentesting Kubernetes from inside a pod.

Here is a demo showing `kdigger v1.3.0` in action, using only four buckets.
Please note that around twenty plugins exist and you can read more about all
the features in the following documentation.

<p align="center">
  <img src="https://user-images.githubusercontent.com/11256051/183686558-0dd9dbaa-ac3d-4ea0-9bcf-81344b91ba3e.gif" alt="Demonstration of kdigger v1.3.0 features"/>
</p>

Please note that this is not an ultimate pentest tool on Kubernetes. Some
plugins perform really simple actions that could be performed manually by
calling the `mount` command or listing all devices present in dev with `ls
/dev` for example. But some others automate scanning processes, such as the
admission controller scanner. In the end, this tool aims to humbly speed up the
pentesting process.

## Table of content

* [Installation](#installation)
    * [Via releases](#via-releases)
    * [Build from source](#build-from-source)
    * [With Nix](#with-nix)
    * [Via Go](#via-go)
* [Usage](#usage)
* [Details](#details)
    * [Updates](#updates)
    * [Usage warning](#usage-warning)
    * [Results warning](#results-warning)
    * [Why another tool?](#why-another-tool)
    * [How is this tool built?](#how-is-this-tool-built)
    * [Areas for improvement](#areas-for-improvement)
    * [How can I experience with this tool?](#how-can-i-experience-with-this-tool)
* [Buckets](#buckets)
    * [Admission](#admission)
    * [API Resources](#api-resources)
    * [Authorization](#authorization)
    * [Capabilities](#capabilities)
    * [Cgroups](#cgroups)
    * [CloudMetadata](#cloudmetadata)
    * [Devices](#devices)
    * [Environment](#environment)
    * [Mount](#mount)
    * [Node](#node)
    * [PIDNamespace](#pidnamespace)
    * [Processes](#processes)
    * [Runtime](#runtime)
    * [Services](#services)
    * [Syscalls](#syscalls)
    * [Token](#token)
    * [UserID](#userid)
    * [UserNamespace](#usernamespace)
    * [Version](#version)
* [Contributing](#contributing)
* [License](#license)

## Installation

`kdigger` is available on Linux amd64, arm64 and macOS amd64.

**Please note that `kdigger` should be mostly run inside of pods on not on your
host machine.**

### Via releases

For installation instructions from binaries please visit the [releases
page](https://github.com/quarkslab/kdigger/releases).

### Build from source

Just type `make` to build with the 
[default build target](https://github.com/quarkslab/kdigger/blob/main/Makefile#L20).
```bash
git clone https://github.com/quarkslab/kdigger
make
```

Then you can move the binary somewhere included in your PATH, for example:

```bash
sudo install kdigger /usr/local/bin
```

Note that you will need [`golangci-lint`](https://github.com/golangci/golangci-lint)
to use the `release` target that will build for the supported architectures.
You can use `make install-linter` to install `golangci-lint` on the host.

### With Nix

You can use Nix with a new shell environments thanks to [a contributor
PR](https://github.com/quarkslab/kdigger/pull/2) that [added kdigger to
nixpkgs](https://github.com/NixOS/nixpkgs/pull/177868).

```bash
nix-shell -p kdigger
```

You can also create docker images with `kdigger` and other tools easily
with [nixery.dev](https://nixery.dev/):

```bash
docker run -it nixery.dev/kubectl/kdigger/bash /bin/bash
```

### Via Go

```bash
go install github.com/quarkslab/kdigger@main
```

## Usage

What you generally want to do is running all the buckets with `dig all` or just
`d a`:
```bash
kdigger dig all
```

Help is provided by the CLI itself, just type `kdigger` to see the options:

```console
$ kdigger
kdigger is an extensible CLI tool to dig around when you are in a Kubernetes
cluster. For that you can use multiples buckets. Buckets are plugins that can
scan specific aspects of a cluster or bring expertise to automate the Kubernetes
pentest process.

Usage:
  kdigger [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  dig         Use all buckets or specific ones
  gen         Generate template for pod with security features disabled
  help        Help about any command
  ls          List available buckets or describe specific ones
  version     Print the version information

Flags:
  -h, --help            help for kdigger
  -o, --output string   Output format. One of: human|json. (default "human")
  -w, --width int       Width for the human output (default 140)

Use "kdigger [command] --help" for more information about a command.

```

Make sure to check out the help on the ``dig`` command to see all the available
flags:

```console
$ kdigger help dig
This command runs buckets, special keyword "all" or "a" runs all registered
buckets. You can find information about all buckets with the list command. To
run one or more specific buckets, just input their names or aliases as
arguments.

Usage:
  kdigger dig [buckets] [flags]

Aliases:
  dig, d

Flags:
      --admission-force     Force creation of pods to scan admission even without cleaning rights. (this flag is specific to the admission bucket)
  -c, --color               Enable color in output. (default true if output is human)
  -h, --help                help for dig
      --kubeconfig string   (optional) absolute path to the kubeconfig file (default "/home/vagrant/.kube/config")
  -n, --namespace string    Kubernetes namespace to use. (default to the namespace in the context)
  -s, --side-effects        Enable all buckets that might have side effect on environment.

Global Flags:
  -o, --output string   Output format. One of: human|json. (default "human")
  -w, --width int       Width for the human output (default 140)
```

You can also generate useful templates for pods with security features disabled
to escalate privileges when you can create such a pod. See the help for this
specific command for more information.

```console
$ kdigger help gen
This command generates templates for pod with security features disabled.
You can customize the pods with some of the string flags and activate
boolean flags to disabled security features. Examples:

  # Generate a very simple template in json
  kdigger gen -o json

  # Create a very simple pod
  kdigger gen | kubectl apply -f -

  # Create a pod named mypod with most security features disabled
  kdigger gen -all mypod | kubectl apply -f -

  # Create a custom privileged pod
  kdigger gen --privileged --image bash --command watch --command date | kubectl apply -f -

Usage:
  kdigger gen [name] [flags]

Aliases:
  gen, generate

Flags:
      --all                   Enable everything
      --command stringArray   Container command used (default [sleep,infinitely])
  -h, --help                  help for gen
      --hostnetwork           Add the hostNetwork flag on the whole pod
      --hostpath              Add a hostPath volume to the container
      --hostpid               Add the hostPid flag on the whole pod
      --image string          Container image used (default "busybox")
      --privileged            Add the security flag to the security context of the pod
      --tolerations           Add tolerations to be schedulable on most nodes

Global Flags:
  -o, --output string   Output format. One of: human|json. (default "human")
  -w, --width int       Width for the human output (default 140)
```

## Details

### Updates

I updates this tool from time to time, when I have new ideas after reading a
book or doing CTF challenges, you can find information in the
[changelog](https://github.com/quarkslab/kdigger/blob/master/CHANGELOG.md).

### Usage warning

Be careful when running this tool, some checks have side effects, like scanning
your available syscalls or trying to create pods to scan the admission control.
By default these checks will **not** run without adding the `--side-effects` or
`-s` flag.

For example, syscalls scans may succeed to perform some syscalls with
empty arguments, and it can alter your environment or configuration. For
instance, if the `hostname` syscall is successful, it will replace the
hostname with the empty string. So please, **NEVER** run with
sufficient permissions (as root for example) directly on your machine.

### Results warning

Some tests are based on details of implementation or side effects on the
environment that might be subject to changes in the future. So be cautious with
the results. For example, CoreDNS is considering removing the wildcard feature,
see [CoreDNS issue 4984]( https://github.com/coredns/coredns/issues/4984).

On top of that, some results might need some experience to be understood and
analyzed. To take a specific example, if you are granted the ``CAP_SYS_ADMIN``
capability inside a Kubernetes container, there is a good chance that it is
because you are running in a privileged container. But you should definitely
confirm that by looking at the number of devices available or the other
capabilities that you are granted. Indeed, it might be necessary to get
``CAP_SYS_ADMIN`` to be privileged but it’s not sufficient and if it is your
goal, you can easily trick the results by crafting very specific pods that
might look confusing regarding this tool results.

It might not be the most sophisticated tool to pentest a Kubernetes cluster,
but you can see this as a *Kubernetes pentest 101 compilation*!

### Why another tool?

I started researching Kubernetes security a few months ago and participated in
the [2021 Europe KubeCon Cloud-Native Security Day
CTF](https://controlplaneio.github.io/kubecon-2021-sig-security-day-ctf/). I
learned a lot by watching various security experts conferences and
demonstrations and this CTF was a really beginner-friendly entry point to
practice what I learned in theory. During a live solving session, I had the
opportunity to see how Kubernetes security experts were trying to solve the
challenge, how they were thinking, what they were looking for.

So I decided to create a tool that compiles most of the checks we usually do as
pentesters when in a Kubernetes pod to acquire information very quickly. There
already are various tools out there. For example, a lot of experts were using
[amicontained](https://github.com/genuinetools/amicontained), a famous
container introspection tool by Jessie Frazelle. This tool is truly awesome,
but some features are outdated, like the PID namespace detection, and it is not
specialized in Kubernetes, it is only a container tool that can already give a
lot of hints about your Kubernetes situation.

That is why, in kdigger, I included most of amicontained features. You can:

- Try to guess your container runtime.
- See your capabilities.
- Scan for namespace activation and configuration.
- Scan for the allowed syscalls.

But you can also do more Kubernetes specific operations:

- Retrieve service account token.
- Scan token permissions.
- List interesting environment variables.
- List available devices.
- Retrieve all available services in a cluster.
- Scan the admission controller chain!

Anyway, this tool is obviously not an *automatically hack your Kubernetes
cluster* application, it is mostly just a compilation of tedious tasks that can
be performed automatically very quickly. You still need a lot of expertise to
interpret the digest and understand what the various outputs mean. And also,
during pentest and challenges, you do not always have Internet access to pull
your favorite toolchain, so you can also see this compilation as a checklist
that you can somehow perform manually with a basic installation and a shell.

### How is this tool built?

In addition to all the available features, this tool was built with a plugin
design so that it can be easily extended by anyone that wants to bring some
expertise.

For example, you are a security researcher on Kubernetes, and when you are
doing CTFs or pentesting real infrastructure, you are often performing specific
repetitive actions that could be automated or at least compiled with others.
You can take a look at `/pkg/plugins/template/template.go` to bootstrap your
own plugins and propose them to the project to extend the features! You only
need a name, optionally some aliases, a description and filling the `Run()`
function with the actual logic.

### Areas for improvement

The expertise proposed by the tool could be refined and more precise. For now
it's mostly dumping raw data for most of the buckets and rely on the user to
understand what it implies.

Generally, the output format is not the best and could be reworked. The human
format via array lines does not fit all the use cases perfectly but is simple to
generalize without having each plugin to implement their format. The tool also
proposes a JSON output format.

### How can I experience with this tool?

Good news! We created a mini Kubernetes CTF with basic steps to experience with
the tool and resolve quick challenges. For more information go to the
[minik8s-ctf repository](https://github.com/quarkslab/minik8s-ctf).

## Buckets

You can list and describe the available buckets (or plugins) with `kdigger
list` or `kdigger ls`:
```console
$ kdigger ls
+---------------+----------------------------+--------------------------------------+-------------+---------------+
|      NAME     |           ALIASES          |              DESCRIPTION             | SIDEEFFECTS | REQUIRECLIENT |
+---------------+----------------------------+--------------------------------------+-------------+---------------+
| admission     | [admissions adm]           | Admission scans the admission        | true        | true          |
|               |                            | controller chain by creating         |             |               |
|               |                            | specific pods to find what is        |             |               |
|               |                            | prevented or not.                    |             |               |
| apiresources  | [api apiresource]          | APIResources discovers the available | false       | true          |
|               |                            | APIs of the cluster.                 |             |               |
| authorization | [authorizations auth]      | Authorization checks your API        | false       | true          |
|               |                            | permissions with the current context |             |               |
|               |                            | or the available token.              |             |               |
| capabilities  | [capability cap]           | Capabilities lists all capabilities  | false       | false         |
|               |                            | in all sets and displays dangerous   |             |               |
|               |                            | capabilities in red.                 |             |               |
| cgroups       | [cgroup cg]                | Cgroups reads the /proc/self/cgroup  | false       | false         |
|               |                            | files that can leak information      |             |               |
|               |                            | under cgroups v1.                    |             |               |
| cloudmetadata | [cloud meta]               | Cloudmetadata scans the usual        | false       | false         |
|               |                            | metadata endpoints in public clouds. |             |               |
| devices       | [device dev]               | Devices shows the list of devices    | false       | false         |
|               |                            | available in the container.          |             |               |
| environment   | [environments environ env] | Environment checks the presence of   | false       | false         |
|               |                            | kubernetes related environment       |             |               |
|               |                            | variables and shows them.            |             |               |
| mount         | [mounts mn]                | Mount shows all mounted devices in   | false       | false         |
|               |                            | the container.                       |             |               |
| node          | [nodes n]                  | Node retrieves various information   | false       | false         |
|               |                            | in /proc about the current host.     |             |               |
| pidnamespace  | [pidnamespaces pidns]      | PIDnamespace analyses the PID        | false       | false         |
|               |                            | namespace of the container in the    |             |               |
|               |                            | context of Kubernetes.               |             |               |
| processes     | [process ps]               | Processes analyses the running       | false       | false         |
|               |                            | processes in your PID namespace      |             |               |
| runtime       | [runtimes rt]              | Runtime finds clues to identify      | false       | false         |
|               |                            | which container runtime is running   |             |               |
|               |                            | the container.                       |             |               |
| services      | [service svc]              | Services uses CoreDNS wildcards      | false       | false         |
|               |                            | feature to discover every service    |             |               |
|               |                            | available in the cluster.            |             |               |
| syscalls      | [syscall sys]              | Syscalls scans most of the syscalls  | true        | false         |
|               |                            | to detect which are blocked and      |             |               |
|               |                            | allowed.                             |             |               |
| token         | [tokens tk]                | Token checks for the presence of a   | false       | false         |
|               |                            | service account token in the         |             |               |
|               |                            | filesystem.                          |             |               |
| userid        | [userids id]               | UserID retrieves UID, GID and their  | false       | false         |
|               |                            | corresponding names.                 |             |               |
| usernamespace | [usernamespaces userns]    | UserNamespace analyses the user      | false       | false         |
|               |                            | namespace configuration.             |             |               |
| version       | [versions v]               | Version dumps the API server version | false       | true          |
|               |                            | informations.                        |             |               |
+---------------+----------------------------+--------------------------------------+-------------+---------------+
```

### Admission

Admission scans the admission controller chain by creating specific pods to
find what is prevented or not. The idea behind this bucket is to check, after
you learned that you have `create pods` ability, if no admission controller
like a PodSecurityPolicy or another is blocking you to create node privilege
escalation pods. Like mounting the host filesystem, or the host PID namespace,
or just a privileged container, for example.

This bucket currently automatically tries to create:
- a privileged pod
- a privilege escalation pod
- a host network pod
- a host path pod
- a run as root pod
- a host PID pod

So, if you are granted rights to `create pods`, you can check the presence of
any admission controller that might restrict you.

### API Resources

APIResources discovers the available APIs of the cluster. These endpoints are
interesting because it only requires authentication and can leak some sensitive
information. Indeed, you can learn about CRDs installed in the cluster, leaking
for example the presence of Prometheus using Prometheus Operator, or the
installation of the Falco, etc. Generally, it can inform you on the stack used
in the cluster if the components are installed using CRDs.

This is equivalent of running `kubectl api-resources`, the output is just
abbreviated.

Prior v1.14, [the endpoints used required not
authentication](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#discovery-roles).
Now the API discovery (and version, health, ready) endpoints are grouped in the
`system:discovery` ClusterRole (visible with `kubectl get clusterrole
system:discovery -o yaml`). The rest is in `system:basic-user` for the self
authorization review with SelfSubjectAccessReviews and SelfSubjectSulesSeviews
and `system:public-info-viewer`. The only part remaining for unauthenticated
user is version, health, ready, live endpoints (see with `kubectl get
clusterrolebinding system:public-info-viewer -o yaml`).

### Authorization

Authorization checks your API permissions with the current context or the
available token. If you use kdigger inside a pod as planned, it will check and
use the service account token that is normally mounted inside the pod. Then it
will basically operate exactly the same operation as if you do `kubectl auth
can-i --list` and display the result.

### Capabilities

Capabilities lists all capabilities in all sets and displays dangerous
capabilities in red.

Basically, in a non-privileged container, the result might look like that:
```text
### CAPABILITIES ###
Comment: The bounding set contains 14 caps, it seems that you are running a non-privileged container.
+-------------+----------------------------------------------------+
|     SET     |                    CAPABILITIES                    |
+-------------+----------------------------------------------------+
| effective   | [chown dac_override fowner fsetid kill setgid      |
|             | setuid setpcap net_bind_service net_raw sys_chroot |
|             | mknod audit_write setfcap]                         |
| permitted   | [chown dac_override fowner fsetid kill setgid      |
|             | setuid setpcap net_bind_service net_raw sys_chroot |
|             | mknod audit_write setfcap]                         |
| inheritable | [chown dac_override fowner fsetid kill setgid      |
|             | setuid setpcap net_bind_service net_raw sys_chroot |
|             | mknod audit_write setfcap]                         |
| bounding    | [chown dac_override fowner fsetid kill setgid      |
|             | setuid setpcap net_bind_service net_raw sys_chroot |
|             | mknod audit_write setfcap]                         |
| ambient     | []                                                 |
+-------------+----------------------------------------------------+
```

This bucket might be especially useful to spot critical capabilities that can
help you to escalate your privileges. This can be a good hint on whether you
are running inside a privileged container or not.

This bucket also checks for the `NoNewPrivs` flag in `/proc/self/status` that
will be set to 1 if `allowPrivilegeEscalation` is set to false in the
`SecurityContext` of the Pod. Note that `allowPrivilegeEscalation` is always
true when the container:
- is run as privileged, or
- has `CAP_SYS_ADMIN`

### Cgroups

Cgroups reads the /proc/self/cgroup files that can leak information under
cgroups v1. The CgroupPath can leak many information, for example, you can see
these kinds of paths:

```text
/docker/744ba5524431986481406cfd13382cbeb7e59a4739dec5ff8b458bc821969662/kubelet/kubepods/besteffort/pod3716cd27-4591-40fc-9dae-3851142b1d51/e06769b086acea415152f733574033db4dc5e81c9a8f6563648fd5f3b60227af
...
/docker/744ba5524431986481406cfd13382cbeb7e59a4739dec5ff8b458bc821969662/system.slice/containerd.service
```

These strings can give you information about the kind of container used, here
the `kubelet/kubepods` part can confirm we are in a Kubernetes pod's container.
Also you can learn about the container runtime used, here docker with
containerd, also the container ID, the pod UID, the snapshot key, etc.

Knowing the container ID and the container runtime can be interesting to
determine where on the host will be stored the files written to the container
filesystem. This is for example useful when exploiting hooks that will be called
in the context of the host, and thus escape the container, see these [kinds of
escapes](http://blog.bofh.it/debian/id_413) for more information.

For kernels using cgroups v2, `/proc/self/mountinfo` can still leak information
about the container ID. See [this Stackoverflow
thread](https://stackoverflow.com/a/69005753) and its related threads for more
information.

### CloudMetadata

Cloudmetadata scans the usual metadata endpoints in public clouds. It is usually
quite simple to find at which service provider a VM come from, because of many
leaks in the filesystems, the environment variables, etc. But from a containers
in a VM, it can be harder, that's why this plugin performs a scan on the network
via the usual service running at `169.254.169.254` or alike. See the source code
for endpoints used and links to more endpoints.

This plugin only gives you the information of the availability of the main
endpoints, which means that you might running in a specific public cloud. If
that's the case, further research, using available endpoints for that cloud, can
be conducted. You can potentially retrieve an authentication token or simply
more metadata to pivot within the cloud account.

### Devices

Devices show the list of devices available in the container. This one is
straightforward, it's equivalent to just `ls /dev`. Nevertheless, the number of
available devices can also be a good hint on running in a privileged container
or not.

### Environment

Environment checks the presence of Kubernetes related environment variables and
shows them. Like always, it's not sufficient, but detecting Kubernetes related
environment variables can give you a pretty good idea that you are running in a
Kubernetes cluster. That might be useful if you want to quickly find out where
you are. Of course, this one is easy to confuse, by just exporting some
environment variable or removing some.

### Mount

Mount show all mounted devices in the container. This is equivalent to use the
`mount` command directly but the number of mounted devices and reading path can
show you mounted volumes, configmap or even secrets inside the pod.

### Node

Node retrieves various information in /proc about the current host. It seeks
information in `/proc/cpuinfo`, `/proc/meminfo` and `/proc/version` to
understand the context of execution of the containers. These files are not
namespaced and thus leak information about the reality of the underlying node.

It lists information about the CPU model, the number of cores, the total memory
available, the part that is used, the kernel version and some compilation
details about it.

### PIDNamespace

PIDNamespace analyzes the PID namespace of the container in the context of
Kubernetes. Detecting the PID namespace is almost impossible so the idea of
this bucket is to scan the `/proc` folder to search for specific processes
such as:
* `pause`: it might signify that you are sharing the PID namespace between all
  the containers composing the pod.
* `kubelet`: it might signify that you are sharing the PID namespace with the
  host.

By the way, the detection in
[amicontained](https://github.com/genuinetools/amicontained) is based on the
device number of the namespace file, a detail of implementation which is no
longer reliable and most of the time wrong. This is why I tried a different
approach.

### Processes

Processes analyzes the running processes in your PID namespace. It is similar
to any `ps` command that list all processes like `ps -e` or `ps -A`. It gives
you the information of the number of running processes and if the first one is
systemd.

### Runtime

Runtime finds clues to identify which container runtime is running the
container. This one is calling exactly the same code that the one in
[amicontained](https://github.com/genuinetools/amicontained). It is using a
package of the [genuinetools/bpfd](https://github.com/genuinetools/bpfd)
project to spot artifacts about container runtime that could betray their
presence.

Please note that this is a 3-year-old part of that code and that it makes no
distinction between Docker and containerd.

### Services

Services uses CoreDNS wildcard feature to discover every service available in
the cluster. In fact, it appears that CoreDNS, that is now widely used in
Kubernetes cluster proposes a wildcard feature. You can learn more about it
[~~here in the
documentation~~](https://github.com/coredns/coredns/blob/master/plugin/kubernetes/README.md#wildcards).

This bucket is extremely useful to perform discovery really fast in a
Kubernetes cluster. The DNS will kindly give you every service domain present
in the cluster.

Note: [This feature was removed](https://github.com/coredns/coredns/pull/5019)
starting as of CoreDNS v1.9.0 because it was mostly used by bad actors (like
this tool). See the associated discussion on [the corresponding Github
issue](https://github.com/coredns/coredns/issues/4984). Kubernetes v1.24 was 
still using CoreDNS v1.8.6, but the v1.25 version updated CoreDNS to v1.9.3.
That's why this plugin no longer works on v1.25 and above.

### Syscalls

Syscalls scans most of the syscalls to detect which are blocked and allowed.
This one is also using a lot of the
[amicontained](https://github.com/genuinetools/amicontained) code base except
that it also banned the `SYS_PTRACE` scan that causes a racing condition that
can hang the program forever.

This is one really nice way to see if you are in a privileged container with a
lot of capabilities quickly: the list of blocked syscall might be almost empty.

This bucket also checks the `Seccomp` flag in `/proc/self/status`, it will
display if Seccomp is disabled, running in strict or in filter mode.

### Token

Token checks for the presence of a service account token in the filesystem.
Then it dumps the stuff it finds in `/run/secrets/kubernetes.io/serviceaccount`
which is composed of the service account token itself, the namespace and the CA
certificate of the kube API server.

You might want to use the `-o json` flag here and use `jq` to get that token
fast!

### UserID

UserID retrieves UID, GID and their corresponding names. It also gives
`homeDir` as a bonus! Unfortunately, we can list all group IDs without CGO
enabled. This is almost (because `id` is better) equivalent to run the `id`
command directly.

### UserNamespace

UserNamespace analyses the user namespace configuration. The user namespace is
transparent and can be easily detected. It is even possible to read the mapping
between the current user namespace and the outer namespace. Unfortunately for
now, user namespaces cannot be used with Kubernetes.

### Version

Version dumps the API server version information. It access the `/version`
path that is accessible even by unauthenticated users. So even without a
service account token you can request this information. You can get more
information on what you can access as an `system:unauthenticated` user with
`kubectl describe clusterrolebinding | grep unauthenticated -B 9` for example.
You may encounter the `system:public-info-viewer` cluster role, you can
describe it with `kubectl describe clusterrole system:public-info-viewer` and
display:

```text
Name:         system:public-info-viewer
Labels:       kubernetes.io/bootstrapping=rbac-defaults
Annotations:  rbac.authorization.kubernetes.io/autoupdate: true
PolicyRule:
  Resources  Non-Resource URLs  Resource Names  Verbs
  ---------  -----------------  --------------  -----
             [/healthz]         []              [get]
             [/livez]           []              [get]
             [/readyz]          []              [get]
             [/version/]        []              [get]
             [/version]         []              [get]
```

## Contributing

As kdigger is a security checklist when pentesting from inside a pod's
container, please consider adding a plugin if you have some checks that are not
covered by the tool. To do that, you can use the
[template](https://github.com/quarkslab/kdigger/blob/master/pkg/plugins/template/template.go)
and load your plugin into the project
[here](https://github.com/quarkslab/kdigger/blob/master/commands/root.go#L71).
I will be happy to see your PR!

If you have any other ideas or advice, consider opening an issue or directly
contact me @mtardy_ on twitter or by email.

![A small digger trying to move the evergreen stuck cruise ship in the suez
canal](https://i.servimg.com/u/f41/11/93/81/35/digger10.jpg)

## License

[Apache License 2.0](./LICENSE)
