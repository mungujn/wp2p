package remote

import (
	"context"
	"testing"
	"time"

	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"github.com/mungujn/web-exp/app"

	"github.com/stretchr/testify/assert"
)

func Test_GetFile(t *testing.T) {
	// prep
	networkName := "test_network"
	protocolId := "test_protocol"
	protocolV := "0.1"

	// prep host 1
	dcfg1 := app.Config{
		Username:        "host1",
		LocalRootFolder: "test_data/host_1",
		LocalNodeHost:   "0.0.0.0",
		LocalNodePort:   4042,
		NetworkName:     networkName,
		ProtocolId:      protocolId,
		ProtocolVersion: protocolV,
		RunGlobal:       false,
	}
	h1Ctx, h1Cancel := context.WithCancel(context.Background())
	host1 := New(dcfg1)
	err := host1.StartHost(h1Ctx)
	assert.NoError(t, err)
	defer h1Cancel()

	// prep host 2
	dcfg2 := app.Config{
		Username:        "host2",
		LocalRootFolder: "test_data/host_2",
		LocalNodeHost:   "0.0.0.0",
		LocalNodePort:   4043,
		NetworkName:     networkName,
		ProtocolId:      protocolId,
		ProtocolVersion: protocolV,
		RunGlobal:       false,
	}
	h2Ctx, h2Cancel := context.WithCancel(context.Background())
	host2 := New(dcfg2)
	err = host2.StartHost(h2Ctx)
	assert.NoError(t, err)
	defer h2Cancel()

	// prep test cases
	cases := []struct {
		Name          string
		Username      string
		Path          string
		Host          int
		ExpectedBytes []byte
	}{
		{
			Name:          "root element",
			Username:      "host2",
			Path:          "error.png",
			Host:          1,
			ExpectedBytes: getFile("test_data/expected/error.png"),
		},
		{
			Name:          "nested element",
			Username:      "host2",
			Path:          "nested/file.js",
			Host:          1,
			ExpectedBytes: getFile("test_data/expected/file.js"),
		},
		{
			Name:          "root element reversed",
			Username:      "host1",
			Path:          "success.png",
			Host:          2,
			ExpectedBytes: getFile("test_data/expected/success.png"),
		},
		{
			Name:          "nested element reversed",
			Username:      "host1",
			Path:          "nested/file.css",
			Host:          2,
			ExpectedBytes: getFile("test_data/expected/file.css"),
		},
		{
			Name:          "self read",
			Username:      "",
			Path:          "success.png",
			Host:          1,
			ExpectedBytes: getFile("test_data/expected/success.png"),
		},
	}

	start := time.Now()
	for {
		if time.Since(start) > time.Second*5 {
			t.Fatal("timed out before successful host handshake")
		}
		usernames1 := host1.GetOnlineNodes()
		usernames2 := host2.GetOnlineNodes()
		if len(usernames1) == 1 && len(usernames2) == 1 {
			if usernames1[0] == "host2" && usernames2[0] == "host1" {
				break
			}
		}
		log.Info("u1 ", usernames1, "u2 ", usernames2)
		time.Sleep(time.Second)

	}

	// run tests
	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			log.Debug("testing host case: ", testCase.Path)
			var host *RemoteFilesystem
			var ctx context.Context
			if testCase.Host == 1 {
				host = host1
				ctx = h1Ctx
			} else {
				host = host2
				ctx = h2Ctx
			}
			data, err := host.GetFile(ctx, testCase.Username, testCase.Path)
			// writeFile(testCase.Name, data)
			assert.Nil(t, err)
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
