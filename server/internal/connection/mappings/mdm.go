package mappings

var (
	MDMRestMappings = ServiceRestMappings{
		"getEndpoints": {
			Params: []Param{
				{
					Name:       "filter",
					JSONSchema: `{"type":"string"}`,
				}},
		},
		"getEndpointByMacAddress": {
			Params: []Param{
				{
					Name:       "macAddress",
					JSONSchema: `{"type":"string"}`,
				}},
		},
		"getEndpointsByType": {
			Params: []Param{
				{
					Name: "type",
					JSONSchema: `{
						"type":"string",
						"enum": ["NON_COMPLIANT", "REGISTERED", "DISCONNECTED"]
					}`,
				}},
		},
		"getEndpointsByOsType": {
			Params: []Param{
				{
					Name: "osType",
					JSONSchema: `{
						"type":"string",
						"enum": ["ANDROID", "IOS", "WINDOWS"]
					}`,
				}},
		},
	}
)

func init() {
	for _, mapping := range MDMRestMappings {
		for i, param := range mapping.Params {
			mapping.Params[i].JSONSchema = cleanJSONSchema(param.JSONSchema)
		}
	}
}
