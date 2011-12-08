// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The template uses the function "code" to inject program
// source into the output by extracting code from files and
// injecting them as HTML-escaped <pre> blocks.
//
// The syntax is simple: 1, 2, or 3 space-separated arguments:
//
// Whole file:
//	{{code "foo.go"}}
// One line (here the signature of main):
//	{{code "foo.go" `/^func.main/`}}
// Block of text, determined by start and end (here the body of main):
//	{{code "foo.go" `/^func.main/` `/^}/`
//
// Patterns can be `/regular expression/`, a decimal number, or "$"
// to signify the end of the file.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
)

func Usage() {
	fmt.Fprintf(os.Stderr, "usage: tmpltohtml file\n")
	os.Exit(2)
}

var templateFuncs = template.FuncMap{
	"code":      code,
	"donotedit": donotedit,
}

func main() {
	flag.Usage = Usage
	flag.Parse()
	if len(flag.Args()) != 1 {
		Usage()
	}

	// Read and parse the input.
	name := flag.Args()[0]
	tmpl := template.New(name).Funcs(templateFuncs)
	if _, err := tmpl.ParseFiles(name); err != nil {
		log.Fatal(err)
	}

	// Execute the template.
	if err := tmpl.Execute(os.Stdout, 0); err != nil {
		log.Fatal(err)
	}
}

// contents reads a file by name and returns its contents as a string.
func contents(name string) string {
	file, err := ioutil.ReadFile(name)
	if err != nil {
		log.Fatal(err)
	}
	return string(file)
}

// format returns a textual representation of the arg, formatted according to its nature.
func format(arg interface{}) string {
	switch arg := arg.(type) {
	case int:
		return fmt.Sprintf("%d", arg)
	case string:
		if len(arg) > 2 && arg[0] == '/' && arg[len(arg)-1] == '/' {
			return fmt.Sprintf("%#q", arg)
		}
		return fmt.Sprintf("%q", arg)
	default:
		log.Fatalf("unrecognized argument: %v type %T", arg, arg)
	}
	return ""
}

func donotedit() string {
	// No editing please.
	return fmt.Sprintf("<!--\n  DO NOT EDIT: created by\n    tmpltohtml %s\n-->\n", flag.Args()[0])
}

func code(file string, arg ...interface{}) (string, error) {
	text := contents(file)
	var command string
	switch len(arg) {
	case 0:
		// text is already whole file.
		command = fmt.Sprintf("code %q", file)
	case 1:
		command = fmt.Sprintf("code %q %s", file, format(arg[0]))
		text = oneLine(file, text, arg[0])
	case 2:
		command = fmt.Sprintf("code %q %s %s", file, format(arg[0]), format(arg[1]))
		text = multipleLines(file, text, arg[0], arg[1])
	default:
		return "", fmt.Errorf("incorrect code invocation: code %q %q", file, arg)
	}
	// Replace tabs by spaces, which work better in HTML.
	text = strings.Replace(text, "\t", "    ", -1)
	// Escape the program text for HTML.
	text = template.HTMLEscapeString(text)
	// Include the command as a comment.
	text = fmt.Sprintf("<pre><!--{{%s}}\n-->%s</pre>", command, text)
	return text, nil
}

// parseArg returns the integer or string value of the argument and tells which it is.
func parseArg(arg interface{}, file string, max int) (ival int, sval string, isInt bool) {
	switch n := arg.(type) {
	case int:
		if n <= 0 || n > max {
			log.Fatalf("%q:%d is out of range", file, n)
		}
		return n, "", true
	case string:
		return 0, n, false
	}
	log.Fatalf("unrecognized argument %v type %T", arg, arg)
	return
}

// oneLine returns the single line generated by a two-argument code invocation.
func oneLine(file, text string, arg interface{}) string {
	lines := strings.SplitAfter(contents(file), "\n")
	line, pattern, isInt := parseArg(arg, file, len(lines))
	if isInt {
		return lines[line-1]
	}
	return lines[match(file, 0, lines, pattern)-1]
}

// multipleLines returns the text generated by a three-argument code invocation.
func multipleLines(file, text string, arg1, arg2 interface{}) string {
	lines := strings.SplitAfter(contents(file), "\n")
	line1, pattern1, isInt1 := parseArg(arg1, file, len(lines))
	line2, pattern2, isInt2 := parseArg(arg2, file, len(lines))
	if !isInt1 {
		line1 = match(file, 0, lines, pattern1)
	}
	if !isInt2 {
		line2 = match(file, line1, lines, pattern2)
	} else if line2 < line1 {
		log.Fatalf("lines out of order for %q: %d %d", text, line1, line2)
	}
	return strings.Join(lines[line1-1:line2], "")
}

// match identifies the input line that matches the pattern in a code invocation.
// If start>0, match lines starting there rather than at the beginning.
// The return value is 1-indexed.
func match(file string, start int, lines []string, pattern string) int {
	// $ matches the end of the file.
	if pattern == "$" {
		if len(lines) == 0 {
			log.Fatalf("%q: empty file", file)
		}
		return len(lines)
	}
	// /regexp/ matches the line that matches the regexp.
	if len(pattern) > 2 && pattern[0] == '/' && pattern[len(pattern)-1] == '/' {
		re, err := regexp.Compile(pattern[1 : len(pattern)-1])
		if err != nil {
			log.Fatal(err)
		}
		for i := start; i < len(lines); i++ {
			if re.MatchString(lines[i]) {
				return i + 1
			}
		}
		log.Fatalf("%s: no match for %#q", file, pattern)
	}
	log.Fatalf("unrecognized pattern: %q", pattern)
	return 0
}
