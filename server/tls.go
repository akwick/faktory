package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/mperham/faktory/util"
)

var (
	tlsDirectories = []string{
		os.ExpandEnv("$HOME") + "/.faktory/tls",
		"/etc/faktory/tls",
	}
)

func tlsConfig(binding string, forceTLS bool) (*tls.Config, error) {
	return findTlsConfigIn(binding, forceTLS, tlsDirectories)
}

func findTlsConfigIn(binding string, disableTls bool, dirs []string) (*tls.Config, error) {
	// TLS is optional when binding to something that matches "localhost:"
	optional, err := regexp.Match("\\Alocalhost:", []byte(binding))
	if err != nil {
		return nil, err
	}

	if disableTls {
		optional = true
	}

	if optional {
		return nil, nil
	}

	// Notably, we do not provide a flag to disable TLS.
	// Infrastructure should never provide footguns.

	var tlscfg *tls.Config

	for _, dir := range dirs {
		if dir == "" {
			continue
		}

		exists, err := util.FileExists(dir + "/public.crt")
		if err != nil {
			return nil, err
		}
		if !exists {
			continue
		}

		cert, err := tls.LoadX509KeyPair(dir+"/public.crt", dir+"/private.key")
		if err != nil {
			return nil, err
		}
		tlscfg = &tls.Config{
			RootCAs:      nil,
			Certificates: []tls.Certificate{cert},
		}
		exists, err = util.FileExists(dir + "/ca.crt")
		if err != nil {
			return nil, err
		}
		if exists {
			pemData, err := ioutil.ReadFile(dir + "/ca.crt")
			if err != nil {
				return nil, err
			}
			cas := x509.NewCertPool()
			cas.AppendCertsFromPEM(pemData)
			tlscfg.RootCAs = cas
		}
		return tlscfg, nil
	}

	return nil, fmt.Errorf("TLS certificates not found in %v", dirs)
}