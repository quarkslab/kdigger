package syscalls

import (
	"fmt"
	"syscall"
	"time"

	"github.com/quarkslab/kdigger/pkg/bucket"
	"golang.org/x/sys/unix"
)

const (
	bucketName        = "syscalls"
	bucketDescription = "Syscalls scans most of the syscalls to detect which are blocked and allowed."
)

var bucketAliases = []string{"syscall", "sys"}

type SyscallsBucket struct{}

func (n SyscallsBucket) Run() (bucket.Results, error) {
	// executes here the code of your plugin
	res := bucket.NewResults(bucketName)

	results := syscallIter()
	var allowed []string
	var blocked []string
	for _, r := range results {
		if r.allowed {
			allowed = append(allowed, r.name)
		} else {
			blocked = append(blocked, r.name)
		}
	}

	res.SetHeaders([]string{"blocked", "allowed"})
	res.AddContent([]interface{}{blocked, allowed})

	skippedNames := make([]string, 0)
	for _, s := range skippedSyscalls {
		skippedNames = append(skippedNames, syscallName(s))
	}
	res.SetComment(fmt.Sprint(skippedNames) + " were not scanned because they cause hang or for obvious reasons.")

	return *res, nil
}

// Register registers a plugin
func Register(b *bucket.Buckets) {
	b.Register(bucketName, bucketAliases, bucketDescription, true, func(config bucket.Config) (bucket.Interface, error) {
		return NewSyscallsBucket(config)
	})
}

func NewSyscallsBucket(config bucket.Config) (*SyscallsBucket, error) {
	return &SyscallsBucket{}, nil
}

type scanResult struct {
	name    string
	allowed bool
}

// syscallIter is modified copy of the amicontained code that you can find here:
// https://github.com/genuinetools/amicontained/blob/568b0d35e60cb2bfc228ecade8b0ba62c49a906a/main.go#L181
func syscallIter() []scanResult {
	var results []scanResult

	c := make(chan scanResult, unix.SYS_RSEQ-len(skippedSyscalls))
	for id := 0; id <= unix.SYS_RSEQ; id++ {
		go scanSyscall(id, c)
	}

	for i := 0; i < cap(c); i++ {
		results = append(results, <-c)
	}

	return results
}

var skippedSyscalls = []int{
	unix.SYS_RT_SIGRETURN,
	unix.SYS_SELECT,
	unix.SYS_PAUSE,
	unix.SYS_PSELECT6,
	unix.SYS_PPOLL,
	unix.SYS_WAITID,
	unix.SYS_EXIT,
	unix.SYS_EXIT_GROUP,
	unix.SYS_CLONE,
	unix.SYS_FORK,
	unix.SYS_VFORK,
	unix.SYS_SECCOMP,
	unix.SYS_PTRACE,
	unix.SYS_VHANGUP,
}

func scanSyscall(id int, c chan (scanResult)) {
	// these cause a hang, so just skip
	// rt_sigreturn, select, pause, pselect6, ppoll
	if id == unix.SYS_RT_SIGRETURN || id == unix.SYS_SELECT || id == unix.SYS_PAUSE || id == unix.SYS_PSELECT6 || id == unix.SYS_PPOLL || id == unix.SYS_WAITID {
		return
	}
	// exit_group and exit -- causes us to exit.. doh!
	if id == unix.SYS_EXIT || id == unix.SYS_EXIT_GROUP {
		return
	}

	// things currently break horribly if  CLONE, FORK or VFORK are called and the call succeeds
	// guess it should be straight forward to kill the forks
	if id == unix.SYS_CLONE || id == unix.SYS_FORK || id == unix.SYS_VFORK {
		return
	}

	// Skip seccomp itself.
	if id == unix.SYS_SECCOMP {
		return
	}

	// ptrace causes a weird race condition that can make the process hang
	if id == unix.SYS_PTRACE {
		return
	}

	// skip vhangup, it exits a container terminal and there is no interest to test it
	if id == unix.SYS_VHANGUP {
		return
	}

	// The call may block, so invoke asynchronously and rely on rejections being fast.
	errs := make(chan error)
	go func() {
		_, _, err := syscall.Syscall(uintptr(id), 0, 0, 0)
		errs <- err
	}()

	var err error
	select {
	case err = <-errs:
	case <-time.After(100 * time.Millisecond):
		// The syscall was allowed, but it didn't return
	}
	// fmt.Println(err.Error())

	if err == syscall.EPERM || err == syscall.EACCES {
		c <- scanResult{name: syscallName(id), allowed: false}
	} else if err != syscall.EOPNOTSUPP {
		c <- scanResult{name: syscallName(id), allowed: true}
	}
}

func syscallName(e int) string {
	switch e {
	case unix.SYS_READ:
		return "read"
	case unix.SYS_WRITE:
		return "write"
	case unix.SYS_OPEN:
		return "open"
	case unix.SYS_CLOSE:
		return "close"
	case unix.SYS_STAT:
		return "stat"
	case unix.SYS_FSTAT:
		return "fstat"
	case unix.SYS_LSTAT:
		return "lstat"
	case unix.SYS_POLL:
		return "poll"
	case unix.SYS_LSEEK:
		return "lseek"
	case unix.SYS_MMAP:
		return "mmap"
	case unix.SYS_MPROTECT:
		return "mprotect"
	case unix.SYS_MUNMAP:
		return "munmap"
	case unix.SYS_BRK:
		return "brk"
	case unix.SYS_RT_SIGACTION:
		return "rt_sigaction"
	case unix.SYS_RT_SIGPROCMASK:
		return "rt_sigprocmask"
	case unix.SYS_RT_SIGRETURN:
		return "rt_sigreturn"
	case unix.SYS_IOCTL:
		return "ioctl"
	case unix.SYS_PREAD64:
		return "pread64"
	case unix.SYS_PWRITE64:
		return "pwrite64"
	case unix.SYS_READV:
		return "readv"
	case unix.SYS_WRITEV:
		return "writev"
	case unix.SYS_ACCESS:
		return "access"
	case unix.SYS_PIPE:
		return "pipe"
	case unix.SYS_SELECT:
		return "select"
	case unix.SYS_SCHED_YIELD:
		return "sched_yield"
	case unix.SYS_MREMAP:
		return "mremap"
	case unix.SYS_MSYNC:
		return "msync"
	case unix.SYS_MINCORE:
		return "mincore"
	case unix.SYS_MADVISE:
		return "madvise"
	case unix.SYS_SHMGET:
		return "shmget"
	case unix.SYS_SHMAT:
		return "shmat"
	case unix.SYS_SHMCTL:
		return "shmctl"
	case unix.SYS_DUP:
		return "dup"
	case unix.SYS_DUP2:
		return "dup2"
	case unix.SYS_PAUSE:
		return "pause"
	case unix.SYS_NANOSLEEP:
		return "nanosleep"
	case unix.SYS_GETITIMER:
		return "getitimer"
	case unix.SYS_ALARM:
		return "alarm"
	case unix.SYS_SETITIMER:
		return "setitimer"
	case unix.SYS_GETPID:
		return "getpid"
	case unix.SYS_SENDFILE:
		return "sendfile"
	case unix.SYS_SOCKET:
		return "socket"
	case unix.SYS_CONNECT:
		return "connect"
	case unix.SYS_ACCEPT:
		return "accept"
	case unix.SYS_SENDTO:
		return "sendto"
	case unix.SYS_RECVFROM:
		return "recvfrom"
	case unix.SYS_SENDMSG:
		return "sendmsg"
	case unix.SYS_RECVMSG:
		return "recvmsg"
	case unix.SYS_SHUTDOWN:
		return "shutdown"
	case unix.SYS_BIND:
		return "bind"
	case unix.SYS_LISTEN:
		return "listen"
	case unix.SYS_GETSOCKNAME:
		return "getsockname"
	case unix.SYS_GETPEERNAME:
		return "getpeername"
	case unix.SYS_SOCKETPAIR:
		return "socketpair"
	case unix.SYS_SETSOCKOPT:
		return "setsockopt"
	case unix.SYS_GETSOCKOPT:
		return "getsockopt"
	case unix.SYS_CLONE:
		return "clone"
	case unix.SYS_FORK:
		return "fork"
	case unix.SYS_VFORK:
		return "vfork"
	case unix.SYS_EXECVE:
		return "execve"
	case unix.SYS_EXIT:
		return "exit"
	case unix.SYS_WAIT4:
		return "wait4"
	case unix.SYS_KILL:
		return "kill"
	case unix.SYS_UNAME:
		return "uname"
	case unix.SYS_SEMGET:
		return "semget"
	case unix.SYS_SEMOP:
		return "semop"
	case unix.SYS_SEMCTL:
		return "semctl"
	case unix.SYS_SHMDT:
		return "shmdt"
	case unix.SYS_MSGGET:
		return "msgget"
	case unix.SYS_MSGSND:
		return "msgsnd"
	case unix.SYS_MSGRCV:
		return "msgrcv"
	case unix.SYS_MSGCTL:
		return "msgctl"
	case unix.SYS_FCNTL:
		return "fcntl"
	case unix.SYS_FLOCK:
		return "flock"
	case unix.SYS_FSYNC:
		return "fsync"
	case unix.SYS_FDATASYNC:
		return "fdatasync"
	case unix.SYS_TRUNCATE:
		return "truncate"
	case unix.SYS_FTRUNCATE:
		return "ftruncate"
	case unix.SYS_GETDENTS:
		return "getdents"
	case unix.SYS_GETCWD:
		return "getcwd"
	case unix.SYS_CHDIR:
		return "chdir"
	case unix.SYS_FCHDIR:
		return "fchdir"
	case unix.SYS_RENAME:
		return "rename"
	case unix.SYS_MKDIR:
		return "mkdir"
	case unix.SYS_RMDIR:
		return "rmdir"
	case unix.SYS_CREAT:
		return "creat"
	case unix.SYS_LINK:
		return "link"
	case unix.SYS_UNLINK:
		return "unlink"
	case unix.SYS_SYMLINK:
		return "symlink"
	case unix.SYS_READLINK:
		return "readlink"
	case unix.SYS_CHMOD:
		return "chmod"
	case unix.SYS_FCHMOD:
		return "fchmod"
	case unix.SYS_CHOWN:
		return "chown"
	case unix.SYS_FCHOWN:
		return "fchown"
	case unix.SYS_LCHOWN:
		return "lchown"
	case unix.SYS_UMASK:
		return "umask"
	case unix.SYS_GETTIMEOFDAY:
		return "gettimeofday"
	case unix.SYS_GETRLIMIT:
		return "getrlimit"
	case unix.SYS_GETRUSAGE:
		return "getrusage"
	case unix.SYS_SYSINFO:
		return "sysinfo"
	case unix.SYS_TIMES:
		return "times"
	case unix.SYS_PTRACE:
		return "ptrace"
	case unix.SYS_GETUID:
		return "getuid"
	case unix.SYS_SYSLOG:
		return "syslog"
	case unix.SYS_GETGID:
		return "getgid"
	case unix.SYS_SETUID:
		return "setuid"
	case unix.SYS_SETGID:
		return "setgid"
	case unix.SYS_GETEUID:
		return "geteuid"
	case unix.SYS_GETEGID:
		return "getegid"
	case unix.SYS_SETPGID:
		return "setpgid"
	case unix.SYS_GETPPID:
		return "getppid"
	case unix.SYS_GETPGRP:
		return "getpgrp"
	case unix.SYS_SETSID:
		return "setsid"
	case unix.SYS_SETREUID:
		return "setreuid"
	case unix.SYS_SETREGID:
		return "setregid"
	case unix.SYS_GETGROUPS:
		return "getgroups"
	case unix.SYS_SETGROUPS:
		return "setgroups"
	case unix.SYS_SETRESUID:
		return "setresuid"
	case unix.SYS_GETRESUID:
		return "getresuid"
	case unix.SYS_SETRESGID:
		return "setresgid"
	case unix.SYS_GETRESGID:
		return "getresgid"
	case unix.SYS_GETPGID:
		return "getpgid"
	case unix.SYS_SETFSUID:
		return "setfsuid"
	case unix.SYS_SETFSGID:
		return "setfsgid"
	case unix.SYS_GETSID:
		return "getsid"
	case unix.SYS_CAPGET:
		return "capget"
	case unix.SYS_CAPSET:
		return "capset"
	case unix.SYS_RT_SIGPENDING:
		return "rt_sigpending"
	case unix.SYS_RT_SIGTIMEDWAIT:
		return "rt_sigtimedwait"
	case unix.SYS_RT_SIGQUEUEINFO:
		return "rt_sigqueueinfo"
	case unix.SYS_RT_SIGSUSPEND:
		return "rt_sigsuspend"
	case unix.SYS_SIGALTSTACK:
		return "sigaltstack"
	case unix.SYS_UTIME:
		return "utime"
	case unix.SYS_MKNOD:
		return "mknod"
	case unix.SYS_USELIB:
		return "uselib"
	case unix.SYS_PERSONALITY:
		return "personality"
	case unix.SYS_USTAT:
		return "ustat"
	case unix.SYS_STATFS:
		return "statfs"
	case unix.SYS_FSTATFS:
		return "fstatfs"
	case unix.SYS_SYSFS:
		return "sysfs"
	case unix.SYS_GETPRIORITY:
		return "getpriority"
	case unix.SYS_SETPRIORITY:
		return "setpriority"
	case unix.SYS_SCHED_SETPARAM:
		return "sched_setparam"
	case unix.SYS_SCHED_GETPARAM:
		return "sched_getparam"
	case unix.SYS_SCHED_SETSCHEDULER:
		return "sched_setscheduler"
	case unix.SYS_SCHED_GETSCHEDULER:
		return "sched_getscheduler"
	case unix.SYS_SCHED_GET_PRIORITY_MAX:
		return "sched_get_priority_max"
	case unix.SYS_SCHED_GET_PRIORITY_MIN:
		return "sched_get_priority_min"
	case unix.SYS_SCHED_RR_GET_INTERVAL:
		return "sched_rr_get_interval"
	case unix.SYS_MLOCK:
		return "mlock"
	case unix.SYS_MUNLOCK:
		return "munlock"
	case unix.SYS_MLOCKALL:
		return "mlockall"
	case unix.SYS_MUNLOCKALL:
		return "munlockall"
	case unix.SYS_VHANGUP:
		return "vhangup"
	case unix.SYS_MODIFY_LDT:
		return "modify_ldt"
	case unix.SYS_PIVOT_ROOT:
		return "pivot_root"
	case unix.SYS__SYSCTL:
		return "_sysctl"
	case unix.SYS_PRCTL:
		return "prctl"
	case unix.SYS_ARCH_PRCTL:
		return "arch_prctl"
	case unix.SYS_ADJTIMEX:
		return "adjtimex"
	case unix.SYS_SETRLIMIT:
		return "setrlimit"
	case unix.SYS_CHROOT:
		return "chroot"
	case unix.SYS_SYNC:
		return "sync"
	case unix.SYS_ACCT:
		return "acct"
	case unix.SYS_SETTIMEOFDAY:
		return "settimeofday"
	case unix.SYS_MOUNT:
		return "mount"
	case unix.SYS_UMOUNT2:
		return "umount2"
	case unix.SYS_SWAPON:
		return "swapon"
	case unix.SYS_SWAPOFF:
		return "swapoff"
	case unix.SYS_REBOOT:
		return "reboot"
	case unix.SYS_SETHOSTNAME:
		return "sethostname"
	case unix.SYS_SETDOMAINNAME:
		return "setdomainname"
	case unix.SYS_IOPL:
		return "iopl"
	case unix.SYS_IOPERM:
		return "ioperm"
	case unix.SYS_CREATE_MODULE:
		return "create_module"
	case unix.SYS_INIT_MODULE:
		return "init_module"
	case unix.SYS_DELETE_MODULE:
		return "delete_module"
	case unix.SYS_GET_KERNEL_SYMS:
		return "get_kernel_syms"
	case unix.SYS_QUERY_MODULE:
		return "query_module"
	case unix.SYS_QUOTACTL:
		return "quotactl"
	case unix.SYS_NFSSERVCTL:
		return "nfsservctl"
	case unix.SYS_GETPMSG:
		return "getpmsg"
	case unix.SYS_PUTPMSG:
		return "putpmsg"
	case unix.SYS_AFS_SYSCALL:
		return "afs_syscall"
	case unix.SYS_TUXCALL:
		return "tuxcall"
	case unix.SYS_SECURITY:
		return "security"
	case unix.SYS_GETTID:
		return "gettid"
	case unix.SYS_READAHEAD:
		return "readahead"
	case unix.SYS_SETXATTR:
		return "setxattr"
	case unix.SYS_LSETXATTR:
		return "lsetxattr"
	case unix.SYS_FSETXATTR:
		return "fsetxattr"
	case unix.SYS_GETXATTR:
		return "getxattr"
	case unix.SYS_LGETXATTR:
		return "lgetxattr"
	case unix.SYS_FGETXATTR:
		return "fgetxattr"
	case unix.SYS_LISTXATTR:
		return "listxattr"
	case unix.SYS_LLISTXATTR:
		return "llistxattr"
	case unix.SYS_FLISTXATTR:
		return "flistxattr"
	case unix.SYS_REMOVEXATTR:
		return "removexattr"
	case unix.SYS_LREMOVEXATTR:
		return "lremovexattr"
	case unix.SYS_FREMOVEXATTR:
		return "fremovexattr"
	case unix.SYS_TKILL:
		return "tkill"
	case unix.SYS_TIME:
		return "time"
	case unix.SYS_FUTEX:
		return "futex"
	case unix.SYS_SCHED_SETAFFINITY:
		return "sched_setaffinity"
	case unix.SYS_SCHED_GETAFFINITY:
		return "sched_getaffinity"
	case unix.SYS_SET_THREAD_AREA:
		return "set_thread_area"
	case unix.SYS_IO_SETUP:
		return "io_setup"
	case unix.SYS_IO_DESTROY:
		return "io_destroy"
	case unix.SYS_IO_GETEVENTS:
		return "io_getevents"
	case unix.SYS_IO_SUBMIT:
		return "io_submit"
	case unix.SYS_IO_CANCEL:
		return "io_cancel"
	case unix.SYS_GET_THREAD_AREA:
		return "get_thread_area"
	case unix.SYS_LOOKUP_DCOOKIE:
		return "lookup_dcookie"
	case unix.SYS_EPOLL_CREATE:
		return "epoll_create"
	case unix.SYS_EPOLL_CTL_OLD:
		return "epoll_ctl_old"
	case unix.SYS_EPOLL_WAIT_OLD:
		return "epoll_wait_old"
	case unix.SYS_REMAP_FILE_PAGES:
		return "remap_file_pages"
	case unix.SYS_GETDENTS64:
		return "getdents64"
	case unix.SYS_SET_TID_ADDRESS:
		return "set_tid_address"
	case unix.SYS_RESTART_SYSCALL:
		return "restart_syscall"
	case unix.SYS_SEMTIMEDOP:
		return "semtimedop"
	case unix.SYS_FADVISE64:
		return "fadvise64"
	case unix.SYS_TIMER_CREATE:
		return "timer_create"
	case unix.SYS_TIMER_SETTIME:
		return "timer_settime"
	case unix.SYS_TIMER_GETTIME:
		return "timer_gettime"
	case unix.SYS_TIMER_GETOVERRUN:
		return "timer_getoverrun"
	case unix.SYS_TIMER_DELETE:
		return "timer_delete"
	case unix.SYS_CLOCK_SETTIME:
		return "clock_settime"
	case unix.SYS_CLOCK_GETTIME:
		return "clock_gettime"
	case unix.SYS_CLOCK_GETRES:
		return "clock_getres"
	case unix.SYS_CLOCK_NANOSLEEP:
		return "clock_nanosleep"
	case unix.SYS_EXIT_GROUP:
		return "exit_group"
	case unix.SYS_EPOLL_WAIT:
		return "epoll_wait"
	case unix.SYS_EPOLL_CTL:
		return "epoll_ctl"
	case unix.SYS_TGKILL:
		return "tgkill"
	case unix.SYS_UTIMES:
		return "utimes"
	case unix.SYS_VSERVER:
		return "vserver"
	case unix.SYS_MBIND:
		return "mbind"
	case unix.SYS_SET_MEMPOLICY:
		return "set_mempolicy"
	case unix.SYS_GET_MEMPOLICY:
		return "get_mempolicy"
	case unix.SYS_MQ_OPEN:
		return "mq_open"
	case unix.SYS_MQ_UNLINK:
		return "mq_unlink"
	case unix.SYS_MQ_TIMEDSEND:
		return "mq_timedsend"
	case unix.SYS_MQ_TIMEDRECEIVE:
		return "mq_timedreceive"
	case unix.SYS_MQ_NOTIFY:
		return "mq_notify"
	case unix.SYS_MQ_GETSETATTR:
		return "mq_getsetattr"
	case unix.SYS_KEXEC_LOAD:
		return "kexec_load"
	case unix.SYS_WAITID:
		return "waitid"
	case unix.SYS_ADD_KEY:
		return "add_key"
	case unix.SYS_REQUEST_KEY:
		return "request_key"
	case unix.SYS_KEYCTL:
		return "keyctl"
	case unix.SYS_IOPRIO_SET:
		return "ioprio_set"
	case unix.SYS_IOPRIO_GET:
		return "ioprio_get"
	case unix.SYS_INOTIFY_INIT:
		return "inotify_init"
	case unix.SYS_INOTIFY_ADD_WATCH:
		return "inotify_add_watch"
	case unix.SYS_INOTIFY_RM_WATCH:
		return "inotify_rm_watch"
	case unix.SYS_MIGRATE_PAGES:
		return "migrate_pages"
	case unix.SYS_OPENAT:
		return "openat"
	case unix.SYS_MKDIRAT:
		return "mkdirat"
	case unix.SYS_MKNODAT:
		return "mknodat"
	case unix.SYS_FCHOWNAT:
		return "fchownat"
	case unix.SYS_FUTIMESAT:
		return "futimesat"
	case unix.SYS_NEWFSTATAT:
		return "newfstatat"
	case unix.SYS_UNLINKAT:
		return "unlinkat"
	case unix.SYS_RENAMEAT:
		return "renameat"
	case unix.SYS_LINKAT:
		return "linkat"
	case unix.SYS_SYMLINKAT:
		return "symlinkat"
	case unix.SYS_READLINKAT:
		return "readlinkat"
	case unix.SYS_FCHMODAT:
		return "fchmodat"
	case unix.SYS_FACCESSAT:
		return "faccessat"
	case unix.SYS_PSELECT6:
		return "pselect6"
	case unix.SYS_PPOLL:
		return "ppoll"
	case unix.SYS_UNSHARE:
		return "unshare"
	case unix.SYS_SET_ROBUST_LIST:
		return "set_robust_list"
	case unix.SYS_GET_ROBUST_LIST:
		return "get_robust_list"
	case unix.SYS_SPLICE:
		return "splice"
	case unix.SYS_TEE:
		return "tee"
	case unix.SYS_SYNC_FILE_RANGE:
		return "sync_file_range"
	case unix.SYS_VMSPLICE:
		return "vmsplice"
	case unix.SYS_MOVE_PAGES:
		return "move_pages"
	case unix.SYS_UTIMENSAT:
		return "utimensat"
	case unix.SYS_EPOLL_PWAIT:
		return "epoll_pwait"
	case unix.SYS_SIGNALFD:
		return "signalfd"
	case unix.SYS_TIMERFD_CREATE:
		return "timerfd_create"
	case unix.SYS_EVENTFD:
		return "eventfd"
	case unix.SYS_FALLOCATE:
		return "fallocate"
	case unix.SYS_TIMERFD_SETTIME:
		return "timerfd_settime"
	case unix.SYS_TIMERFD_GETTIME:
		return "timerfd_gettime"
	case unix.SYS_ACCEPT4:
		return "accept4"
	case unix.SYS_SIGNALFD4:
		return "signalfd4"
	case unix.SYS_EVENTFD2:
		return "eventfd2"
	case unix.SYS_EPOLL_CREATE1:
		return "epoll_create1"
	case unix.SYS_DUP3:
		return "dup3"
	case unix.SYS_PIPE2:
		return "pipe2"
	case unix.SYS_INOTIFY_INIT1:
		return "inotify_init1"
	case unix.SYS_PREADV:
		return "preadv"
	case unix.SYS_PWRITEV:
		return "pwritev"
	case unix.SYS_RT_TGSIGQUEUEINFO:
		return "rt_tgsigqueueinfo"
	case unix.SYS_PERF_EVENT_OPEN:
		return "perf_event_open"
	case unix.SYS_RECVMMSG:
		return "recvmmsg"
	case unix.SYS_FANOTIFY_INIT:
		return "fanotify_init"
	case unix.SYS_FANOTIFY_MARK:
		return "fanotify_mark"
	case unix.SYS_PRLIMIT64:
		return "prlimit64"
	case unix.SYS_NAME_TO_HANDLE_AT:
		return "name_to_handle_at"
	case unix.SYS_OPEN_BY_HANDLE_AT:
		return "open_by_handle_at"
	case unix.SYS_CLOCK_ADJTIME:
		return "clock_adjtime"
	case unix.SYS_SYNCFS:
		return "syncfs"
	case unix.SYS_SENDMMSG:
		return "sendmmsg"
	case unix.SYS_SETNS:
		return "setns"
	case unix.SYS_GETCPU:
		return "getcpu"
	case unix.SYS_PROCESS_VM_READV:
		return "process_vm_readv"
	case unix.SYS_PROCESS_VM_WRITEV:
		return "process_vm_writev"
	case unix.SYS_KCMP:
		return "kcmp"
	case unix.SYS_FINIT_MODULE:
		return "finit_module"
	case unix.SYS_SCHED_SETATTR:
		return "sched_setattr"
	case unix.SYS_SCHED_GETATTR:
		return "sched_getattr"
	case unix.SYS_RENAMEAT2:
		return "renameat2"
	case unix.SYS_SECCOMP:
		return "seccomp"
	case unix.SYS_GETRANDOM:
		return "getrandom"
	case unix.SYS_MEMFD_CREATE:
		return "memfd_create"
	case unix.SYS_KEXEC_FILE_LOAD:
		return "kexec_file_load"
	case unix.SYS_BPF:
		return "bpf"
	case unix.SYS_EXECVEAT:
		return "execveat"
	case unix.SYS_USERFAULTFD:
		return "userfaultfd"
	case unix.SYS_MEMBARRIER:
		return "membarrier"
	case unix.SYS_MLOCK2:
		return "mlock2"
	case unix.SYS_COPY_FILE_RANGE:
		return "copy_file_range"
	case unix.SYS_PREADV2:
		return "preadv2"
	case unix.SYS_PWRITEV2:
		return "pwritev2"
	case unix.SYS_PKEY_MPROTECT:
		return "pkey_mprotect"
	case unix.SYS_PKEY_ALLOC:
		return "pkey_alloc"
	case unix.SYS_PKEY_FREE:
		return "pkey_free"
	case unix.SYS_STATX:
		return "statx"
	case unix.SYS_IO_PGETEVENTS:
		return "io_pgetevents"
	case unix.SYS_RSEQ:
		return "rseq"
	}
	return fmt.Sprintf("%d - ERR_UNKNOWN_SYSCALL", e)
}
