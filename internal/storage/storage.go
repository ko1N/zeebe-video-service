package storage

import (
	"fmt"
	"net/url"

	"github.com/ko1N/zeebe-video-service/internal/environment"
)

// File -
type File struct {
	Name        string
	ContentType string
	Size        int64
}

// Storage -
type Storage interface {
	List(bucket string) ([]File, error)
	CreateBucket(bucket string) error
	DeleteBucket(bucket string) error
	GetFile(bucket string, objname string, outname string) error
	PutFile(bucket string, inpath string, outfile string) error
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
	default:
		return nil, fmt.Errorf("invalid scheme")
	}
}
