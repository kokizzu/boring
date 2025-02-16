package daemon

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/alebeck/boring/internal/tunnel"
)

const (
	Flag        = "--daemon"
	sockName    = "boringd.sock"
	logFileName = "boringd.log"
	initWait    = 2 * time.Millisecond
)

type CmdKind int

const (
	Nop CmdKind = iota
	Open
	Close
	List
)

var cmdKindNames = map[CmdKind]string{
	Nop:   "Nop",
	Open:  "Open",
	Close: "Close",
	List:  "List",
}

var sock, logFile, executableFile string

func init() {
	if sock = os.Getenv("BORING_SOCK"); sock == "" {
		sock = filepath.Join(os.TempDir(), sockName)
	}
	if logFile = os.Getenv("BORING_LOG_FILE"); logFile == "" {
		logFile = filepath.Join(os.TempDir(), logFileName)
	}
}

func (k CmdKind) String() string {
	n, ok := cmdKindNames[k]
	if !ok {
		return fmt.Sprintf("%d", int(k))
	}
	return n
}

type Cmd struct {
	Kind   CmdKind           `json:"kind"`
	Tunnel tunnel.TunnelDesc `json:"tunnel,omitempty"`
}

type Resp struct {
	Success bool                         `json:"success"`
	Error   string                       `json:"error"`
	Tunnels map[string]tunnel.TunnelDesc `json:"tunnels"`
}

// Ensure starts the daemon if it is not already running.
func Ensure(ctx context.Context) error {
	starting := false
	wait := time.NewTimer(0.)
	waitTime := initWait

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-wait.C:
			if conn, err := Connect(); err == nil {
				go func() { conn.Close() }()
				return nil
			}
			if !starting {
				if err := startDaemon(); err != nil {
					return fmt.Errorf("Failed to start daemon: %v", err)
				}
				starting = true
			}
			wait.Reset(waitTime)
			waitTime *= 2
		}
	}
}

func Connect() (net.Conn, error) {
	return net.Dial("unix", sock)
}

func startDaemon() error {
	ex, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine executable path: %v", err)
	}
	if err = runDaemonized(ex, Flag); err != nil {
		return err
	}
	return nil
}
