package authorization

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPeerCertificate_MarshalJSON(t *testing.T) {
	template := &x509.Certificate{
		IsCA: true,
		BasicConstraintsValid: true,
		SubjectKeyId:          []byte{1, 2, 3},
		SerialNumber:          big.NewInt(1234),
		Subject: pkix.Name{
			Country:      []string{"Earth"},
			Organization: []string{"Mother Nature"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(5, 5, 5),

		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	// generate private key
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		fmt.Println(err)
	}

	publickey := &privatekey.PublicKey

	// create a self-signed certificate. template = parent
	var parent = template
	raw, err := x509.CreateCertificate(rand.Reader, template, parent, publickey, privatekey)

	if err != nil {
		fmt.Println(err)
	}

	cert, _ := x509.ParseCertificate(raw)
	var certs = []*x509.Certificate{cert}

	addr := "www.authz.com/auth"
	req, _ := http.NewRequest("GET", addr, nil)
	req.RequestURI = addr
	req.TLS = &tls.ConnectionState{}
	req.TLS.PeerCertificates = certs
	req.Header.Add("header", "value")

	for _, c := range req.TLS.PeerCertificates {
		pcObj := PeerCertificate(*c)

		t.Run("Marshalling :", func(t *testing.T) {
			x, err := pcObj.MarshalJSON()
			raw = x
			require.NotNil(t, x)
			require.Nil(t, err)
		})

		t.Run("UnMarshalling :", func(t *testing.T) {
			err := pcObj.UnmarshalJSON(raw)
			require.Nil(t, err)
			require.Equal(t, "Earth", pcObj.Subject.Country[0])
			require.Equal(t, true, pcObj.IsCA)

		})

	}

}
