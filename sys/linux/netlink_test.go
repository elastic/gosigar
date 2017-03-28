// +build linux

package linux

import (
	"bytes"
	"encoding/binary"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Validate NetlinkClient implements NetlinkSendReceiver.
var _ NetlinkSendReceiver = &NetlinkClient{}

func TestParseNetlinkErrorDataTooShort(t *testing.T) {
	assert.Error(t, ParseNetlinkError(nil), "too short")
}

func TestParseNetlinkErrorErrno(t *testing.T) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, -1*int32(NLE_MSG_TOOSHORT))
	assert.Equal(t, ParseNetlinkError(buf.Bytes()), NLE_MSG_TOOSHORT)
}

func TestNewNetlinkClient(t *testing.T) {
	c, err := NewNetlinkClient(syscall.NETLINK_AUDIT, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	assert.Len(t, c.readBuf, os.Getpagesize())

	// First PID assigned by the kernel will be our actual PID.
	assert.EqualValues(t, os.Getpid(), c.pid)

	c2, err := NewNetlinkClient(syscall.NETLINK_AUDIT, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c2.Close()

	// Second PID assigned by kernel will be random.
	assert.NotEqual(t, 0, c2.pid)
	assert.NotEqual(t, c.pid, c2.pid)
}
