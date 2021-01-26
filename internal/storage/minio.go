package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/ko1N/zeebe-video-service/internal/environment"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage - describes a minio storage
type MinIOStorage struct {
	environment environment.Environment
	client      *minio.Client
	location    string
}

// MinIOConfig - config entry describing a storage config
type MinIOConfig struct {
	Endpoint        string `json:"endpoint" toml:"endpoint"`
	AccessKey       string `json:"access_key" toml:"access_key"`
	AccessKeySecret string `json:"access_key_secret" toml:"access_key_secret"`
	UseSSL          bool   `json:"use_ssl" toml:"use_ssl"`
	Location        string `json:"location" toml:"location"`
}

// ConnectMinIO - opens a connection to minio and returns the connection object
func ConnectMinIO(env environment.Environment, conf *MinIOConfig) (*MinIOStorage, error) {
	client, err := minio.New(conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.AccessKey, conf.AccessKeySecret, ""),
		Secure: conf.UseSSL,
	})
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	return &MinIOStorage{
		environment: env,
		client:      client,
		location:    conf.Location,
	}, nil
}

func parseFilename(filename string) (string, string) {
	cleanFilename := strings.TrimLeft(path.Clean(filename), string(os.PathSeparator))
	split := strings.Split(cleanFilename, string(os.PathSeparator))
	if len(split) == 1 {
		return split[0], ""
	} else {
		return split[0], strings.Join(split[1:], string(os.PathSeparator))
	}
}

// List - lists files in a remote location
func (self *MinIOStorage) List(folder string) ([]File, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bucket, prefix := parseFilename(folder)
	if prefix != "" {
		prefix += string(os.PathSeparator)
	}

	objectCh := self.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	var files []File
	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			continue
		}
		files = append(files, File{
			Name:        object.Key,
			ContentType: object.ContentType,
			Size:        object.Size,
		})
	}
	return files, nil
}

// CreateBucket - creates a new storage bucket
func (self *MinIOStorage) CreateFolder(folder string) error {
	bucket, _ := parseFilename(folder)

	found, err := self.client.BucketExists(context.Background(), bucket)
	if err != nil {
		return err
	}
	if found {
		err = self.DeleteFolder(bucket)
		if err != nil {
			return err
		}
	}
	return self.client.MakeBucket(
		context.Background(),
		bucket,
		minio.MakeBucketOptions{
			Region:        self.location,
			ObjectLocking: false,
		})
}

// DeleteBucket - deletes the given storage bucket
func (self *MinIOStorage) DeleteFolder(folder string) error {
	bucket, _ := parseFilename(folder)

	objectsCh := make(chan minio.ObjectInfo)

	// send objects to the remove channel
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		defer close(objectsCh)
		// List all objects from a bucket-name with a matching prefix.
		for object := range self.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
			Prefix:    "",
			Recursive: true,
		}) {
			if object.Err != nil {
				//log.Fatalln(object.Err)
			}
			objectsCh <- object
		}
	}()

	for rErr := range self.client.RemoveObjects(
		context.Background(),
		bucket,
		objectsCh,
		minio.RemoveObjectsOptions{
			GovernanceBypass: false,
		}) {
		fmt.Println("Error detected during deletion: ", rErr)
	}

	return self.client.RemoveBucket(context.Background(), bucket)
}

// TODO: clean up
type MinIOFileReader struct {
	object *minio.Object
}

func (self *MinIOFileReader) Read(p []byte) (n int, err error) {
	return self.object.Read(p)
}

func (self *MinIOFileReader) Seek(offset int64, whence int) (int64, error) {
	return self.object.Seek(offset, whence)
}

func (self *MinIOFileReader) Size() (int64, error) {
	info, err := self.object.Stat()
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

func (self *MinIOFileReader) Close() error {
	return self.object.Close()
}

// TODO: clean above up

func (self *MinIOStorage) GetFileReader(filename string) (VirtualFileReader, error) {
	bucket, remotefilename := parseFilename(filename)

	object, err := self.client.GetObject(context.Background(), bucket, remotefilename, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return &MinIOFileReader{
		object: object,
	}, nil
}

// DownloadFile - copies a file from the minio storage to the environment writer
func (self *MinIOStorage) DownloadFile(remotefile string, localfile string) error {
	bucket, remotefilename := parseFilename(remotefile)

	object, err := self.client.GetObject(context.Background(), bucket, remotefilename, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer object.Close()

	writer, err := self.environment.FileWriter(localfile)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, object)
	return err
}

// UploadFile - copies a file to the minio storage
func (self *MinIOStorage) UploadFile(localfile string, remotefile string) error {
	bucket, remotefilename := parseFilename(remotefile)

	reader, err := self.environment.FileReader(localfile)
	if err != nil {
		return err
	}
	defer reader.Close()

	_, err = self.client.PutObject(context.Background(), bucket, remotefilename, reader, -1, minio.PutObjectOptions{
		//ContentType: "application/octet-stream",
	})
	return err
}

// DeleteFile - deletes a file on the minio storage
func (self *MinIOStorage) DeleteFile(remotefile string) error {
	bucket, remotefilename := parseFilename(remotefile)
	return self.client.RemoveObject(context.Background(), bucket, remotefilename, minio.RemoveObjectOptions{})
}

// Close - closes the minio connection
func (self *MinIOStorage) Close() {
	// no-op
}
