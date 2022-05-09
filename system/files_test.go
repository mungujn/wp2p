package system

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/mungujn/web-exp/local"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test_GetFile(t *testing.T) {
	// prep test config
	ctx := context.Background()
	cfg := Config{
		Username:        "me",
		LocalRootFolder: "test_data/server_root",
	}

	// prep test distributed services provider
	provider := local.New(cfg.LocalRootFolder)

	// prep test system
	sys, err := New(ctx, cfg, provider)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	cases := []struct {
		Name                string
		Path                string
		ExpectedContentType string
		ExpectedBytes       []byte
	}{
		{
			Name:                "home page",
			Path:                "",
			ExpectedContentType: htmlContent,
			ExpectedBytes:       getFile("test_data/expected/home.html"),
		},
		{
			Name:                "html content type",
			Path:                "file.html",
			ExpectedContentType: htmlContent,
			ExpectedBytes:       getFile("test_data/expected/file.html"),
		},
		{
			Name:                "png content type",
			Path:                "file.png",
			ExpectedContentType: pngContent,
			ExpectedBytes:       getFile("test_data/expected/file.png"),
		},
		{
			Name:                "js content type",
			Path:                "file.js",
			ExpectedContentType: jsContent,
			ExpectedBytes:       getFile("test_data/expected/file.js"),
		},
		{
			Name:                "root level element, type inferred",
			Path:                "favicon.ico",
			ExpectedContentType: inferContent,
			ExpectedBytes:       getFile("test_data/expected/favicon.ico"),
		},
		{
			Name:                "nested element, css content type",
			Path:                "sub-path/file.css",
			ExpectedContentType: cssContent,
			ExpectedBytes:       getFile("test_data/expected/file.css"),
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			log.Debug("testing file case: ", testCase.Path)
			data, contentType, err := sys.GetFile(ctx, testCase.Path)
			// writeFile(testCase.Name, data)
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedContentType, contentType)
			assert.Equal(t, testCase.ExpectedBytes, data)
		})
	}
}

func getFile(path string) []byte {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return []byte("will fail")
	}
	return contents
}

// func writeFile(path string, contents []byte) {
// 	err := ioutil.WriteFile(path, contents, 0644)
// 	if err != nil {
// 		log.Error(err)
// 	}
// }
