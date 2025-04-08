package connection

import gopxgrid "github.com/vkumov/go-pxgrid"

type rawAttributes struct {
	State       string   `json:"state,omitempty"`
	DNS         string   `json:"dns,omitempty"`
	DNSStrategy int      `json:"dns_strategy,omitempty"`
	Description string   `json:"description,omitempty"`
	Verify      string   `json:"verify,omitempty"`
	CA          []string `json:"ca,omitempty"`
}

func strategyFromRaw(strategy int) gopxgrid.INETFamilyStrategy {
	switch strategy {
	case 1:
		return gopxgrid.IPv4
	case 2:
		return gopxgrid.IPv46
	case 3:
		return gopxgrid.IPv64
	case 4:
		return gopxgrid.IPv6
	default:
		return gopxgrid.DefaultINETFamilyStrategy
	}
}

func (c *Connection) getRawAttributes() rawAttributes {
	attributes := rawAttributes{
		State:       string(c.state),
		Description: c.description,
		DNS:         c.dns,
		DNSStrategy: int(c.dnsStrategy),
		CA:          make([]string, len(c.tlsCfg.CA)),
		Verify:      "all",
	}
	if c.tlsCfg.InsecureSkipVerify {
		attributes.Verify = "none"
	}
	copy(attributes.CA, c.tlsCfg.CA)

	return attributes
}
