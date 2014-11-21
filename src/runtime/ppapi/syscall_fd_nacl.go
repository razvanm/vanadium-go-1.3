package ppapi

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

// files is the table indexed by file descriptor.
var files struct {
	sync.RWMutex
	tab []*file
}

// A file is an open file, something with a file descriptor.
// A particular *file may appear in the files table multiple times, due to use of Dup or Dup2.
type file struct {
	fdref int      // used in files.tab
	impl  fileImpl // underlying implementation
}

// A fileImpl is the implementation of something that can be a file.
type fileImpl interface {
	// Standard operations.
	// These can be called concurrently from multiple goroutines.
	stat(*syscall.Stat_t) error
	read([]byte) (int, error)
	write([]byte) (int, error)
	seek(int64, int) (int64, error)
	pread([]byte, int64) (int, error)
	pwrite([]byte, int64) (int, error)

	// Close is called when the last reference to a *file is removed
	// from the file descriptor table. It may be called concurrently
	// with active operations such as blocked read or write calls.
	close() error
}

// newFD adds impl to the file descriptor table,
// returning the new file descriptor.
// Like Unix, it uses the lowest available descriptor.
func newFD(impl fileImpl) int {
	files.Lock()
	defer files.Unlock()
	f := &file{impl: impl, fdref: 1}
	for fd, oldf := range files.tab {
		// Look if there is a space for a new file
		if oldf == nil {
			files.tab[fd] = f
			return fd
		}
	}
	fd := len(files.tab)
	files.tab = append(files.tab, f)
	return fd
}

var stdin = &defaultFileImpl{}
var stdout = &consoleLogFile{logLevel: PP_LOGLEVEL_LOG}
var stderr = &consoleLogFile{logLevel: PP_LOGLEVEL_WARNING}

// Install Native Client stdin, stdout, stderr.
func init_fds() {
	if len(files.tab) != 0 {
		return
	}
	newFD(stdin)
	newFD(stdout)
	newFD(stderr)
}

func fdToFileRep(fd int) (*file, error) {
	files.Lock()
	defer files.Unlock()
	if fd < 0 || fd >= len(files.tab) || files.tab[fd] == nil {
		return nil, syscall.EBADF
	}
	return files.tab[fd], nil
}

// fdToFile retrieves the *file corresponding to a file descriptor.
func fdToFile(fd int) (fileImpl, error) {
	file, err := fdToFileRep(fd)
	if err != nil {
		return nil, err
	}
	return file.impl, nil
}

var symLinks map[string]string
var tmpFiles map[string]*bytesBufFileData

func initSymlinks() {
	if symLinks == nil {
		symLinks = map[string]string{}
	}
}

func initTmpFiles() {
	if tmpFiles == nil {
		tmpFiles = map[string]*bytesBufFileData{}
	}
}

func resolveSymlink(path string) string {
	initSymlinks()
	seen := map[string]bool{}
	for {
		if _, ok := seen[path]; ok {
			panic(fmt.Sprintf("Infinite loop in symlink %s %v", path, seen))
		}
		seen[path] = true
		if nextPath, ok := symLinks[path]; ok {
			path = nextPath
		} else {
			break
		}
	}
	return path
}

func (PPAPISyscallImpl) Open(path string, mode int, perm uint32) (fd int, err error) {
	path = resolveSymlink(path)
	//fmt.Printf("Opening: \"%s\" (resolved to \"%s\") %d %d\n", origPath, path, mode, perm)
	// TODO(bprosnitz) Handle modes better
	initEnvVars()
	if strings.HasPrefix(path, envVars["TMPDIR"]) {
		initTmpFiles()
		data, ok := tmpFiles[path]
		if !ok {
			data = newByteBufFileData(path)
			tmpFiles[path] = data
		}
		bbFile := &bytesBufFile{
			dat:   data,
			index: 0,
		}
		return newFD(bbFile), nil
	}
	if strings.HasPrefix(path, "/usr/local/google/home") {
		return 0, fmt.Errorf("File not found: %s", path)
	}
	switch path {
	case "/etc/localtime":
		data, err := base64.StdEncoding.DecodeString(base64_localtime)
		if err != nil {
			panic(fmt.Sprintf("Error decoding base64: %v", err))
		}
		return newFD(&bytesReadFile{data, 0, sync.Mutex{}}), nil
	case "/dev/urandom":
		return newFD(&randomImpl{}), nil
	default:
		panic(fmt.Sprintf("Open() not implemented. Path: %s", path))
	}
}

func (PPAPISyscallImpl) Close(fd int) error {
	files.Lock()
	if fd < 0 || fd >= len(files.tab) || files.tab[fd] == nil {
		files.Unlock()
		return syscall.EBADF
	}
	f := files.tab[fd]
	files.tab[fd] = nil
	f.fdref--
	fdref := f.fdref
	if fdref > 0 {
		files.Unlock()
		panic("Shouldn't get here until Dup/Dup2 is implemented")
		return nil
	}
	files.Unlock()
	return f.impl.close()
}

func (PPAPISyscallImpl) CloseOnExec(fd int) {
	// nothing to do - no exec
}

func (PPAPISyscallImpl) Dup(fd int) (int, error) {
	panic("Dup not yet uncommented")
	/*
	   files.Lock()
	   defer files.Unlock()
	   if fd < 0 || fd >= len(files.tab) || files.tab[fd] == nil {
	       return -1, EBADF
	   }
	   f := files.tab[fd]
	   f.fdref++
	   for newfd, oldf := range files.tab {
	       if oldf == nil {
	           files.tab[newfd] = f
	           return newfd, nil
	       }
	   }
	   newfd := len(files.tab)
	   files.tab = append(files.tab, f)
	   return newfd, nil*/
}

func (PPAPISyscallImpl) Dup2(fd, newfd int) error {
	panic("Dup2 not yet uncommented")
	/*
	   files.Lock()
	   defer files.Unlock()
	   if fd < 0 || fd >= len(files.tab) || files.tab[fd] == nil || newfd < 0 || newfd >= len(files.tab)+100 {
	       files.Unlock()
	       return EBADF
	   }
	   f := files.tab[fd]
	   f.fdref++
	   for cap(files.tab) <= newfd {
	       files.tab = append(files.tab[:cap(files.tab)], nil)
	   }
	   oldf := files.tab[newfd]
	   var oldfdref int
	   if oldf != nil {
	       oldf.fdref--
	       oldfdref = oldf.fdref
	   }
	   files.tab[newfd] = f
	   files.Unlock()
	   if oldf != nil {
	       if oldfdref == 0 {
	           oldf.impl.close()
	       }
	   }
	   return nil*/
}

var dirMade map[string]bool

func (PPAPISyscallImpl) Stat(path string, stat *syscall.Stat_t) (err error) {
	switch path {
	case "/tmp/revocation_dir/caveat_dir", "/tmp/revocation_dir", "/tmp/revocation_dir/revocation_dir":
		if dirMade != nil {
			if _, ok := dirMade[path]; ok {
				stat.Mode = syscall.S_IFDIR
				return nil
			}
		}
		return fmt.Errorf("path %s does not exist", path)
	case "/tmp":
		stat.Mode = syscall.S_IFDIR
		return nil
	default:
		panic(fmt.Sprintf("Stat() not implemented for %s", path))
	}
}

func (PPAPISyscallImpl) Mkdir(path string, mode uint32) (err error) {
	switch path {
	case "/tmp/revocation_dir/caveat_dir", "/tmp/revocation_dir/revocation_dir", "/tmp/revocation_dir":
		if dirMade == nil {
			dirMade = map[string]bool{}
		}
		dirMade[path] = true
		return nil

	default:
		panic(fmt.Sprintf("Mkdir() not implemented for %s", path))
	}
}

func (PPAPISyscallImpl) Fstat(fd int, st *syscall.Stat_t) error {
	panic("Fstat not yet uncommented")
	/*
	   f, err := fdToFile(fd)
	   if err != nil {
	       return err
	   }
	   return f.impl.stat(st)*/
}

func (PPAPISyscallImpl) Read(fd int, b []byte) (int, error) {
	f, err := fdToFile(fd)
	if err != nil {
		return 0, err
	}
	return f.read(b)
}

func (PPAPISyscallImpl) Write(fd int, b []byte) (int, error) {
	f, err := fdToFile(fd)
	if err != nil {
		return 0, err
	}
	return f.write(b)
}

func (PPAPISyscallImpl) Pread(fd int, b []byte, offset int64) (int, error) {
	panic("Pread not yet uncommented")
	/*
	   f, err := fdToFile(fd)
	   if err != nil {
	       return 0, err
	   }
	   return f.impl.pread(b, offset)*/
}

func (PPAPISyscallImpl) Pwrite(fd int, b []byte, offset int64) (int, error) {
	panic("Pwrite not yet uncommented")
	/*
	   f, err := fdToFile(fd)
	   if err != nil {
	       return 0, err
	   }
	   return f.impl.pwrite(b, offset)*/
}

func (PPAPISyscallImpl) Seek(fd int, offset int64, whence int) (int64, error) {
	panic("Seek not yet uncommented")
	/*f, err := fdToFile(fd)
	  if err != nil {
	      return 0, err
	  }
	  return f.impl.seek(offset, whence)*/
}

func (PPAPISyscallImpl) Unlink(path string) (err error) {
	//fmt.Printf("Unlinking: %s Links: %v\n", path, symLinks)
	initSymlinks()
	if _, ok := symLinks[path]; ok {
		delete(symLinks, path)
		return nil
	}
	return fmt.Errorf("File %s not a symlink", path)
}

func (PPAPISyscallImpl) Rmdir(path string) (err error) {
	//fmt.Printf("Attempting to remove dir: %s\n", path)
	path = resolveSymlink(path)
	switch path {
	case "/tmp/NaClMain.FATAL", "/tmp/NaClMain.ERROR", "/tmp/NaClMain.WARNING", "/tmp/NaClMain.INFO":
		return fmt.Errorf("Dir %s not found", path)
	default:
		panic(fmt.Sprintf("Rmdir() %s not implemented]", path))
	}
}

func (PPAPISyscallImpl) Symlink(oldpath string, newpath string) (err error) {
	//fmt.Printf("Symlinked %s to %s\n", newpath, oldpath)
	initSymlinks()
	symLinks[newpath] = oldpath
	return nil
}

func (PPAPISyscallImpl) Fsync(fd int) (err error) {
	// This is a no-op because everything is in memory right now.
	// Implement if this changes.
	return err
}

type consoleLogFile struct {
	impl     PPAPISyscallImpl
	logLevel LogLevel
}

func (*consoleLogFile) close() error { return nil }
func (*consoleLogFile) stat(*syscall.Stat_t) error {
	panic("stat not implemented")
}
func (*consoleLogFile) read([]byte) (int, error) {
	panic("Cannot read from log file")
}
func (c *consoleLogFile) writeLine(b []byte) (int, error) {
	_, file, line, _ := runtime.Caller(3)
	loc := file + ":" + strconv.Itoa(line)
	// Unfortunately nacl truncates logs at 128 chars.
	batchSize := 128
	for i := 0; i*batchSize < len(b); i++ {
		min := i * batchSize
		max := (i + 1) * batchSize
		if max > len(b) {
			max = len(b)
		}
		c.impl.Instance.LogWithSourceString(c.logLevel, loc, string(b[min:max]))
	}
	return len(b), nil
}
func (c *consoleLogFile) write(b []byte) (int, error) {
	s := string(b)
	parts := strings.Split(s, "\n")
	written := len(parts) - 1 // newlines
	for _, part := range parts {
		additionalWrite, err := c.writeLine([]byte(part))
		if err != nil {
			return 0, err
		}
		written += additionalWrite
	}
	return written, nil
}
func (*consoleLogFile) seek(int64, int) (int64, error) {
	panic("Cannot seek in log file.")
}
func (c *consoleLogFile) pread(b []byte, offset int64) (int, error) {
	return c.read(b[offset:])
}
func (c *consoleLogFile) pwrite(b []byte, offset int64) (int, error) {
	return c.write(b[offset:])
}

// defaulFileImpl imlements fileImpl.
// It can be embedded to complete a partial fileImpl implementation.
type defaultFileImpl struct{}

func (*defaultFileImpl) close() error                      { return nil }
func (*defaultFileImpl) stat(*syscall.Stat_t) error        { return syscall.ENOSYS }
func (*defaultFileImpl) read([]byte) (int, error)          { return 0, syscall.ENOSYS }
func (*defaultFileImpl) write([]byte) (int, error)         { return 0, syscall.ENOSYS }
func (*defaultFileImpl) seek(int64, int) (int64, error)    { return 0, syscall.ENOSYS }
func (*defaultFileImpl) pread([]byte, int64) (int, error)  { return 0, syscall.ENOSYS }
func (*defaultFileImpl) pwrite([]byte, int64) (int, error) { return 0, syscall.ENOSYS }

type randomImpl struct{}

func (*randomImpl) close() error               { return nil }
func (*randomImpl) stat(*syscall.Stat_t) error { return syscall.ENOSYS }
func (*randomImpl) read(b []byte) (int, error) {
	// TODO(bprosnitz) Make cryptographically secure?
	for i, _ := range b {
		b[i] = byte(rand.Uint32())
	}
	return len(b), nil
}
func (*randomImpl) write([]byte) (int, error)      { return 0, syscall.ENOSYS }
func (*randomImpl) seek(int64, int) (int64, error) { return 0, syscall.ENOSYS }
func (r *randomImpl) pread(b []byte, offset int64) (int, error) {
	return r.read(b[offset:])
}
func (*randomImpl) pwrite([]byte, int64) (int, error) { return 0, syscall.ENOSYS }

type bytesReadFile struct {
	bytes     []byte
	indexRead int
	lock      sync.Mutex
}

func (*bytesReadFile) close() error { return nil }
func (*bytesReadFile) stat(*syscall.Stat_t) error {
	panic("Stat not implemented")
}
func (brf *bytesReadFile) read(b []byte) (int, error) {
	brf.lock.Lock()
	amt := copy(b, brf.bytes[brf.indexRead:])
	brf.indexRead += amt
	brf.lock.Unlock()
	return amt, nil
}
func (*bytesReadFile) write([]byte) (int, error) {
	panic("Cannot write to a read file")
}
func (*bytesReadFile) seek(int64, int) (int64, error) {
	panic("Seek not implemented")
}
func (brf *bytesReadFile) pread(b []byte, offset int64) (int, error) {
	return brf.read(b[offset:])
}
func (*bytesReadFile) pwrite([]byte, int64) (int, error) {
	panic("Cannot write to a read file")
}

type bytesBufFileData struct {
	bytes []byte
	lock  sync.Mutex
	path  string // This is just for debugging, can be removed.
}

type bytesBufFile struct {
	dat   *bytesBufFileData
	index int
}

func newByteBufFileData(path string) *bytesBufFileData {
	return &bytesBufFileData{
		bytes: []byte{},
		path:  path,
	}
}
func (bbf *bytesBufFile) close() error {
	bbf.dat = nil
	return nil
}
func (*bytesBufFile) stat(*syscall.Stat_t) error {
	panic("Stat not implemented")
}
func (bbf *bytesBufFile) read(b []byte) (int, error) {
	if bbf.dat == nil {
		panic("Cannot read closed file")
	}
	bbf.dat.lock.Lock()
	amt := copy(b, bbf.dat.bytes[bbf.index:])
	bbf.index += amt
	bbf.dat.lock.Unlock()
	return amt, nil
}
func (bbf *bytesBufFile) write(b []byte) (int, error) {
	if bbf.dat == nil {
		panic("Cannot write to closed file")
	}
	bbf.dat.lock.Lock()
	neededSize := bbf.index + len(b) + 1
	if neededSize >= len(bbf.dat.bytes) {
		newBuf := make([]byte, neededSize)
		copy(newBuf, bbf.dat.bytes)
		bbf.dat.bytes = newBuf
	}
	fmt.Printf("temp file %s is now %d bytes long", bbf.dat.path, len(bbf.dat.bytes))
	if copy(bbf.dat.bytes[bbf.index:], b) != len(b) {
		panic("Invalid copy during write")
	}
	bbf.index += len(b)
	bbf.dat.lock.Unlock()
	return len(b), nil
}
func (*bytesBufFile) seek(int64, int) (int64, error) {
	panic("Seek not implemented")
}
func (bbf *bytesBufFile) pread(b []byte, offset int64) (int, error) {
	return bbf.read(b[offset:])
}
func (bbf *bytesBufFile) pwrite(b []byte, offset int64) (int, error) {
	return bbf.write(b[offset:])
}

const base64_localtime = "VFppZjIAAAAAAAAAAAAAAAAAAAAAAAAEAAAABAAAAAAAAAC5AAAABAAAABCepkig" +
	"n7sVkKCGKqChmveQy4kaoNIj9HDSYSYQ1v50INiArZDa/tGg28CQENzes6DdqayQ" +
	"3r6VoN+JjpDgnneg4WlwkOJ+WaDjSVKQ5F47oOUpNJDmR1gg5xJREOgnOiDo8jMQ" +
	"6gccIOrSFRDr5v4g7LH3EO3G4CDukdkQ76/8oPBxuxDxj96g8n/BkPNvwKD0X6OQ" +
	"9U+ioPY/hZD3L4Sg+CiiEPkPZqD6CIQQ+viDIPvoZhD82GUg/chIEP64RyD/qCoQ" +
	"AJgpIAGIDBACeAsgA3EokARhJ6AFUQqQBkEJoAcw7JAHjUOgCRDOkAmtvyAK8LCQ" +
	"C+CvoAzZzRANwJGgDrmvEA+priAQmZEQEYmQIBJ5cxATaXIgFFlVEBVJVCAWOTcQ" +
	"Fyk2IBgiU5AZCRggGgI1kBryNKAb4heQHNIWoB3B+ZAesfigH6HbkCB2KyAhgb2Q" +
	"IlYNICNq2hAkNe8gJUq8ECYV0SAnKp4QJ/7toCkKgBAp3s+gKupiECu+saAs036Q" +
	"LZ6ToC6zYJAvfnWgMJNCkDFnkiAycySQM0d0IDRTBpA1J1YgNjLokDcHOCA4HAUQ" +
	"OOcaIDn75xA6xvwgO9vJEDywGKA9u6sQPo/6oD+bjRBAb9ygQYSpkEJPvqBDZIuQ" +
	"RC+goEVEbZBF89MgRy2KEEfTtSBJDWwQSbOXIErtThBLnLOgTNZqkE18laBOtkyQ" +
	"T1x3oFCWLpBRPFmgUnYQkFMcO6BUVfKQVPwdoFY11JBW5TogWB7xEFjFHCBZ/tMQ" +
	"WqT+IFvetRBchOAgXb6XEF5kwiBfnnkQYE3eoGGHlZBiLcCgY2d3kGQNoqBlR1mQ" +
	"Ze2EoGcnO5BnzWagaQcdkGmtSKBq5v+Qa5ZlIGzQHBBtdkcgbq/+EG9WKSBwj+AQ" +
	"cTYLIHJvwhBzFe0gdE+kEHT/CaB2OMCQdt7roHgYopB4vs2gefiEkHqer6B72GaQ" +
	"fH6RoH24SJB+XnOgf5gqkAABAAECAwEAAQABAAEAAQABAAEAAQABAAEAAQABAAEA" +
	"AQABAAEAAQABAAEAAQABAAEAAQABAAEAAQABAAEAAQABAAEAAQABAAEAAQABAAEA" +
	"AQABAAEAAQABAAEAAQABAAEAAQABAAEAAQABAAEAAQABAAEAAQABAAEAAQABAAEA" +
	"AQABAAEAAQABAAEAAQABAAEAAQABAAEAAQABAAEAAQABAAEAAQABAAEAAQABAAEA" +
	"AQABAAEAAQAB//+dkAEA//+PgAAE//+dkAEI//+dkAEMUERUAFBTVABQV1QAUFBU" +
	"AAAAAAEAAAABVFppZjIAAAAAAAAAAAAAAAAAAAAAAAAFAAAABQAAAAAAAAC6AAAA" +
	"BQAAABT/////XgQawP////+epkig/////5+7FZD/////oIYqoP////+hmveQ////" +
	"/8uJGqD/////0iP0cP/////SYSYQ/////9b+dCD/////2ICtkP/////a/tGg////" +
	"/9vAkBD/////3N6zoP/////dqayQ/////96+laD/////34mOkP/////gnneg////" +
	"/+FpcJD/////4n5ZoP/////jSVKQ/////+ReO6D/////5Sk0kP/////mR1gg////" +
	"/+cSURD/////6Cc6IP/////o8jMQ/////+oHHCD/////6tIVEP/////r5v4g////" +
	"/+yx9xD/////7cbgIP/////ukdkQ/////++v/KD/////8HG7EP/////xj96g////" +
	"//J/wZD/////82/AoP/////0X6OQ//////VPoqD/////9j+FkP/////3L4Sg////" +
	"//goohD/////+Q9moP/////6CIQQ//////r4gyD/////++hmEP/////82GUg////" +
	"//3ISBD//////rhHIP//////qCoQAAAAAACYKSAAAAAAAYgMEAAAAAACeAsgAAAA" +
	"AANxKJAAAAAABGEnoAAAAAAFUQqQAAAAAAZBCaAAAAAABzDskAAAAAAHjUOgAAAA" +
	"AAkQzpAAAAAACa2/IAAAAAAK8LCQAAAAAAvgr6AAAAAADNnNEAAAAAANwJGgAAAA" +
	"AA65rxAAAAAAD6muIAAAAAAQmZEQAAAAABGJkCAAAAAAEnlzEAAAAAATaXIgAAAA" +
	"ABRZVRAAAAAAFUlUIAAAAAAWOTcQAAAAABcpNiAAAAAAGCJTkAAAAAAZCRggAAAA" +
	"ABoCNZAAAAAAGvI0oAAAAAAb4heQAAAAABzSFqAAAAAAHcH5kAAAAAAesfigAAAA" +
	"AB+h25AAAAAAIHYrIAAAAAAhgb2QAAAAACJWDSAAAAAAI2raEAAAAAAkNe8gAAAA" +
	"ACVKvBAAAAAAJhXRIAAAAAAnKp4QAAAAACf+7aAAAAAAKQqAEAAAAAAp3s+gAAAA" +
	"ACrqYhAAAAAAK76xoAAAAAAs036QAAAAAC2ek6AAAAAALrNgkAAAAAAvfnWgAAAA" +
	"ADCTQpAAAAAAMWeSIAAAAAAycySQAAAAADNHdCAAAAAANFMGkAAAAAA1J1YgAAAA" +
	"ADYy6JAAAAAANwc4IAAAAAA4HAUQAAAAADjnGiAAAAAAOfvnEAAAAAA6xvwgAAAA" +
	"ADvbyRAAAAAAPLAYoAAAAAA9u6sQAAAAAD6P+qAAAAAAP5uNEAAAAABAb9ygAAAA" +
	"AEGEqZAAAAAAQk++oAAAAABDZIuQAAAAAEQvoKAAAAAARURtkAAAAABF89MgAAAA" +
	"AEctihAAAAAAR9O1IAAAAABJDWwQAAAAAEmzlyAAAAAASu1OEAAAAABLnLOgAAAA" +
	"AEzWapAAAAAATXyVoAAAAABOtkyQAAAAAE9cd6AAAAAAUJYukAAAAABRPFmgAAAA" +
	"AFJ2EJAAAAAAUxw7oAAAAABUVfKQAAAAAFT8HaAAAAAAVjXUkAAAAABW5TogAAAA" +
	"AFge8RAAAAAAWMUcIAAAAABZ/tMQAAAAAFqk/iAAAAAAW961EAAAAABchOAgAAAA" +
	"AF2+lxAAAAAAXmTCIAAAAABfnnkQAAAAAGBN3qAAAAAAYYeVkAAAAABiLcCgAAAA" +
	"AGNnd5AAAAAAZA2ioAAAAABlR1mQAAAAAGXthKAAAAAAZyc7kAAAAABnzWagAAAA" +
	"AGkHHZAAAAAAaa1IoAAAAABq5v+QAAAAAGuWZSAAAAAAbNAcEAAAAABtdkcgAAAA" +
	"AG6v/hAAAAAAb1YpIAAAAABwj+AQAAAAAHE2CyAAAAAAcm/CEAAAAABzFe0gAAAA" +
	"AHRPpBAAAAAAdP8JoAAAAAB2OMCQAAAAAHbe66AAAAAAeBiikAAAAAB4vs2gAAAA" +
	"AHn4hJAAAAAAep6voAAAAAB72GaQAAAAAHx+kaAAAAAAfbhIkAAAAAB+XnOgAAAA" +
	"AH+YKpACAQIBAgMEAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIB" +
	"AgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIB" +
	"AgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIB" +
	"AgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQL/" +
	"/5EmAAD//52QAQT//4+AAAj//52QAQz//52QARBMTVQAUERUAFBTVABQV1QAUFBU" +
	"AAAAAAABAAAAAAEKUFNUOFBEVCxNMy4yLjAsTTExLjEuMAo="
