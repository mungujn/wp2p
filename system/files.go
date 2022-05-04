package system

import (
	"context"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

func (s *System) GetFile(ctx context.Context, path string) ([]byte, string, error) {
	log.Debug("GetFile: ", path)
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		err := fmt.Errorf("invalid path: %s", path)
		log.Error(err)
		return nil, "", err
	}
	username := parts[0]
	filename := parts[1]
	str := fmt.Sprintf("getting file %s of user %s", filename, username)
	log.Info(str)
	return []byte(str), plainText, nil
}
