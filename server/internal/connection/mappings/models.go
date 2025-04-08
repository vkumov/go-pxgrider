package mappings

import pxgrider_proto "github.com/vkumov/go-pxgrider/pkg"

type (
	Param struct {
		Name       string
		JSONSchema string
	}

	ParamValue struct {
		Name  string
		Value any
	}

	RestMapping struct {
		Name        string
		Description string  `json:"description,omitempty"`
		Params      []Param `json:"params,omitempty"`
	}

	ServiceRestMappings map[string]RestMapping

	MethodSlice []RestMapping
)

func (m MethodSlice) ToProto() []*pxgrider_proto.Method {
	var res []*pxgrider_proto.Method
	for _, v := range m {
		var params []*pxgrider_proto.Param
		for _, p := range v.Params {
			params = append(params, &pxgrider_proto.Param{
				Name:   p.Name,
				Schema: p.JSONSchema,
			})
		}
		res = append(res, &pxgrider_proto.Method{
			Name:        v.Name,
			Params:      params,
			Description: v.Description,
		})
	}
	return res
}
