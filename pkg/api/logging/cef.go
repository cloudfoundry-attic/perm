package logging

import (
	"bytes"
	"code.cloudfoundry.org/perm/cmd/contextx"
	"context"
	"fmt"
	"github.com/xoebus/ceflog"
	"google.golang.org/grpc/peer"
	"io"
	"net"
	"strconv"
)

type Vendor string
type Product string
type Version string
type Hostname string
type CEFLogger struct {
	logger   *ceflog.Logger
	hostname string
	destPort int
}

const CEFTimeFormat = "Jan 2 2006 15:04:05"

func NewCEFLogger(writer io.Writer, vendor Vendor, product Product, version Version, hostname Hostname, destPort int) *CEFLogger {
	return &CEFLogger{
		logger:   ceflog.New(writer, string(vendor), string(product), string(version)),
		hostname: string(hostname),
		destPort: destPort,
	}
}

func (l *CEFLogger) Log(ctx context.Context, signature string, name string, args ...CustomExtension) {
	peer, ok := peer.FromContext(ctx)

	var srcAddr net.IP
	var srcPort int
	if ok {
		switch addr := peer.Addr.(type) {
		case *net.TCPAddr:
			srcAddr = addr.IP
			srcPort = addr.Port
		default:

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

	counter := 1
	invalidFound := false
	var msgBuffer bytes.Buffer
	for _, ce := range args {
		if ce.Key == "" || ce.Value == "" && invalidFound == false {
			msgBuffer.WriteString("ERROR:invalid-custom-extension;")
			invalidFound = true
		} else {
			extension = append(extension, ceflog.Pair{Key: fmt.Sprintf("cs%dLabel", counter), Value: fmt.Sprintf("%s", ce.Key)})
			extension = append(extension, ceflog.Pair{Key: fmt.Sprintf("cs%d", counter), Value: fmt.Sprintf("%s", ce.Value)})
			counter++
			if counter > 6 {
				msgBuffer.WriteString("ERROR:too-many-custom-extensions;")
				break
			}
		}
	}
	if msgBuffer.String() != "" {
		extension = append(extension, ceflog.Pair{Key: "msg", Value: msgBuffer.String()})
	}
	l.logger.LogEvent(signature, name, 0, extension)
}

type CustomExtension struct {
	Key   string
	Value string
}
