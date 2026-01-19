// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package cmdutil

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/elastic/harp/build/version"

	exec "golang.org/x/sys/execabs"
)

// BugReport generates a bug report body
//
//nolint:godox // Bug not allow in documentation
func BugReport() string {
	var buf bytes.Buffer
	buf.WriteString(bugHeader)
	printGoVersion(&buf)
	buf.WriteString("### Does this issue reproduce with the latest release?\n\n\n")
	printEnvDetails(&buf)
	printAppDetails(&buf)
	buf.WriteString(bugFooter)

	return buf.String()
}

const bugHeader = `<!-- Please answer these questions before submitting your issue. Thanks! -->
`

const bugFooter = `### What did you do?
<!--
If possible, provide a recipe for reproducing the error.
A complete runnable program is good.
-->

### What did you expect to see?
### What did you see instead?
`

func printGoVersion(w io.Writer) {
	_, _ = fmt.Fprintf(w, "### What version of Go are you using (`go version`)?\n\n")
	_, _ = fmt.Fprintf(w, "<pre>\n")
	_, _ = fmt.Fprintf(w, "$ go version\n")
	printCmdOut(w, "", "go", "version")
	_, _ = fmt.Fprintf(w, "</pre>\n")
	_, _ = fmt.Fprintf(w, "\n")
}

func printEnvDetails(w io.Writer) {
	_, _ = fmt.Fprintf(w, "### What operating system and processor architecture are you using (`go env`)?\n\n")
	_, _ = fmt.Fprintf(w, "<details><summary><code>go env</code> Output</summary><br><pre>\n")
	_, _ = fmt.Fprintf(w, "$ go env\n")
	printCmdOut(w, "", "go", "env")
	printGoDetails(w)
	printOSDetails(w)
	printCDetails(w)
	_, _ = fmt.Fprintf(w, "</pre></details>\n\n")
}

func printGoDetails(w io.Writer) {
	// Get GOROOT from `go env GOROOT` instead of deprecated runtime.GOROOT()
	cmd := exec.Command("go", "env", "GOROOT")
	out, err := cmd.Output()
	if err != nil {
		return
	}
	goroot := string(bytes.TrimSpace(out))
	printCmdOut(w, "GOROOT/bin/go version: ", filepath.Join(goroot, "bin", "go"), "version")
	printCmdOut(w, "GOROOT/bin/go tool compile -V: ", filepath.Join(goroot, "bin", "go"), "tool", "compile", "-V")
}

func printOSDetails(w io.Writer) {
	switch runtime.GOOS {
	case "darwin":
		printCmdOut(w, "uname -v: ", "uname", "-v")
		printCmdOut(w, "", "sw_vers")
	case "linux":
		printCmdOut(w, "uname -sr: ", "uname", "-sr")
		printCmdOut(w, "", "lsb_release", "-a")
		printGlibcVersion(w)
	case "openbsd", "netbsd", "freebsd", "dragonfly":
		printCmdOut(w, "uname -v: ", "uname", "-v")
	case "illumos", "solaris":
		// Be sure to use the OS-supplied uname, in "/usr/bin":
		printCmdOut(w, "uname -srv: ", "/usr/bin/uname", "-srv")
		out, err := os.ReadFile("/etc/release")
		if err == nil {
			_, _ = fmt.Fprintf(w, "/etc/release: %s\n", out)
		}
	}
}

func printAppDetails(w io.Writer) {
	bi := version.NewInfo()
	_, _ = fmt.Fprintf(w, "### What version of Secret are you using (`harp version`)?\n\n")
	_, _ = fmt.Fprintf(w, "<pre>\n")
	_, _ = fmt.Fprintf(w, "$ harp version\n")
	_, _ = fmt.Fprintf(w, "%s\n", bi.String())
	_, _ = fmt.Fprintf(w, "</pre>\n")
	_, _ = fmt.Fprintf(w, "\n")
}

func printCDetails(w io.Writer) {
	printCmdOut(w, "lldb --version: ", "lldb", "--version")
	cmd := exec.Command("gdb", "--version")
	out, err := cmd.Output()
	if err == nil {
		// There's apparently no combination of command line flags
		// to get gdb to spit out its version without the license and warranty.
		// Print up to the first newline.
		_, _ = fmt.Fprintf(w, "gdb --version: %s\n", firstLine(out))
	}
}

// printCmdOut prints the output of running the given command.
// It ignores failures; 'go bug' is best effort.
func printCmdOut(w io.Writer, prefix, path string, args ...string) {
	cmd := exec.Command(path, args...)
	out, err := cmd.Output()
	if err != nil {
		return
	}
	_, _ = fmt.Fprintf(w, "%s%s\n", prefix, bytes.TrimSpace(out))
}

// firstLine returns the first line of a given byte slice.
func firstLine(buf []byte) []byte {
	idx := bytes.IndexByte(buf, '\n')
	if idx > 0 {
		buf = buf[:idx]
	}
	return bytes.TrimSpace(buf)
}

// printGlibcVersion prints information about the glibc version.
// It ignores failures.
func printGlibcVersion(w io.Writer) {
	tempdir := os.TempDir()
	if tempdir == "" {
		return
	}
	src := []byte(`int main() {}`)
	srcfile := filepath.Join(tempdir, "go-bug.c")
	outfile := filepath.Join(tempdir, "go-bug")
	err := os.WriteFile(srcfile, src, 0o600)
	if err != nil {
		return
	}
	defer func() { _ = os.Remove(srcfile) }()
	//nolint:gosec // G204: srcfile/outfile are generated temp paths, not user input
	cmd := exec.Command("gcc", "-o", outfile, srcfile)
	if _, err = cmd.CombinedOutput(); err != nil {
		return
	}
	defer func() { _ = os.Remove(outfile) }()

	//nolint:gosec // G204: outfile is a generated temp path, not user input
	cmd = exec.Command("ldd", outfile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return
	}
	re := regexp.MustCompile(`libc\.so[^ ]* => ([^ ]+)`)
	m := re.FindStringSubmatch(string(out))
	if m == nil {
		return
	}
	//nolint:gosec // controlled input
	cmd = exec.Command(m[1])
	out, err = cmd.Output()
	if err != nil {
		return
	}
	_, _ = fmt.Fprintf(w, "%s: %s\n", m[1], firstLine(out))

	// print another line (the one containing version string) in case of musl libc
	if idx := bytes.IndexByte(out, '\n'); bytes.Contains(out, []byte("musl")) {
		_, _ = fmt.Fprintf(w, "%s\n", firstLine(out[idx+1:]))
	}
}
