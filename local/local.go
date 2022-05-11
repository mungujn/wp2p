package local

import (
	"context"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

type LocalFilesystem struct {
	rootFolder string
}

func New(rootFolder string) *LocalFilesystem {
	log.Debug("setting up local file system to serve from root folder: ", rootFolder)
	return &LocalFilesystem{
		rootFolder: rootFolder,
	}
}

func (lfs *LocalFilesystem) StartHost(ctx context.Context) error {
	return nil
}

func (lfs *LocalFilesystem) GetFile(ctx context.Context, username, path string) ([]byte, error) {
	var fullPath string
	if username == "" {
		fullPath = lfs.rootFolder + "/" + path
	} else {
		fullPath = lfs.rootFolder + "/" + username + "/" + path
	}
	log.Debug("reading file: ", fullPath)
	return ioutil.ReadFile(fullPath)
}

func (lfs *LocalFilesystem) GetOnlineNodes(ctx context.Context) ([]string, error) {
	return []string{"user_2", "user_3"}, nil
}
