package cef_test

import (
	"context"
	"net"
	"time"

	"code.cloudfoundry.org/perm/cmd/contextx"
	"code.cloudfoundry.org/perm/pkg/logx"
	. "code.cloudfoundry.org/perm/pkg/logx/cef"
	"code.cloudfoundry.org/perm/pkg/logx/logxfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"google.golang.org/grpc/peer"
)

var _ = Describe("Logger", func() {
	var (
		logOutput *Buffer
		errLogger *logxfakes.FakeLogger

		logger *Logger

		p   *peer.Peer
		ctx context.Context
	)

	BeforeEach(func() {
		logOutput = NewBuffer()
		errLogger = new(logxfakes.FakeLogger)

		logger = NewLogger(logOutput, "cloud_foundry", "unittest", "0.0.1", "hook", 443, errLogger)

		p = &peer.Peer{
			Addr: &net.TCPAddr{IP: net.IPv4(1, 1, 1, 1), Port: 12345},
		}
		rt := time.Date(1999, 12, 31, 23, 59, 59, 59, time.UTC)
		ctx = contextx.WithReceiptTime(peer.NewContext(context.Background(), p), rt)
	})

	Describe("#Log", func() {
		Context("when all fields are available", func() {
			It("logs source and destination hostnames and ports", func() {
				logger.Log(ctx, "test-signature", "test-name")

				Eventually(logOutput).Should(Say("test-signature"))
				Eventually(logOutput).Should(Say("test-name"))
				Eventually(logOutput).Should(Say("dst=hook"))
				Eventually(logOutput).Should(Say("src=1.1.1.1"))
				Eventually(logOutput).Should(Say("dpt=443"))
				Eventually(logOutput).Should(Say("spt=12345"))
				Eventually(logOutput).Should(Say("rt=\"Dec 31 1999 23:59:59\""))
			})
		})

		Context("when the receipt time is not available", func() {
			It("does not log rt", func() {
				noReceiptContext := peer.NewContext(context.Background(), p)
				logger.Log(noReceiptContext, "test-signature", "test-name")

				Consistently(logOutput).ShouldNot(Say("rt="))
			})
		})

		Context("when there are custom extensions", func() {
			Context("when the custom extensions are valid", func() {
				var (
					customExtension1 logx.SecurityData
					customExtension2 logx.SecurityData
				)

				BeforeEach(func() {
					customExtension1 = logx.SecurityData{Key: "roleName", Value: "my-role-name"}
					customExtension2 = logx.SecurityData{Key: "roleBlame", Value: "my-role-blame"}
				})

				It("logs each extension", func() {
					logger.Log(ctx, "test-signature", "test-name", customExtension1, customExtension2)

					Eventually(logOutput).Should(Say("cs1Label=roleName"))
					Eventually(logOutput).Should(Say("cs1=my-role-name"))
					Eventually(logOutput).Should(Say("cs2Label=roleBlame"))
					Eventually(logOutput).Should(Say("cs2=my-role-blame"))
				})

				It("does not call error logger when no errors occur", func() {
					logger.Log(ctx, "test-signature", "test-name", customExtension1, customExtension2)

					Expect(errLogger.ErrorCallCount()).To(Equal(0))
				})

				Context("when the custom extension is a 'msg' pair", func() {
					It("does not use custom labels for the extension key pair", func() {
						msgExtension := logx.SecurityData{Key: "msg", Value: "some-msg"}
						logger.Log(ctx, "test-signature", "test-name", msgExtension)

						Eventually(logOutput).Should(Say("msg=some-msg"))

						Consistently(logOutput).ShouldNot(Say("cs1"))
					})
				})
			})

			Context("when the extension provided is invalid", func() {
				var (
					invalidExtension logx.SecurityData
					validExtension   logx.SecurityData
				)

				BeforeEach(func() {
					validExtension = logx.SecurityData{Key: "key", Value: "value"}
				})

				Context("because there is no key", func() {
					BeforeEach(func() {
						invalidExtension = logx.SecurityData{Value: "no-key"}
						logger.Log(ctx, "test-signature", "test-name", invalidExtension, validExtension)
					})

					It("should log that there were invalid extensions", func() {
						Consistently(logOutput).ShouldNot(Say("no-key"))

						Expect(errLogger.ErrorCallCount()).To(Equal(1))
						msg, err, _ := errLogger.ErrorArgsForCall(0)
						Expect(msg).To(Equal("invalid-cef-custom-extension"))
						Expect(err).To(MatchError("the extension key and/or value is empty"))
					})

					It("should still log correct extensions", func() {
						Eventually(logOutput).Should(Say("cs1Label=key"))
						Eventually(logOutput).Should(Say("cs1=value"))
					})
				})

				Context("because there is no value", func() {
					BeforeEach(func() {
						invalidExtension = logx.SecurityData{Key: "no-value"}
						logger.Log(ctx, "test-signature", "test-name", invalidExtension, validExtension)
					})

					It("should log that there were invalid extensions", func() {
						Consistently(logOutput).ShouldNot(Say("no-value"))

						Expect(errLogger.ErrorCallCount()).To(Equal(1))
						msg, err, _ := errLogger.ErrorArgsForCall(0)
						Expect(msg).To(Equal("invalid-cef-custom-extension"))
						Expect(err).To(MatchError("the extension key and/or value is empty"))
					})

					It("should still log correct extensions", func() {
						Eventually(logOutput).Should(Say("cs1Label=key"))
						Eventually(logOutput).Should(Say("cs1=value"))
					})
				})
			})

			Context("when there are more than 6 custom extensions", func() {
				var (
					customExtension1 logx.SecurityData
					customExtension2 logx.SecurityData
					customExtension3 logx.SecurityData
					customExtension4 logx.SecurityData
					customExtension5 logx.SecurityData
					customExtension6 logx.SecurityData
					extraExtension   logx.SecurityData
				)

				BeforeEach(func() {
					customExtension1 = logx.SecurityData{Key: "roleName", Value: "my-role-name"}
					customExtension2 = logx.SecurityData{Key: "roleBlame", Value: "my-role-blame"}
					customExtension3 = logx.SecurityData{Key: "roleDame", Value: "my-role-dame"}
					customExtension4 = logx.SecurityData{Key: "roleFame", Value: "my-role-fame"}
					customExtension5 = logx.SecurityData{Key: "msg", Value: "some-msg"}
					customExtension6 = logx.SecurityData{Key: "roleEndgame", Value: "my-role-endgame"}
					extraExtension = logx.SecurityData{Key: "dog", Value: "cat"}
				})

				It("should only log the first 6 custom extensions", func() {
					args := []logx.SecurityData{
						customExtension1,
						customExtension2,
						customExtension3,
						customExtension4,
						customExtension5,
						customExtension6,
						extraExtension,
					}
					logger.Log(ctx, "test-signature", "test-name", args...)

					Eventually(logOutput).Should(Say("cs1Label=roleName"))
					Eventually(logOutput).Should(Say("cs1=my-role-name"))
					Eventually(logOutput).Should(Say("cs2Label=roleBlame"))
					Eventually(logOutput).Should(Say("cs2=my-role-blame"))
					Eventually(logOutput).Should(Say("cs3Label=roleDame"))
					Eventually(logOutput).Should(Say("cs3=my-role-dame"))
					Eventually(logOutput).Should(Say("cs4Label=roleFame"))
					Eventually(logOutput).Should(Say("cs4=my-role-fame"))
					Eventually(logOutput).Should(Say("msg=some-msg"))
					Eventually(logOutput).Should(Say("cs5Label=roleEndgame"))
					Eventually(logOutput).Should(Say("cs5=my-role-endgame"))

					Consistently(logOutput).ShouldNot(Say("cs6Label=dog"))
					Consistently(logOutput).ShouldNot(Say("cs6=cat"))

					Expect(errLogger.ErrorCallCount()).To(Equal(1))
					msg, err, _ := errLogger.ErrorArgsForCall(0)
					Expect(msg).To(Equal("invalid-cef-custom-extension"))
					Expect(err).To(MatchError("cannot provide more than 6 custom extensions"))
				})

				Context("when there is also as an invalid extension", func() {
					var badExtension logx.SecurityData

					BeforeEach(func() {
						badExtension = logx.SecurityData{Value: "no-key"}
					})

					It("logs both errors in the message", func() {
						args := []logx.SecurityData{
							badExtension,
							customExtension1,
							customExtension2,
							customExtension3,
							customExtension4,
							customExtension5,
							customExtension6,
							extraExtension,
						}
						logger.Log(ctx, "test-signature", "test-name", args...)

						Consistently(logOutput).ShouldNot(Say("no-key"))

						Eventually(logOutput).Should(Say("cs5Label=roleEndgame"))
						Eventually(logOutput).Should(Say("cs5=my-role-endgame"))

						Expect(errLogger.ErrorCallCount()).To(Equal(2))

						msg, err, _ := errLogger.ErrorArgsForCall(0)
						Expect(msg).To(Equal("invalid-cef-custom-extension"))
						Expect(err).To(MatchError("the extension key and/or value is empty"))

						msg, err, _ = errLogger.ErrorArgsForCall(1)
						Expect(msg).To(Equal("invalid-cef-custom-extension"))
						Expect(err).To(MatchError("cannot provide more than 6 custom extensions"))
					})
				})
			})
		})
	})
})
