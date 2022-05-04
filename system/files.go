package system

import (
	"context"
	"errors"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

// GetFile returns file content
func (s *System) GetFile(ctx context.Context, path string) ([]byte, error) {
	log.Debug("GetFile: ", path)
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		err := fmt.Errorf("invalid path: %s", path)
		log.Error(err)
		return nil, err
	}
	username := parts[0]
	filename := parts[1]
	str := fmt.Sprintf("reading file %s of user %s", filename, username)
	log.Info(str)

	parts = strings.Split(filename, ".")
	lenParts := len(parts)
	if lenParts != 2 {
		if lenParts == 1 {
			filename = parts[0] + ".html"
		} else {
			errStr := fmt.Sprintf("invalid filename: %s, no extension provided", filename)
			err := errors.New(errStr)
			log.Error(err)
			return []byte(errStr), err
		}
	}

	file, err := s.fileProvider.GetFile(ctx, username, filename)
	if err != nil {
		log.Error(err)
		return []byte(err.Error()), err
	}
	
	log.Info("file read")
	return file, nil
}
