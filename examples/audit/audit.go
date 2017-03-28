// +build linux

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/elastic/gosigar/sys/linux"
	"github.com/pkg/errors"
)

var (
	fs     = flag.NewFlagSet("audit", flag.ExitOnError)
	debug  = fs.Bool("d", false, "enable debug output to stderr")
	diag   = fs.String("diag", "", "dump raw information from kernel to file")
	pretty = fs.Bool("pretty", false, "pretty print json output")
)

func enableLogger() {
	log.SetOutput(os.Stderr)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	})
}

func main() {
	fs.Parse(os.Args[1:])

	if *debug {
		enableLogger()
	}

	if err := read(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func read() error {
	if os.Geteuid() != 0 {
		return errors.New("you must be root to receive audit data")
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

	log.Debugln("starting netlink client")
	client, err := linux.NewAuditClient(diagWriter)
	if err != nil {
		return err
	}

	log.Debugln("sending message to kernel registering our PID as the audit daemon")
	if err = client.SetPortID(0); err != nil {
		return errors.Wrap(err, "failed to set audit PID")
	}

	for {
		m, err := client.Receive(false)
		if err != nil {
			return errors.Wrap(err, "receive failed")
		}

		if m.MessageType < 1300 || m.MessageType >= 2100 {
			continue
		}

		// Ignore AUDIT_EOE.
		if m.MessageType == 1320 {
			continue
		}

		event := map[string]interface{}{
			"type": m.MessageType,
			"msg":  string(m.RawData),
		}

		var jsonEvent []byte
		if *pretty {
			jsonEvent, err = json.MarshalIndent(event, "", "  ")
		} else {
			jsonEvent, err = json.Marshal(event)
		}
		if err != nil {
			log.WithError(err).Warn("Failed to marshal event to JSON")
		}

		fmt.Println(string(jsonEvent))
	}

	return nil
}
