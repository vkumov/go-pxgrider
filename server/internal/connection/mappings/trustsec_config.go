package mappings

var (
	TrustSecConfigRestMappings = ServiceRestMappings{
		"getSecurityGroups": {
			Params: []Param{
				{
					Name:       "id",
					JSONSchema: `{"type":"string"}`,
				},
				{
					Name:       "startIndex",
					JSONSchema: `{"type":"integer"}`,
				},
				{
					Name:       "recordCount",
					JSONSchema: `{"type":"integer"}`,
				},
				{
					Name: "startTimestamp",
					JSONSchema: `{
						"$comment":"ISO8601 Datetime, e.g. 2021-01-01T00:00:00Z",
						"type":"string"
					}`,
				},
				{
					Name: "endTimestamp",
					JSONSchema: `{
						"$comment":"ISO8601 Datetime, e.g. 2021-01-01T00:00:00Z",
						"type":"string"
					}`,
				}},
		},
		"getSecurityGroupAcls": {
			Params: []Param{
				{
					Name:       "id",
					JSONSchema: `{"type":"string"}`,
				},
				{
					Name:       "startIndex",
					JSONSchema: `{"type":"integer"}`,
				},
				{
					Name:       "recordCount",
					JSONSchema: `{"type":"integer"}`,
				},
				{
					Name: "startTimestamp",
					JSONSchema: `{
						"$comment":"ISO8601 Datetime, e.g. 2021-01-01T00:00:00Z",
						"type":"string"
					}`,
				},
				{
					Name: "endTimestamp",
					JSONSchema: `{
						"$comment":"ISO8601 Datetime, e.g. 2021-01-01T00:00:00Z",
						"type":"string"
					}`,
				}},
		},
		"getVirtualNetwork": {
			Params: []Param{
				{
					Name:       "id",
					JSONSchema: `{"type":"string"}`,
				},
				{
					Name:       "startIndex",
					JSONSchema: `{"type":"integer"}`,
				},
				{
					Name:       "recordCount",
					JSONSchema: `{"type":"integer"}`,
				},
				{
					Name: "startTimestamp",
					JSONSchema: `{
						"$comment":"ISO8601 Datetime, e.g. 2021-01-01T00:00:00Z",
						"type":"string"
					}`,
				},
				{
					Name: "endTimestamp",
					JSONSchema: `{
						"$comment":"ISO8601 Datetime, e.g. 2021-01-01T00:00:00Z",
						"type":"string"
					}`,
				}},
		},
		"getEgressPolicies": {
			Params: []Param{
				{
					Name:       "id",
					JSONSchema: `{"type":"string"}`,
				},
				{
					Name:       "matrixId",
					JSONSchema: `{"type":"string"}`,
				},
				{
					Name:       "startIndex",
					JSONSchema: `{"type":"integer"}`,
				},
				{
					Name:       "recordCount",
					JSONSchema: `{"type":"integer"}`,
				},
				{
					Name: "startTimestamp",
					JSONSchema: `{
						"$comment":"ISO8601 Datetime, e.g. 2021-01-01T00:00:00Z",
						"type":"string"
					}`,
				},
				{
					Name: "endTimestamp",
					JSONSchema: `{
						"$comment":"ISO8601 Datetime, e.g. 2021-01-01T00:00:00Z",
						"type":"string"
					}`,
				}},
		},
		"getEgressMatrices": {},
	}
)

func init() {
	for _, mapping := range TrustSecConfigRestMappings {
		for i, param := range mapping.Params {
			mapping.Params[i].JSONSchema = cleanJSONSchema(param.JSONSchema)
		}
	}
}
