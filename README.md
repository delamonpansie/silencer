## About

Silencer is a simple replacement for [fail2ban](https://www.fail2ban.org) written in Go.

After several hours of unsuccessful configuring of fail2ban I gave up and decided to build my own.

## Running
```

silencer [-config silencer.yaml] [-debug-rule] [-dry-run]

```

## Configuration
The configuration is stored in YAML file. During startup silencer will
try to read "silencer.yaml" in the current directory. It is possible
to override location via `-config` option.


`log_file` section defines a collection of log files to monitor and
rules attached to them. Rules are used to match and extract IP address
from a log line.


Rule matching works by using a sequence of regexes to match and trim
line until only IP remains. If the regex fails to match, then the rule
is considered failed, and no more matching is performed. If regex
contains capture group, then log line will be replaced with the value
of capture group.


`env` section defines commons strings. All regexes are expanded using
these strings.

### Duration

A duration value uses [Go syntax](https://pkg.go.dev/time#ParseDuration).

There are three ways to configure a block duration in config:
1. top level `duration` defines the default duration. It will be used
   if there is no file or rule level config.

2. file level `duration` defines duration for a matches coming from a
   given file.  It has priority over the default duration.

3. rule level `duration` defines duration for a matches of a given
   rule. It has highest priority.

## Debugging rules

Sometimes it is helpful to check if specific rules correctly parses
log lines.  There is a debug mode for that: to check rule `auth` in
`silencer.yaml` run the following command `./silencer -config
silencer.yaml -debug-rule auth`.

## Building & testing

```
git clone https://github.com/delamonpansie/silencer
cd silencer
go install go.uber.org/mock/mockgen@latest
go generate ./...
go test ./...
go build .
```


## Example configuration

This is a minimal working example with nftables. The nft configuration
is very minimal and would require extension for local needs (e.g. add
FORWARD chain). The log monitoring is configured for exim, ssh, and
named. Attempts to use smtp auth, guess ssh usernames or passwords,
use dns recursion will be blocked for one week.


1. `/etc/nftables.conf`
```
#!/usr/sbin/nft -f
flush ruleset

table ip inet {
	set silence {
		type ipv4_addr
		flags interval,timeout
		comment "drop all packets from these hosts"
	}

	chain INPUT {
		type filter hook input priority filter; policy drop;
		iif { "lo", "eth0" } counter accept comment "always allow local network"
		iifname { "ppp0" } ip saddr @silence counter drop comment "filter out garbage only on WAN interface"
		tcp dport { 22, 25, 80, 443, 465 } counter accept comment "public accessible services"
   }
}
```

2. `/etc/silencer.yaml`
```yaml
filter:
  nft:
    set: silence
    table: inet

# default duration
duration: 168h

# never block hosts from these networks
whitelist:
  - ip: 192.168.0.0
    mask: [255, 255, 0, 0]
  - ip: 10.0.0.0
    mask: [255, 0, 0, 0]

# some useful regexes
env:
  ip:         (?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)
  date_time1: \d{4}-(?:0\d|10|11|12)-(?:[012]\d|3[01]) \d{2}:\d{2}:\d{2}.\d+
  date_time:  \d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}.\d+
  date_time_iso: \d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}

log_file:
  - file_name: /var/log/exim4/mainlog
    rule:
    # block whoever tries using AUTH LOGIN on smtp port. We do not use authentication, so
    # who ever tries auth is picking passowrds
    - name: auth
      re:
          # example line: `2024-09-10 02:20:07.641 [20764] SMTP protocol error in "AUTH LOGIN" H=(User) [xx.xx.xx.xx]:xx I=[xx.xx.xx.xx]:25 AUTH command used when not advertised`

          # Rule matching works by using a sequence of regexes to match and trim line until only IP remains.
          # If the regex fails to match, then the rule is considered failed, and no more matching
          # is performed. If regex contains capture group, then log line will be replaced with the value
          # of capture group.

          # match begging of the line with /^$date_time \[\d+\] SMTP protocol error in "AUTH LOGIN"/ and select tail for latter processing.
          # line buffer would be `H=(User) [xx.xx.xx.xx]:xx I=[xx.xx.xx.xx]:25 AUTH command used when not advertised` after this regex has been matched
          - ^$date_time \[\d+\] SMTP protocol error in "AUTH LOGIN" (.*)

          # trim tail ` AUTH command used when not advertised`
          # line buffer would be `H=(User) [xx.xx.xx.xx]:xx I=[xx.xx.xx.xx]:25` after this regex has been matched.
          - (.*) AUTH command used when not advertised$$

          # trim H=(User) part. Sometimes it contains ip idderss inside (), so we should be careful
          # line buffer would be `[xx.xx.xx.xx]:xx I=[xx.xx.xx.xx]:25` after this regex has been matched
          - ^(?:H=(?:[\w.-]+ )?(?:\(\S+\) )?)?(.*)

          # finnaly, match the offender's ip
          - \[($ip)\]

  - file_name: /var/log/auth.log
    duration: 360h # use 360h block duration for matches in this file
    rule:
      # block whoever tries to guess valid usernames
      - name: sshd
        duration: 720h # this specific rule will result in 720h block
        re:
          # we're matching generic auth.log here, so we need to pick lines coming from sshd first.
          # note, that we also dropping matched prefix, because we use capture group to match tail of the line. Contents of this capture group will be used as next line buffer.
          - ^$date_time_iso \S+ sshd\[\d+\][:]\s(.*)

          # match the offender ip. Since this is the only ip address in the log line, the rule is very simple
          - ^Disconnected from invalid user \S+ ($ip) port \d+ \[preauth\]

      # block whoever guessed right username but failed to provide correct pasword. It's safe since we have
      # ssh key authentication as our primary access method. Might be risky otherwise.
      - name: sshd-pam
        re:
          - ^$date_time_iso \S+ sshd\[\d+\][:]\s(.*)
          - ^pam_unix\(sshd:auth\)[:] authentication failure; logname= uid=0 euid=0 tty=ssh ruser= rhost=($ip)\s*(?:\s+user=\S*)?$$
  - file_name: /var/log/daemon.log
    rule:
      # we do not provide recursive DNS service, so who tries to do recursion should be blocked
      - name: named
        re:
          # pick lines from named first
          - ^$date_time_iso \S+ named\[\d+\][:]\s(.*)
          - (.*) \(\S+\)[:] view public[:] query \(cache\) '\S+/[A-Z]+/IN' denied$$
          - client @0x[\da-f]+ ($ip)#\d+
```
