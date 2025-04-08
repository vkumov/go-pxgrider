package mappings

var (
	TrustSecRestMappings = ServiceRestMappings{}
)

func init() {
	for _, mapping := range TrustSecRestMappings {
		for i, param := range mapping.Params {
			mapping.Params[i].JSONSchema = cleanJSONSchema(param.JSONSchema)
		}
	}
}
