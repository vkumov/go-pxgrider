package dns

import (
	"context"
	"net"
	"strconv"

	gopxgrid "github.com/vkumov/go-pxgrid"
	pb "github.com/vkumov/go-pxgrider/pkg"
)

type CustomResolver struct {
	cfg      *gopxgrid.DNSConfig
	resolver *net.Resolver
}

func dnsCfg(cfg *gopxgrid.DNSConfig) *gopxgrid.DNSConfig {
	if cfg == nil {
		return &gopxgrid.DNSConfig{
			FamilyStrategy: gopxgrid.DefaultINETFamilyStrategy,
		}
	}
	if cfg.FamilyStrategy == gopxgrid.IPUnknown {
		cfg.FamilyStrategy = gopxgrid.DefaultINETFamilyStrategy
	}
	return cfg
}

func NewCustomResolver(cfg *gopxgrid.DNSConfig) (*CustomResolver, error) {
	ensuredCfg := dnsCfg(cfg)
	var r *net.Resolver

	if ensuredCfg.Server != "" {
		ip, err := gopxgrid.ParseDNSHost(ensuredCfg.Server)
		if err != nil {
			return nil, err
		}

		r = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{}
				return d.DialContext(ctx, "udp", ip.String())
			},
		}
	}

	return &CustomResolver{
		cfg:      ensuredCfg,
		resolver: r,
	}, nil
}

func NewCustomResolverFromPBRequest(dns *pb.DNS, strategy *pb.FamilyPreference) (*CustomResolver, error) {
	var cfg *gopxgrid.DNSConfig

	if dns != nil {
		var host string
		if dns.Port != 0 {
			host = net.JoinHostPort(dns.Ip, strconv.Itoa(int(dns.Port)))
		} else {
			host = net.JoinHostPort(dns.Ip, "53")
		}

		cfg = &gopxgrid.DNSConfig{
			Server: host,
		}
	}

	if strategy != nil {
		if cfg == nil {
			cfg = &gopxgrid.DNSConfig{}
		}

		switch strategy {
		case pb.FamilyPreference_FamilyPreference_IPv4.Enum():
			cfg.FamilyStrategy = gopxgrid.IPv4
		case pb.FamilyPreference_FamilyPreference_IPv6.Enum():
			cfg.FamilyStrategy = gopxgrid.IPv6
		case pb.FamilyPreference_FamilyPreference_IPv4AndIPv6.Enum():
			cfg.FamilyStrategy = gopxgrid.IPv46
		case pb.FamilyPreference_FamilyPreference_IPv6AndIPv4.Enum():
			cfg.FamilyStrategy = gopxgrid.IPv64
		}
	}

	return NewCustomResolver(cfg)
}

func (r *CustomResolver) Lookup(ctx context.Context, host string) ([]net.IPAddr, net.IPAddr, error) {
	if host == "" {
		return nil, net.IPAddr{}, &net.DNSError{Err: "empty host", Name: host}
	}

	var (
		addrs []net.IPAddr
		err   error
	)

	if r.resolver == nil {
		addrs, err = net.DefaultResolver.LookupIPAddr(ctx, host)
	} else {
		addrs, err = r.resolver.LookupIPAddr(ctx, host)
	}

	if err != nil {
		return nil, net.IPAddr{}, err
	}

	singleAddr, err := r.getOneIPAddr(addrs, err)
	if err != nil {
		return nil, net.IPAddr{}, err
	}

	return addrs, singleAddr, nil
}

func (r *CustomResolver) getOneIPAddr(addrs []net.IPAddr, err error) (net.IPAddr, error) {
	if err != nil {
		return net.IPAddr{}, err
	}

	switch r.cfg.FamilyStrategy {
	case gopxgrid.IPv4:
		for _, i := range addrs {
			if i.IP.To4() != nil {
				return i, nil
			}
		}
		return net.IPAddr{}, &net.AddrError{Err: "no IPv4 address found", Addr: ""}
	case gopxgrid.IPv6:
		for _, i := range addrs {
			if i.IP.To4() == nil {
				return i, nil
			}
		}
		return net.IPAddr{}, &net.AddrError{Err: "no IPv6 address found", Addr: ""}
	case gopxgrid.IPv46:
		var firstIPv6 net.IPAddr
		for _, i := range addrs {
			if i.IP.To4() != nil {
				return i, nil
			}
			if firstIPv6.IP == nil {
				firstIPv6 = i
			}
		}
		if firstIPv6.IP != nil {
			return firstIPv6, nil
		}
		return net.IPAddr{}, &net.AddrError{Err: "no IPv4 or IPv6 address found", Addr: ""}
	case gopxgrid.IPv64:
		var firstIPv4 net.IPAddr
		for _, i := range addrs {
			if i.IP.To4() == nil {
				return i, nil
			}
			if firstIPv4.IP == nil {
				firstIPv4 = i
			}
		}
		if firstIPv4.IP != nil {
			return firstIPv4, nil
		}
		return net.IPAddr{}, &net.AddrError{Err: "no IPv4 or IPv6 address found", Addr: ""}
	}

	return net.IPAddr{}, &net.AddrError{Err: "unknown strategy", Addr: ""}
}

func IPToProto(ip net.IPAddr) *pb.IP {
	i := &pb.IP{
		Ip: ip.IP.String(),
	}

	if ip.IP.To4() != nil {
		i.Family = pb.Family_INET_IPv4
	} else {
		i.Family = pb.Family_INET_IPv6
	}

	return i
}
