package remote

import (
	"context"
	"testing"
	"time"

	"fmt"

	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func Test_GetFile(t *testing.T) {
	// prep
	networkName := "test_network"
	protocolId := "test_protocol"
	protocolV := "0.1"

	// prep host 1
	h1Ctx, h1Cancel := context.WithCancel(context.Background())
	host1 := New(
		"host1",
		"test_data/host_1",
		"0.0.0.0",
		4042,
		networkName,
		fmt.Sprintf("/%s/%s", protocolId, protocolV),
	)
	err := host1.StartHost(h1Ctx)
	assert.NoError(t, err)
	defer h1Cancel()

	// prep host 2
	h2Ctx, h2Cancel := context.WithCancel(context.Background())
	host2 := New(
		"host2",
		"test_data/host_2",
		"0.0.0.0",
		4043,
		networkName,
		fmt.Sprintf("/%s/%s", protocolId, protocolV),
	)
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
