package ppapi

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"
	"unsafe"
	"log"
)

var (
	errFileIOCreateFailed     = errors.New("FileIO creation failed")
	errFileSystemCreateFailed = errors.New("filesystem creation failed")
	errFileRefParentFailed    = errors.New("FileRef.Parent failed")
	errNegativeFilePosition   = errors.New("negative file position")
)

// FileSystem specifies the file system type associated with a file.  For
// example, the filesystem specifies whether a file is persistent or temporary.
type FileSystem struct {
	Resource
	instance Instance
}

// Type returns the type of the file system.
func (fs FileSystem) Type() FileSystemType {
	return ppb_filesystem_get_type(fs.id)
}

// OpenFS opens the file system.
//
// A file system must be opened before running any other operation on it.
//
// Note that this does not request quota; to do that, you must either invoke
// requestQuota from JavaScript:
// http://www.html5rocks.com/en/tutorials/file/filesystem/#toc-requesting-quota
// or set the unlimitedStorage permission for Chrome Web Store apps:
// http://code.google.com/chrome/extensions/manifest.html#permissions.
func (fs FileSystem) OpenFS(expectedSize int64) error {
	code := ppb_filesystem_open(fs.id, expectedSize, ppNullCompletionCallback)
	return decodeError(Error(code))
}

// Remove deletes a file or directory.
//
// If the name refers to a directory, then the directory must be empty. It is an
// error to delete a file or directory that is in use. It is not valid to delete
// a file in the external file system.
func (fs FileSystem) Remove(name string) error {
	ref, err := fs.CreateFileRef(name)
	if err != nil {
		return &os.PathError{Op: "Remove", Path: name, Err: err}
	}
	defer ref.Release()
	if err := ref.Delete(); err != nil {
		return &os.PathError{Op: "Remove", Path: name, Err: err}
	}
	return nil
}

// RemoveAll removes path and any children it contains. It removes everything it
// can but returns the first error it encounters. If the path does not exist,
// RemoveAll returns nil (no error).
func (fs FileSystem) RemoveAll(name string) error {
	ref, err := fs.CreateFileRef(name)
	if err != nil {
		return &os.PathError{Op: "RemoveAll", Path: name, Err: err}
	}
	defer ref.Release()

	// Stat the file.
	info, err := ref.Stat()
	if err != nil {
		if err == ppErrors[PP_ERROR_FILENOTFOUND] {
			return nil
		}
		return &os.PathError{Op: "RemoveAll", Path: name, Err: err}
	}

	// If it is a directory, delete the entries recursively.
	if info.IsDir() {
		names, err := fs.Readdirnames(name)
		if err != nil {
			return err
		}
		for _, name := range names {
			if err := fs.RemoveAll(name); err != nil {
				return err
			}
		}
	}

	// Remove the file.
	if err := ref.Delete(); err != nil {
		return &os.PathError{Op: "RemoveAll", Path: name, Err: err}
	}
	return nil
}

// Rename renames a file.
//
// It is an error to rename a file or directory that is in use. It is not valid
// to rename a file in the external file system.
func (fs FileSystem) Rename(fromName, toName string) error {
	fromRef, err := fs.CreateFileRef(fromName)
	if err != nil {
		return err
	}
	defer fromRef.Release()

	toRef, err := fs.CreateFileRef(toName)
	if err != nil {
		return err
	}
	defer toRef.Release()

	return fromRef.Rename(toRef)
}

// Mkdir creates a new directory with the specified name and permission bits. If
// there is an error, it will be of type *PathError.
//
// It is not valid to make a directory in the external file system.
func (fs FileSystem) Mkdir(name string) error {
	ref, err := fs.CreateFileRef(name)
	if err != nil {
		return &os.PathError{Op: "Mkdir", Path: name, Err: err}
	}
	defer ref.Release()

	if err := ref.MakeDirectory(PP_MAKEDIRECTORYFLAG_NONE); err != nil {
		return &os.PathError{Op: "Mkdir", Path: name, Err: err}
	}
	return nil
}

// MkdirAll creates a directory named path, along with any necessary parents,
// and returns nil, or else returns an error. The permission bits perm are used
// for all directories that MkdirAll creates. If path is already a directory,
// MkdirAll does nothing and returns nil.
func (fs FileSystem) MkdirAll(name string) error {
	ref, err := fs.CreateFileRef(name)
	if err != nil {
		return &os.PathError{Op: "Mkdir", Path: name, Err: err}
	}
	defer ref.Release()

	if err := ref.MakeDirectory(PP_MAKEDIRECTORYFLAG_WITH_ANCESTORS); err != nil {
		return &os.PathError{Op: "Mkdir", Path: name, Err: err}
	}
	return nil
}

// Readdirnames reads and returns a slice of names from the directory f.
func (fs FileSystem) Readdirnames(name string) ([]string, error) {
	ref, err := fs.CreateFileRef(name)
	if err != nil {
		return nil, &os.PathError{Op: "Readdirnames", Path: name, Err: err}
	}
	defer ref.Release()

	entries, err := ref.ReadDirectoryEntries()
	if err != nil {
		return nil, &os.PathError{Op: "Readdirnames", Path: name, Err: err}
	}

	names := make([]string, len(entries))
	for i, entry := range entries {
		name, err := entry.File.Path()
		if err != nil {
			return nil, &os.PathError{Op: "Readdirnames", Path: name, Err: err}
		}
		names[i] = name
	}
	return names, nil
}

// Stat queries info about a file or directory.
//
// You must have access to read this file or directory if it exists in the
// external filesystem.
func (fs FileSystem) Stat(name string) (info FileInfo, err error) {
	ref, e := fs.CreateFileRef(name)
	if e != nil {
		err = &os.PathError{Op: "Stat", Path: name, Err: e}
		return
	}
	defer ref.Release()

	info, e = ref.Stat()
	if e != nil {
		err = &os.PathError{Op: "Stat", Path: name, Err: e}
		return
	}
	return
}

// Chtimes changes the access and modification times of the named file, similar
// to the Unix utime() or utimes() functions.
//
// The underlying filesystem may truncate or round the values to a less precise
// time unit. If there is an error, it will be of type *PathError.
func (fs FileSystem) Chtimes(name string, atime, mtime time.Time) error {
	ref, err := fs.CreateFileRef(name)
	if err != nil {
		return &os.PathError{Op: "Chtimes", Path: name, Err: err}
	}
	defer ref.Release()
	if err := ref.Touch(atime, mtime); err != nil {
		return &os.PathError{Op: "Chtimes", Path: name, Err: err}
	}
	return nil
}

// Open opens the named file for reading. If successful, methods on the returned
// file can be used for reading; the associated file has mode O_RDONLY. If there
// is an error, it will be of type *PathError.
func (fs FileSystem) Open(name string) (*FileIO, error) {
	return fs.OpenFile(name, os.O_RDONLY)
}

// Create creates the named file mode 0666 (before umask), truncating it if it
// already exists. If successful, methods on the returned File can be used for
// I/O; the associated file has mode O_RDWR. If there is an error, it will be of
// type *PathError.
func (fs FileSystem) Create(name string) (*FileIO, error) {
	return fs.OpenFile(name, os.O_RDWR|os.O_TRUNC|os.O_CREATE)
}

// OpenFile is the generalized open call; most users will use Open or Create
// instead. It opens the named file with specified flag (O_RDONLY etc.) and
// perm, (0666 etc.) if applicable. If successful, methods on the returned File
// can be used for I/O. If there is an error, it will be of type *PathError.
func (fs FileSystem) OpenFile(name string, flag int) (*FileIO, error) {
	ref, err := fs.CreateFileRef(name)
	if err != nil {
		return nil, &os.PathError{Op: "CreateFileRef", Path: name, Err: err}
	}
	defer ref.Release()

	file, err := fs.instance.CreateFileIO()
	if err != nil {
		return nil, &os.PathError{Op: "CreateFileIO", Path: name, Err: err}
	}

	var pflag FileOpenFlag
	if flag&os.O_RDONLY != 0 {
		pflag |= PP_FILEOPENFLAG_READ
	}
	if flag&os.O_WRONLY != 0 {
		pflag |= PP_FILEOPENFLAG_WRITE
	}
	if flag&os.O_RDWR != 0 {
		pflag |= PP_FILEOPENFLAG_READ | PP_FILEOPENFLAG_WRITE
	}
	if flag&os.O_CREATE != 0 {
		pflag |= PP_FILEOPENFLAG_CREATE
	}
	if flag&os.O_EXCL != 0 {
		pflag |= PP_FILEOPENFLAG_EXCLUSIVE
	}
	if flag&os.O_TRUNC != 0 {
		pflag |= PP_FILEOPENFLAG_TRUNCATE
	}
	if flag&os.O_APPEND != 0 {
		pflag |= PP_FILEOPENFLAG_APPEND
	}
	if err := file.Open(ref, pflag); err != nil {
		file.Release()
		return nil, &os.PathError{Op: "Open", Path: name, Err: err}
	}
	return &file, nil
}

// FileInfo represents information about a file, such as size, type, and
// creation time.
//
// Implements os.FileInfo.
type FileInfo struct {
	Filename            string
	Len                 int64
	Type                FileType
	FSType              FileSystemType
	CTime, ATime, MTime time.Time
}

var _ os.FileInfo = &FileInfo{}

func (in FileInfo) toPP(out *pp_FileInfo) {
	*(*int64)(unsafe.Pointer(&out[0])) = in.Len
	*(*FileType)(unsafe.Pointer(&out[8])) = in.Type
	*(*FileSystemType)(unsafe.Pointer(&out[12])) = in.FSType
	*(*pp_Time)(unsafe.Pointer(&out[16])) = toPPTime(in.CTime)
	*(*pp_Time)(unsafe.Pointer(&out[24])) = toPPTime(in.ATime)
	*(*pp_Time)(unsafe.Pointer(&out[32])) = toPPTime(in.MTime)
}

func (out *FileInfo) fromPP(name string, in pp_FileInfo) {
	out.Filename = name
	out.Len = *(*int64)(unsafe.Pointer(&in[0]))
	out.Type = *(*FileType)(unsafe.Pointer(&in[8]))
	out.FSType = *(*FileSystemType)(unsafe.Pointer(&in[12]))
	out.CTime = fromPPTime(*(*pp_Time)(unsafe.Pointer(&in[16])))
	out.ATime = fromPPTime(*(*pp_Time)(unsafe.Pointer(&in[24])))
	out.MTime = fromPPTime(*(*pp_Time)(unsafe.Pointer(&in[32])))
}

func (info *FileInfo) Name() string {
	return info.Filename
}

func (info *FileInfo) Size() int64 {
	return info.Len
}

func (info *FileInfo) Mode() os.FileMode {
	return 0666
}

func (info *FileInfo) ModTime() time.Time {
	return info.MTime
}

func (info *FileInfo) IsDir() bool {
	return info.Type == PP_FILETYPE_DIRECTORY
}

func (info *FileInfo) Sys() interface{} {
	return nil
}

// DirectoryEntry is an entry in a directory.
type DirectoryEntry struct {
	File FileRef
	Type FileType
}

func (in DirectoryEntry) toPP(out *pp_DirectoryEntry) {
	*(*pp_Resource)(unsafe.Pointer(&out[0])) = in.File.id
	*(*FileType)(unsafe.Pointer(&out[4])) = in.Type
}

func (out *DirectoryEntry) fromPP(in pp_DirectoryEntry) {
	out.File.id = *(*pp_Resource)(unsafe.Pointer(&in[0]))
	out.Type = *(*FileType)(unsafe.Pointer(&in[4]))
}

// FileRef represents a "weak pointer" to a file in a file system.
type FileRef struct {
	Resource
}

// CreateFileRef creates a weak pointer to a file in the given file system.
// The returned ref must be released explicitly with the Release method.
func (fs FileSystem) CreateFileRef(path string) (ref FileRef, err error) {
	b := append([]byte(path), byte(0))
	id := ppb_fileref_create(fs.id, &b[0])
	if id == 0 {
		err = fmt.Errorf("can't create file %q", path)
		return
	}
	ref.id = id
	return
}

// Delete deletes a file or directory.
//
// If the ref refers to a directory, then the directory must be empty. It is an
// error to delete a file or directory that is in use. It is not valid to delete
// a file in the external file system.
func (ref FileRef) Delete() error {
	code := ppb_fileref_delete(ref.id, ppNullCompletionCallback)
	return decodeError(Error(code))
}

// GetParent returns the parent directory of this file.
//
// If file_ref points to the root of the filesystem, then the root is returned.
func (ref FileRef) Parent() (parent FileRef, err error) {
	id := ppb_fileref_get_parent(ref.id)
	if id == 0 {
		err = errFileRefParentFailed
		return
	}
	parent.id = id
	return
}

// Name returns the name of the file.
func (ref FileRef) Name() (string, error) {
	var ppVar pp_Var
	ppb_fileref_get_name(&ppVar, ref.id)
	v := makeVar(ppVar)
	s, err := v.AsString()
	v.Release()
	return s, err
}

// Path returns the full path of the file.
func (ref FileRef) Path() (string, error) {
	var ppVar pp_Var
	ppb_fileref_get_path(&ppVar, ref.id)
	v := makeVar(ppVar)
	s, err := v.AsString()
	v.Release()
	return s, err
}

// Rename renames a file or directory.
//
// Arguments file_ref and new_file_ref must both refer to files in the same file
// system. It is an error to rename a file or directory that is in use. It is
// not valid to rename a file in the external file system.
func (ref FileRef) Rename(newName FileRef) error {
	code := ppb_fileref_rename(ref.id, newName.id, ppNullCompletionCallback)
	return decodeError(Error(code))
}

// MakeDirectory makes a new directory in the file system according to the given
// flags, which is a bit-mask of the MakeDirectoryFlag values.
//
// It is not valid to make a directory in the external file system.
func (ref FileRef) MakeDirectory(flags MakeDirectoryFlag) error {
	code := ppb_fileref_make_directory(ref.id, int32(flags), ppNullCompletionCallback)
	return decodeError(Error(code))
}

// Stat queries info about a file or directory.
//
// You must have access to read this file or directory if it exists in the
// external filesystem.
func (ref FileRef) Stat() (info FileInfo, err error) {
	var name string
	name, err = ref.Name()
	if err != nil {
		return
	}
	var ppInfo pp_FileInfo
	code := ppb_fileref_query(ref.id, &ppInfo, ppNullCompletionCallback)
	if code < 0 {
		err = decodeError(Error(code))
		return
	}
	info.fromPP(name, ppInfo)
	return
}

// Touch Updates time stamps for a file.
//
// You must have write access to the file if it exists in the external filesystem.
func (ref FileRef) Touch(atime, mtime time.Time) error {
	code := ppb_fileref_touch(ref.id, toPPTime(atime), toPPTime(mtime), ppNullCompletionCallback)
	return decodeError(Error(code))
}

// ReadDirectoryEntries reads all entries in a directory.
func (ref FileRef) ReadDirectoryEntries() (entries []DirectoryEntry, err error) {
	var aout pp_ArrayOutput
	var alloc arrayOutputBuffer
	init_array_output(&aout, &alloc)
	log.Printf("AAAAAAAA2: %p", &alloc)
	code := ppb_fileref_read_directory_entries(ref.id, aout, ppNullCompletionCallback)
	if code < 0 {
		err = decodeError(Error(code))
		return
	}

	count := alloc.count
	for i := uint32(0); i < count; i++ {
		ppEntry := (*pp_DirectoryEntry)(unsafe.Pointer(&alloc.buffer[i*alloc.size]))
		var entry DirectoryEntry
		entry.fromPP(*ppEntry)
		entries = append(entries, entry)
	}
	return
}

// FileIO is used to operate on regular files.
type FileIO struct {
	Resource
	name     string
	position int64
}

// Open opens the specified regular file for I/O according to the given open
// flags, which is a bit-mask of the PP_FileOpenFlags values.
//
// Upon success, the corresponding file is classified as "in use" by this FileIO
// object until such time as the FileIO object is closed or destroyed.
func (file *FileIO) Open(ref FileRef, openFlags FileOpenFlag) error {
	name, err := ref.Name()
	if err != nil {
		return err
	}
	code := ppb_fileio_open(file.id, ref.id, int32(openFlags), ppNullCompletionCallback)
	if code >= 0 {
		file.name = name
	}
	if openFlags&PP_FILEOPENFLAG_APPEND != 0 {
		info, err := file.Stat()
		if err != nil {
			return err
		}
		file.position = info.Len
	}
	return decodeError(Error(code))
}

// Close cancels any IO that may be pending, and closes the FileIO object.
//
// Any pending callbacks will still run, reporting PP_ERROR_ABORTED if pending
// IO was interrupted. It is not valid to call Open() again after a call to this
// method. Note: If the FileIO object is destroyed, and it is still open, then
// it will be implicitly closed, so you are not required to call Close().
func (file *FileIO) Close() error {
	ppb_fileio_close(file.id)
	return nil
}

// Sync flushes changes to disk.
//
// This call can be very expensive! The FileIO object must have been opened with
// write access and there must be no other operations pending.
func (file *FileIO) Sync() error {
	code := ppb_fileio_flush(file.id, ppNullCompletionCallback)
	return decodeError(Error(code))
}

// Stat queries info about the file opened by this FileIO object.
//
// The FileIO object must be opened, and there must be no other operations
// pending.
func (file *FileIO) Stat() (info FileInfo, err error) {
	var ppInfo pp_FileInfo
	code := ppb_fileio_query(file.id, &ppInfo, ppNullCompletionCallback)
	if code < 0 {
		err = decodeError(Error(code))
		return
	}
	info.fromPP(file.name, ppInfo)
	return
}

// Read reads up to len(b) bytes from the File. It returns the number of bytes
// read and an error, if any. EOF is signaled by a zero count with err set to
// io.EOF.
func (file *FileIO) Read(buf []byte) (n int, err error) {
	amount, err := file.ReadAt(buf, file.position)
	file.position += int64(amount)
	return amount, err
}

// Write writes len(b) bytes to the File. It returns the number of bytes written
// and an error, if any. Write returns a non-nil error when n != len(b).
func (file *FileIO) Write(buf []byte) (n int, err error) {
	amount, err := file.WriteAt(buf, file.position)
	file.position += int64(amount)
	return amount, err
}

// WriteString is like Write, but writes the contents of string s rather than a
// slice of bytes.
func (file *FileIO) WriteString(s string) (n int, err error) {
	return file.Write([]byte(s))
}

// ReadAtOffset reads from an offset in the file.  Does not affect the current
// file position.
//
// The size of the buf must be large enough to hold the specified number of
// bytes to read. This function might perform a partial read, meaning all the
// requested bytes might not be returned, even if the end of the file has not
// been reached. The FileIO object must have been opened with read access.
func (file *FileIO) ReadAt(buf []byte, off int64) (amount int, err error) {
	code := ppb_fileio_read(file.id, off, &buf[0], int32(len(buf)), ppNullCompletionCallback)
	if code < 0 {
		err = decodeError(Error(code))
		return
	}
	if code == 0 {
		err = io.EOF
		return
	}
	amount = int(code)
	return
}

// WriteAt writes to an offset in the file.  Does not affect the current
// file position.
//
// This function might perform a partial write. The FileIO object must have been
// opened with write access.
func (file *FileIO) WriteAt(buf []byte, off int64) (amount int, err error) {
	code := ppb_fileio_write(file.id, off, &buf[0], int32(len(buf)), ppNullCompletionCallback)
	if code < 0 {
		err = decodeError(Error(code))
		return
	}
	if code == 0 {
		err = io.EOF
		return
	}
	amount = int(code)
	return
}

// Seek sets the offset for the next Read or Write on file to offset,
// interpreted according to whence: 0 means relative to the origin of the file,
// 1 means relative to the current offset, and 2 means relative to the end. It
// returns the new offset and an error, if any.
func (file *FileIO) Seek(offset int64, whence int) (ret int64, err error) {
	switch whence {
	case 0:
		file.position = offset
	case 1:
		if offset > file.position {
			err = errNegativeFilePosition
			return
		}
		file.position += offset
	case 2:
		var info FileInfo
		info, err = file.Stat()
		if err != nil {
			return
		}
		p := info.Len + offset
		if p < 0 {
			err = errNegativeFilePosition
			return
		}
		file.position = p
	}
	ret = file.position
	return
}

// Truncate sets the length of the file.
//
// If the file size is extended, then the extended area of the file is
// zero-filled. The FileIO object must have been opened with write access and
// there must be no other operations pending.
func (file *FileIO) Truncate(length int64) error {
	code := ppb_fileio_set_length(file.id, length, ppNullCompletionCallback)
	return decodeError(Error(code))
}

// Touch Updates time stamps for the file opened by this FileIO object.
//
// This function will fail if the FileIO object has not been opened. The FileIO
// object must be opened, and there must be no other operations pending.
func (file *FileIO) Touch(atime, mtime time.Time) error {
	code := ppb_fileio_touch(file.id, toPPTime(atime), toPPTime(mtime), ppNullCompletionCallback)
	return decodeError(Error(code))
}

// CreateFileSystem creates a file system with the given type.
func (inst Instance) CreateFileSystem(ty FileSystemType) (fs FileSystem, err error) {
	id := ppb_filesystem_create(inst.id, ty)
	if id == 0 {
		err = errFileSystemCreateFailed
		return
	}
	fs.instance = inst
	fs.id = id
	return
}

// CreateFileIO creates a new FileIO object.
func (inst Instance) CreateFileIO() (file FileIO, err error) {
	id := ppb_fileio_create(inst.id)
	if id == 0 {
		err = errFileIOCreateFailed
		return
	}
	file.id = id
	return
}
