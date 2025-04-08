package connection

import gopxgrid "github.com/vkumov/go-pxgrid"

func (c *Connection) normalizeServiceName(sname string) (ServiceName, error) {
	switch sname {
	case string(knownServicesMap.ANCConfig), gopxgrid.ANCConfigServiceName:
		return gopxgrid.ANCConfigServiceName, nil
	case string(knownServicesMap.EndpointAsset), gopxgrid.EndpointAssetServiceName:
		return gopxgrid.EndpointAssetServiceName, nil
	case string(knownServicesMap.MDM), gopxgrid.MDMServiceName:
		return gopxgrid.MDMServiceName, nil
	case string(knownServicesMap.SessionDirectory), gopxgrid.SessionDirectoryServiceName:
		return gopxgrid.SessionDirectoryServiceName, nil
	case string(knownServicesMap.ProfilerConfiguration), gopxgrid.ProfilerConfigurationServiceName:
		return gopxgrid.ProfilerConfigurationServiceName, nil
	case string(knownServicesMap.RadiusFailure), gopxgrid.RadiusFailureServiceName:
		return gopxgrid.RadiusFailureServiceName, nil
	case string(knownServicesMap.SystemHealth), gopxgrid.SystemHealthServiceName:
		return gopxgrid.SystemHealthServiceName, nil
	case string(knownServicesMap.TrustSecConfiguration), gopxgrid.TrustSecConfigurationServiceName:
		return gopxgrid.TrustSecConfigurationServiceName, nil
	case string(knownServicesMap.TrustSecSXP), gopxgrid.TrustSecSXPServiceName:
		return gopxgrid.TrustSecSXPServiceName, nil
	case string(knownServicesMap.TrustSec), gopxgrid.TrustSecServiceName:
		return gopxgrid.TrustSecServiceName, nil
	default:
		return "", ErrServiceNotFound
	}
}
