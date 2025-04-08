package server

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"strconv"

	gopxgrid "github.com/vkumov/go-pxgrid"

	pb "github.com/vkumov/go-pxgrider/pkg"
	"github.com/vkumov/go-pxgrider/server/internal/connection"
)

func (s *server) GetConnections(ctx context.Context, req *pb.GetConnectionsRequest) (*pb.GetConnectionsResponse, error) {
	s.app.Log().Debug().Str("uid", req.GetUser().Uid).Msg("GetConnections")
	u := s.app.Users().GetUser(ctx, req.GetUser().Uid)
	if u == nil {
		return nil, ErrUserNotFound
	}

	connections := u.GetConnections()
	resp := &pb.GetConnectionsResponse{
		Connections: make([]*pb.Connection, 0, len(connections)),
	}

	for _, c := range connections {
		resp.Connections = append(resp.Connections, c.ToProto())
	}
	s.app.Log().Debug().Int("total", len(resp.Connections)).Msg("Connections found")

	return resp, nil
}

func (s *server) GetConnectionsTotal(ctx context.Context, req *pb.GetConnectionsTotalRequest) (*pb.GetConnectionsTotalResponse, error) {
	s.app.Log().Debug().Str("uid", req.GetUser().Uid).Msg("GetConnectionsTotal")
	u := s.app.Users().GetUser(ctx, req.GetUser().Uid)
	if u == nil {
		return nil, ErrUserNotFound
	}

	connections := u.GetConnections()
	return &pb.GetConnectionsTotalResponse{Total: int64(len(connections))}, nil
}

func (s *server) CreateConnection(ctx context.Context, req *pb.CreateConnectionRequest) (*pb.CreateConnectionResponse, error) {
	s.app.Log().Debug().Str("uid", req.GetUser().Uid).Msg("CreateConnection")
	u := s.app.Users().GetUser(ctx, req.GetUser().Uid)
	if u == nil {
		return nil, ErrUserNotFound
	}

	var crreq connection.ConnectionCreate

	if req.FriendlyName == "" {
		return nil, errors.New("friendly name is required")
	}
	crreq.FriendlyName = req.FriendlyName

	if len(req.Nodes) == 0 {
		return nil, errors.New("at least one node is required")
	}
	crreq.PrimaryNode = connection.Node{
		FQDN:        req.Nodes[0].Fqdn,
		ControlPort: int(req.Nodes[0].ControlPort),
	}
	if len(req.Nodes) > 1 {
		for _, n := range req.Nodes[1:] {
			crreq.SecondaryNodes = append(crreq.SecondaryNodes, connection.Node{
				FQDN:        n.Fqdn,
				ControlPort: int(n.ControlPort),
			})
		}
	}

	if req.Credentials == nil || req.Credentials.GetKind() == nil {
		return nil, errors.New("credentials are required")
	}
	cr := req.Credentials.GetKind()
	crreq.Credentials = connection.Credentials{
		NodeName: req.Credentials.NodeName,
	}
	switch cr := cr.(type) {
	case *pb.Credentials_Certificate:
		crreq.Credentials.Type = connection.CredentialsTypeCertificate
		crreq.Credentials.Certificate = cr.Certificate.GetCertificate()
		crreq.Credentials.PrivateKey = cr.Certificate.GetPrivateKey()
		crreq.Credentials.Chain = make([]string, 0, len(cr.Certificate.GetCaCertificates()))
		for _, c := range cr.Certificate.GetCaCertificates() {
			crreq.Credentials.Chain = append(crreq.Credentials.Chain, c)
		}
	case *pb.Credentials_Password:
		crreq.Credentials.Type = connection.CredentialsTypePassword
		crreq.Credentials.Password = cr.Password.GetPassword()
	}

	crreq.Description = req.Description
	if dns := req.GetDnsDetails(); dns != nil {
		crreq.DNS = dns.Dns.GetIp()
		if p := dns.Dns.GetPort(); p != 0 {
			crreq.DNS = net.JoinHostPort(crreq.DNS, strconv.Itoa(int(p)))
		}

		switch dns.Strategy {
		case pb.FamilyPreference_FamilyPreference_IPv4:
			crreq.DNSStrategy = gopxgrid.IPv4
		case pb.FamilyPreference_FamilyPreference_IPv6:
			crreq.DNSStrategy = gopxgrid.IPv6
		case pb.FamilyPreference_FamilyPreference_IPv4AndIPv6:
			crreq.DNSStrategy = gopxgrid.IPv46
		case pb.FamilyPreference_FamilyPreference_IPv6AndIPv4:
			crreq.DNSStrategy = gopxgrid.IPv64
		}
	}

	if req.ClientName == "" {
		return nil, errors.New("client name is required")
	}
	crreq.ClientName = req.ClientName
	crreq.InsecureTLS = req.InsecureTls

	for _, cert := range req.CaCertificates {
		crreq.CA = append(crreq.CA, cert)
	}

	cn, err := u.AddConnection(ctx, crreq)
	if err != nil {
		return nil, err
	}

	return &pb.CreateConnectionResponse{Connection: cn.ToProto()}, nil
}

func (s *server) GetConnection(ctx context.Context, req *pb.GetConnectionRequest) (*pb.GetConnectionResponse, error) {
	s.app.Log().Debug().Str("uid", req.GetUser().Uid).Str("id", req.Id).Msg("GetConnection")
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.GetConnectionResponse{Connection: c.ToProto()}, nil
}

func (s *server) UpdateConnection(ctx context.Context, req *pb.UpdateConnectionRequest) (*pb.UpdateConnectionResponse, error) {
	s.app.Log().Debug().Str("uid", req.GetUser().Uid).Str("id", req.Id).Msg("UpdateConnection")
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.Id)
	if err != nil {
		return nil, err
	}

	upd := connection.ConnectionUpdate{}
	switch v := req.GetFriendlyName().GetKind().(type) {
	case *pb.NullableString_Value:
		upd.FriendlyName = sql.Null[string]{V: v.Value, Valid: true}
	}

	switch v := req.GetNodes().GetKind().(type) {
	case *pb.NullableNodeList_Value:
		nodes := v.Value.GetNodes()
		if len(nodes) == 0 {
			return nil, errors.New("at least one node is required")
		}
		upd.PrimaryNode = sql.Null[connection.Node]{V: connection.Node{
			FQDN:        nodes[0].Fqdn,
			ControlPort: int(nodes[0].ControlPort),
		}, Valid: true}
		if len(nodes) > 1 {
			nds := make([]connection.Node, 0, len(nodes[1:]))
			for _, n := range nodes[1:] {
				nds = append(nds, connection.Node{
					FQDN:        n.Fqdn,
					ControlPort: int(n.ControlPort),
				})
			}
			upd.SecondaryNodes = sql.Null[[]connection.Node]{V: nds, Valid: true}
		}
	}

	switch v := req.GetCredentials().GetKind().(type) {
	case *pb.NullableCredentials_Value:
		newCreds := connection.Credentials{
			NodeName: v.Value.GetNodeName(),
		}

		switch cr := v.Value.GetKind().(type) {
		case *pb.Credentials_Certificate:
			newCreds.Type = connection.CredentialsTypeCertificate
			newCreds.Certificate = cr.Certificate.GetCertificate()
			newCreds.PrivateKey = cr.Certificate.GetPrivateKey()
			newCreds.Chain = make([]string, 0, len(cr.Certificate.GetCaCertificates()))
			for _, c := range cr.Certificate.GetCaCertificates() {
				newCreds.Chain = append(newCreds.Chain, c)
			}
		case *pb.Credentials_Password:
			newCreds.Type = connection.CredentialsTypePassword
			newCreds.Password = cr.Password.GetPassword()
		}

		upd.Credentials = sql.Null[connection.Credentials]{V: newCreds, Valid: true}
	}

	switch v := req.GetDescription().GetKind().(type) {
	case *pb.NullableString_Value:
		upd.Description = sql.Null[string]{V: v.Value, Valid: true}
	}

	switch v := req.GetDns().GetKind().(type) {
	case *pb.NullableDNS_Value:
		host := v.Value.GetIp()
		if p := v.Value.GetPort(); p != 0 {
			host = net.JoinHostPort(host, strconv.Itoa(int(p)))
		}
		upd.DNS = sql.Null[string]{V: host, Valid: true}
	}

	switch v := req.GetDnsStrategy().GetKind().(type) {
	case *pb.NullableFamilyPreference_Value:
		switch v.Value {
		case pb.FamilyPreference_FamilyPreference_IPv4:
			upd.DNSStrategy = sql.Null[gopxgrid.INETFamilyStrategy]{V: gopxgrid.IPv4, Valid: true}
		case pb.FamilyPreference_FamilyPreference_IPv6:
			upd.DNSStrategy = sql.Null[gopxgrid.INETFamilyStrategy]{V: gopxgrid.IPv6, Valid: true}
		case pb.FamilyPreference_FamilyPreference_IPv4AndIPv6:
			upd.DNSStrategy = sql.Null[gopxgrid.INETFamilyStrategy]{V: gopxgrid.IPv46, Valid: true}
		case pb.FamilyPreference_FamilyPreference_IPv6AndIPv4:
			upd.DNSStrategy = sql.Null[gopxgrid.INETFamilyStrategy]{V: gopxgrid.IPv64, Valid: true}
		}
	}

	switch v := req.GetClientName().GetKind().(type) {
	case *pb.NullableString_Value:
		upd.ClientName = sql.Null[string]{V: v.Value, Valid: true}
	}

	switch v := req.GetOwner().GetKind().(type) {
	case *pb.NullableString_Value:
		upd.Owner = sql.Null[string]{V: v.Value, Valid: true}
	}

	switch v := req.GetInsecureTls().GetKind().(type) {
	case *pb.NullableBool_Value:
		upd.InsecureTLS = sql.Null[bool]{V: v.Value, Valid: true}
	}

	switch v := req.GetCa().GetKind().(type) {
	case *pb.NullableStringList_Value:
		upd.CA = sql.Null[[]string]{V: v.Value.GetStrings(), Valid: true}
	}

	if err = c.Update(ctx, upd); err != nil {
		return nil, err
	}
	return &pb.UpdateConnectionResponse{}, nil
}

func (s *server) DeleteConnection(ctx context.Context, req *pb.DeleteConnectionRequest) (*pb.DeleteConnectionResponse, error) {
	s.app.Log().Debug().Str("uid", req.GetUser().Uid).Str("id", req.Id).Msg("DeleteConnection")
	u := s.app.Users().GetUser(ctx, req.GetUser().Uid)
	if u == nil {
		return nil, ErrUserNotFound
	}

	if err := u.DeleteConnection(ctx, req.Id); err != nil {
		return nil, err
	}

	return &pb.DeleteConnectionResponse{}, nil
}

func (s *server) RefreshConnection(context.Context, *pb.RefreshConnectionRequest) (*pb.RefreshConnectionResponse, error) {
	// FIXME: implement
	return nil, errors.New("not implemented")
}

func (s *server) RefreshAccountState(ctx context.Context, req *pb.RefreshAccountStateRequest) (*pb.RefreshAccountStateResponse, error) {
	s.app.Log().Debug().Str("uid", req.GetUser().Uid).Str("id", req.ConnectionId).Msg("RefreshAccountState")
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.ConnectionId)
	if err != nil {
		return nil, err
	}

	r, err := c.RefreshAccountState(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.RefreshAccountStateResponse{State: string(r.AccountState), Version: r.Version}, nil
}
