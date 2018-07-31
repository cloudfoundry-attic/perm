package cef

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"strconv"

	"code.cloudfoundry.org/perm/cmd/contextx"
	"code.cloudfoundry.org/perm/pkg/logx"
	"github.com/xoebus/ceflog"
	"google.golang.org/grpc/peer"
)

const CEFTimeFormat = "Jan 2 2006 15:04:05"

type Vendor string
type Product string
type Version string
type Hostname string

type Logger struct {
	logger   *ceflog.Logger
	hostname string
	destPort int
}

func NewLogger(writer io.Writer, vendor Vendor, product Product, version Version, hostname Hostname, destPort int) *Logger {
	return &Logger{
		logger:   ceflog.New(writer, string(vendor), string(product), string(version)),
		hostname: string(hostname),
		destPort: destPort,
	}
}

func (l *Logger) Log(ctx context.Context, signature string, name string, args ...logx.SecurityData) {
	var srcAddr net.IP
	var srcPort int

	peer, ok := peer.FromContext(ctx)
	if ok {
		switch addr := peer.Addr.(type) {
		case *net.TCPAddr:
			srcAddr = addr.IP
			srcPort = addr.Port
		}
	}

	extension := ceflog.Extension{
		ceflog.Pair{Key: "dst", Value: l.hostname},
		ceflog.Pair{Key: "src", Value: srcAddr.String()},
		ceflog.Pair{Key: "dpt", Value: strconv.FormatInt(int64(l.destPort), 10)},
		ceflog.Pair{Key: "spt", Value: strconv.FormatInt(int64(srcPort), 10)},
	}

	if rt, ok := contextx.ReceiptTimeFromContext(ctx); ok {
		extension = append(extension, ceflog.Pair{Key: "rt", Value: fmt.Sprintf("\"%s\"", rt.Format(CEFTimeFormat))})
	}

	var msgBuffer bytes.Buffer

	if len(args) == 1 {
		ce := args[0]
		if ce.Key == "" || ce.Value == "" {
			msgBuffer.WriteString("ERROR:invalid-custom-extension;")
		} else {
			extension = append(extension, ceflog.Pair{Key: ce.Key, Value: ce.Value})
		}
	} else {
		counter := 1
		invalidFound := false

		for _, ce := range args {
			if ce.Key == "" || ce.Value == "" && invalidFound == false {
				msgBuffer.WriteString("ERROR:invalid-custom-extension;")
				invalidFound = true
			} else {
				extension = append(extension, ceflog.Pair{Key: fmt.Sprintf("cs%dLabel", counter), Value: ce.Key})
				extension = append(extension, ceflog.Pair{Key: fmt.Sprintf("cs%d", counter), Value: ce.Value})
				counter++
				if counter > 6 {
					msgBuffer.WriteString("ERROR:too-many-custom-extensions;")
					break
				}
			}
		}
	}

	if msgBuffer.String() != "" {
		extension = append(extension, ceflog.Pair{Key: "msg", Value: msgBuffer.String()})
	}

	l.logger.LogEvent(signature, name, 0, extension)
}
