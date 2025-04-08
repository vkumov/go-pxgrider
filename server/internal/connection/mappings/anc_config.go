package mappings

var (
	ANCConfigRestMappings = ServiceRestMappings{
		"getPolicies": {},
		"getPolicyByName": {
			Params: []Param{
				{
					Name:       "name",
					JSONSchema: `{"type":"string"}`,
				}},
		},
		"createPolicy": {
			// Params: []Param{
			// 	{
			// 		Name: "policy",
			// 		JSONSchema: `{
			// 			"type":"object",
			// 			"properties":{
			// 				"name":{"type":"string"},
			// 				"actions":{
			// 					"type":"array",
			// 					"items":{
			// 						"type":"string",
			// 						"enum":["QUARANTINE","SHUT_DOWN","PORT_BOUNCE","RE_AUTHENTICATE"]
			// 					}
			// 				},
			// 			}
			// 		}`,
			// 	}},
			Params: []Param{
				{
					Name:       "name",
					JSONSchema: `{"type":"string"}`,
				},
				{
					Name: "actions",
					JSONSchema: `{
						"type":"array",
						"items":{
							"type":"string",
							"enum":["QUARANTINE","SHUT_DOWN","PORT_BOUNCE","RE_AUTHENTICATE"]
						}
					}`,
				}},
		},
		"deletePolicyByName": {
			Params: []Param{
				{
					Name:       "name",
					JSONSchema: `{"type":"string"}`,
				}},
		},
		"getEndpoints":        {},
		"getEndpointPolicies": {},
		"getEndpointByMacAddress": {
			Params: []Param{
				{
					Name:       "macAddress",
					JSONSchema: `{"type":"string"}`,
				}},
		},
		"getEndpointByIpAddress": {
			Params: []Param{
				{
					Name:       "macAddress",
					JSONSchema: `{"type":"string"}`,
				},
				{
					Name:       "nasIpAddress",
					JSONSchema: `{"type":"string"}`,
				}},
		},
		"applyEndpointByIpAddress": {
			Params: []Param{
				{
					Name:       "policyName",
					JSONSchema: `{"type":"string"}`,
				},
				{
					Name:       "ipAddress",
					JSONSchema: `{"type":"string"}`,
				}},
		},
		"applyEndpointByMacAddress": {
			Params: []Param{
				{
					Name:       "policyName",
					JSONSchema: `{"type":"string"}`,
				},
				{
					Name:       "macAddress",
					JSONSchema: `{"type":"string"}`,
				}},
		},
		"clearEndpointByMacAddress": {
			Params: []Param{
				{
					Name:       "macAddress",
					JSONSchema: `{"type":"string"}`,
				}},
		},
		"applyEndpointPolicy": {
			Params: []Param{
				{
					Name:       "policy",
					JSONSchema: `{"type":"string"}`,
				},
				{
					Name:       "macAddress",
					JSONSchema: `{"type":"string"}`,
				},
				{
					Name:       "nasIpAddress",
					JSONSchema: `{"type":"string"}`,
				},
				{
					Name:       "sessionId",
					JSONSchema: `{"type":"string"}`,
				},
				{
					Name:       "nasPortId",
					JSONSchema: `{"type":"string"}`,
				},
				{
					Name:       "ipAddress",
					JSONSchema: `{"type":"string"}`,
				},
				{
					Name:       "userName",
					JSONSchema: `{"type":"string"}`,
				}},
		},
		"clearEndpointPolicy": {
			Params: []Param{
				{
					Name:       "macAddress",
					JSONSchema: `{"type":"string"}`,
				},
				{
					Name:       "nasIpAddress",
					JSONSchema: `{"type":"string"}`,
				}},
		},
		"getOperationStatus": {
			Params: []Param{
				{
					Name:       "operationId",
					JSONSchema: `{"type":"string"}`,
				}},
		},
	}
)

func init() {
	for _, mapping := range ANCConfigRestMappings {
		for i, param := range mapping.Params {
			mapping.Params[i].JSONSchema = cleanJSONSchema(param.JSONSchema)
		}
	}
}
