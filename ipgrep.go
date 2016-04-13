// ipgrep scans one or more input files for valid IPv4 or IPv6 addresses and
// prints the result.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"unicode"

	"sync"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

const (
	prog       = "ipgrep"
	banlistURL = `https://www.binarydefense.com/banlist.txt`
)

const usage = `
usage: %[1]v [file ...]

%[1]v scans one or more input files for valid IPv4 or IPv6 addresses and prints
the result. It accepts text files in any format (newline-delimited, JSON, YAML,
etc.) so long as the files contain IPs separated either by whitespace or by any
punctuation character other than '.' or ':'.

For example, these are all valid input:

	10.10.10.2 https://webserver.com
	{"ip": "172.16.2.84"}
	log -> time=13:10, event=foo, addr=192.168.0.2, desc="a foo went bar"
	IP address 8.8.8.8 is for Google DNS.

ipgrep would extract 10.10.10.2, 172.16.2.84, 192.168.0.2, and 8.8.8.8 from the
above. However, given this input — 

	There’s no place like 127.0.0.1.

— ipgrep extracts nothing: The final '.' renders the address invalid, and this
utility doesn’t try quite that hard.
`

// scanResult stores the results of processing a single input file.
type scanResult struct {
	File string   // path to the input file.
	IPs  []net.IP // list of IPs parsed from the file.
	Err  error    // set if an I/O error occurs or the file is empty.
}

// Error satisfies the error interface.
func (r scanResult) Error() string {
	if r.Err == nil {
		return ""
	}
	return fmt.Sprintf("error: %v: %v", r.File, r.Err)
}

func main() {
	if len(os.Args) < 2 {
		help()
	}
	switch os.Args[1] {
	case "-h", "-help", "--help":
		help()
	}

	// If any of the input files cannot be read, quit with an error.
	var files []*os.File
	for _, fn := range os.Args[1:] {
		fp, err := os.Open(fn)
		if err != nil {
			die(err)
		}
		files = append(files, fp)
	}

	var (
		results = make(chan *scanResult, len(files))
		wg      sync.WaitGroup
	)
	// Create a goroutine to scan and parse each input file, collecting the
	// results in a channel.
	for _, fp := range files {
		wg.Add(1)
		go func(fp *os.File) {
			defer wg.Done()
			defer fp.Close()
			results <- scan(fp)
		}(fp)
	}
	wg.Wait()
	close(results)

	var failed []*scanResult
	for r := range results {
		// Show successfully extracted IPs first; display errors later.
		if r.Err != nil {
			failed = append(failed, r)
			continue
		}
		fmt.Printf("# results for %v:\n", r.File)
		for _, ip := range r.IPs {
			fmt.Println(ip)
		}
		fmt.Println()
	}
	if len(failed) > 0 {
		fmt.Println("# errors:")
		for _, r := range failed {
			printError(r)
		}
	}
}

// split is used to divide file content into “words” that might be valid IP
// addresses.
func split(r rune) bool {
	if unicode.IsSpace(r) || unicode.IsPunct(r) && r != '.' && r != ':' {
		return true
	}
	return false
}

// scan reads a file, splits its content in “words,” and tests each word to see
// if it is a valid IPv4 or IPv6 address. If reading the file causes an I/O
// error, or if the file is empty, *scanResult will have a non-nil Err field.
func scan(fp *os.File) *scanResult {
	var (
		res = &scanResult{File: fp.Name()}
		b   []byte
	)
	if b, res.Err = ioutil.ReadAll(fp); res.Err != nil {
		return res
	}
	if len(b) == 0 {
		res.Err = fmt.Errorf("empty file")
		return res
	}
	for _, word := range bytes.FieldsFunc(b, split) {
		if ip := net.ParseIP(string(word)); ip != nil {
			res.IPs = append(res.IPs, ip)
		}
	}
	return res
}

func help() {
	fmt.Fprintf(os.Stderr, usage, prog)
	os.Exit(0)
}

func printError(errMsg interface{}) {
	var (
		stderr = colorable.NewColorableStderr()
		red    = color.New(color.FgRed).SprintfFunc()
	)
	fmt.Fprintf(stderr, red("\n%v: error: %v\n", prog, errMsg))
}

func die(errMsg interface{}) {
	printError(errMsg)
	os.Exit(1)
}
