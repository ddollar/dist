package dist

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/inconshreveable/go-update"
	"io/ioutil"
	"net/http"
	"runtime"
)

var DigicertHighAssuranceCert = `-----BEGIN CERTIFICATE-----
MIIDxTCCAq2gAwIBAgIQAqxcJmoLQJuPC3nyrkYldzANBgkqhkiG9w0BAQUFADBs
MQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3
d3cuZGlnaWNlcnQuY29tMSswKQYDVQQDEyJEaWdpQ2VydCBIaWdoIEFzc3VyYW5j
ZSBFViBSb290IENBMB4XDTA2MTExMDAwMDAwMFoXDTMxMTExMDAwMDAwMFowbDEL
MAkGA1UEBhMCVVMxFTATBgNVBAoTDERpZ2lDZXJ0IEluYzEZMBcGA1UECxMQd3d3
LmRpZ2ljZXJ0LmNvbTErMCkGA1UEAxMiRGlnaUNlcnQgSGlnaCBBc3N1cmFuY2Ug
RVYgUm9vdCBDQTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMbM5XPm
+9S75S0tMqbf5YE/yc0lSbZxKsPVlDRnogocsF9ppkCxxLeyj9CYpKlBWTrT3JTW
PNt0OKRKzE0lgvdKpVMSOO7zSW1xkX5jtqumX8OkhPhPYlG++MXs2ziS4wblCJEM
xChBVfvLWokVfnHoNb9Ncgk9vjo4UFt3MRuNs8ckRZqnrG0AFFoEt7oT61EKmEFB
Ik5lYYeBQVCmeVyJ3hlKV9Uu5l0cUyx+mM0aBhakaHPQNAQTXKFx01p8VdteZOE3
hzBWBOURtCmAEvF5OYiiAhF8J2a3iLd48soKqDirCmTCv2ZdlYTBoSUeh10aUAsg
EsxBu24LUTi4S8sCAwEAAaNjMGEwDgYDVR0PAQH/BAQDAgGGMA8GA1UdEwEB/wQF
MAMBAf8wHQYDVR0OBBYEFLE+w2kD+L9HAdSYJhoIAu9jZCvDMB8GA1UdIwQYMBaA
FLE+w2kD+L9HAdSYJhoIAu9jZCvDMA0GCSqGSIb3DQEBBQUAA4IBAQAcGgaX3Nec
nzyIZgYIVyHbIUf4KmeqvxgydkAQV8GK83rZEWWONfqe/EW1ntlMMUu4kehDLI6z
eM7b41N5cdblIZQB2lWHmiRk9opmzN6cN82oNLFpmyPInngiK3BD41VHMWEZ71jF
hS9OMPagMRYjyOfiZRYzy78aG6A9+MpeizGLYAiJLQwGXFK3xPkKmNEVX58Svnw2
Yzi9RKR/5CYrCsSXaQ3pjOLAEFe4yHYSkVXySGnYvCoCWw9E1CAx2/S6cCZdkGCe
vEsXCS+0yx5DaMkHJ8HSXPfqIbloEpw8nL+e/IBcm2PN7EeqJSdnoDfzAIJ9VNep
+OkuE6N36B9K
-----END CERTIFICATE-----`

type Dist struct {
	Host    string
	Project string
}

type DistRelease struct {
	Version string
	Url     string
}

func NewDist(project string) (d *Dist) {
	d = new(Dist)
	d.Host = "https://godist.herokuapp.com"
	d.Project = project
	return
}

func (d *Dist) Update() (version string, err error) {
	releases, err := d.fetchReleases()
	if len(releases) < 1 {
		return "", errors.New("no releases")
	}
	d.updateFromUrl(releases[0].Url)
	version = releases[0].Version
	return
}

func (d *Dist) UpdateTo(version string) (err error) {
	releases, err := d.fetchReleases()
	for _, release := range releases {
		if release.Version == version {
			d.updateFromUrl(release.Url)
			return
		}
	}
	return errors.New(fmt.Sprintf("no such version: %s", version))
}

func (d *Dist) fetchReleases() (releases []DistRelease, err error) {
	url := fmt.Sprintf("%s/projects/%s/releases/%s-%s", d.Host, d.Project, runtime.GOOS, runtime.GOARCH)
	client := d.httpClient()
	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &releases)
	return
}

func (d *Dist) httpClient() (client *http.Client) {
	chain := d.rootCertificate()
	config := tls.Config { }
	config.RootCAs = x509.NewCertPool()
	for _, cert := range chain.Certificate {
		x509Cert, err := x509.ParseCertificate(cert)
		if err != nil {
			panic(err)
		}
		config.RootCAs.AddCert(x509Cert)
	}
	config.BuildNameToCertificate()
	tr := http.Transport{ TLSClientConfig: &config}
	client = &http.Client{Transport: &tr}
	return
}

func (d *Dist) updateFromUrl(url string) (err error) {
	client := d.httpClient()
	res, err := client.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	err, _ = update.FromStream(res.Body)
	return
}

func (d *Dist) rootCertificate() (cert tls.Certificate) {
	certPEMBlock := []byte(DigicertHighAssuranceCert)
	var certDERBlock *pem.Block
	for {
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}
		if certDERBlock.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
		}
	}
	return
}
