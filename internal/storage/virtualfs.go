package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type VirtualFS struct {
	filesystem *fuse.Server
	root       *mountRoot
}

func (self *VirtualFS) AddInput(input string) error {
	url, err := url.Parse(input)
	if err != nil {
		return err
	}
	self.root.inputs = append(self.root.inputs, url)
	return nil
}

func (self *VirtualFS) Close() {
	self.filesystem.Unmount()
}

type mountRoot struct {
	fs.Inode

	inputs []*url.URL
}

func (self *mountRoot) OnAdd(ctx context.Context) {
	for _, input := range self.inputs {
		dir, base := filepath.Split(input.Path)

		p := &self.Inode
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
		ch := p.NewPersistentInode(ctx, &mountFile{
			input: input,
		}, fs.StableAttr{})
		p.AddChild(base, ch, true)
	}
}

var _ = (fs.NodeOnAdder)((*mountRoot)(nil))

/////////////////////////////////////////////////////

// mountFile - implements a file on the mounted filesystem
type mountFile struct {
	fs.Inode
	input *url.URL

	mu      sync.Mutex
	storage Storage
	reader  VirtualFileReader
}

var _ = (fs.NodeOpener)((*mountFile)(nil))
var _ = (fs.NodeGetattrer)((*mountFile)(nil))

// Getattr sets the minimum, which is the size. A more full-featured
// FS would also set timestamps and permissions.
func (self *mountFile) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	self.mu.Lock()
	defer self.mu.Unlock()

	if self.storage == nil {
		storage, err := ConnectStorage(nil, self.input)
		if err != nil {
			fmt.Println("Failed to open storage")
			// TODO: return proper error
			return 1
		}
		self.storage = storage
	}

	if self.reader == nil {
		reader, err := self.storage.GetFileReader(self.input.Path)
		if err != nil {
			fmt.Println("Failed to open storage")
			// TODO: return proper error
			return 1
		}
		self.reader = reader
	}

	out.Mode = 0755 // TODO:
	out.Nlink = 1
	//out.Mtime = uint64(zf.file.ModTime().Unix())
	//out.Atime = out.Mtime
	//out.Ctime = out.Mtime
	size, _ := self.reader.Size()
	// TODO: error handling
	out.Size = uint64(size)
	const bs = 512
	out.Blksize = bs
	out.Blocks = (out.Size + bs - 1) / bs
	return 0
}

// TODO: implement file close?

// Open lazily unpacks zip data
func (self *mountFile) Open(ctx context.Context, flags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	self.mu.Lock()
	defer self.mu.Unlock()

	if self.storage == nil {
		storage, err := ConnectStorage(nil, self.input)
		if err != nil {
			fmt.Println("Failed to open storage")
			// TODO: return proper error
			return nil, 0, 1
		}
		self.storage = storage
	}

	if self.reader == nil {
		reader, err := self.storage.GetFileReader(self.input.Path)
		if err != nil {
			fmt.Println("Failed to open storage")
			// TODO: return proper error
			return nil, 0, 1
		}
		self.reader = reader
	}

	// We don't return a filehandle since we don't really need
	// one.  The file content is immutable, so hint the kernel to
	// cache the data.
	return nil, fuse.FOPEN_KEEP_CACHE, 0
}

// Read simply returns the data that was already unpacked in the Open call
func (self *mountFile) Read(ctx context.Context, f fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	self.mu.Lock()
	defer self.mu.Unlock()

	if self.reader == nil {
		return fuse.ReadResultData([]byte{}), 1
	}

	bytes := len(dest)
	buffer := make([]byte, bytes)

	_, err := self.reader.Seek(off, io.SeekStart)
	if err != nil {
		fmt.Println("big error ", err)
	}

	_, err = self.reader.Read(buffer)
	if err != nil {
		fmt.Println("big error2 ", err)
	}

	return fuse.ReadResultData(buffer), 0
}

/////////////////////////////////////////////////////

// TODO: put into struct and connect with input/output files :)
func MountVirtualFS(inputs []string) (*VirtualFS, error) {
	inputUrls := []*url.URL{}
	for _, input := range inputs {
		url, err := url.Parse(input)
		if err != nil {
			return nil, err
		}
		inputUrls = append(inputUrls, url)
	}

	opts := &fs.Options{}
	opts.Debug = true

	root := &mountRoot{
		inputs: inputUrls,
	}

	server, err := fs.Mount("test2", root, opts)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
		return nil, err
	}
	go func() {
		server.Wait()
	}()

	return &VirtualFS{
		filesystem: server,
		root:       root,
	}, nil
}
