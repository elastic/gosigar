// +build linux

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/elastic/gosigar/sys/linux"
)

var (
	fs    = flag.NewFlagSet("ss", flag.ExitOnError)
	debug = fs.Bool("d", false, "enable debug output to stderr")
	ipv6  = fs.Bool("6", false, "display only IP version 6 sockets")
	v1    = fs.Bool("v1", false, "send inet_diag_msg v1 instead of v2")
	diag  = fs.String("diag", "", "dump raw information about TCP sockets to FILE")
)

func main() {
	log.SetFlags(0)
	fs.Parse(os.Args[1:])

	if !*debug {
		log.SetOutput(ioutil.Discard)
	}

	if err := sockets(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func sockets() error {
	// Set address family based on flags. The requested address family only
	// works with inet_diag_req_v2. v1 returns all tcp sockets.
	af := linux.AF_INET
	if *ipv6 {
		af = linux.AF_INET6
	}

	// For debug purposes allow for sending either inet_diag_req and inet_diag_req_v2.
	var req syscall.NetlinkMessage
	if *v1 {
		req = linux.NewInetDiagReq()
	} else {
		req = linux.NewInetDiagReqV2(af)
	}

	// Write netlink response to a file for further analysis or for writing
	// tests cases.
	var diagWriter io.Writer
	if *diag != "" {
		f, err := os.OpenFile(*diag, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer f.Close()
		diagWriter = f
	}

	log.Println("sending netlink request")
	msgs, err := linux.NetlinkInetDiagWithBuf(req, nil, diagWriter)
	if err != nil {
		return err
	}
	log.Printf("received %d inet_diag_msg responses", len(msgs))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join([]string{
		"State",
		"Recv-Q",
		"Send-Q",
		"Local Address:Port",
		"Remote Address:Port",
		"UID",
		"Inode",
		"PID/Program",
	}, "\t"))
	defer w.Flush()

	inodeToPid := mapInodesToPid()

	for _, diag := range msgs {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v:%v\t%v:%v\t%v\t%v\t%v\n",
			linux.TCPState(diag.State), diag.RQueue, diag.WQueue,
			diag.SrcIP().String(), diag.SrcPort(),
			diag.DstIP().String(), diag.DstPort(),
			diag.UID, diag.Inode, inodeToPid[diag.Inode])
	}

	return nil
}

func mapInodesToPid() (ret map[uint32]string) {
	ret = map[uint32]string{}

	fd, err := os.Open("/proc")
	if err != nil {
		fmt.Printf("Error opening /proc: %v", err)
	}
	defer fd.Close()

	dirContents, err := fd.Readdirnames(0)
	if err != nil {
		fmt.Printf("Error reading files in /proc: %v", err)
	}

	for _, pid := range dirContents {
		_, err := strconv.ParseUint(pid, 10, 32)
		if err != nil {
			// exclude files with a not numeric name. We only want to access pid directories
			continue
		}

		pidDir, err := os.Open("/proc/" + string(pid) + "/fd/")
		if err != nil {
			// ignore errors:
			//   - missing directory, pid has already finished
			//   - permission denied
			continue
		}

		fds, err := pidDir.Readdirnames(0)
		if err != nil {
			continue
		}

		for _, fd := range fds {
			link, err := os.Readlink("/proc/" + string(pid) + "/fd/" + fd)
			if err != nil {
				continue
			}

			var inode uint32

			_, err = fmt.Sscanf(link, "socket:[%d]", &inode)
			if err != nil {
				// this inode is not a socket
				continue
			}

			ret[inode] = pid
		}
	}

	return ret
}
