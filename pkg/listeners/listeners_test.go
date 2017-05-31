// +build linux,cgo

package listeners

import (
	"crypto/tls"
	"flag"
	"net"
	"os"
	"strconv"
	"testing"

	apiserver "github.com/docker/docker/api/server"
	"github.com/docker/docker/cmd/dockerd/hack"
	"github.com/docker/docker/dockerversion"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var rootEnabled = false

func init() {
	flag.BoolVar(&rootEnabled, "test.root", false, "enable tests that require root")
}
func RequiresRoot(t *testing.T) {
	if !rootEnabled {
		t.Skip("skipping test that requires root")
		return
	}
	require.Equal(t, 0, os.Getuid(), "This test must be run as root.")
}
func TestInitForUnix(t *testing.T) {
	RequiresRoot(t)

	proto := "unix"
	addr := "/tmp/example.sock"
	tlsConfig := tlsconfig.ServerDefault()

	garbageGroupName := "garbage"

	ls, serverConfig, err := commonConfig(proto, addr, garbageGroupName, tlsConfig)
	assert.Equal(t, "group "+garbageGroupName+" not found", err.Error())

	ls, serverConfig, err = commonConfig(proto, addr, "root", tlsConfig)
	require.NoError(t, err)

	serverCommonConfig(ls, proto, addr, serverConfig)

	errClient := validateClient(proto, addr)
	require.NoError(t, errClient)
}

func TestInitForTCP(t *testing.T) {
	proto := "tcp"
	addr := "127.0.0.1:4004"

	tlsConfig := tlsconfig.ServerDefault()
	ls, serverConfig, err := commonConfig(proto, addr, "root", tlsConfig)
	require.NoError(t, err)

	serverCommonConfig(ls, proto, addr, serverConfig)

	errClient := validateClient(proto, addr)
	require.NoError(t, errClient)

}

func TestInitForFD(t *testing.T) {
	pid := os.Getpid()
	os.Setenv("LISTEN_PID", strconv.Itoa(pid))
	os.Setenv("LISTEN_FDS", strconv.Itoa(2))

	proto := "fd"
	addr := "*"
	garbageAddr := "garbage"
	tlsConfig := tlsconfig.ServerDefault()

	ls, serverConfig, err := commonConfig(proto, "10", "root", tlsConfig)
	assert.Equal(t, "too few socket activated files passed in by systemd", err.Error())

	ls, serverConfig, err = commonConfig(proto, "garbage", "root", tlsConfig)
	assert.Equal(t, "failed to parse systemd fd address: should be a number: "+garbageAddr, err.Error())

	ls, serverConfig, err = commonConfig(proto, addr, "root", tlsConfig)
	require.NoError(t, err)

	ls = wrapListeners(proto, ls)
	apiServer := apiserver.New(serverConfig)
	apiServer.Accept(addr, ls...)

}

func wrapListeners(proto string, ls []net.Listener) []net.Listener {
	switch proto {
	case "unix":
		ls[0] = &hack.MalformedHostHeaderOverride{ls[0]}
	case "fd":
		for i := range ls {
			ls[i] = &hack.MalformedHostHeaderOverride{ls[i]}
		}
	}
	return ls
}

func validateClient(proto string, address string) error {
	conn, err := net.Dial(proto, address)
	if err != nil && conn != nil {
		return err
	}
	return nil
}

func commonConfig(proto string, addr string, socketGrp string, tlsConfig *tls.Config) ([]net.Listener, *apiserver.Config, error) {
	serverConfig := &apiserver.Config{
		Logging:     true,
		SocketGroup: socketGrp,
		Version:     dockerversion.Version,
		EnableCors:  false,
		CorsHeaders: "",
	}

	tlsConfig = tlsconfig.ServerDefault()
	serverConfig.TLSConfig = tlsConfig

	if proto == "fd" {
		ls, err := Init(proto, addr, serverConfig.SocketGroup, nil)
		return ls, serverConfig, err
	}
	ls, err := Init(proto, addr, serverConfig.SocketGroup, tlsConfig)
	return ls, serverConfig, err
}

func serverCommonConfig(ls []net.Listener, proto string, addr string, serverConfig *apiserver.Config) {
	ls = wrapListeners(proto, ls)
	apiServer := apiserver.New(serverConfig)
	apiServer.Accept(addr, ls...)

	serveAPIWait := make(chan error)
	go apiServer.Wait(serveAPIWait)
	defer apiServer.Close()
}
