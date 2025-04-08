package connection

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/ettle/strcase"
	"github.com/rs/zerolog"
	gopxgrid "github.com/vkumov/go-pxgrid"

	pxgrider_proto "github.com/vkumov/go-pxgrider/pkg"
	"github.com/vkumov/go-pxgrider/server/internal/connection/mappings"
)

type (
	ServiceNameWithFriendlyName struct {
		ServiceName
		FriendlyName string
	}

	ServiceNameWithFriendlyNameSlice []ServiceNameWithFriendlyName

	knownServices struct {
		ANCConfig             ServiceName
		EndpointAsset         ServiceName
		MDM                   ServiceName
		ProfilerConfiguration ServiceName
		RadiusFailure         ServiceName
		SessionDirectory      ServiceName
		SystemHealth          ServiceName
		TrustSec              ServiceName
		TrustSecConfiguration ServiceName
		TrustSecSXP           ServiceName
	}
)

var (
	ErrServiceNotFound = errors.New("service not found")

	knownServicesMap = knownServices{
		ANCConfig:             "ANCConfig",
		EndpointAsset:         "EndpointAsset",
		MDM:                   "MDM",
		ProfilerConfiguration: "ProfilerConfiguration",
		RadiusFailure:         "RadiusFailure",
		SessionDirectory:      "SessionDirectory",
		SystemHealth:          "SystemHealth",
		TrustSec:              "TrustSec",
		TrustSecConfiguration: "TrustSecConfiguration",
		TrustSecSXP:           "TrustSecSXP",
	}
)

func (c *Connection) GetServices() (ServiceNameWithFriendlyNameSlice, error) {
	px, err := c.PX()
	if err != nil {
		return nil, err
	}

	return []ServiceNameWithFriendlyName{
		{ServiceName: knownServicesMap.ANCConfig, FriendlyName: px.ANCConfig().Name()},
		{ServiceName: knownServicesMap.EndpointAsset, FriendlyName: px.EndpointAsset().Name()},
		{ServiceName: knownServicesMap.MDM, FriendlyName: px.MDM().Name()},
		{ServiceName: knownServicesMap.ProfilerConfiguration, FriendlyName: px.ProfilerConfiguration().Name()},
		{ServiceName: knownServicesMap.RadiusFailure, FriendlyName: px.RadiusFailure().Name()},
		{ServiceName: knownServicesMap.SessionDirectory, FriendlyName: px.SessionDirectory().Name()},
		{ServiceName: knownServicesMap.SystemHealth, FriendlyName: px.SystemHealth().Name()},
		{ServiceName: knownServicesMap.TrustSec, FriendlyName: px.TrustSec().Name()},
		{ServiceName: knownServicesMap.TrustSecConfiguration, FriendlyName: px.TrustSecConfiguration().Name()},
		{ServiceName: knownServicesMap.TrustSecSXP, FriendlyName: px.TrustSecSXP().Name()},
	}, nil
}

func (c *Connection) GetMethodsOfService(service string) (mappings.MethodSlice, error) {
	var methods []mappings.RestMapping

	svc, err := c.getServiceByName(service)
	if err != nil {
		return nil, err
	}

	t := reflect.TypeOf(svc)
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		if m.Name != "Rest" {
			continue
		}

		rs := m.Func.Call([]reflect.Value{reflect.ValueOf(svc)})
		if len(rs) != 1 {
			c.log.Warn().Str("service", string(service)).Msg("invalid method signature")
			continue
		}

		mt := rs[0].Type()
		for j := 0; j < mt.NumMethod(); j++ {
			m := mt.Method(j)

			c.log.Debug().Str("service", string(service)).Str("method", m.Name).Str("reported", svc.Name()).
				Msg("Getting method mapping")
			mapping, err := c.getMethodMappings(service, m.Name)
			if err != nil {
				c.log.Warn().AnErr("err", err).Str("service", string(service)).Str("method", m.Name).Msg("failed to get method params")
				continue
			}

			methods = append(methods, mapping)
		}
	}

	return methods, nil
}

func (c *Connection) GetTopicsOfService(service string) ([]string, error) {
	svc, err := c.getServiceByName(service)
	if err != nil {
		return nil, err
	}

	topics := make([]string, 0)
	t := reflect.TypeOf(svc)
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		if m.Type.NumOut() != 1 || !strings.HasPrefix(m.Type.Out(0).String(), "gopxgrid.Subscriber") {
			continue
		}
		if m.Name == "On" {
			continue
		}
		name := m.Name
		topics = append(topics, strcase.ToCamel(name[2:]))
	}

	return topics, nil
}

func (c *Connection) GetAllTopics() (map[string][]string, error) {
	allServices, err := c.GetServices()
	if err != nil {
		return nil, err
	}

	topics := make(map[string][]string, len(allServices))
	for _, s := range allServices {
		t, err := c.GetTopicsOfService(string(s.ServiceName))
		if err != nil {
			return nil, err
		}

		topics[s.FriendlyName] = t
	}

	return topics, nil
}

func (c *Connection) CallServiceMethod(ctx context.Context, service, method, node string, params []mappings.ParamValue) (any, error) {
	svc, err := c.getServiceByName(service)
	if err != nil {
		return nil, err
	}

	pMap := make(map[string]any)
	for _, p := range params {
		if p.Value == nil {
			continue
		}
		pMap[p.Name] = p.Value
	}

	node = strings.TrimSpace(node)

	c.log.Debug().
		Str("service", service).Str("method", method).Interface("params", pMap).Str("node", node).
		Msg("Calling service method")
	caller := svc.AnyREST(method, pMap)

	var res gopxgrid.FullResponse[any]
	if node != "" {
		res, err = caller.DoOnNodeByName(ctx, node)
	} else {
		res, err = caller.Do(ctx)
	}

	if err != nil {
		return nil, err
	}

	if c.log.GetLevel() <= zerolog.DebugLevel {
		c.log.Debug().
			Int("status_code", res.StatusCode).Str("body", res.Body).
			Interface("result", res.Result).Msg("Service method result")
	} else {
		c.log.Info().
			Int("status_code", res.StatusCode).Msg("Service method executed")
	}

	return res, nil
}

func (c *Connection) getServiceByName(name string) (gopxgrid.PxGridService, error) {
	px, err := c.PX()
	if err != nil {
		return nil, err
	}

	normalized, err := c.normalizeServiceName(name)
	if err != nil {
		return nil, err
	}

	// c.log.Debug().Str("name", name).Str("normalized", string(normalized)).Msg("Normalized name")

	switch normalized {
	case gopxgrid.ANCConfigServiceName:
		return px.ANCConfig(), nil

	case gopxgrid.EndpointAssetServiceName:
		return px.EndpointAsset(), nil

	case gopxgrid.MDMServiceName:
		return px.MDM(), nil

	case gopxgrid.ProfilerConfigurationServiceName:
		return px.ProfilerConfiguration(), nil

	case gopxgrid.RadiusFailureServiceName:
		return px.RadiusFailure(), nil

	case gopxgrid.SessionDirectoryServiceName:
		return px.SessionDirectory(), nil

	case gopxgrid.SystemHealthServiceName:
		return px.SystemHealth(), nil

	case gopxgrid.TrustSecConfigurationServiceName:
		return px.TrustSecConfiguration(), nil

	case gopxgrid.TrustSecSXPServiceName:
		return px.TrustSecSXP(), nil

	case gopxgrid.TrustSecServiceName:
		return px.TrustSec(), nil

	default:
		return nil, ErrServiceNotFound
	}
}

func (c *Connection) getMethodMappings(service, method string) (mappings.RestMapping, error) {
	svc, err := c.normalizeServiceName(service)
	if err != nil {
		return mappings.RestMapping{}, err
	}

	var ms mappings.ServiceRestMappings
	switch svc {
	case gopxgrid.ANCConfigServiceName:
		ms = mappings.ANCConfigRestMappings
	case gopxgrid.EndpointAssetServiceName:
		ms = mappings.EndpointAssetRestMappings
	case gopxgrid.MDMServiceName:
		ms = mappings.MDMRestMappings
	case gopxgrid.ProfilerConfigurationServiceName:
		ms = mappings.ProfilerConfigurationRestMappings
	case gopxgrid.RadiusFailureServiceName:
		ms = mappings.RadiusFailureRestMappings
	case gopxgrid.SessionDirectoryServiceName:
		ms = mappings.SessionDirectoryRestMappings
	case gopxgrid.SystemHealthServiceName:
		ms = mappings.SystemHealthRestMappings
	case gopxgrid.TrustSecConfigurationServiceName:
		ms = mappings.TrustSecConfigRestMappings
	case gopxgrid.TrustSecSXPServiceName:
		ms = mappings.TrustSecSXPRestMappings
	case gopxgrid.TrustSecServiceName:
		ms = mappings.TrustSecRestMappings
	default:
		return mappings.RestMapping{}, ErrServiceNotFound
	}

	c.log.Debug().Str("service", service).Str("method", method).Str("normalized", string(svc)).
		Msg("Getting method mapping")
	m, realName, err := ms.GetMapping(method)
	if err != nil {
		return mappings.RestMapping{}, err
	}
	m.Name = realName

	return m, nil
}

func (s ServiceNameWithFriendlyNameSlice) ToProto() []*pxgrider_proto.ServiceNameWithFriendlyName {
	var res []*pxgrider_proto.ServiceNameWithFriendlyName
	for _, v := range s {
		res = append(res, &pxgrider_proto.ServiceNameWithFriendlyName{
			ServiceName:  string(v.ServiceName),
			FriendlyName: v.FriendlyName,
		})
	}
	return res
}
