package connection

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
)

func x509PoolFromCA(ca []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()

	for _, cert := range ca {
		if ok := pool.AppendCertsFromPEM([]byte(cert)); !ok {
			return nil, fmt.Errorf("failed to append certificate to pool")
		}
	}

	return pool, nil
}

func getX509Pair(certPEMBlock, keyPEMBlock string) (*tls.Certificate, error) {
	cert, err := tls.X509KeyPair([]byte(certPEMBlock), []byte(keyPEMBlock))
	if err != nil {
		return nil, err
	}
	return &cert, nil
}
