// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"reflect"
	"runtime"
	"runtime/ppapi"
	"sort"
	"time"
)

type testInstance struct{
	ppapi.Instance
	errors int
}

// Errorf writes an error message to the console and increments the error count.
func (inst *testInstance) Errorf(format string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	loc := ppapi.VarFromString(fmt.Sprintf("%s:%d", file, line))
	v := ppapi.VarFromString(fmt.Sprintf(format, args...))
	inst.LogWithSource(ppapi.PP_LOGLEVEL_ERROR, loc, v)
	loc.Release()
	v.Release()
	inst.errors++
}

type byValue []string

func (a byValue) Len() int           { return len(a) }
func (a byValue) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byValue) Less(i, j int) bool { return a[i] < a[j] }

// equalStringSets compares two sets of strings, represented as slices.
func equalStringSets(s1, s2 []string) bool {
	sort.Sort(byValue(s1))
	sort.Sort(byValue(s2))
	return reflect.DeepEqual(s1, s2)
}

// directoryEntryNames returns a slice of entry names.
func (inst *testInstance) directoryEntryNames(entries []ppapi.DirectoryEntry) []string {
	var names []string
	for _, entry := range entries {
		name, err := entry.File.Path()
		if err != nil {
			inst.Errorf("Path failed: %s", err)
		} else {
			names = append(names, name)
		}
	}
	return names
}

// TestFileSystem tests ppapi.FileSystem operations.
func (inst *testInstance) TestFileSystem() {
	// Create a filesystem.
	fs, err := inst.CreateFileSystem(ppapi.PP_FILESYSTEMTYPE_LOCALTEMPORARY)
	if err != nil {
		inst.Errorf("CreateFileSystem failed: %s", err)
		return
	}
	defer fs.Release()
	ty := fs.Type()
	if ty != ppapi.PP_FILESYSTEMTYPE_LOCALTEMPORARY {
		inst.Errorf("Unexpected filesystem type: %d", ty)
	}

	// Open the filesystem with expected size 64K.
	err = fs.OpenFS(1 << 16)
	if err != nil {
		inst.Errorf("Can't open filesystem: %s", err)
		return
	}

	// Create a directory.
	err = fs.MkdirAll("/tfs/a/b/c")
	if err != nil {
		inst.Errorf("Can't create directory: %s", err)
	}
	info, err := fs.Stat("/tfs/a/b/c")
	if err != nil {
		inst.Errorf("Stat failed: %s", err)
	}
	if !info.IsDir() {
		inst.Errorf("Should be a directory")
	}

	// Create some entries.
	fs.Mkdir("/tfs/a/b/c/d")
	fs.Mkdir("/tfs/a/b/c/e")

	// List the directory.
	names, err := fs.Readdirnames("/tfs/a/b/c")
	if err != nil {
		inst.Errorf("Readdirnames failed: %s", err)
	}
	expected := []string{"/tfs/a/b/c/d", "/tfs/a/b/c/e"}
	if !equalStringSets(expected, names) {
		inst.Errorf("Expected %v, got %v", expected, names)
	}

	// Rename.
	if err := fs.Rename("/tfs/a/b/c/e", "/tfs/a/b/c/f"); err != nil {
		inst.Errorf("Rename failed: %s", err)
	}

	// List the directory.
	names, err = fs.Readdirnames("/tfs/a/b/c")
	if err != nil {
		inst.Errorf("Readdirnames failed: %s", err)
	}
	expected = []string{"/tfs/a/b/c/d", "/tfs/a/b/c/f"}
	if !equalStringSets(expected, names) {
		inst.Errorf("Expected %v, got %v", expected, names)
	}

	// Remove and list.
	if err := fs.Remove("/tfs/a/b/c/d"); err != nil {
		inst.Errorf("Remove failed: %s", err)
	}
	names, err = fs.Readdirnames("/tfs/a/b/c")
	if err != nil {
		inst.Errorf("Readdirnames failed: %s", err)
	}
	expected = []string{"/tfs/a/b/c/f"}
	if !equalStringSets(expected, names) {
		inst.Errorf("Expected %v, got %v", expected, names)
	}

	// Stat.
	stat, err := fs.Stat("/tfs/a/b/c/f")
	if err != nil {
		inst.Errorf("Stat failed: %s", err)
	}
	atime := stat.ATime.Add(time.Second)
	mtime := stat.MTime.Add(100 * time.Second)
	if err := fs.Chtimes("/tfs/a/b/c/f", atime, mtime); err != nil {
		inst.Errorf("Chtimes failed: %s", err)
	}
	stat, err = fs.Stat("/tfs/a/b/c/f")
	if err != nil {
		inst.Errorf("Stat failed: %s", err)
	}
	if stat.ATime != atime {
		// TOTO(jyh): Figure out whether this is allowed.
		inst.Warningf("Expected atime %v, got %v", atime, stat.ATime)
	}
	if stat.MTime != mtime {
		inst.Errorf("Expected mtime %v, got %v", mtime, stat.MTime)
	}

	if err := fs.RemoveAll("/tfs/a"); err != nil {
		inst.Errorf("RemoveAll failed: %s", err)
	}
}

// TestFileRef tests ppapi.FileRef operations.
func (inst *testInstance) TestFileRef() {
	// Create a filesystem.
	fs, err := inst.CreateFileSystem(ppapi.PP_FILESYSTEMTYPE_LOCALTEMPORARY)
	if err != nil {
		inst.Errorf("CreateFileSystem failed: %s", err)
		return
	}
	defer fs.Release()

	// Open the filesystem with expected size 64K.
	err = fs.OpenFS(1 << 16)
	if err != nil {
		inst.Errorf("Can't open filesystem: %s", err)
		return
	}

	// Create a file reference.
	ref, err := fs.CreateFileRef("/tfr/a/b/c")
	if err != nil {
		inst.Errorf("Can't create file: %s", err)
		return
	}
	defer ref.Release()

	// FileRef naming methods.
	name, err := ref.Name()
	if err != nil {
		inst.Errorf("Can't get FileRef name: %s", err)
	}
	if name != "c" {
		inst.Errorf("Expected c, got %q", name)
	}
	path, err := ref.Path()
	if err != nil {
		inst.Errorf("Can't get FileRef path: %s", err)
	}
	if path != "/tfr/a/b/c" {
		inst.Errorf("Expected /tfr/a/b/c, got %q", path)
	}

	// FileRef Parent method.
	parent, err := ref.Parent()
	if err != nil {
		inst.Errorf("Can't get FileRef parent: %s", err)
		return
	}
	name, err = parent.Name()
	if err != nil {
		inst.Errorf("Can't get FileRef name: %s", err)
	}
	if name != "b" {
		inst.Errorf("Expected b, got %q", name)
	}
	path, err = parent.Path()
	if err != nil {
		inst.Errorf("Can't get FileRef path: %s", err)
	}
	if path != "/tfr/a/b" {
		inst.Errorf("Expected /tfr/a/b, got %q", path)
	}

	// Create the directory and check stat info.
	err = ref.MakeDirectory(ppapi.PP_MAKEDIRECTORYFLAG_WITH_ANCESTORS)
	if err != nil {
		inst.Errorf("MakeDirectory failed: %s", err)
	}
	info, err := ref.Stat()
	if err != nil {
		inst.Errorf("Stat failed: %s", err)
	}
	newATime := info.ATime.Add(100 * time.Second)
	newMTime := info.MTime.Add(1234 * time.Second)
	err = ref.Touch(newATime, newMTime)
	if err != nil {
		inst.Errorf("Touch failed: %s", err)
	}
	newInfo, err := ref.Stat()
	if err != nil {
		inst.Errorf("Stat failed: %s", err)
	}
	if newInfo.ATime != newATime {
		// TODO(jyh): This fails.  Is access time mutable?
		inst.Warningf("Expected atime %v, got %v", newATime, newInfo.ATime)
	}
	if newInfo.MTime != newMTime {
		inst.Errorf("Expected atime %v, got %v", newMTime, newInfo.MTime)
	}

	// Create some entries.
	fs.Mkdir("/tfr/a/b/c/d")
	fs.Mkdir("/tfr/a/b/c/e")

	// List the directory.
	entries, err := ref.ReadDirectoryEntries()
	if err != nil {
		inst.Errorf("ReadDirectoryEntries failed: %s", err)
	}
	expected := []string{"/tfr/a/b/c/d", "/tfr/a/b/c/e"}
	actual := inst.directoryEntryNames(entries)
	if !equalStringSets(expected, actual) {
		inst.Errorf("Expected %v, got %v", expected, actual)
	}

	// Rename.
	fs.Rename("/tfr/a/b/c/e", "/tfr/a/b/c/f")

	// List the directory.
	entries, err = ref.ReadDirectoryEntries()
	if err != nil {
		inst.Errorf("ReadDirectoryEntries failed: %s", err)
	}
	expected = []string{"/tfr/a/b/c/d", "/tfr/a/b/c/f"}
	actual = inst.directoryEntryNames(entries)
	if !equalStringSets(expected, actual) {
		inst.Errorf("Expected %v, got %v", expected, actual)
	}

	if err := fs.RemoveAll("/tfr/a"); err != nil {
		inst.Errorf("RemoveAll failed: %s", err)
	}
}

// TestFileIO tests ppapi.FileIO operations.
func (inst *testInstance) TestFileIO() {
	// Create a filesystem.
	fs, err := inst.CreateFileSystem(ppapi.PP_FILESYSTEMTYPE_LOCALTEMPORARY)
	if err != nil {
		inst.Errorf("CreateFileSystem failed: %s", err)
		return
	}
	defer fs.Release()

	// Open the filesystem with expected size 1M.
	if err := fs.OpenFS(1 << 20); err != nil {
		inst.Errorf("Can't open filesystem: %s", err)
		return
	}
	if err := fs.MkdirAll("/tfi/a/b/c"); err != nil {
		inst.Errorf("Mkdir failed: %s", err)
	}

	// Open a file.
	wfile, err := fs.Create("/tfi/a/b/c/file.txt")
	defer wfile.Release()
	if err != nil {
		inst.Errorf("Create failed: %s", err)
		return
	}
	if n, err := wfile.Write([]byte("Hello")); n != 5 || err != nil {
		inst.Errorf("Write failed: %d, %s", n, err)
	}
	if n, err := wfile.Write([]byte(" world\n")); n != 7 || err != nil {
		inst.Errorf("Write failed: %d, %s", n, err)
	}

	rfile, err := fs.Open("/tfi/a/b/c/file.txt")
	if err != nil {
		inst.Errorf("Open failed: %s", err)
		return
	}
	defer rfile.Release()

	var buf [100]byte
	n, err := rfile.Read(buf[:])
	if err != nil {
		inst.Errorf("Read failed: %d", err)
	}
	s := string(buf[:n])
	if s != "Hello world\n" {
		inst.Errorf("Garbled file: %q", s)
	}

	if _, err := wfile.WriteAt([]byte("abc"), 4); err != nil {
		inst.Errorf("Write failed: %s", err)
	}
	n, err = rfile.ReadAt(buf[:], 0)
	if err != nil {
		inst.Errorf("Read failed: %d", err)
	}
	s = string(buf[:n])
	if s != "Hellabcorld\n" {
		inst.Errorf("Garbled file: %q", s)
	}

	if _, err := wfile.WriteAt([]byte("ello World, Go as you are"), 1); err != nil {
		inst.Errorf("Write failed: %s", err)
	}
	n, err = rfile.ReadAt(buf[:], 0)
	if err != nil {
		inst.Errorf("Read failed: %d", err)
	}
	s = string(buf[:n])
	if s != "Hello World, Go as you are" {
		inst.Errorf("Garbled file: %q", s)
	}

	if err := wfile.Truncate(11); err != nil {
		inst.Errorf("Truncate failed: %s", err)
	}
	if n, err := rfile.Seek(2, 0); n != 2 || err != nil {
		inst.Errorf("Seek failed: %d, %s", n, err)
	}
	n, err = rfile.Read(buf[:])
	if err != nil {
		inst.Errorf("Read failed: %d", err)
	}
	s = string(buf[:n])
	if s != "llo World" {
		inst.Errorf("Garbled file: %q", s)
	}

	if n, err := wfile.Seek(0, 2); n != 11 || err != nil {
		inst.Errorf("Seek failed: %d, %s", n, err)
	}
	if _, err := wfile.Write([]byte(" ZYX")); err != nil {
		inst.Errorf("Write failed: %s", err)
	}
	n, err = wfile.ReadAt(buf[:], 0)
	if err != nil {
		inst.Errorf("Read failed: %d", err)
	}
	if n, err := wfile.Seek(0, 1); n != 15 || err != nil {
		inst.Errorf("Seek failed: %d, %s", n, err)
	}
	s = string(buf[:n])
	if s != "Hello World ZYX" {
		inst.Errorf("Garbled file: %q", s)
	}

	if err := fs.RemoveAll("/tfi/a"); err != nil {
		inst.Errorf("RemoveAll failed: %s", err)
	}
}

func (inst testInstance) RunAllTests() {
	inst.TestFileSystem()
	inst.TestFileRef()
	inst.TestFileIO()
	if inst.errors == 0 {
		inst.Printf("All tests passed")
	} else {
		inst.Errorf("Tests failed with %d errors", inst.errors)
	}
}

func (inst testInstance) DidCreate(args map[string]string) bool {
	go inst.RunAllTests()
	return true
}

func (testInstance) DidDestroy() {
}

func (testInstance) DidChangeView(view ppapi.View) {
}

func (testInstance) DidChangeFocus(has_focus bool) {
}

func (testInstance) HandleDocumentLoad(url_loader ppapi.Resource) bool {
	return true
}

func (testInstance) HandleInputEvent(event ppapi.InputEvent) bool {
	return true
}

func (testInstance) Graphics3DContextLost() {
}

func (testInstance) HandleMessage(message ppapi.Var) {
}

func (testInstance) MouseLockLost() {
}

func newTestInstance(inst ppapi.Instance) ppapi.InstanceHandlers {
	return testInstance{Instance: inst}
}

func main() {
	ppapi.Init(newTestInstance)
}
