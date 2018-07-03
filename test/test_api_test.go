package perm_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"time"

	"code.cloudfoundry.org/perm/pkg/api"
	"code.cloudfoundry.org/perm/pkg/api/rpc/rpcfakes"
	"code.cloudfoundry.org/perm/pkg/monitor/monitorfakes"
	"code.cloudfoundry.org/perm/pkg/perm"
	"code.cloudfoundry.org/perm/pkg/permstats"
	oidc "github.com/coreos/go-oidc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/oauth2"
	jose "gopkg.in/square/go-jose.v2"
)

const (
	validPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA918Nv+kmlGF1uz2MMJaJ8TFXzV9E5bFVVotKHxHl1HEQhjxF
FVLkBiagjh61pu/eC5tjDdyA0gkYWpLfEvAnAatV/t+HxggjGb8fpA0babKztGfz
RG59GmquRqzQQFwmpr/NClLdCcg0npmStJeGCFh0PRH/TVVClDs6dcUsoIDFSjvL
S9SIxCF0p/TLxiFF3m/5y+6ODN7S9CdasND1Rjbbg4+c20krXO/YTtKyfjc9MS3U
7EXaVgu5KolTDyUSITwVNAgEhSPTKp+IxDx2fgZ4l6bydN+xHI0cE4cX4d10lb5G
tD6vK8UsukW7QfTB5xLOUX7lJc6xNSka4xDSlQIDAQABAoIBABsX9B+S38Dcs9Jg
OVyRAGbEasN5rcgiliA2fVXN1ghgodix/TcKryLlVCx8vJSeLQnEaSL5hbp7eIlj
EL+4Qe1y4KZbwTk1ZvLI9iQ3s0ruYbRetkxGdblQ+emPv/dsoGcfFswMq10I6op8
c48IEYwUdBbEQ9wqfHJT0mFXyT2C5fhD3A0jSVTzKHec1QNycJwxIEmqmJY5RuPx
P4S5zv5BJ4w3jYjV0ILQOjecV5rGQD4gHzD4+r0bSrTQCbljm9/DOv6r1EBpTz8r
uf0vbCX1YCNlQ6rBf3MgW+K3z6egYoDNiVETVoydI5qeMvtCVEqnbMRXMp1K0hfM
XFrJFDkCgYEA/iAkN0DoF6xbR0Rjg3PgNZJDLI0xQLf3Xn9n86139LhVQvuq1bDc
/AYkbji1BvXypAiE9Cp/tJ+z/gNVQAD6V8gum57moSC7vgSa4wb0h+WD7Uf0tDCT
9cLFckypoMzCU9+MY7j8PLO11phIxC7zOwAZKbee7NYk3n2kBnYrPdsCgYEA+TIo
ixdJgmra4i49BUMUn6V9B6Kj1B1FtlITqUdZJ6PqqWMMciCsecjAiR7MT8Nnp+kX
rSWg7IDYRdAnyWSR96RWNByOD5Jt8ZwIdTP+WB0OasH8mf7Hn4KWRHN+A3pkToiv
fiv0Nmlgko6hG4rvSGIHJRK1I3Kif4Zt5HOM9E8CgYEAouPHUwNvwXzhJVVY1DG6
TZxrImt+XpWNIi1YXIGcmmhtfnoCjubHP2RQhbYjk0qjNTGgx0FWiliz7uYEBvqZ
fRr7hRTdj/qDXNFm1o7mvxUG81lkKPvaW3V1SkaJlGCrT0fDnUg9pksrC1qhid7u
Was+ddcVL4o0J8kxElM8dHECgYEAzbBwNLbpD0RCHbXK2mAPUuNXO4ksrzXmR+Kh
pfVlisnLNTuzlzSPCQsCmWwZerExCzDkQSAxH2YOnjl9zcc8kOtN2D/FpubX5zlC
5fMfuv1o3Af5B+d8QJaakC/AUQCicQxzxrJjJtJ+Sxp9su1QKy/288voRjUmGhsM
9CfIrhMCgYB4DOuJEP99zbclE1pnPIiBRmPJdE6nTuBmBu9C0jTsI58YMl5Kddbm
LG3ujlOKoS6FpkV1mH+D6GsfNwnbP9JrtXoBwMaQIGHQtTI7Tdxe4GFPgBcXq+ZH
89SLmAbBj+ca9wx0YTo6/isBzMJNLPFThbscXjReS1K5UZ/EN87uFw==
-----END RSA PRIVATE KEY-----`

	foreignPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQDQCRZ9s6DCFpp6rk+wZJ7mrkCby8r+h9oY/7ouc7VK6hp88nw7
J9y3xqaIQSu/Tg8Wx6+ShuR9QM3yFajxs0t+xeEhrOkQvPJ0lxwdDdT1nSwC9xiY
/gD5xUwDViBs4WrB/1EUBUazUFjlT3WQQ+2EqqbqkBdRX0fN+KaiV9ll9QIDAQAB
AoGAFucZOcd/yD5SzXTJQyMgt0axyDUcaP8tzJjCt4B3kgLJ3b2YXa7axsSw6sk5
9rqyQJDFTH1bRErRIXiu+8UAZ4ayJPlvE9XKmS/JzAcPqXNqPx2VET3H00sLOl23
xjMeCAPX2z2ZMQfA9h34uDKwTnjf1mriti0m0uasXO4JfIECQQDp2/qdFb0oDTJY
+uU5HgQ8RpExv2dsR/Y3ZCfj+cuQR2BwAaIdibnSJAsSxT8zpQF2kPR3AUUGZtK0
7qJ0K70VAkEA47s4dnjWufVeQmxyGMeHtiJGzwdF1ars0LL5m3JQvfPCkjBcLKDQ
uf5JTAXJMJ/l6C341wPF4rCC/+dqyAX9YQJAXh9ccabTN/B/yBJK+b8cA0p/m58m
uA0Kiuazq2zZQluH8+ykW/EXqf05u7dJpbaOrTLQQalwJ5Bw08OL/OextQJBAM+j
xgCnf0mAwtgXnxSe4UudByj+/ZqrRU+o0FP+sEXx+wdmFrUOUCI2C8jIQcAXGv5O
5GPP6d8eh+Missb8RyECQFCbPrZx7SSAulFsZyF61RBR3tIP8tvzkGq9ZqTzC0U/
d7awuT2TT95mId/sDODb2YftWPnH76RBDwl4QhTJJU4=
-----END RSA PRIVATE KEY-----`
)

var (
	validExpiryDate = time.Now().AddDate(50, 0, 0) // 50 years from now
)

type clientConfig struct {
	addr      string
	tlsConfig *tls.Config
}

type testTokenSource struct {
	token     *oauth2.Token
	callCount int
}

func (t *testTokenSource) Token() (*oauth2.Token, error) {
	t.callCount += 1
	return t.token, nil
}

func testAPI(serverOptsFactory func() []api.ServerOption) {
	var (
		serverOpts  []api.ServerOption
		clientConf  clientConfig
		fakeStatter *monitorfakes.FakePermStatter
	)

	BeforeEach(func() {
		serverOpts = serverOptsFactory()
		fakeStatter = new(monitorfakes.FakePermStatter)

		permServerCert, err := tls.X509KeyPair([]byte(testCert), []byte(testCertKey))
		Expect(err).NotTo(HaveOccurred())

		permServerTLSConfig := &tls.Config{
			Certificates: []tls.Certificate{permServerCert},
		}

		serverOpts = append(
			serverOpts,
			api.WithTLSConfig(permServerTLSConfig),
			api.WithStats(permstats.NewHandler(fakeStatter)),
		)

		rootCAPool := x509.NewCertPool()
		ok := rootCAPool.AppendCertsFromPEM([]byte(testCA))
		Expect(ok).To(BeTrue())

		clientConf = clientConfig{
			tlsConfig: &tls.Config{
				RootCAs: rootCAPool,
			},
		}
	})

	Describe("With Authentication", func() {
		var (
			subject *api.Server

			client *perm.Client

			oauthServer        *httptest.Server
			validIssuer        string
			fakeSecurityLogger *rpcfakes.FakeSecurityLogger
			expectSecurityLog  func(logID, logName string, extensions map[string]string)
		)

		BeforeEach(func() {
			fakeSecurityLogger = new(rpcfakes.FakeSecurityLogger)

			logLineID := 0
			expectSecurityLog = func(logID, logName string, extensions map[string]string) {
				_, actualLogID, actualLogName, actualExtensions := fakeSecurityLogger.LogArgsForCall(logLineID)
				Expect(logID).To(Equal(actualLogID))
				Expect(logName).To(Equal(actualLogName))

				for _, actualExt := range actualExtensions {
					var expectedVal string
					var ok bool
					if expectedVal, ok = extensions[actualExt.Key]; !ok {
						Fail(fmt.Sprintf("Found unexpected CEF extension '%s' with value '%s'", actualExt.Key, actualExt.Value))
					}
					if !strings.HasPrefix(actualExt.Value, expectedVal) {
						Fail(fmt.Sprintf("Value for extension %s doesn't match. Expected prefix: %s; Actual: %s", actualExt.Key, expectedVal, actualExt.Value))
					}
				}
				if len(extensions) != len(actualExtensions) {
					failureMsg := fmt.Sprintf("Expected %d extensions, but got %d", len(extensions), len(actualExtensions))
					Fail(failureMsg)
				}

				logLineID++
			}
			oauthServer = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch req.URL.Path {
				case "/oauth/token/.well-known/openid-configuration":
					w.Write([]byte(fmt.Sprintf(`
{
  "issuer": "https://%s/oauth/token",
  "jwks_uri": "https://%s/token_keys"
}`, req.Host, req.Host)))
				case "/token_keys":
					w.Write([]byte(`
{
	"keys": [
		{
			"kty": "RSA",
			"e": "AQAB",
			"n": "APdfDb_pJpRhdbs9jDCWifExV81fROWxVVaLSh8R5dRxEIY8RRVS5AYmoI4etabv3gubYw3cgNIJGFqS3xLwJwGrVf7fh8YIIxm_H6QNG2mys7Rn80RufRpqrkas0EBcJqa_zQpS3QnINJ6ZkrSXhghYdD0R_01VQpQ7OnXFLKCAxUo7y0vUiMQhdKf0y8YhRd5v-cvujgze0vQnWrDQ9UY224OPnNtJK1zv2E7Ssn43PTEt1OxF2lYLuSqJUw8lEiE8FTQIBIUj0yqfiMQ8dn4GeJem8nTfsRyNHBOHF-HddJW-RrQ-ryvFLLpFu0H0wecSzlF-5SXOsTUpGuMQ0pU"
		}
	]
}`))
				default:
					out, err := httputil.DumpRequest(req, true)
					Expect(err).NotTo(HaveOccurred())
					Fail(fmt.Sprintf("unexpected request: %s", out))
				}
			}))

			oauthServer.StartTLS()

			validIssuer = fmt.Sprintf("%s/oauth/token", oauthServer.URL)
		})

		BeforeEach(func() {
			oauthServerCA, err := x509.ParseCertificate(oauthServer.TLS.Certificates[0].Certificate[0])
			Expect(err).NotTo(HaveOccurred())
			certpool := x509.NewCertPool()
			certpool.AddCert(oauthServerCA)

			oidcContext := oidc.ClientContext(context.Background(), &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						RootCAs: certpool,
					},
				},
			})

			oidcProvider, err := oidc.NewProvider(oidcContext, validIssuer)
			Expect(err).NotTo(HaveOccurred())

			serverOpts = append(serverOpts, api.WithOIDCProvider(oidcProvider))
			serverOpts = append(serverOpts, api.WithSecurityLogger(fakeSecurityLogger))
			subject = api.NewServer(serverOpts...)

			listener, err := net.Listen("tcp", "localhost:")
			Expect(err).NotTo(HaveOccurred())

			clientConf.addr = listener.Addr().String()

			go func() {
				err := subject.Serve(listener)
				Expect(err).NotTo(HaveOccurred())
			}()
		})

		AfterEach(func() {
			oauthServer.Close()

			subject.Stop()

			err := client.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("creates the new role when no errors occur during authentication", func() {
			validToken, err := getSignedToken(validPrivateKey, validIssuer, validExpiryDate)
			Expect(err).ToNot(HaveOccurred())

			client, err = perm.Dial(
				clientConf.addr,
				perm.WithTLSConfig(clientConf.tlsConfig),
				perm.WithTokenSource(&testTokenSource{token: validToken}))
			Expect(err).NotTo(HaveOccurred())

			_, err = client.CreateRole(context.Background(), uuid.NewV4().String())
			Expect(err).ToNot(HaveOccurred())
			expectSecurityLog("Auth", "Auth", map[string]string{"msg": "authentication succeeded", "subject": "subject_id"})
		})

		It("returns a unauthenticated error when the client does not send a JWT token", func() {
			var err error
			client, err = perm.Dial(clientConf.addr, perm.WithTLSConfig(clientConf.tlsConfig))
			Expect(err).NotTo(HaveOccurred())

			_, err = client.CreateRole(context.Background(), uuid.NewV4().String())
			Expect(err).To(MatchError("perm: unauthenticated"))
			Expect(fakeSecurityLogger.LogCallCount()).To(Equal(1))
			expectSecurityLog("Auth", "Auth", map[string]string{"msg": "no token"})
		})

		It("returns a malformed token error when the client's token is malformed", func() {
			var err error
			client, err = perm.Dial(
				clientConf.addr,
				perm.WithTLSConfig(clientConf.tlsConfig),
				perm.WithTokenSource(&testTokenSource{token: &oauth2.Token{AccessToken: "hello, world"}}))
			Expect(err).NotTo(HaveOccurred())

			_, err = client.CreateRole(context.Background(), uuid.NewV4().String())
			Expect(err).To(MatchError("perm: unauthenticated"))
			Expect(fakeSecurityLogger.LogCallCount()).To(Equal(1))
			expectSecurityLog("Auth", "Auth", map[string]string{"msg": "oidc: malformed jwt: square/go-jose: compact JWS format must have three parts"})
		})

		It("returns a token invalid error when the client's token is signed by an unknown key", func() {
			invalidSignedToken, err := getSignedToken(foreignPrivateKey, validIssuer, validExpiryDate)
			Expect(err).ToNot(HaveOccurred())

			client, err = perm.Dial(
				clientConf.addr,
				perm.WithTLSConfig(clientConf.tlsConfig),
				perm.WithTokenSource(&testTokenSource{token: invalidSignedToken}))
			Expect(err).NotTo(HaveOccurred())

			_, err = client.CreateRole(context.Background(), uuid.NewV4().String())
			Expect(err).To(MatchError("perm: unauthenticated"))
			Expect(fakeSecurityLogger.LogCallCount()).To(Equal(1))
			expectSecurityLog("Auth", "Auth", map[string]string{"msg": "failed to verify signature: failed to verify id token signature"})
		})

		It("returns a token invalid error when the client's token is expired", func() {
			expiry := time.Now().AddDate(0, 0, -1) // 1 hour ago
			expiredToken, err := getSignedToken(validPrivateKey, validIssuer, expiry)
			Expect(err).NotTo(HaveOccurred())

			client, err = perm.Dial(
				clientConf.addr,
				perm.WithTLSConfig(clientConf.tlsConfig),
				perm.WithTokenSource(&testTokenSource{token: expiredToken}))
			Expect(err).NotTo(HaveOccurred())

			_, err = client.CreateRole(context.Background(), uuid.NewV4().String())
			Expect(err).To(MatchError("perm: unauthenticated"))
			Expect(fakeSecurityLogger.LogCallCount()).To(Equal(1))
			expectSecurityLog("Auth", "Auth", map[string]string{"msg": "oidc: token is expired"})
		})

		It("returns an unauthenticated error when the token issuer doesn't match provider issuer", func() {
			invalidToken, err := getSignedToken(validPrivateKey, "https://uaa.run.pivotal.io:443/oauth/token", validExpiryDate)
			Expect(err).NotTo(HaveOccurred())

			client, err = perm.Dial(
				clientConf.addr,
				perm.WithTLSConfig(clientConf.tlsConfig),
				perm.WithTokenSource(&testTokenSource{token: invalidToken}))
			Expect(err).NotTo(HaveOccurred())

			_, err = client.CreateRole(context.Background(), uuid.NewV4().String())
			Expect(err).To(MatchError("perm: unauthenticated"))
			Expect(fakeSecurityLogger.LogCallCount()).To(Equal(1))
			expectSecurityLog("Auth", "Auth", map[string]string{"msg": "oidc: id token issued by a different provider"})
		})

		It("refreshes the token when it does not have AccessToken and therefore invalid", func() {
			tokenSource := &testTokenSource{token: &oauth2.Token{}}

			var err error
			client, err = perm.Dial(
				clientConf.addr,
				perm.WithTLSConfig(clientConf.tlsConfig),
				perm.WithTokenSource(tokenSource))
			Expect(err).NotTo(HaveOccurred())

			_, err = client.CreateRole(context.Background(), uuid.NewV4().String())
			Expect(tokenSource.callCount).To(Equal(1))

			_, err = client.CreateRole(context.Background(), uuid.NewV4().String())
			Expect(tokenSource.callCount).To(Equal(2))

			Expect(fakeSecurityLogger.LogCallCount()).To(Equal(2))
			// TODO: should the second message be a success?
			expectSecurityLog("Auth", "Auth", map[string]string{"msg": "oidc: malformed jwt: square/go-jose: compact JWS format must have three parts"})
			expectSecurityLog("Auth", "Auth", map[string]string{"msg": "oidc: malformed jwt: square/go-jose: compact JWS format must have three parts"})
		})

		It("refreshes the token when it is expired", func() {
			validToken, err := getSignedToken(validPrivateKey, validIssuer, time.Now().Add(-10*time.Second))
			Expect(err).ToNot(HaveOccurred())
			tokenSource := &testTokenSource{token: validToken}

			client, err = perm.Dial(
				clientConf.addr,
				perm.WithTLSConfig(clientConf.tlsConfig),
				perm.WithTokenSource(tokenSource))
			Expect(err).NotTo(HaveOccurred())

			_, err = client.CreateRole(context.Background(), uuid.NewV4().String())
			Expect(tokenSource.callCount).To(Equal(1))

			_, err = client.CreateRole(context.Background(), uuid.NewV4().String())
			Expect(tokenSource.callCount).To(Equal(2))

			Expect(fakeSecurityLogger.LogCallCount()).To(Equal(2))
			// TODO: should the second message be a success?
			expectSecurityLog("Auth", "Auth", map[string]string{"msg": "oidc: token is expired"})
			expectSecurityLog("Auth", "Auth", map[string]string{"msg": "oidc: token is expired"})
		})
	})

	Describe("Without Authentication", func() {
		var (
			subject *api.Server
			client  *perm.Client
		)

		BeforeEach(func() {
			subject = api.NewServer(serverOpts...)

			listener, err := net.Listen("tcp", "localhost:")
			Expect(err).NotTo(HaveOccurred())

			clientConf.addr = listener.Addr().String()

			client, err = perm.Dial(clientConf.addr, perm.WithTLSConfig(clientConf.tlsConfig))
			Expect(err).NotTo(HaveOccurred())

			go func() {
				err := subject.Serve(listener)
				Expect(err).NotTo(HaveOccurred())
			}()
		})

		AfterEach(func() {
			subject.Stop()

			err := client.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("#CreateRole", func() {
			It("returns the new role", func() {
				name := uuid.NewV4().String()

				role, err := client.CreateRole(context.Background(), name)
				Expect(err).NotTo(HaveOccurred())

				Expect(role.Name).To(Equal(name))
			})

			It("fails when a role with the same name already exists", func() {
				name := uuid.NewV4().String()

				_, err := client.CreateRole(context.Background(), name)
				Expect(err).NotTo(HaveOccurred())

				_, err = client.CreateRole(context.Background(), name)
				Expect(err).To(MatchError("role already exists"))
			})
		})

		Describe("#DeleteRole", func() {
			It("succeeds when the role exists", func() {
				role, err := client.CreateRole(context.Background(), uuid.NewV4().String())
				Expect(err).NotTo(HaveOccurred())

				err = client.DeleteRole(context.Background(), role.Name)
				Expect(err).NotTo(HaveOccurred())
			})

			It("fails when the role does not exist", func() {
				err := client.DeleteRole(context.Background(), uuid.NewV4().String())
				Expect(err).To(MatchError("role not found"))
			})

		})

		Describe("#AssignRole", func() {
			It("succeeds when the role exists and the actor has not yet been assigned to it", func() {
				role, err := client.CreateRole(context.Background(), uuid.NewV4().String())
				Expect(err).NotTo(HaveOccurred())

				actor := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: uuid.NewV4().String(),
				}

				err = client.AssignRole(context.Background(), role.Name, actor)
				Expect(err).NotTo(HaveOccurred())
			})

			It("allows assignments with the same actor ID but different namespaces", func() {
				role, err := client.CreateRole(context.Background(), uuid.NewV4().String())
				Expect(err).NotTo(HaveOccurred())

				id := uuid.NewV4().String()
				actor1 := perm.Actor{
					ID:        id,
					Namespace: uuid.NewV4().String(),
				}
				actor2 := perm.Actor{
					ID:        id,
					Namespace: uuid.NewV4().String(),
				}

				err = client.AssignRole(context.Background(), role.Name, actor1)
				Expect(err).NotTo(HaveOccurred())

				err = client.AssignRole(context.Background(), role.Name, actor2)
				Expect(err).NotTo(HaveOccurred())
			})

			It("allows assignments with the same namespace but different actor IDs", func() {
				role, err := client.CreateRole(context.Background(), uuid.NewV4().String())
				Expect(err).NotTo(HaveOccurred())

				namespace := uuid.NewV4().String()
				actor1 := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: namespace,
				}
				actor2 := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: namespace,
				}

				err = client.AssignRole(context.Background(), role.Name, actor1)
				Expect(err).NotTo(HaveOccurred())

				err = client.AssignRole(context.Background(), role.Name, actor2)
				Expect(err).NotTo(HaveOccurred())
			})

			It("fails when the role does not exist", func() {
				actor := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: uuid.NewV4().String(),
				}

				err := client.AssignRole(context.Background(), uuid.NewV4().String(), actor)
				Expect(err).To(MatchError("role not found"))
			})

			It("fails when the actor has already been assigned to the role", func() {
				role, err := client.CreateRole(context.Background(), uuid.NewV4().String())
				Expect(err).NotTo(HaveOccurred())

				actor := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: uuid.NewV4().String(),
				}

				err = client.AssignRole(context.Background(), role.Name, actor)
				Expect(err).NotTo(HaveOccurred())

				err = client.AssignRole(context.Background(), role.Name, actor)
				Expect(err).To(MatchError("assignment already exists"))
			})

			It("fails when the actor namespace has not been specified", func() {
				role, err := client.CreateRole(context.Background(), uuid.NewV4().String())
				Expect(err).NotTo(HaveOccurred())

				actor := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: "",
				}

				err = client.AssignRole(context.Background(), role.Name, actor)
				Expect(err).To(MatchError("actor namespace cannot be empty"))
			})
		})

		Describe("#UnassignRole", func() {
			It("succeeds when the role exists and the actor has been assigned to it", func() {
				actor := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: uuid.NewV4().String(),
				}

				role, err := client.CreateRole(context.Background(), uuid.NewV4().String())
				Expect(err).NotTo(HaveOccurred())

				err = client.AssignRole(context.Background(), role.Name, actor)
				Expect(err).NotTo(HaveOccurred())

				err = client.UnassignRole(context.Background(), role.Name, actor)
				Expect(err).NotTo(HaveOccurred())
			})

			It("can only be called once per assignment", func() {
				actor := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: uuid.NewV4().String(),
				}

				role, err := client.CreateRole(context.Background(), uuid.NewV4().String())
				Expect(err).NotTo(HaveOccurred())

				err = client.AssignRole(context.Background(), role.Name, actor)
				Expect(err).NotTo(HaveOccurred())

				err = client.UnassignRole(context.Background(), role.Name, actor)
				Expect(err).NotTo(HaveOccurred())

				err = client.UnassignRole(context.Background(), role.Name, actor)
				Expect(err).To(MatchError("assignment not found"))
			})

			It("fails when the role does not exist", func() {
				actor := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: uuid.NewV4().String(),
				}

				err := client.UnassignRole(context.Background(), uuid.NewV4().String(), actor)
				Expect(err).To(MatchError("assignment not found"))
			})

			It("fails when the actor has not been assigned to the role", func() {
				actor := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: uuid.NewV4().String(),
				}

				role, err := client.CreateRole(context.Background(), uuid.NewV4().String())
				Expect(err).NotTo(HaveOccurred())

				err = client.UnassignRole(context.Background(), role.Name, actor)
				Expect(err).To(MatchError("assignment not found"))
			})
		})

		Describe("#HasPermission", func() {
			It("returns true when the actor has a single role that matches the permission", func() {
				permission := perm.Permission{
					Action:          "test.read",
					ResourcePattern: uuid.NewV4().String(),
				}

				role, err := client.CreateRole(context.Background(), uuid.NewV4().String(), permission)
				Expect(err).NotTo(HaveOccurred())

				actor := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: uuid.NewV4().String(),
				}

				err = client.AssignRole(context.Background(), role.Name, actor)
				Expect(err).NotTo(HaveOccurred())

				hasPermission, err := client.HasPermission(context.Background(), actor, permission.Action, permission.ResourcePattern)
				Expect(err).NotTo(HaveOccurred())

				Expect(hasPermission).To(Equal(true))
			})

			It("returns true when the actor has multiple roles that match the permission", func() {
				permission := perm.Permission{
					Action:          "test.read",
					ResourcePattern: uuid.NewV4().String(),
				}

				role1, err := client.CreateRole(context.Background(), uuid.NewV4().String(), permission)
				Expect(err).NotTo(HaveOccurred())

				role2, err := client.CreateRole(context.Background(), uuid.NewV4().String(), permission)
				Expect(err).NotTo(HaveOccurred())

				actor := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: uuid.NewV4().String(),
				}

				err = client.AssignRole(context.Background(), role1.Name, actor)
				Expect(err).NotTo(HaveOccurred())

				err = client.AssignRole(context.Background(), role2.Name, actor)
				Expect(err).NotTo(HaveOccurred())

				hasPermission, err := client.HasPermission(context.Background(), actor, permission.Action, permission.ResourcePattern)
				Expect(err).NotTo(HaveOccurred())

				Expect(hasPermission).To(Equal(true))
			})

			It("returns false when the actor has not been assigned to a role with the matching permission", func() {
				permission1 := perm.Permission{
					Action:          "test.read",
					ResourcePattern: uuid.NewV4().String(),
				}

				role1, err := client.CreateRole(context.Background(), uuid.NewV4().String(), permission1)
				Expect(err).NotTo(HaveOccurred())

				permission2 := perm.Permission{
					Action:          "test.read",
					ResourcePattern: uuid.NewV4().String(),
				}

				_, err = client.CreateRole(context.Background(), uuid.NewV4().String(), permission2)
				Expect(err).NotTo(HaveOccurred())

				actor := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: uuid.NewV4().String(),
				}

				err = client.AssignRole(context.Background(), role1.Name, actor)
				Expect(err).NotTo(HaveOccurred())

				hasPermission, err := client.HasPermission(context.Background(), actor, permission2.Action, permission2.ResourcePattern)
				Expect(err).NotTo(HaveOccurred())

				Expect(hasPermission).To(Equal(false))
			})

			It("returns false when the actor has been assigned to no roles", func() {
				permission := perm.Permission{
					Action:          "test.read",
					ResourcePattern: uuid.NewV4().String(),
				}

				_, err := client.CreateRole(context.Background(), uuid.NewV4().String(), permission)
				Expect(err).NotTo(HaveOccurred())

				actor := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: uuid.NewV4().String(),
				}

				hasPermission, err := client.HasPermission(context.Background(), actor, permission.Action, permission.ResourcePattern)
				Expect(err).NotTo(HaveOccurred())

				Expect(hasPermission).To(Equal(false))
			})

			It("returns false when no roles have the matching permission", func() {
				role, err := client.CreateRole(context.Background(), uuid.NewV4().String())
				Expect(err).NotTo(HaveOccurred())

				actor := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: uuid.NewV4().String(),
				}

				err = client.AssignRole(context.Background(), role.Name, actor)
				Expect(err).NotTo(HaveOccurred())

				permission := perm.Permission{
					Action:          "test.read",
					ResourcePattern: uuid.NewV4().String(),
				}

				hasPermission, err := client.HasPermission(context.Background(), actor, permission.Action, permission.ResourcePattern)
				Expect(err).NotTo(HaveOccurred())

				Expect(hasPermission).To(Equal(false))
			})
		})

		Describe("#ListResourcePatterns", func() {
			It("returns the list of resource patterns on which the actor can perform the action", func() {
				action := uuid.NewV4().String()
				permission1 := perm.Permission{
					Action:          action,
					ResourcePattern: uuid.NewV4().String(),
				}
				permission2 := perm.Permission{
					Action:          uuid.NewV4().String(),
					ResourcePattern: uuid.NewV4().String(),
				}

				role1, err := client.CreateRole(context.Background(), uuid.NewV4().String(), permission1, permission2)
				Expect(err).NotTo(HaveOccurred())

				permission3 := perm.Permission{
					Action:          action,
					ResourcePattern: uuid.NewV4().String(),
				}
				role2, err := client.CreateRole(context.Background(), uuid.NewV4().String(), permission3)
				Expect(err).NotTo(HaveOccurred())

				actor := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: uuid.NewV4().String(),
				}

				err = client.AssignRole(context.Background(), role1.Name, actor)
				Expect(err).NotTo(HaveOccurred())

				err = client.AssignRole(context.Background(), role2.Name, actor)
				Expect(err).NotTo(HaveOccurred())

				resourcePatterns, err := client.ListResourcePatterns(context.Background(), actor, action)
				Expect(err).NotTo(HaveOccurred())

				Expect(resourcePatterns).To(HaveLen(2))
				Expect(resourcePatterns).To(ContainElement(permission1.ResourcePattern))
				Expect(resourcePatterns).To(ContainElement(permission3.ResourcePattern))
			})

			It("de-dupes the results if the user has access to the same resource pattern via multiple roles", func() {
				action := uuid.NewV4().String()
				permission := perm.Permission{
					Action:          action,
					ResourcePattern: uuid.NewV4().String(),
				}

				role1, err := client.CreateRole(context.Background(), uuid.NewV4().String(), permission)
				Expect(err).NotTo(HaveOccurred())

				role2, err := client.CreateRole(context.Background(), uuid.NewV4().String(), permission)
				Expect(err).NotTo(HaveOccurred())

				actor := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: uuid.NewV4().String(),
				}

				err = client.AssignRole(context.Background(), role1.Name, actor)
				Expect(err).NotTo(HaveOccurred())

				err = client.AssignRole(context.Background(), role2.Name, actor)
				Expect(err).NotTo(HaveOccurred())

				resourcePatterns, err := client.ListResourcePatterns(context.Background(), actor, action)
				Expect(err).NotTo(HaveOccurred())

				Expect(resourcePatterns).To(HaveLen(1))
				Expect(resourcePatterns).To(ContainElement(permission.ResourcePattern))
			})

			It("returns an empty list if the actor is not assigned to any roles with a relevant permission", func() {
				action := uuid.NewV4().String()
				permission1 := perm.Permission{
					Action:          action,
					ResourcePattern: uuid.NewV4().String(),
				}

				_, err := client.CreateRole(context.Background(), uuid.NewV4().String(), permission1)
				Expect(err).NotTo(HaveOccurred())

				permission2 := perm.Permission{
					Action:          uuid.NewV4().String(),
					ResourcePattern: uuid.NewV4().String(),
				}

				role, err := client.CreateRole(context.Background(), uuid.NewV4().String(), permission2)
				Expect(err).NotTo(HaveOccurred())

				actor := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: uuid.NewV4().String(),
				}

				err = client.AssignRole(context.Background(), role.Name, actor)
				Expect(err).NotTo(HaveOccurred())

				resourcePatterns, err := client.ListResourcePatterns(context.Background(), actor, action)
				Expect(err).NotTo(HaveOccurred())

				Expect(resourcePatterns).To(BeEmpty())
			})

			It("returns an empty list if the actor is not assigned to any roles", func() {
				actor := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: uuid.NewV4().String(),
				}
				action := uuid.NewV4().String()

				resourcePatterns, err := client.ListResourcePatterns(context.Background(), actor, action)
				Expect(err).NotTo(HaveOccurred())

				Expect(resourcePatterns).To(BeEmpty())
			})
		})

		Describe("Statter", func() {
			BeforeEach(func() {
				role := uuid.NewV4().String()
				permission := perm.Permission{
					Action:          uuid.NewV4().String(),
					ResourcePattern: uuid.NewV4().String(),
				}
				returnedRole, err := client.CreateRole(context.Background(), role, permission)
				Expect(returnedRole.Name).To(Equal(role))
				Expect(err).NotTo(HaveOccurred())
			})

			It("records request count", func() {
				Eventually(fakeStatter.IncCallCount()).Should(Equal(1))
				methodName, increment, rate := fakeStatter.IncArgsForCall(0)
				Expect(methodName).To(Equal("count.CreateRole"))
				Expect(increment).To(Equal(int64(1)))
				Expect(rate).To(Equal(float32(1)))
			})

			It("records the time taken to serve the rpc call", func() {
				Eventually(fakeStatter.TimingDurationCallCount()).Should(Equal(1))
				methodName, duration, rate := fakeStatter.TimingDurationArgsForCall(0)
				Expect(methodName).To(Equal("rpcduration.CreateRole"))
				Expect(duration).To(BeNumerically(">", 0))
				Expect(rate).To(Equal(float32(1)))
			})

			It("records request size", func() {
				Eventually(fakeStatter.RawCallCount()).Should(Equal(3))
				methodName, size, rate := fakeStatter.RawArgsForCall(0)
				Expect(methodName).To(Equal("requestsize.CreateRole"))
				Expect(size).To(Equal("116")) //This is the length of Create Role request
				Expect(rate).To(Equal(float32(1)))
			})

			It("records response size", func() {
				Eventually(fakeStatter.RawCallCount()).Should(Equal(3))
				methodName, size, rate := fakeStatter.RawArgsForCall(1)
				Expect(methodName).To(Equal("responsesize.CreateRole"))
				Expect(size).To(Equal("40")) //This is the length of Create Role response
				Expect(rate).To(Equal(float32(1)))
			})

			It("records success", func() {
				Eventually(fakeStatter.RawCallCount()).Should(Equal(3))
				methodName, increment, rate := fakeStatter.RawArgsForCall(2)
				Expect(methodName).To(Equal("success.CreateRole"))
				Expect(increment).To(Equal("1"))
				Expect(rate).To(Equal(float32(1)))
			})
		})
	})
}

func getSignedToken(privateKey, issuer string, expiry time.Time) (*oauth2.Token, error) {
	issuedAt := time.Now().AddDate(-50, 0, 0).Unix() // 50 years ago

	payload := fmt.Sprintf(`
{
	"scope": [
		"openid"
	],
	"sub": "subject_id",
	"iat": %d,
	"exp": %d,
	"iss": "%s",
	"aud": [
		"abc",
		"password",
		"perm",
		"foobar",
		"openid"
	]
}`, issuedAt, expiry.Unix(), issuer)

	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		Fail("unable to parse the test private key")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	Expect(err).NotTo(HaveOccurred())

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: privKey}, nil)
	Expect(err).NotTo(HaveOccurred())

	signedJWTToken, err := signer.Sign([]byte(payload))
	Expect(err).NotTo(HaveOccurred())

	serialized := signedJWTToken.FullSerialize()

	var token struct {
		Protected string `json:"protected"`
		Payload   string `json:"payload"`
		Signature string `json:"signature"`
	}

	err = json.Unmarshal([]byte(serialized), &token)
	Expect(err).NotTo(HaveOccurred())

	return &oauth2.Token{
		AccessToken: fmt.Sprintf("%s.%s.%s", token.Protected, token.Payload, token.Signature),
		Expiry:      expiry,
	}, nil
}
