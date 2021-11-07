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

type DockerResolver struct {
	Resolver     string        `koanf:"resolver"`
	Startup      time.Duration `koanf:"startup" default:"5s"`
	Verbose      bool          `koanf:"verbose" default:"false"`
	InsecureHttp bool          `koanf:"insecure_http" default:"false"`
	netResolver  *net.Resolver
}

func (self *DockerResolver) Init() error {
	if !strings.Contains(self.Resolver, ":") {
		self.Resolver += ":53"
	}
	for i := 0; i < int(self.Startup.Seconds()); i++ {
		r, rErr := intGetResolver(self)
		if rErr == nil {
			self.netResolver = r
			self.VPrint("Resolver initialized")
			return nil
		} else {
			self.VPrint("Resolver lookup failed")
			time.Sleep(time.Second)
		}
	}
	return fmt.Errorf("Can't get resolver for %s", self.Resolver)
}

func (self *DockerResolver) LookUp(domain string) ([]string, error) {
	if self.netResolver != nil {
		res, resErr := intLookUp(self.netResolver, domain)
		self.VPrint("LookUp: " + domain + " Result: " + res[0])
		return res, resErr
	} else {
		return nil, fmt.Errorf("Resolver not initialized")
	}
}

func (self *DockerResolver) GetHttpClient() (*http.Client, error) {
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

func (self *DockerResolver) GetDialer() (*net.Dialer, error) {
	if self.netResolver != nil {
		return &net.Dialer{
			Timeout:  5 * time.Second,
			Resolver: self.netResolver,
		}, nil
	} else {
		return nil, fmt.Errorf("Resolver not initialized")
	}
}

func (self *DockerResolver) VPrint(msg string) {
	if self.Verbose {
		fmt.Println(msg)
	}
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

func intGetResolver(resolver *DockerResolver) (*net.Resolver, error) {
	addr := net.ParseIP(resolver.Resolver)
	ip := resolver.Resolver
	if addr != nil {
		resolver.VPrint("GetResolverEx: " + ip)
		return intBaseResolver(addr.String()), nil
	} else {
		tReso := intBaseResolver("127.0.0.11:53")
		dRes, dResErr := intLookUp(tReso, resolver.Resolver)
		if dResErr == nil && len(dRes) > 0 {
			ip = dRes[0]
			return intBaseResolver(ip), nil
		}
	}
	return nil, fmt.Errorf("Can't get resolver for %s", resolver.Resolver)
}
