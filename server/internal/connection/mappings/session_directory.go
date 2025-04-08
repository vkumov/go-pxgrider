package mappings

var (
	SessionDirectoryRestMappings = ServiceRestMappings{
		"getSessions": {
			Params: []Param{
				{
					Name: "startTimestamp",
					JSONSchema: `{
						"$comment":"ISO8601 Datetime, e.g. 2021-01-01T00:00:00Z",
						"type":"string"
					}`,
				},
				{
					Name: "filter",
					JSONSchema: `{
						"$comment":"JMESFilter for each session object. e.g. \"nasIpAddress == '10.0.0.1'\"",
						"type":"string"
					}`,
				}},
		},
		"getSessionsForRecovery": {
			Params: []Param{
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
		"getSessionByIpAddress": {
			Params: []Param{
				{
					Name:       "ipAddress",
					JSONSchema: `{"type":"string"}`,
				}},
		},
		"getSessionByMacAddress": {
			Params: []Param{
				{
					Name:       "macAddress",
					JSONSchema: `{"type":"string"}`,
				}},
		},
		"getUserGroups": {
			Params: []Param{
				{
					Name: "filter",
					JSONSchema: `{
						"$comment":"JMESFilter for each user group object (optional. since ISE 3.4) (e.g. \"(groups[].name.contains(@, 'User Identity Groups: Employee')).contains(@, ` + "`true`" + `)\" )",
						"type":"string"
					}`,
				}},
		},
		"getUserGroupByUserName": {
			Params: []Param{
				{
					Name:       "userName",
					JSONSchema: `{"type":"string"}`,
				}},
		},
	}
)

func init() {
	for _, mapping := range SessionDirectoryRestMappings {
		for i, param := range mapping.Params {
			mapping.Params[i].JSONSchema = cleanJSONSchema(param.JSONSchema)
		}
	}
}
