package server

import (
	"context"
	"encoding/json"
	"errors"

	pb "github.com/vkumov/go-pxgrider/pkg"
	"github.com/vkumov/go-pxgrider/server/internal/connection"
	"github.com/vkumov/go-pxgrider/server/internal/connection/mappings"
)

var (
	ErrServiceNameRequired = errors.New("service name is required")
)

func (s *server) GetConnectionServices(ctx context.Context, req *pb.GetConnectionServicesRequest) (*pb.GetConnectionServicesResponse, error) {
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.GetConnectionId())
	if err != nil {
		return nil, err
	}

	ss, err := c.GetServices()
	if err != nil {
		return nil, err
	}

	return &pb.GetConnectionServicesResponse{Services: ss.ToProto()}, nil
}

func (s *server) GetServiceMethods(ctx context.Context, req *pb.GetServiceMethodsRequest) (*pb.GetServiceMethodsResponse, error) {
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.GetConnectionId())
	if err != nil {
		return nil, err
	}

	sname := req.GetServiceName()
	if sname == "" {
		return nil, ErrServiceNameRequired
	}

	ms, err := c.GetMethodsOfService(sname)
	if err != nil {
		return nil, err
	}

	return &pb.GetServiceMethodsResponse{Methods: ms.ToProto()}, nil
}

func (s *server) CallServiceMethod(ctx context.Context, req *pb.CallServiceMethodRequest) (*pb.CallServiceMethodResponse, error) {
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.GetConnectionId())
	if err != nil {
		return nil, err
	}

	s.app.Log().Debug().Str("service", req.GetServiceName()).Str("method", req.GetMethodName()).Msg("Call service method")

	sname := req.GetServiceName()
	if sname == "" {
		return nil, ErrServiceNameRequired
	}

	mname := req.GetMethodName()
	if mname == "" {
		return nil, errors.New("method name is required")
	}

	s.app.Log().Debug().Str("service", sname).Str("method", mname).Msg("Getting params")
	params := req.GetParams()
	decodedParams := make([]mappings.ParamValue, 0, len(params))
	for _, p := range params {
		var value interface{}
		err := json.Unmarshal([]byte(p.JsonValue), &value)
		if err != nil {
			return nil, err
		}

		decodedParams = append(decodedParams, mappings.ParamValue{
			Name:  p.GetName(),
			Value: value,
		})
	}

	s.app.Log().Debug().Int("params", len(decodedParams)).Msg("Calling service method")
	res, err := c.CallServiceMethod(ctx, sname, mname, req.GetNode(), decodedParams)
	if err != nil {
		return nil, err
	}

	var jsonRes string
	if res != nil {
		b, err := json.Marshal(res)
		if err != nil {
			return nil, err
		}
		jsonRes = string(b)
	}

	return &pb.CallServiceMethodResponse{JsonResponse: jsonRes}, nil
}

func (s *server) GetConnectionService(ctx context.Context, req *pb.GetConnectionServiceRequest) (*pb.GetConnectionServiceResponse, error) {
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.GetConnectionId())
	if err != nil {
		return nil, err
	}

	sname := req.GetServiceName()
	if sname == "" {
		return nil, ErrServiceNameRequired
	}

	svc, err := c.GetServiceByName(sname)
	if err != nil {
		return nil, err
	}

	return &pb.GetConnectionServiceResponse{Service: connection.ServiceToProto(svc)}, nil
}

func (s *server) ServiceCheckNodes(ctx context.Context, req *pb.ServiceCheckNodesRequest) (*pb.ServiceCheckNodesResponse, error) {
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.GetConnectionId())
	if err != nil {
		return nil, err
	}

	sname := req.GetServiceName()
	if sname == "" {
		return nil, ErrServiceNameRequired
	}

	svc, err := c.GetServiceByName(sname)
	if err != nil {
		return nil, err
	}

	err = svc.CheckNodes(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.ServiceCheckNodesResponse{}, nil
}

func (s *server) ServiceLookup(ctx context.Context, req *pb.ServiceLookupRequest) (*pb.ServiceLookupResponse, error) {
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.GetConnectionId())
	if err != nil {
		return nil, err
	}

	sname := req.GetServiceName()
	if sname == "" {
		return nil, ErrServiceNameRequired
	}

	svc, err := c.GetServiceByName(sname)
	if err != nil {
		return nil, err
	}

	s.app.Log().Debug().Str("service", sname).Msg("Lookup service")
	err = svc.Lookup(ctx)
	if err != nil {
		return nil, err
	}

	svc, err = c.GetServiceByName(sname)
	if err != nil {
		return nil, err
	}

	return &pb.ServiceLookupResponse{
		Service: connection.ServiceToProto(svc),
	}, nil
}

func (s *server) ServiceUpdateSecrets(ctx context.Context, req *pb.ServiceUpdateSecretsRequest) (*pb.ServiceUpdateSecretsResponse, error) {
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.GetConnectionId())
	if err != nil {
		return nil, err
	}

	sname := req.GetServiceName()
	if sname == "" {
		return nil, ErrServiceNameRequired
	}

	svc, err := c.GetServiceByName(sname)
	if err != nil {
		return nil, err
	}

	err = svc.UpdateSecrets(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.ServiceUpdateSecretsResponse{}, nil
}
