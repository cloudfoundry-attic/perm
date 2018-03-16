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
		Context("when all fields are available", func() {
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

		Context("when the receipt time is not available", func() {
			It("does not log rt", func() {
				p := &peer.Peer{}
				p.Addr = &net.TCPAddr{IP: net.IPv4(1, 1, 1, 1), Port: 12345}
				ctx := peer.NewContext(context.Background(), p)
				securityLogger.Log(ctx, "test-signature", "test-name")

				output := string(logOutput.Contents())
				Expect(output).NotTo(ContainSubstring("rt="))
			})
		})

		Context("when there are custom extensions", func() {
			var (
				customExtension1 CustomExtension
				customExtension2 CustomExtension
			)
			BeforeEach(func() {
				customExtension1 = CustomExtension{Key: "roleName", Value: "my-role-name"}
				customExtension2 = CustomExtension{Key: "roleBlame", Value: "my-role-blame"}
			})

			It("logs each extension", func() {
				p := &peer.Peer{}
				p.Addr = &net.TCPAddr{IP: net.IPv4(1, 1, 1, 1), Port: 12345}
				ctx := peer.NewContext(context.Background(), p)
				securityLogger.Log(ctx, "test-signature", "test-name", customExtension1, customExtension2)

				Eventually(logOutput).Should(gbytes.Say("cs1Label=roleName"))
				Eventually(logOutput).Should(gbytes.Say("cs1=my-role-name"))
				Eventually(logOutput).Should(gbytes.Say("cs2Label=roleBlame"))
				Eventually(logOutput).Should(gbytes.Say("cs2=my-role-blame"))
			})

			Context("when the extension provided is invalid", func() {
				var invalidExtension CustomExtension
				var validExtension CustomExtension
				BeforeEach(func() {
					validExtension = CustomExtension{Key: "key", Value: "value"}
				})

				Context("because there is no key", func() {
					BeforeEach(func() {
						invalidExtension = CustomExtension{Value: "no-key"}

						p := &peer.Peer{}
						p.Addr = &net.TCPAddr{IP: net.IPv4(1, 1, 1, 1), Port: 12345}
						ctx := peer.NewContext(context.Background(), p)
						securityLogger.Log(ctx, "test-signature", "test-name", invalidExtension, validExtension)
					})

					It("should log that there were invalid extensions", func() {
						Eventually(logOutput).Should(gbytes.Say("msg=ERROR:invalid-custom-extension;"))
						Consistently(logOutput).ShouldNot(gbytes.Say("cs1=no-key"))
					})

					It("should still log correct extensions", func() {
						Eventually(logOutput).Should(gbytes.Say("cs1Label=key"))
						Eventually(logOutput).Should(gbytes.Say("cs1=value"))
					})
				})

				Context("because there is no value", func() {
					BeforeEach(func() {
						invalidExtension = CustomExtension{Key: "noValue"}

						p := &peer.Peer{}
						p.Addr = &net.TCPAddr{IP: net.IPv4(1, 1, 1, 1), Port: 12345}
						ctx := peer.NewContext(context.Background(), p)
						securityLogger.Log(ctx, "test-signature", "test-name", invalidExtension, validExtension)
					})

					It("should log that there were invalid extensions", func() {
						p := &peer.Peer{}
						p.Addr = &net.TCPAddr{IP: net.IPv4(1, 1, 1, 1), Port: 12345}
						ctx := peer.NewContext(context.Background(), p)
						securityLogger.Log(ctx, "test-signature", "test-name", invalidExtension)

						Eventually(logOutput).Should(gbytes.Say("msg=ERROR:invalid-custom-extension"))
						Consistently(logOutput).ShouldNot(gbytes.Say("cs1Label=noValue"))
					})

					It("should still log correct extensions", func() {
						Eventually(logOutput).Should(gbytes.Say("cs1Label=key"))
						Eventually(logOutput).Should(gbytes.Say("cs1=value"))
					})
				})

			})

			Context("when there are more than 6 custom extensions", func() {
				var (
					customExtension3 CustomExtension
					customExtension4 CustomExtension
					customExtension5 CustomExtension
					customExtension6 CustomExtension
					extraExtension   CustomExtension
				)

				BeforeEach(func() {
					customExtension3 = CustomExtension{Key: "roleDame", Value: "my-role-dame"}
					customExtension4 = CustomExtension{Key: "roleFame", Value: "my-role-fame"}
					customExtension5 = CustomExtension{Key: "roleBeshame", Value: "my-role-beshame"}
					customExtension6 = CustomExtension{Key: "roleEndgame", Value: "my-role-endgame"}
					extraExtension = CustomExtension{Key: "dog", Value: "cat"}
				})

				It("should only log the first 6 custom extensions", func() {
					p := &peer.Peer{}
					p.Addr = &net.TCPAddr{IP: net.IPv4(1, 1, 1, 1), Port: 12345}
					ctx := peer.NewContext(context.Background(), p)
					args := []CustomExtension{
						customExtension1,
						customExtension2,
						customExtension3,
						customExtension4,
						customExtension5,
						customExtension6,
						extraExtension,
					}

					securityLogger.Log(ctx, "test-signature", "test-name", args...)

					Eventually(logOutput).Should(gbytes.Say("cs1Label=roleName"))
					Eventually(logOutput).Should(gbytes.Say("cs1=my-role-name"))
					Eventually(logOutput).Should(gbytes.Say("cs2Label=roleBlame"))
					Eventually(logOutput).Should(gbytes.Say("cs2=my-role-blame"))
					Eventually(logOutput).Should(gbytes.Say("cs3Label=roleDame"))
					Eventually(logOutput).Should(gbytes.Say("cs3=my-role-dame"))
					Eventually(logOutput).Should(gbytes.Say("cs4Label=roleFame"))
					Eventually(logOutput).Should(gbytes.Say("cs4=my-role-fame"))
					Eventually(logOutput).Should(gbytes.Say("cs5Label=roleBeshame"))
					Eventually(logOutput).Should(gbytes.Say("cs5=my-role-beshame"))
					Eventually(logOutput).Should(gbytes.Say("cs6Label=roleEndgame"))
					Eventually(logOutput).Should(gbytes.Say("cs6=my-role-endgame"))

					Eventually(logOutput).Should(gbytes.Say("msg=ERROR:too-many-custom-extensions;"))
					Consistently(logOutput).ShouldNot(gbytes.Say("cs7Label=dog"))
					Consistently(logOutput).ShouldNot(gbytes.Say("cs7=cat"))
				})

				Context("when there is also as an invalid extension", func() {
					var badExtension CustomExtension
					BeforeEach(func() {
						badExtension = CustomExtension{Value: "no-key"}
					})

					It("logs both errors in the message", func() {
						p := &peer.Peer{}
						p.Addr = &net.TCPAddr{IP: net.IPv4(1, 1, 1, 1), Port: 12345}
						ctx := peer.NewContext(context.Background(), p)
						args := []CustomExtension{
							badExtension,
							customExtension1,
							customExtension2,
							customExtension3,
							customExtension4,
							customExtension5,
							customExtension6,
							extraExtension,
						}

						securityLogger.Log(ctx, "test-signature", "test-name", args...)
						Consistently(logOutput).ShouldNot(gbytes.Say("cs1=no-key"))
						Eventually(logOutput).Should(gbytes.Say("cs6Label=roleEndgame"))
						Eventually(logOutput).Should(gbytes.Say("cs6=my-role-endgame"))
						Eventually(logOutput).Should(gbytes.Say("msg=ERROR:invalid-custom-extension;ERROR:too-many-custom-extensions;"))
					})
				})
			})

		})
	})
})
