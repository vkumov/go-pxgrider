package connection

import (
	"encoding/json"

	gopxgrid "github.com/vkumov/go-pxgrid"
	pxgrider_proto "github.com/vkumov/go-pxgrider/pkg"
)

func ServiceToProto(s gopxgrid.PxGridService) *pxgrider_proto.Service {
	nodes := make([]*pxgrider_proto.ServiceNode, 0, len(s.Nodes()))
	for _, n := range s.Nodes() {
		nodes = append(nodes, &pxgrider_proto.ServiceNode{
			Name:       n.Name,
			NodeName:   n.NodeName,
			Properties: mapPropsToProto(n.Properties),
		})
	}

	return &pxgrider_proto.Service{
		Nodes: nodes,
	}
}

func mapPropsToProto(props map[string]any) map[string]string {
	m := make(map[string]string, len(props))
	for k, v := range props {
		switch t := v.(type) {
		case string:
			m[k] = t
		default:
			if t == nil {
				continue
			}
			bt, _ := json.Marshal(t)
			m[k] = string(bt)
		}
	}
	return m
}
