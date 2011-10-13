package main

import (
	"io/ioutil"
	"log"
	"path/filepath"
)

func init() {
	addTestCases(reflectTests(), reflectFn)
}

func reflectTests() []testCase {
	var tests []testCase

	names, _ := filepath.Glob("testdata/reflect.*.in")
	for _, in := range names {
		out := in[:len(in)-len(".in")] + ".out"
		inb, err := ioutil.ReadFile(in)
		if err != nil {
			log.Fatal(err)
		}
		outb, err := ioutil.ReadFile(out)
		if err != nil {
			log.Fatal(err)
		}
		tests = append(tests, testCase{Name: in, In: string(inb), Out: string(outb)})
	}

	return tests
}
