package net

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type PreResolver struct {
	Resolver          string        `koanf:"resolver"`
	BootstrapResolver string        `koanf:"bootstrapResolver" default:"127.0.0.11:53"`
	Startup           time.Duration `koanf:"startup" default:"5s"`
	InsecureHttp      bool          `koanf:"insecure_http" default:"false"`
	verbose           bool          `default:"false"`
	netResolver       *net.Resolver
}

func (self *PreResolver) Init(verbose bool) error {
	self.Resolver = ensurePort(self.Resolver)
	self.BootstrapResolver = ensurePort(self.BootstrapResolver)
	for i := 0; i < int(self.Startup.Seconds()); i++ {
		r, rErr := intGetResolver(self)
		if rErr == nil {
			self.netResolver = r
			self.vPrint("Resolver initialized")
			return nil
		} else {
			self.vPrint("Resolver lookup failed")
			time.Sleep(time.Second)
		}
	}
	return fmt.Errorf("Can't get resolver for %s", self.Resolver)
}

func (self *PreResolver) LookUp(domain string) ([]string, error) {
	if self.netResolver != nil {
		res, resErr := intLookUp(self.netResolver, domain)
		self.vPrint("LookUp: " + domain + " Result: " + res[0])
		return res, resErr
	} else {
		return nil, fmt.Errorf("Resolver not initialized")
	}
}

func (self *PreResolver) GetHttpClient() (*http.Client, error) {
	if self.netResolver != nil {
		dialer, _ := self.GetDialer()
		tr := &http.Transport{
			Dial: dialer.Dial,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: self.InsecureHttp,
			},
		}

		client := http.Client{Transport: tr}
		return &client, nil
	} else {
		return nil, fmt.Errorf("Resolver not initialized")
	}
}

func (self *PreResolver) GetDialer() (*net.Dialer, error) {
	if self.netResolver != nil {
		return &net.Dialer{
			Timeout:  5 * time.Second,
			Resolver: self.netResolver,
		}, nil
	} else {
		return nil, fmt.Errorf("Resolver not initialized")
	}
}

func (self *PreResolver) vPrint(a ...interface{}) {
	if self.verbose {
		fmt.Println(a...)
	}
}

func ensurePort(adr string) string {
	if !strings.Contains(adr, ":") {
		adr += ":53"
	}
	return adr
}

func intLookUp(resolver *net.Resolver, domain string) ([]string, error) {
	return resolver.LookupHost(context.Background(), domain)
}

func intBaseResolver(resolver string) *net.Resolver {
	res := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, resolver)
		},
	}
	return res
}

func intGetResolver(resolver *PreResolver) (*net.Resolver, error) {
	addr := net.ParseIP(resolver.Resolver)
	ip := resolver.Resolver
	if addr != nil {
		resolver.vPrint("GetResolverEx: " + ip)
		return intBaseResolver(addr.String()), nil
	} else {
		tReso := intBaseResolver(resolver.BootstrapResolver)
		dRes, dResErr := intLookUp(tReso, resolver.Resolver)
		if dResErr == nil && len(dRes) > 0 {
			ip = dRes[0]
			return intBaseResolver(ip), nil
		}
	}
	return nil, fmt.Errorf("Can't get resolver for %s", resolver.Resolver)
}
