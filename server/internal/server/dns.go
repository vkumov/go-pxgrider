package server

import (
	"context"

	pb "github.com/vkumov/go-pxgrider/pkg"
	"github.com/vkumov/go-pxgrider/server/internal/dns"
)

func (s *server) CheckFQDN(ctx context.Context, req *pb.CheckFQDNRequest) (*pb.CheckFQDNResponse, error) {
	s.app.Log().Debug().Str("fqdn", req.Fqdn).
		Str("dns_ip", req.GetDns().GetIp()).Uint32("dns_port", req.GetDns().GetPort()).Msg("CheckFQDN")

	resolv, err := dns.NewCustomResolverFromPBRequest(req.GetDns(), &req.FamilyPreference)
	if err != nil {
		return &pb.CheckFQDNResponse{Error: err.Error(), IsValid: false}, nil
	}

	all, ip, err := resolv.Lookup(ctx, req.Fqdn)
	if err != nil {
		return &pb.CheckFQDNResponse{Error: err.Error(), IsValid: false}, nil
	}

	ips := make([]*pb.IP, 0, len(all))
	for _, v := range all {
		ips = append(ips, dns.IPToProto(v))
	}

	s.app.Log().Debug().Str("fqdn", req.Fqdn).Int("total", len(ips)).Msg("IPs found")

	return &pb.CheckFQDNResponse{Ips: ips, Candidate: dns.IPToProto(ip), IsValid: true}, nil
}
