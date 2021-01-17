package storage

import (
	"context"
	"fmt"
	"io"
	"log"

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

// List - lists files in a remote location
func (self *MinIOStorage) List(bucket string) ([]File, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	objectCh := self.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    "",
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
func (self *MinIOStorage) CreateBucket(bucket string) error {
	found, err := self.client.BucketExists(context.Background(), bucket)
	if err != nil {
		return err
	}
	if found {
		err = self.DeleteBucket(bucket)
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
func (self *MinIOStorage) DeleteBucket(bucket string) error {
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

// GetFile - copies a file from the minio storage to the environment writer
func (self *MinIOStorage) GetFile(bucket string, objname string, outpath string) error {
	object, err := self.client.GetObject(context.Background(), bucket, objname, minio.GetObjectOptions{})
	if err != nil {
		return err
	}

	writer, err := self.environment.FileWriter(outpath)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, object)
	return err
}

// PutFile - copies a file to the minio storage
func (self *MinIOStorage) PutFile(bucket string, inpath string, objname string) error {
	reader, err := self.environment.FileReader(inpath)
	if err != nil {
		return err
	}
	defer reader.Close()

	_, err = self.client.PutObject(context.Background(), bucket, objname, reader, -1, minio.PutObjectOptions{
		//ContentType: "application/octet-stream",
	})
	return err
}
