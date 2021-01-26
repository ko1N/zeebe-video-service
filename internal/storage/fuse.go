package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type InputFile struct {
	filename string
	reader   VirtualFileReader
}

type mountRoot struct {
	fs.Inode

	// input reader
	input InputFile

	// TODO: output writer
}

func (mr *mountRoot) OnAdd(ctx context.Context) {
	dir, base := filepath.Split(filepath.Clean(mr.input.filename))

	p := &mr.Inode
	for _, component := range strings.Split(dir, "/") {
		if len(component) == 0 {
			continue
		}
		ch := p.GetChild(component)
		if ch == nil {
			ch = p.NewPersistentInode(ctx, &fs.Inode{},
				fs.StableAttr{Mode: fuse.S_IFDIR})
			p.AddChild(component, ch, true)
		}

		p = ch
	}
	ch := p.NewPersistentInode(ctx, &mountFile{input: mr.input}, fs.StableAttr{})
	p.AddChild(base, ch, true)
}

var _ = (fs.NodeOnAdder)((*mountRoot)(nil))

/////////////////////////////////////////////////////

// mountFile - implements a file on the mounted filesystem
type mountFile struct {
	fs.Inode
	input InputFile

	//mu   sync.Mutex
	//data []byte
}

var _ = (fs.NodeOpener)((*mountFile)(nil))
var _ = (fs.NodeGetattrer)((*mountFile)(nil))

// Getattr sets the minimum, which is the size. A more full-featured
// FS would also set timestamps and permissions.
func (mf *mountFile) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0755 // TODO:
	out.Nlink = 1
	//out.Mtime = uint64(zf.file.ModTime().Unix())
	//out.Atime = out.Mtime
	//out.Ctime = out.Mtime
	size, _ := mf.input.reader.Size()
	// TODO: error handling
	out.Size = uint64(size)
	const bs = 512
	out.Blksize = bs
	out.Blocks = (out.Size + bs - 1) / bs
	return 0
}

// TODO: implement file close?

// Open lazily unpacks zip data
func (mf *mountFile) Open(ctx context.Context, flags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	// TODO: close reader + writer at one point?

	// store reader
	// TODO: mutex in mf?

	// We don't return a filehandle since we don't really need
	// one.  The file content is immutable, so hint the kernel to
	// cache the data.
	return nil, fuse.FOPEN_KEEP_CACHE, 0
}

// Read simply returns the data that was already unpacked in the Open call
func (mf *mountFile) Read(ctx context.Context, f fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	bytes := len(dest)
	buffer := make([]byte, bytes)
	fmt.Println("READING CHUNK WITH SIZE ", bytes)
	fmt.Println("READING CHUNK WITH SIZE ", bytes)
	fmt.Println("READING CHUNK WITH SIZE ", bytes)
	fmt.Println("READING CHUNK WITH SIZE ", bytes)

	_, err := mf.input.reader.Seek(off, io.SeekStart)
	if err != nil {
		fmt.Println("big error ", err)
	}

	_, err = mf.input.reader.Read(buffer)
	if err != nil {
		fmt.Println("big error2 ", err)
	}

	return fuse.ReadResultData(buffer), 0
}

/////////////////////////////////////////////////////

// TODO: put into struct and connect with input/output files :)
func RunFuse() {
	// TEST
	url, err := url.Parse("minio://minio:miniominio@172.17.0.1:9000/test/untitled.mp4")

	store, err := ConnectStorage(nil, url)
	defer store.Close()

	reader, err := store.GetFileReader(url.Path)
	defer reader.Close()
	// TEST

	opts := &fs.Options{}
	opts.Debug = true

	vfs := &mountRoot{
		input: InputFile{
			filename: "untitled.mp4",
			reader:   reader,
		},
	}

	server, err := fs.Mount("test", vfs, opts)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}
	server.Wait()
}
