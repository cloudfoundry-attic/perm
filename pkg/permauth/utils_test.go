package permauth_test

import (
	"fmt"

	"net/http"

	"code.cloudfoundry.org/perm/pkg/permauth"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("GetOIDCIssuer", func() {
	var (
		server   *ghttp.Server
		tokenURL string
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		tokenURL = fmt.Sprintf("%s/oauth/token", server.URL())
	})

	AfterEach(func() {
		server.Close()
	})

	It("fetches the issuer from .well-known/openid-configuration", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/oauth/token/.well-known/openid-configuration"),
				ghttp.RespondWith(200, `{"issuer": "foo"}`),
			),
		)
		issuer, err := permauth.GetOIDCIssuer(http.DefaultClient, tokenURL)
		Expect(err).NotTo(HaveOccurred())
		Expect(issuer).To(Equal("foo"))
	})

	It("returns an error on bad get status", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/oauth/token/.well-known/openid-configuration"),
				ghttp.RespondWith(404, `{"error": "not found"}`),
			),
		)
		_, err := permauth.GetOIDCIssuer(http.DefaultClient, tokenURL)
		Expect(err).To(MatchError("HTTP bad response: 404 Not Found"))
	})

	It("returns an error on unparseable endpoint content", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/oauth/token/.well-known/openid-configuration"),
				ghttp.RespondWith(200, `{"issuer": "foo....`),
			),
		)
		_, err := permauth.GetOIDCIssuer(http.DefaultClient, tokenURL)
		Expect(err).To(MatchError("unexpected EOF"))
	})
})
