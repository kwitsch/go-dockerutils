package net

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

type DockerResolver struct {
	Resolver    string        `koanf:"resolver"`
	Startup     time.Duration `koanf:"startup"`
	Verbose     bool          `koanf:"verbose"`
	netResolver *net.Resolver
}

func (resolver *DockerResolver) Init() error {
	if resolver.Startup < 5*time.Second {
		resolver.Startup = 5 * time.Second
	}
	for i := 0; i < int(resolver.Startup.Seconds()); i++ {
		r, rErr := intGetResolver(resolver)
		if rErr == nil {
			resolver.netResolver = r
			resolver.VPrint("Resolver initialized")
			return nil
		} else {
			resolver.VPrint("Resolver lookup failed")
			time.Sleep(time.Second)
		}
	}
	return errors.New("Can't get resolver for " + resolver.Resolver)
}

func (r *DockerResolver) LookUp(domain string) ([]string, error) {
	if r.netResolver != nil {
		res, resErr := intLookUp(r.netResolver, domain)
		r.VPrint("LookUp: " + domain + " Result: " + res[0])
		return res, resErr
	} else {
		return nil, errors.New("Resolver not initialized")
	}
}

func (r *DockerResolver) VPrint(msg string) {
	if r.Verbose {
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
			return d.DialContext(ctx, network, resolver+":53")
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
		tReso := intBaseResolver("127.0.0.11")
		dRes, dResErr := intLookUp(tReso, resolver.Resolver)
		if dResErr == nil && len(dRes) > 0 {
			ip = dRes[0]
			return intBaseResolver(ip), nil
		}
	}
	return nil, errors.New("Can't get resolver for " + resolver.Resolver)
}
