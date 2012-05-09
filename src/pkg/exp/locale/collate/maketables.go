// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// Collation table generator.
// Data read from the web.

package main

import (
	"bufio"
	"exp/locale/collate"
	"exp/locale/collate/build"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var ducet = flag.String("ducet",
	"http://unicode.org/Public/UCA/"+unicode.Version+"/allkeys.txt",
	"URL of the Default Unicode Collation Element Table (DUCET).")
var localFiles = flag.Bool("local",
	false,
	"data files have been copied to the current directory; for debugging only")

func failonerror(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

// openReader opens the url or file given by url and returns it as an io.ReadCloser
// or nil on error.
func openReader(url string) (io.ReadCloser, error) {
	if *localFiles {
		pwd, _ := os.Getwd()
		url = "file://" + path.Join(pwd, path.Base(url))
	}
	t := &http.Transport{}
	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
	c := &http.Client{Transport: t}
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(`bad GET status for "%s": %s`, url, resp.Status)
	}
	return resp.Body, nil
}

// parseUCA parses a Default Unicode Collation Element Table of the format
// specified in http://www.unicode.org/reports/tr10/#File_Format.
// It returns the variable top.
func parseUCA(builder *build.Builder) int {
	maxVar, minNonVar := 0, 1<<30
	r, err := openReader(*ducet)
	failonerror(err)
	defer r.Close()
	input := bufio.NewReader(r)
	colelem := regexp.MustCompile(`\[([.*])([0-9A-F.]+)\]`)
	for i := 1; err == nil; i++ {
		l, prefix, e := input.ReadLine()
		err = e
		line := string(l)
		if prefix {
			log.Fatalf("%d: buffer overflow", i)
		}
		if err != nil && err != io.EOF {
			log.Fatalf("%d: %v", i, err)
		}
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		if line[0] == '@' {
			// parse properties
			switch {
			case strings.HasPrefix(line[1:], "version "):
				a := strings.Split(line[1:], " ")
				if a[1] != unicode.Version {
					log.Fatalf("incompatible version %s; want %s", a[1], unicode.Version)
				}
			case strings.HasPrefix(line[1:], "backwards "):
				log.Fatalf("%d: unsupported option backwards", i)
			default:
				log.Printf("%d: unknown option %s", i, line[1:])
			}
		} else {
			// parse entries
			part := strings.Split(line, " ; ")
			if len(part) != 2 {
				log.Fatalf("%d: production rule without ';': %v", i, line)
			}
			lhs := []rune{}
			for _, v := range strings.Split(part[0], " ") {
				if v == "" {
					continue
				}
				lhs = append(lhs, rune(convHex(i, v)))
			}
			var n int
			rhs := [][]int{}
			for _, m := range colelem.FindAllStringSubmatch(part[1], -1) {
				n += len(m[0])
				elem := []int{}
				for _, h := range strings.Split(m[2], ".") {
					elem = append(elem, convHex(i, h))
				}
				if p := elem[0]; m[1] == "*" {
					if p > maxVar {
						maxVar = p
					}
				} else if p > 0 && p < minNonVar {
					minNonVar = p
				}
				rhs = append(rhs, elem)
			}
			if len(part[1]) < n+3 || part[1][n+1] != '#' {
				log.Fatalf("%d: expected comment; found %s", i, part[1][n:])
			}
			builder.Add(lhs, rhs)
		}
	}
	if maxVar >= minNonVar {
		log.Fatalf("found maxVar > minNonVar (%d > %d)", maxVar, minNonVar)
	}
	return maxVar
}

func convHex(line int, s string) int {
	r, e := strconv.ParseInt(s, 16, 32)
	if e != nil {
		log.Fatalf("%d: %v", line, e)
	}
	return int(r)
}

// TODO: move this functionality to exp/locale/collate/build.
func printCollators(c *collate.Collator, vartop int) {
	const name = "Root"
	fmt.Printf("var _%s = Collator{\n", name)
	fmt.Printf("\tStrength: %v,\n", c.Strength)
	fmt.Printf("\tvariableTop: 0x%X,\n", vartop)
	fmt.Printf("\tf: norm.NFD,\n")
	fmt.Printf("\tt: &%sTable,\n", strings.ToLower(name))
	fmt.Printf("}\n\n")
	fmt.Printf("var (\n")
	fmt.Printf("\t%s = _%s\n", name, name)
	fmt.Printf(")\n\n")
}

func main() {
	flag.Parse()
	b := build.NewBuilder()
	vartop := parseUCA(b)
	_, err := b.Build("")
	failonerror(err)

	fmt.Println("// Generated by running")
	fmt.Printf("//  maketables --ducet=%s\n", *ducet)
	fmt.Println("// DO NOT EDIT")
	fmt.Println("// TODO: implement more compact representation for sparse blocks.")
	fmt.Println("")
	fmt.Println("package collate")
	fmt.Println("")
	fmt.Println(`import "exp/norm"`)
	fmt.Println("")

	c := &collate.Collator{}
	c.Strength = collate.Quaternary
	printCollators(c, vartop)

	_, err = b.Print(os.Stdout)
	failonerror(err)
}
