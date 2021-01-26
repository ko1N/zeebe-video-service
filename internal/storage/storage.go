package storage

import (
	"fmt"
	"io"
	"net/url"

	"github.com/ko1N/zeebe-video-service/internal/environment"
)

// TODO: refactor
type File struct {
	Name        string
	ContentType string
	Size        int64
}

// TODO: custom interface
type VirtualFileReader interface {
	io.Reader
	io.Seeker
	Size() (int64, error)
	io.Closer
}

// Storage -
type Storage interface {
	List(folder string) ([]File, error)
	CreateFolder(folder string) error
	DeleteFolder(folder string) error
	DownloadFile(remotefile string, localfile string) error
	UploadFile(localfile string, remotefile string) error
	DeleteFile(remotefile string) error

	// new behavior
	GetFileReader(filename string) (VirtualFileReader, error)

	Close()
}

func ConnectStorage(env environment.Environment, url *url.URL) (Storage, error) {
	// connect
	switch url.Scheme {
	case "minio":
		conf := MinIOConfig{
			Endpoint:  url.Host,
			AccessKey: url.User.Username(),
			UseSSL:    false,       // TODO: fix ssl?
			Location:  "us-east-1", // TODO: from path?
		}
		if passwd, ok := url.User.Password(); ok {
			conf.AccessKeySecret = passwd
		}
		return ConnectMinIO(env, &conf)
		/*
			case "smb":
				conf := SmbConfig{
					Server: url.Host,
					User:   url.User.Username(),
				}
				if passwd, ok := url.User.Password(); ok {
					conf.Password = passwd
				}
				return ConnectSmb(env, &conf)
		*/
	default:
		return nil, fmt.Errorf("invalid scheme")
	}
}
