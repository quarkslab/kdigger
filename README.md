# kdigger

![A small digger trying to move the evergreen stuck cruise ship in the suez
canal](digger.jpg)

`kdigger` for "Kubernetes digger" is a Kubernetes pentest tool. This tool is a
compilation of various plugins called buckets to facilitate pentesting
Kubernetes from inside a pod.

Some plugins perform really simple actions, that could be performed just by
calling the `mount` command or listing all devices present in dev with `ls
/dev` for example, but some others automate scanning processes, such as the
admission controller scanner. In the end, this tool aims to speed up the
pentesting process.

## Installation

### Via Go

```bash
$ go get github.com/mtardy/kdigger
```

## Usage


What you generally want to do is running all the buckets with `dig` or just `d`:
```bash
$ kdigger dig
```

Help is provided by the CLI itself, just type `kdigger` to see the options:

```text
kdigger is an extensible CLI tool to dig around when you are in a Kubernetes
cluster. For that you can use multiples buckets. Buckets are plugins that can
scan specific aspects of a cluster or bring expertise to automate the Kubernetes
pentest process.

Usage:
  kdigger [command]

Available Commands:
  dig         Use all buckets or specific ones
  help        Help about any command
  ls          List available buckets or describe specific ones

Flags:
  -h, --help            help for kdigger
  -o, --output string   Output format. One of: human|json. (default "human")

Use "kdigger [command] --help" for more information about a command.
```

Especially on the dig command to browse the available flags:

```text
This command, with no arguments, runs all registered buckets. You can find
information about all buckets with the list command. To run one or more
specific buckets, just input their names or aliases as arguments.

Usage:
  kdigger dig [flags]

Aliases:
  dig, d

Flags:
  -c, --color               Enable color in output. (default true if output is human)
  -h, --help                help for dig
      --kubeconfig string   (optional) absolute path to the kubeconfig file (default "/home/mahe/.kube/config")
  -n, --namespace string    Kubernetes namespace to use. (default to the namespace in the context)

Global Flags:
  -o, --output string   Output format. One of: human|json. (default "human")
```

The current main flags that you can use are:
- `output` to format the results in a human way or in json to exploit the
  results.
- `namespace` to run scans and requests in specific namespace.
- `color` to enable or disable color in results.

## Buckets

You can list and describe the available buckets (or plugins) with `kdigger ls`:
```text
+---------------+----------------------------+----------------------------------------------------+
|      NAME     |           ALIASES          |                     DESCRIPTION                    |
+---------------+----------------------------+----------------------------------------------------+
| admission     | [admissions adm]           | Admission scans the admission controller chain by  |
|               |                            | creating specific pods to find what is prevented   |
|               |                            | or not.                                            |
| authorization | [authorizations auth]      | Authorization checks your API permissions with the |
|               |                            | current context or the available token.            |
| capabilities  | [capability cap]           | Capabilities list all capabilities in all sets and |
|               |                            | displays dangerous capabilities in red.            |
| devices       | [device dev]               | Devices shows the list of devices available in the |
|               |                            | container.                                         |
| environment   | [environments environ env] | Environment checks the presence of kubernetes      |
|               |                            | related environment variables and shows them.      |
| mount         | [mounts mn]                | Mount shows all mounted devices in the container.  |
| namespaces    | [namespace ns]             | Namespaces analyses namespaces of the container in |
|               |                            | the context of Kubernetes.                         |
| runtime       | [runtimes rt]              | Runtime finds clues to identify which container    |
|               |                            | runtime is running the container.                  |
| services      | [service svc]              | Services uses CoreDNS wildcards feature to         |
|               |                            | discover every service available in the cluster.   |
| syscalls      | [syscall sys]              | Syscalls scans most of the syscalls to detect      |
|               |                            | which are blocked and allowed.                     |
| token         | [tokens tk]                | Token checks for the presence of a service account |
|               |                            | token in the filesystem.                           |
+---------------+----------------------------+----------------------------------------------------+
```

### Admission

Admission scans the admission controller chain by creating specific pods to
find what is prevented or not. The idea behind this bucket is to check, after
you learned that you have `create pods` ability, if no admission controller
like a PodSecurityPolicy or another is blocking you to create node privilege
escalation pods. Like mounting the host filesystem, or the host PID namespace,
or just a privileged container for example.

This bucket currently automatically tries to create:
- a privileged pod
- a privilege escalation pod
- a host network pod
- a host path pod
- a run as root pod
- a host PID pod

So, if you are granted rights to `create pods`, you can check the presence of
any admission controller that might restrict you.

### Authorization

Authorization checks your API permissions with the current context or the
available token. If you use kdigger inside a pod as planned, it will check and
use the service account token that is normally mounted inside the pod. Then it
will basically operate exactly the same operation as if you do `kubectl auth
can-i --list` and display the result.

### Capabilities

Capabilities list all capabilities in all sets and displays dangerous
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

### Namespaces

Namespaces analyses namespaces of the container in the context of Kubernetes.
Nevertheless, detecting namespaces, except for the user namespace, is quite
difficult or almost impossible. You always rely on bugs or implementation
details that can quickly change. This bucket detects the user namespace and
exposes the mappings but try to give some information about the PID namespace
just by detecting some process present in `/proc`.

Indeed, the detection in
[amicontained](https://github.com/genuinetools/amicontained) is based on the
device number of the namespace file, a detail of implementation which is no
longer reliable and most of the time wrong. This is why I tried a different
approach.

To explain what this bucket scans for the PID namespace, if you see a process
named `pause` in `/proc`, you might be sharing the PID namespace between all
the containers composing the pod. Identically, if you spot a process named
`kubelet` you might be sharing the host PID namespace.

### Runtime

Runtime finds clues to identify which container runtime is running the
container. This one is calling exactly the same code that the one in
[amicontained](https://github.com/genuinetools/amicontained). It is using a
package of the [genuinetools/bpfd](https://github.com/genuinetools/bpfd)
project to spot artefacts about container runtime that could betray their
presence.

Please note that this is a 3 year old part of that code and that it makes no
distinction between Docker and containerd.

### Services

Services uses CoreDNS wildcards feature to discover every service available in
the cluster. In fact, it appears that CoreDNS, that is now widely used in
Kubernetes cluster proposes a wildcards features. You can learn more about it
[here in the
documentation](https://github.com/coredns/coredns/blob/master/plugin/kubernetes/README.md#wildcards).

This bucket is extremely useful to perform discovery really fast in a
Kubernetes cluster. The DNS will kindly give you every service domain present
in the cluster.

### Syscalls

Syscalls scans most of the syscalls to detect which are blocked and allowed.
This one is also using a lot of the
[amicontained](https://github.com/genuinetools/amicontained) code base except
that it also banned the `SYS_PTRACE` scan that causes a racing condition that
can hang the program forever.

This is one really nice way to see if you are in a privileged container with a
lot of capabilities quickly: the list of blocked syscall might be almost empty.

### Token

Token checks for the presence of a service account token in the filesystem.
Then it dumps the stuff it finds in `/run/secrets/kubernetes.io/serviceaccount`
which is composed of the service account token itself, the namespace and the CA
certificate of the kube API server.

You might want to use the `-o json` flag here and use `jq` to get that token
fast!

## Details

### Warnings

Some tests are based on details of implementation or side effects that
might be subject to changes in the future. So be careful with the results.

On top of that, some results might need some experience to be understood and
analyzed. To take a specific example, if you are granted the `CAP_SYS_ADMIN`
capability inside a Kubernetes container, there is a good chance that it is
because you are running in a privileged container. But you should definitely
confirm that by looking at the number of devices available or the other
capabilities that you are granted. Indeed it might be necessary to get
`CAP_SYS_ADMIN` to be privileged but it's not sufficient and if it is your
goal, you can really easily trick the results by crafting very specific pods
that might look confusing regarding this tool results.

It might not be the most sophisticated tool to pentest a Kubernetes cluster,
but you can see this as a _Kubernetes pentest 101 summer compilation_!

### Why another tool?

I started researching Kubernetes security a few months ago and participated to
the 2021 Europe kubecon security day CTF. I learned a lot by watching various
security experts conferences and demonstration and this CTF was a really
beginner-friendly entry point to practice what I learned in theory.

I had the opportunity to see most of my Kubernetes security mentors live,
trying to solve one of the challenges of this CTF and it was awesome to
understand what their techniques and thinking were.

However during some pentests, you cannot bring your tools with you because you
don't have any direct internet access so you have to work with what is
available. Some pentesters build their checklist to have a rigorous process
and to not forget crucial operations that could give them important
information.

But sometimes, you can fetch your favourite tools from the internet and while
there are various solutions out there, a lot of experts were using
[amicontained](https://github.com/genuinetools/amicontained), a famous
container introspection tool. This tool is truly awesome but some features are
outdated, like the PID namespace detection and it is not specialized on
Kubernetes, it's only a container tool that can give already give a lot of
hints about your Kubernetes situation!

So I decided to write a tool that include most of
[amicontained](https://github.com/genuinetools/amicontained) information, but
also much more. And specifically aimed at Kubernetes pentesting. So with
`kdigger`, you can try to guess your container runtime, see your capabilities,
scan namespace activation and the allowed syscalls, like amicontained, but you
can also automatically retrieve service account token, scan their permissions,
list interesting environment variables, list devices and even scan the
admission controller chain!

It gives you lots of information really fast, like a digest, that you can later
analyse to understand the situation further and conduct more specialized
analyses.

### How this tool is built?

In addition to all the available features, this tool was built with a plugin
design so that it can be easily extended by anyone that wants to bring some
expertises.

For example, you are a security researcher on Kubernetes, and when you are
doing CTFs or pentesting real infrastructure, you are often performing specific
repetitive actions that could be automated or at least compiled with others.
You can take a look at `/pkg/plugins/template/template.go` to bootstrap your
own plugins and propose them to the project to extend the features! You only
need a name, optionally some aliases, a description and filling the `Run()`
function with the actual logic.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to
discuss what you would like to change.

## License

[MIT](https://choosealicense.com/licenses/mit/)
