package connection

import (
	gopxgrid "github.com/vkumov/go-pxgrid"
	"github.com/vkumov/go-pxgrider/server/internal/db/models"
)

func (c *Connection) SetState(state gopxgrid.AccountState) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.state = state
	c.unsaved[models.ClientColumns.Attributes] = struct{}{}
}

func (c *Connection) SetFriendlyName(name string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.friendlyName = name
	c.unsaved[models.ClientColumns.FriendlyName] = struct{}{}
}

func (c *Connection) SetOwner(owner string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.owner = owner
	c.unsaved[models.ClientColumns.Owner] = struct{}{}
}

func (c *Connection) SetDescription(desc string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.description = desc
	c.unsaved[models.ClientColumns.Attributes] = struct{}{}
}

func (c *Connection) SetDNS(dns string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.dns = dns
	c.unsaved[models.ClientColumns.Attributes] = struct{}{}
}

func (c *Connection) SetDNSStrategy(strategy gopxgrid.INETFamilyStrategy) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.dnsStrategy = strategy
	c.unsaved[models.ClientColumns.Attributes] = struct{}{}
}

func (c *Connection) SetCA(ca []string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.tlsCfg.CA = ca
	c.unsaved[models.ClientColumns.Attributes] = struct{}{}
}

func (c *Connection) SetInsecureTLS(skip bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.tlsCfg.InsecureSkipVerify = skip
	c.unsaved[models.ClientColumns.Attributes] = struct{}{}
}

func (c *Connection) SetClientName(name string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.clientName = name
	c.unsaved[models.ClientColumns.ClientName] = struct{}{}
}

func (c *Connection) SetPrimaryNode(fqdn string) {
	c.SetPrimaryNodeWithControlPort(fqdn, DefaultControlPort)
}

func (c *Connection) SetPrimaryNodeWithControlPort(fqdn string, port int) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.primaryNode = Node{FQDN: fqdn, ControlPort: port}
	c.unsaved[models.ClientColumns.Primary] = struct{}{}
}

func (c *Connection) ClearSecondaryNodes() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.secondaryNodes = nil
	c.unsaved[models.ClientColumns.Secondaries] = struct{}{}
}

func (c *Connection) AddSecondaryNode(fqdn string) {
	c.AddSecondaryNodeWithControlPort(fqdn, DefaultControlPort)
}

func (c *Connection) AddSecondaryNodeWithControlPort(fqdn string, port int) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.secondaryNodes = append(c.secondaryNodes, Node{FQDN: fqdn, ControlPort: port})
	c.unsaved[models.ClientColumns.Secondaries] = struct{}{}
}

func (c *Connection) setPasswordBasedAuth(password string) {
	c.credentials = Credentials{
		Type:        CredentialsTypePassword,
		NodeName:    c.clientName,
		Password:    password,
		Certificate: "",
		PrivateKey:  "",
	}
}

func (c *Connection) SetPasswordBasedAuth(password string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.setPasswordBasedAuth(password)
	c.unsaved[models.ClientColumns.Credentials] = struct{}{}
}

func (c *Connection) setCertificateBasedAuth(certPEMBlock, keyPEMBlock string, chain []string) {
	c.credentials = Credentials{
		Type:        CredentialsTypeCertificate,
		Certificate: certPEMBlock,
		PrivateKey:  keyPEMBlock,
		Chain:       chain,
		NodeName:    "",
		Password:    "",
	}
}

func (c *Connection) SetCertificateBasedAuth(certPEMBlock, keyPEMBlock string, chain []string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.setCertificateBasedAuth(certPEMBlock, keyPEMBlock, chain)
	c.unsaved[models.ClientColumns.Credentials] = struct{}{}
}

func (c *Connection) SetLogLevel(level string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.log.Level(level)
}

// func (c *Connection) SetLogger(logger *zerolog.Logger) {
// 	c.lock.Lock()
// 	defer c.lock.Unlock()

// 	c.log = logger
// }
