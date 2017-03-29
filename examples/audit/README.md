audit
=====

This is an example of an audit daemon that is implemented using the `AuditClient`
from the `github.com/elastic/gosigar/sys/linux` package. It connects to the
kernel and registers to receive audit messages.

As this is only an example, it does not manage audit rules. So you must install
them with another tool (like `auditctl`). You must stop the auditd daemon
before running the example (`service stop auditd` for RHEL).

```yaml
go get github.com/elastic/gosigar/examples/audit
sudo $GOPATH/bin/audit -d -format=json
```

This will output one JSON event for each audit message. You can pipe the output
through `jq` if you would like the output to be pretty. Below is an example of
the output.

```json
{
  "@timestamp": "2017-03-31 22:08:25.96 +0000 UTC",
  "a0": "4",
  "a1": "7f808e0c4408",
  "a2": "10",
  "a3": "0",
  "arch": "x86_64",
  "auid": "4294967295",
  "comm": "ntpd",
  "egid": "38",
  "euid": "38",
  "exe": "/usr/sbin/ntpd",
  "exit": "0",
  "fsgid": "38",
  "fsuid": "38",
  "gid": "38",
  "items": "0",
  "pid": "1106",
  "ppid": "1",
  "raw_msg": "audit(1490998105.960:595907): arch=c000003e syscall=42 success=yes exit=0 a0=4 a1=7f808e0c4408 a2=10 a3=0 items=0 ppid=1 pid=1106 auid=4294967295 uid=38 gid=38 euid=38 suid=38 fsuid=38 egid=38 sgid=38 fsgid=38 tty=(none) ses=4294967295 comm=\"ntpd\" exe=\"/usr/sbin/ntpd\" subj=system_u:system_r:ntpd_t:s0 key=(null)",
  "record_type": "SYSCALL",
  "sequence": "595907",
  "ses": "4294967295",
  "sgid": "38",
  "subj": "system_u:system_r:ntpd_t:s0",
  "success": "yes",
  "suid": "38",
  "syscall": "connect",
  "tty": "(none)",
  "uid": "38"
}
```