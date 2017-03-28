// +build linux

package linux

import (
	"encoding/hex"
	"flag"
	"io"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

var hexdump = flag.Bool("hexdump", false, "dump kernel responses to stdout in hexdump -C format")

func TestAuditClientGetStatus(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("must be root to get audit status")
	}

	status, err := getStatus(t)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Status: %+v", status)
}

func TestAuditClientGetStatusPermissionError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("must be non-root to test permission failure")
	}

	status, err := getStatus(t)
	assert.Nil(t, status, "status should be nil")
	assert.Equal(t, syscall.EPERM, err)
}

func getStatus(t testing.TB) (*AuditStatus, error) {
	var dumper io.WriteCloser
	if *hexdump {
		dumper = hex.Dumper(os.Stdout)
		defer dumper.Close()
	}

	c, err := NewAuditClient(dumper)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	return c.GetStatus()
}

func TestAuditClientSetPID(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("must be root to set audit port id")
	}

	var dumper io.WriteCloser
	if *hexdump {
		dumper = hex.Dumper(os.Stdout)
		defer dumper.Close()
	}

	c, err := NewAuditClient(dumper)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	err = c.SetPortID(0)
	if err != nil {
		t.Fatal(err)
	}
}
