package mappings

var (
	SystemHealthRestMappings = ServiceRestMappings{
		"getHealths": {
			Params: []Param{
				{
					Name: "nodeName",
					JSONSchema: `{
						"$comment":"All nodes if not present",
						"type":"string"
					}`,
				},
				{
					Name: "startTimestamp",
					JSONSchema: `{
						"$comment":"ISO8601 Datetime, e.g. 2021-01-01T00:00:00Z. Last 1 hour if not present",
						"type":"string"
					}`,
				}},
		},
		"getPerformances": {
			Params: []Param{
				{
					Name: "nodeName",
					JSONSchema: `{
						"$comment":"All nodes if not present",
						"type":"string"
					}`,
				},
				{
					Name: "startTimestamp",
					JSONSchema: `{
						"$comment":"ISO8601 Datetime, e.g. 2021-01-01T00:00:00Z. Last 1 hour if not present",
						"type":"string"
					}`,
				}},
		},
	}
)

func init() {
	for _, mapping := range SystemHealthRestMappings {
		for i, param := range mapping.Params {
			mapping.Params[i].JSONSchema = cleanJSONSchema(param.JSONSchema)
		}
	}
}
