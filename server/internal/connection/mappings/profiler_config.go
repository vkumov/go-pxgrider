package mappings

var (
	ProfilerConfigurationRestMappings = ServiceRestMappings{
		"getProfiles": {},
	}
)

func init() {
	for _, mapping := range ProfilerConfigurationRestMappings {
		for i, param := range mapping.Params {
			mapping.Params[i].JSONSchema = cleanJSONSchema(param.JSONSchema)
		}
	}
}
