package storage

/*
import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/hirochachacha/go-smb2"
	"github.com/ko1N/zeebe-video-service/internal/environment"
)

// SmbStorage - describes a smb storage
type SmbStorage struct {
	environment environment.Environment
	conn        net.Conn
	session     *smb2.Session
}

// SmbConfig - config entry describing a storage config
type SmbConfig struct {
	Server   string `json:"server" toml:"server"`
	User     string `json:"user" toml:"user"`
	Password string `json:"password" toml:"password"`
}

// ConnectSmb - opens a connection to smb and returns the connection object
func ConnectSmb(env environment.Environment, conf *SmbConfig) (*SmbStorage, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:445", conf.Server))
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	dialer := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     conf.User,
			Password: conf.Password,
		},
	}
	session, err := dialer.Dial(conn)
	if err != nil {
		log.Fatalln(err)
		conn.Close()
		return nil, err
	}

	return &SmbStorage{
		environment: env,
		conn:        conn,
		session:     session,
	}, nil
}

// List - lists files in a remote location
func (self *SmbStorage) List(folder string) ([]File, error) {
	share, dir := parseFilename(folder)

	mount, err := self.session.Mount(share)
	if err != nil {
		return nil, err
	}
	defer mount.Umount()

	fileInfos, err := mount.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []File
	for _, fileInfo := range fileInfos {
		files = append(files, File{
			Name:        fileInfo.Name(),
			ContentType: "application/octet-stream",
			Size:        fileInfo.Size(),
		})
	}
	return files, nil
}

// CreateFolder - creates a new folder
func (self *SmbStorage) CreateFolder(folder string) error {
	share, dir := parseFilename(folder)

	// open mount
	mount, err := self.session.Mount(share)
	if err != nil {
		return err
	}
	defer mount.Umount()

	// make full path
	err = mount.MkdirAll(dir, os.ModeDir)
	if err != nil {
		return err
	}

	return nil
}

// DeleteFolder - deletes the given folder
func (self *SmbStorage) DeleteFolder(folder string) error {
	share, dir := parseFilename(folder)

	// open mount
	mount, err := self.session.Mount(share)
	if err != nil {
		return err
	}
	defer mount.Umount()

	// double check if dir is actually a directory
	stat, err := mount.Stat(dir)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("cannot delete regular file")
	}

	err = mount.Remove(dir)
	if err != nil {
		return err
	}

	return nil
}

// DownloadFile - copies a file from the smb storage to the environment writer
func (self *SmbStorage) DownloadFile(remotefile string, localfile string) error {
	share, remotefilename := parseFilename(remotefile)

	// open mount
	mount, err := self.session.Mount(share)
	if err != nil {
		return err
	}
	defer mount.Umount()

	// open remote file
	file, err := mount.Open(remotefilename)
	if err != nil {
		return err
	}
	defer file.Close()

	// open local file
	writer, err := self.environment.FileWriter(localfile)
	if err != nil {
		return err
	}
	defer writer.Close()

	// copy file
	_, err = io.Copy(writer, file)
	return err
}

// UploadFile - copies a file to the smb storage
func (self *SmbStorage) UploadFile(localfile string, remotefile string) error {
	share, remotefilename := parseFilename(remotefile)

	// open local file
	reader, err := self.environment.FileReader(localfile)
	if err != nil {
		return err
	}
	defer reader.Close()

	// open mount
	mount, err := self.session.Mount(share)
	if err != nil {
		return err
	}
	defer mount.Umount()

	// open remote file
	file, err := mount.Create(remotefilename)
	if err != nil {
		return err
	}
	defer file.Close()

	// copy file
	_, err = io.Copy(file, reader)
	return err
}

// DeleteFile - deletes a file on the smb storage
func (self *SmbStorage) DeleteFile(remotefile string) error {
	share, remotefilename := parseFilename(remotefile)

	// open mount
	mount, err := self.session.Mount(share)
	if err != nil {
		return err
	}
	defer mount.Umount()

	// double check if dir is actually a file
	stat, err := mount.Stat(remotefilename)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return fmt.Errorf("cannot delete folder")
	}

	err = mount.Remove(remotefilename)
	if err != nil {
		return err
	}

	return nil
}

// Close - closes the samba connection
func (self *SmbStorage) Close() {
	self.session.Logoff()
	self.conn.Close()
}
*/
