package mappings

var (
	RadiusFailureRestMappings = ServiceRestMappings{
		"getFailures": {
			Params: []Param{
				{
					Name: "startTimestamp",
					JSONSchema: `{
						"$comment":"ISO8601 Datetime, e.g. 2021-01-01T00:00:00Z",
						"type":"string"
					}`,
				}},
		},
		"getFailureById": {
			Params: []Param{
				{
					Name:       "id",
					JSONSchema: `{"type":"string"}`,
				}},
		},
	}
)

func init() {
	for _, mapping := range RadiusFailureRestMappings {
		for i, param := range mapping.Params {
			mapping.Params[i].JSONSchema = cleanJSONSchema(param.JSONSchema)
		}
	}
}
