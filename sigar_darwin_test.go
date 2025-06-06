package gosigar_test

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/elastic/gosigar"
	sigar "github.com/elastic/gosigar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var procinfo map[string]string

func setUp(t testing.TB) {
	out, err := exec.Command("/bin/ps", "-p1", "-c", "-opid,comm,stat,ppid,pgid,tty,pri,ni").Output()
	if err != nil {
		t.Fatal(err)
	}
	rdr := bufio.NewReader(bytes.NewReader(out))
	_, err = rdr.ReadString('\n') // skip header
	if err != nil {
		t.Fatal(err)
	}
	data, err := rdr.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	procinfo = make(map[string]string, 8)
	fields := strings.Fields(data)
	procinfo["pid"] = fields[0]
	procinfo["name"] = fields[1]
	procinfo["stat"] = fields[2]
	procinfo["ppid"] = fields[3]
	procinfo["pgid"] = fields[4]
	procinfo["tty"] = fields[5]
	procinfo["prio"] = fields[6]
	procinfo["nice"] = fields[7]

}

func tearDown(t testing.TB) {
}

/* ProcState.Get() call task_info, which on Mac OS X requires root
   or a signed executable. Skip the test if not running as root
   to accommodate automated tests, but let users test locally using
   `sudo -E go test`
*/

func TestDarwinProcState(t *testing.T) {
	setUp(t)
	defer tearDown(t)

	state := sigar.ProcState{}
	usr, err := user.Current()
	if err == nil && usr.Username == "root" {
		if assert.NoError(t, state.Get(1)) {

			ppid, _ := strconv.Atoi(procinfo["ppid"])
			pgid, _ := strconv.Atoi(procinfo["pgid"])

			assert.Equal(t, procinfo["name"], state.Name)
			assert.Equal(t, ppid, state.Ppid)
			assert.Equal(t, pgid, state.Pgid)
			assert.Equal(t, 1, state.Pgid)
			assert.Equal(t, 0, state.Ppid)
		}
	} else {
		t.Skip("Skipping ProcState test; run as root to test")
	}
}

func TestDarwinProcFDUsage(t *testing.T) {
	t.Run("Get(pid) on current pid returns open file count", func(t *testing.T) {
		myPid := os.Getpid()
		fdUsage := &sigar.ProcFDUsage{}

		require.NoError(t, fdUsage.Get(myPid))
		beforeOpen := fdUsage.Open
		f, err := ioutil.TempFile(t.TempDir(), "test-open-1")
		require.NoError(t, err)
		defer f.Close()

		require.NoError(t, fdUsage.Get(myPid))
		assert.Equal(t, beforeOpen+1, fdUsage.Open, "opening file increases Open count")

		require.NoError(t, f.Close())
		require.NoError(t, fdUsage.Get(myPid))
		assert.Equal(t, beforeOpen, fdUsage.Open, "closing file decreases Open count")
	})
	t.Run("Get(pid) on current pid returns rlimit", func(t *testing.T) {
		myPid := os.Getpid()
		fdUsage := &sigar.ProcFDUsage{}
		hardLim, softLim := getRlimitViaShell(t)

		require.NoError(t, fdUsage.Get(myPid))
		assert.Equal(t, hardLim, fdUsage.HardLimit)
		assert.Equal(t, softLim, fdUsage.SoftLimit)

	})
	t.Run("Get(pid) on another process returns an error", func(t *testing.T) {
		fdUsage := &sigar.ProcFDUsage{}
		otherPid := os.Getpid() + 10
		err := fdUsage.Get(otherPid)
		require.Error(t, err)
		assert.Equal(t, gosigar.ErrNotImplemented{runtime.GOOS}, err)
	})
}

func getRlimitViaShell(t *testing.T) (uint64, uint64) {
	out, err := exec.Command("/bin/sh", "-c", "ulimit -n -H").Output()
	require.NoError(t, err)
	hardLimit, err := parseRlimitOutput(string(out))
	require.NoError(t, err)

	out, err = exec.Command("/bin/sh", "-c", "ulimit -n -S").Output()
	require.NoError(t, err)
	softLimit, err := parseRlimitOutput(string(out))
	require.NoError(t, err)

	return hardLimit, softLimit
}

func parseRlimitOutput(output string) (uint64, error) {
	output = strings.TrimSpace(output)
	if output == "unlimited" {
		return syscall.RLIM_INFINITY, nil
	}
	return strconv.ParseUint(output, 10, 64)
}
