package cryptox

import "crypto/x509"

func NewCertPool(certs ...[]byte) (*x509.CertPool, error) {
	pool := x509.NewCertPool()

	for _, cert := range certs {
		if ok := pool.AppendCertsFromPEM(cert); !ok {
			return nil, ErrFailedToAppendCertToPool
		}
	}

	return pool, nil
}
