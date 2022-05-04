package local

import (
	"context"
	"io/ioutil"

	"github.com/mungujn/web-exp/system"

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

func (lfs *LocalFilesystem) SetUp(ctx context.Context, cfg system.Config) error {
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
