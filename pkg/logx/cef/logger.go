package cef

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"

	"code.cloudfoundry.org/perm/cmd/contextx"
	"code.cloudfoundry.org/perm/pkg/logx"
	"github.com/xoebus/ceflog"
	"google.golang.org/grpc/peer"
)

const (
	CEFTimeFormat             = "Jan 2 2006 15:04:05"
	invalidCEFCustomExtension = "invalid-cef-custom-extension"
)

type Vendor string
type Product string
type Version string
type Hostname string

type Logger struct {
	logger    *ceflog.Logger
	hostname  string
	destPort  int
	errLogger logx.Logger
}

func NewLogger(writer io.Writer, vendor Vendor, product Product, version Version, hostname Hostname, destPort int, errLogger logx.Logger) *Logger {
	return &Logger{
		logger:    ceflog.New(writer, string(vendor), string(product), string(version)),
		hostname:  string(hostname),
		destPort:  destPort,
		errLogger: errLogger,
	}
}

func (l *Logger) Log(ctx context.Context, signature string, name string, args ...logx.SecurityData) {
	var (
		srcAddr net.IP
		srcPort int
	)

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

	counter := 1
	invalidFound := false

	for _, ce := range args {
		if ce.Key == "" || ce.Value == "" && invalidFound == false {
			l.errLogger.Error(invalidCEFCustomExtension, errors.New("the extension key and/or value is empty"))
			invalidFound = true
		} else {
			extension = append(extension, ceflog.Pair{Key: fmt.Sprintf("cs%dLabel", counter), Value: fmt.Sprintf("%s", ce.Key)})
			extension = append(extension, ceflog.Pair{Key: fmt.Sprintf("cs%d", counter), Value: fmt.Sprintf("%s", ce.Value)})
			counter++
			if counter > 6 {
				l.errLogger.Error(invalidCEFCustomExtension, errors.New("cannot provide more than 6 custom extensions"))
				break
			}
		}
	}

	l.logger.LogEvent(signature, name, 0, extension)
}
