package mappings

var (
	TrustSecSXPRestMappings = ServiceRestMappings{
		"getBindings": {
			Params: []Param{
				{
					Name: "filter",
					JSONSchema: `{
						"$comment":"JMESFilter for each bindings (optional) (e.g. \"tag == ` + "`5`" + `\")",
						"type":"string"
					}`,
				}},
		},
	}
)

func init() {
	for _, mapping := range TrustSecSXPRestMappings {
		for i, param := range mapping.Params {
			mapping.Params[i].JSONSchema = cleanJSONSchema(param.JSONSchema)
		}
	}
}
