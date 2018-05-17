package permauth_test

import (
	"net/http"

	. "code.cloudfoundry.org/perm/pkg/permauth"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

const UAAResponse = `
{
  "keys": [
    {
      "kty": "RSA",
      "e": "AQAB",
      "use": "sig",
      "kid": "sha2-2017-01-20-key",
      "alg": "RS256",
      "value": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyH6kYCP29faDAUPKtei3\nV/Zh8eCHyHRDHrD0iosvgHuaakK1AFHjD19ojuPiTQm8r8nEeQtHb6mDi1LvZ03e\nEWxpvWwFfFVtCyBqWr5wn6IkY+ZFXfERLn2NCn6sMVxcFV12sUtuqD+jrW8MnTG7\nhofQqxmVVKKsZiXCvUSzfiKxDgoiRuD3MJSoZ0nQTHVmYxlFHuhTEETuTqSPmOXd\n/xJBVRi5WYCjt1aKRRZEz04zVEBVhVkr2H84qcVJHcfXFu4JM6dg0nmTjgd5cZUN\ncwA1KhK2/Qru9N0xlk9FGD2cvrVCCPWFPvZ1W7U7PBWOSBBH6GergA+dk2vQr7Ho\nlQIDAQAB\n-----END PUBLIC KEY-----",
      "n": "AMh-pGAj9vX2gwFDyrXot1f2YfHgh8h0Qx6w9IqLL4B7mmpCtQBR4w9faI7j4k0JvK_JxHkLR2-pg4tS72dN3hFsab1sBXxVbQsgalq-cJ-iJGPmRV3xES59jQp-rDFcXBVddrFLbqg_o61vDJ0xu4aH0KsZlVSirGYlwr1Es34isQ4KIkbg9zCUqGdJ0Ex1ZmMZRR7oUxBE7k6kj5jl3f8SQVUYuVmAo7dWikUWRM9OM1RAVYVZK9h_OKnFSR3H1xbuCTOnYNJ5k44HeXGVDXMANSoStv0K7vTdMZZPRRg9nL61Qgj1hT72dVu1OzwVjkgQR-hnq4APnZNr0K-x6JU"
    }
  ]
}
`

const UAAIssuerResponse = `
{
  "issuer": "https://uaa.cloudfoundry.org/oauth/token",
  "authorization_endpoint": "https://login.run.pivotal.io/oauth/authorize",
  "token_endpoint": "https://login.run.pivotal.io/oauth/token"
}
`

var _ = Describe("Helper functions for JWT verification", func() {
	var server *ghttp.Server

	BeforeEach(func() {
		server = ghttp.NewServer()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("#GetUAAPubKey", func() {
		Context("when the UAA url is functional", func() {
			Context("when at least one key is present", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/token_keys"),
							ghttp.RespondWith(http.StatusOK, UAAResponse),
						),
					)
				})
				It("should return the public key", func() {
					pubkey, err := GetUAAPubKey(server.URL())
					Expect(err).NotTo(HaveOccurred())
					Expect(pubkey).To(ContainSubstring("-----BEGIN PUBLIC KEY-----"))
				})
			})
			Context("when no keys are present in the response", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/token_keys"),
							ghttp.RespondWith(http.StatusOK, `{"keys":[]}`),
						),
					)
				})
				It("should return an error", func() {
					_, err := GetUAAPubKey(server.URL())
					Expect(err).To(MatchError("No public key found on the UAA /token_keys endpoint"))
				})
			})
		})
		Context("when the UAA url is dysfunctional", func() {
			Context("when the endpoint does not exist", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/token_keys"),
							ghttp.RespondWith(http.StatusNotFound, `{}`),
						),
					)
				})
				It("should return an error", func() {
					_, err := GetUAAPubKey(server.URL())
					Expect(err).To(MatchError("No public key found on the UAA /token_keys endpoint"))
				})
			})
			Context("the response JSON isn't valid", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/token_keys"),
							ghttp.RespondWith(http.StatusOK, `{"hello"}`),
						),
					)
				})
				It("should return an error", func() {
					_, err := GetUAAPubKey(server.URL())
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	Describe("#GetUAAIssuer", func() {
		Context("when the UAA url is functional", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/.well-known/openid-configuration"),
						ghttp.RespondWith(http.StatusOK, UAAIssuerResponse),
					),
				)
			})
			It("should return the issuer", func() {
				issuer, err := GetUAAIssuer(server.URL())
				Expect(err).NotTo(HaveOccurred())
				Expect(issuer).To(Equal("https://uaa.cloudfoundry.org/oauth/token"))
			})
		})
		Context("when the UAA url is dysfunctional", func() {
			Context("when the endpoint does not exist", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/.well-known/openid-configuration"),
							ghttp.RespondWith(http.StatusNotFound, `{}`),
						),
					)
				})
				It("should return an error", func() {
					_, err := GetUAAIssuer(server.URL())
					Expect(err).To(MatchError("No issuer found on the UAA /.well-known/openid-configuration endpoint"))
				})
			})
			Context("the response JSON isn't valid", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/.well-known/openid-configuration"),
							ghttp.RespondWith(http.StatusOK, `{"hello"}`),
						),
					)
				})
				It("should return an error", func() {
					_, err := GetUAAIssuer(server.URL())
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})
