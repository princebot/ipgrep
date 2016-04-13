# ipgrep
```usage: ipgrep [file ...]```

**ipgrep** scans one or more input files for valid IPv4 or IPv6 addresses and prints the result. It accepts text files in any format (plaintext, JSON, YAML, etc.) so long as the files contain IPs separated by a delimiter it can recognize (i.e., any whitespace character and any punctuation character other than `.` or `:`).

These are all valid inputs:

	10.10.10.2 https://webserver.com
	{"ip": "172.16.2.84"}
	log -> time=13:10, event=foo, addr=192.168.0.2, desc="a foo went bar"
	IP address 8.8.8.8 is for Google DNS.

**ipgrep** would extract `10.10.10.2`, `172.16.2.84`, `192.168.0.2`, and `8.8.8.8` from the above.

Given this input, however —

	There’s no place like 127.0.0.1.

— **ipgrep** extracts nothing: Technically, the final `.` renders that IP invalid, and this utility does not aspire to robustness.

Use standard Golang-fu to install: `go get -u github.com/princebot/ipgrep`
