package system

import (
	"context"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

// GetFile returns file content
func (s *System) GetFile(ctx context.Context, path string) ([]byte, string, error) {
	log.Debug("GetFile: ", path)
	var username string
	var filename string
	parts := strings.Split(path, "/")
	noParts := len(parts)

	if noParts == 1 && parts[0] == "" {
		log.Info("fetching root path /")
		return s.renderedHomePage(s.GetOnlineNodes())
	}

	if noParts == 1 && parts[0] != "" {
		log.Debug("no username provided, defaulting to current user")
		username = s.cfg.Username
		filename = parts[0]
		filenameParts := strings.Split(filename, ".")
		if len(filenameParts) == 1 {
			log.Debug("no file extension provided, defaulting to current user index.html")
			filename = "index.html"
		}
	} else {
		username = parts[0]
		filename = strings.Join(parts[1:], "/")
	}

	str := fmt.Sprintf("reading file %s of user %s", filename, username)
	log.Info(str)

	if username == s.cfg.Username {
		username = ""
	}

	file, err := s.fileProvider.GetFile(ctx, username, filename)
	if err != nil {
		log.Error(err)
		return []byte(err.Error()), plainTextContent, err
	}

	log.Info("file read")
	return file, inferContentType(filename), nil
}

func (s *System) GetOnlineNodes() []string {
	return s.fileProvider.GetOnlineNodes()
}
