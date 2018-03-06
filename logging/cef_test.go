package logging_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/perm/cmd/contextx"
	. "code.cloudfoundry.org/perm/logging"
	"context"
	"github.com/onsi/gomega/gbytes"
	"google.golang.org/grpc/peer"
	"net"
	"time"
)

var _ = Describe("Logging", func() {
	var securityLogger *CEFLogger
	var logOutput *gbytes.Buffer
	var rt = time.Date(1999, 12, 31, 23, 59, 59, 59, time.UTC)
	BeforeEach(func() {
		logOutput = gbytes.NewBuffer()
		securityLogger = NewCEFLogger(logOutput, "cloud_foundry", "unittest", "0.0.1", "hook", 443)
	})

	Describe("#Log", func() {
		Describe("when all fields are available", func() {
			It("logs source and destination hostnames and ports", func() {
				p := &peer.Peer{}
				p.Addr = &net.TCPAddr{IP: net.IPv4(1, 1, 1, 1), Port: 12345}
				ctx := contextx.WithReceiptTime(peer.NewContext(context.Background(), p), rt)
				securityLogger.Log(ctx, "test-signature", "test-name")

				Eventually(logOutput).Should(gbytes.Say("test-signature"))
				Eventually(logOutput).Should(gbytes.Say("test-name"))
				Eventually(logOutput).Should(gbytes.Say("dst=hook"))
				Eventually(logOutput).Should(gbytes.Say("src=1.1.1.1"))
				Eventually(logOutput).Should(gbytes.Say("dpt=443"))
				Eventually(logOutput).Should(gbytes.Say("spt=12345"))
				Eventually(logOutput).Should(gbytes.Say("rt=\"Dec 31 1999 23:59:59\""))
			})
		})

		Describe("when the receipt time is not available", func() {
			It("does not log rt", func() {
				p := &peer.Peer{}
				p.Addr = &net.TCPAddr{IP: net.IPv4(1, 1, 1, 1), Port: 12345}
				ctx := peer.NewContext(context.Background(), p)
				securityLogger.Log(ctx, "test-signature", "test-name")

				output := string(logOutput.Contents())
				Expect(output).NotTo(ContainSubstring("rt="))
			})
		})
	})

})
