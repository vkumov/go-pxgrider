package connection

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog"
	gopxgrid "github.com/vkumov/go-pxgrid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	pb "github.com/vkumov/go-pxgrider/pkg"
	"github.com/vkumov/go-pxgrider/server/internal/db/models"
	"github.com/vkumov/go-pxgrider/server/internal/logger"
	"github.com/vkumov/go-pxgrider/server/internal/utils"
)

type (
	ServiceName string

	Node struct {
		FQDN        string `json:"fqdn"`
		ControlPort int    `json:"controlPort,omitempty"`
	}

	tlsCfg struct {
		InsecureSkipVerify bool     `json:"insecureSkipVerify"`
		CA                 []string `json:"ca"`
	}

	Connection struct {
		id             string
		friendlyName   string
		primaryNode    Node
		secondaryNodes []Node
		credentials    Credentials
		state          gopxgrid.AccountState
		description    string
		dns            string
		dnsStrategy    gopxgrid.INETFamilyStrategy
		clientName     string
		tlsCfg         tlsCfg
		owner          string
		topics         map[ServiceName]map[TopicName]*Subscription
		unsaved        map[string]struct{}

		lock sync.Mutex

		log    *logger.Logger
		pxCfg  atomic.Pointer[gopxgrid.PxGridConfig]
		pxCnsm atomic.Pointer[gopxgrid.PxGridConsumer]
		db     atomic.Pointer[sql.DB]
	}

	ConnectionCreate struct {
		FriendlyName   string                      `json:"friendlyName"`
		PrimaryNode    Node                        `json:"primaryNode"`
		SecondaryNodes []Node                      `json:"secondaryNodes,omitempty"`
		Credentials    Credentials                 `json:"credentials"`
		Description    string                      `json:"description,omitempty"`
		DNS            string                      `json:"dns,omitempty"`
		DNSStrategy    gopxgrid.INETFamilyStrategy `json:"dnsStrategy,omitempty"`
		ClientName     string                      `json:"clientName"`
		InsecureTLS    bool                        `json:"insecureTLS,omitempty"`
		CA             []string                    `json:"ca,omitempty"`
	}

	ConnectionUpdate struct {
		FriendlyName   sql.Null[string]                      `json:"friendlyName,omitempty"`
		PrimaryNode    sql.Null[Node]                        `json:"primaryNode,omitempty"`
		SecondaryNodes sql.Null[[]Node]                      `json:"secondaryNodes,omitempty"`
		Credentials    sql.Null[Credentials]                 `json:"credentials,omitempty"`
		State          sql.Null[gopxgrid.AccountState]       `json:"state,omitempty"`
		Description    sql.Null[string]                      `json:"description,omitempty"`
		DNS            sql.Null[string]                      `json:"dns,omitempty"`
		DNSStrategy    sql.Null[gopxgrid.INETFamilyStrategy] `json:"dnsStrategy,omitempty"`
		ClientName     sql.Null[string]                      `json:"clientName,omitempty"`
		Owner          sql.Null[string]                      `json:"owner,omitempty"`
		InsecureTLS    sql.Null[bool]                        `json:"insecureTLS,omitempty"`
		CA             sql.Null[[]string]                    `json:"ca,omitempty"`
	}
)

const (
	// AccountStateUnknown is the default state for an account
	AccountStateUnknown gopxgrid.AccountState = "UNKNOWN"

	DefaultControlPort = 8910
)

func New(db *sql.DB, id, owner string, log *zerolog.Logger, logWriter io.Writer) *Connection {
	// logger := log.With().Str("connection_id", id).Logger()
	l := logger.NewCombined(id, log, db, logWriter, logger.ComponentFieldName, "pxgrid:consumer")
	c := &Connection{
		id:      id,
		state:   AccountStateUnknown,
		owner:   owner,
		log:     l,
		unsaved: make(map[string]struct{}),
	}

	c.db.Store(db)
	c.pxCfg.Store(gopxgrid.NewPxGridConfig())

	return c
}

func NewWithRequest(db *sql.DB, id, owner string, req ConnectionCreate, log *zerolog.Logger, logWriter io.Writer) (*Connection, error) {
	c := New(db, id, owner, log, logWriter)
	c.friendlyName = req.FriendlyName
	c.description = req.Description
	c.dns = req.DNS
	c.dnsStrategy = req.DNSStrategy
	c.clientName = req.ClientName
	c.tlsCfg.InsecureSkipVerify = req.InsecureTLS

	if len(req.CA) > 0 {
		c.tlsCfg.CA = make([]string, len(req.CA))
		copy(c.tlsCfg.CA, req.CA)
	}

	if req.PrimaryNode.ControlPort > 0 {
		c.primaryNode = req.PrimaryNode
	} else {
		c.primaryNode = Node{FQDN: req.PrimaryNode.FQDN, ControlPort: DefaultControlPort}
	}

	for _, node := range req.SecondaryNodes {
		if node.ControlPort > 0 {
			c.secondaryNodes = append(c.secondaryNodes, node)
		} else {
			c.secondaryNodes = append(c.secondaryNodes, Node{FQDN: node.FQDN, ControlPort: DefaultControlPort})
		}
	}

	switch req.Credentials.Type {
	case CredentialsTypeCertificate:
		c.setCertificateBasedAuth(req.Credentials.Certificate, req.Credentials.PrivateKey, req.Credentials.Chain)
	case CredentialsTypePassword:
		c.setPasswordBasedAuth(req.Credentials.Password)
	default:
		return nil, fmt.Errorf("unknown credentials type: %s", req.Credentials.Type)
	}

	return c, nil
}

func (c *Connection) WithDBData(cl *models.Client) error {
	c.friendlyName = cl.FriendlyName.String
	c.clientName = cl.ClientName.String

	if cl.Primary.Valid && !utils.IsEmptyJSON(cl.Primary.JSON) {
		if err := cl.Primary.Unmarshal(&c.primaryNode); err != nil {
			return fmt.Errorf("failed to unmarshal primary node for connection %s: %w", c.id, err)
		}
		if c.primaryNode.ControlPort == 0 {
			c.primaryNode.ControlPort = DefaultControlPort
		}
	}
	if cl.Secondaries.Valid && !utils.IsEmptyJSON(cl.Secondaries.JSON) {
		if err := cl.Secondaries.Unmarshal(&c.secondaryNodes); err != nil {
			return fmt.Errorf("failed to unmarshal secondary nodes for connection %s: %w", c.id, err)
		}
		for i := range c.secondaryNodes {
			if c.secondaryNodes[i].ControlPort == 0 {
				c.secondaryNodes[i].ControlPort = DefaultControlPort
			}
		}
	}
	if cl.Credentials.Valid && !utils.IsEmptyJSON(cl.Credentials.JSON) {
		if err := cl.Credentials.Unmarshal(&c.credentials); err != nil {
			return fmt.Errorf("failed to unmarshal credentials for connection %s: %w", c.id, err)
		}
	}
	if cl.Attributes.Valid && !utils.IsEmptyJSON(cl.Attributes.JSON) {
		attr := new(rawAttributes)
		if err := cl.Attributes.Unmarshal(attr); err != nil {
			return fmt.Errorf("failed to unmarshal attributes for connection %s: %w", c.id, err)
		}

		c.state = gopxgrid.AccountState(attr.State)
		c.description = attr.Description
		c.dns = attr.DNS
		c.dnsStrategy = strategyFromRaw(attr.DNSStrategy)

		if len(attr.CA) > 0 {
			c.tlsCfg.CA = make([]string, len(attr.CA))
			copy(c.tlsCfg.CA, attr.CA)
		}

		c.tlsCfg.InsecureSkipVerify = attr.Verify == "none"
	}
	if cl.Topics.Valid && !utils.IsEmptyJSON(cl.Topics.JSON) {
		if err := cl.Topics.Unmarshal(&c.topics); err != nil {
			return fmt.Errorf("failed to unmarshal topics for connection %s: %w", c.id, err)
		}
	}

	c.ensureAfterDBLoad()

	return c.RebuildPxGridConfig()
}

func (c *Connection) ensureAfterDBLoad() {
	if len(c.topics) > 0 {
		for svc, topics := range c.topics {
			for topic, sub := range topics {
				logger := c.log.With().Str("service", string(svc)).Str("topic", string(topic)).Logger()
				sub.log = &logger
			}
		}
	}
}

func (c *Connection) RebuildPxGridConfig() error {
	pxCfg := c.pxCfg.Load()
	pxCfg.
		SetDNS(c.dns, gopxgrid.DefaultINETFamilyStrategy).
		SetInsecureTLS(c.tlsCfg.InsecureSkipVerify).
		SetDescription(c.description).
		SetNodeName(c.clientName)

	if len(c.tlsCfg.CA) > 0 {
		pool, err := x509PoolFromCA(c.tlsCfg.CA)
		if err != nil {
			return fmt.Errorf("failed to get x509 pool for connection %s: %w", c.id, err)
		}
		pxCfg.SetCA(pool)
	}

	if c.credentials.Type == "certificate" {
		pxCfg.SetAuth(c.clientName, "")
		cert, err := getX509Pair(c.credentials.Certificate, c.credentials.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to get x509 pair for connection %s: %w", c.id, err)
		}
		pxCfg.SetClientCertificate(cert)
	} else {
		pxCfg.SetAuth(c.credentials.NodeName, c.credentials.Password)
	}

	pxCfg.Hosts = make([]gopxgrid.Host, 0)

	if c.primaryNode.FQDN != "" {
		pxCfg.AddHost(c.primaryNode.FQDN, c.primaryNode.ControlPort)
	}

	for _, node := range c.secondaryNodes {
		pxCfg.AddHost(node.FQDN, node.ControlPort)
	}

	pxCfg.SetLogger(&logger.PxGridLog{Logger: c.log})

	c.pxCfg.Store(pxCfg)

	return nil
}

func (c *Connection) Update(ctx context.Context, upd ConnectionUpdate) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	shouldRollback := true
	updatedColumns := make([]string, 0)

	if upd.FriendlyName.Valid {
		defer func(old string) {
			if shouldRollback {
				c.friendlyName = old
			}
		}(c.friendlyName)
		c.friendlyName = upd.FriendlyName.V
		updatedColumns = append(updatedColumns, models.ClientColumns.FriendlyName)
	}

	if upd.PrimaryNode.Valid {
		defer func(old Node) {
			if shouldRollback {
				c.primaryNode = old
			}
		}(c.primaryNode)
		c.primaryNode = upd.PrimaryNode.V
		updatedColumns = append(updatedColumns, models.ClientColumns.Primary)
	}

	if upd.SecondaryNodes.Valid {
		defer func(old []Node) {
			if shouldRollback {
				c.secondaryNodes = old
			}
		}(c.secondaryNodes)
		c.secondaryNodes = upd.SecondaryNodes.V
		updatedColumns = append(updatedColumns, models.ClientColumns.Secondaries)
	}

	if upd.Credentials.Valid {
		defer func(old Credentials) {
			if shouldRollback {
				c.credentials = old
			}
		}(c.credentials)
		c.credentials = upd.Credentials.V
		updatedColumns = append(updatedColumns, models.ClientColumns.Credentials)
	}

	if upd.ClientName.Valid {
		defer func(old string) {
			if shouldRollback {
				c.clientName = old
			}
		}(c.clientName)
		c.clientName = upd.ClientName.V
		updatedColumns = append(updatedColumns, models.ClientColumns.ClientName)
	}

	if upd.Owner.Valid {
		defer func(old string) {
			if shouldRollback {
				c.owner = old
			}
		}(c.owner)
		c.owner = upd.Owner.V
		updatedColumns = append(updatedColumns, models.ClientColumns.Owner)
	}

	oldAttributes := c.getRawAttributes()
	updatedAttributes := false

	if upd.State.Valid {
		c.state = upd.State.V
		updatedAttributes = true
	}

	if upd.Description.Valid {
		c.description = upd.Description.V
		updatedAttributes = true
	}

	if upd.DNS.Valid {
		c.dns = upd.DNS.V
		updatedAttributes = true
	}

	if upd.DNSStrategy.Valid {
		c.dnsStrategy = upd.DNSStrategy.V
		updatedAttributes = true
	}

	if upd.InsecureTLS.Valid {
		c.tlsCfg.InsecureSkipVerify = upd.InsecureTLS.V
		updatedAttributes = true
	}

	if upd.CA.Valid {
		c.tlsCfg.CA = make([]string, len(upd.CA.V))
		copy(c.tlsCfg.CA, upd.CA.V)
		updatedAttributes = true
	}

	if updatedAttributes {
		defer func(old rawAttributes) {
			if shouldRollback {
				c.state = gopxgrid.AccountState(old.State)
				c.description = old.Description
				c.dns = old.DNS
				c.dnsStrategy = strategyFromRaw(old.DNSStrategy)
				c.tlsCfg.InsecureSkipVerify = old.Verify == "none"
				c.tlsCfg.CA = make([]string, len(old.CA))
				copy(c.tlsCfg.CA, old.CA)
			}
		}(oldAttributes)
		updatedColumns = append(updatedColumns, models.ClientColumns.Attributes)
	}

	err := c.store(ctx, updatedColumns)
	if err == nil {
		shouldRollback = false
	}

	return err
}

func (c *Connection) store(ctx context.Context, columns []string) error {
	dbRef := models.Client{ID: c.id}
	for _, cl := range columns {
		switch cl {
		case models.ClientColumns.FriendlyName:
			dbRef.FriendlyName = null.StringFrom(c.friendlyName)
		case models.ClientColumns.Primary:
			if err := dbRef.Primary.Marshal(c.primaryNode); err != nil {
				return err
			}
		case models.ClientColumns.Secondaries:
			if err := dbRef.Secondaries.Marshal(c.secondaryNodes); err != nil {
				return err
			}
		case models.ClientColumns.Credentials:
			if err := dbRef.Credentials.Marshal(c.credentials); err != nil {
				return err
			}
		case models.ClientColumns.ClientName:
			dbRef.ClientName = null.StringFrom(c.clientName)
		case models.ClientColumns.Owner:
			dbRef.Owner = c.owner
		case models.ClientColumns.Attributes:
			attributes := c.getRawAttributes()
			if err := dbRef.Attributes.Marshal(attributes); err != nil {
				return err
			}
		case models.ClientColumns.Topics:
			if err := dbRef.Topics.Marshal(c.topics); err != nil {
				return err
			}
			// case models.ClientColumns.Services:
			// 	if err := dbRef.Services.Marshal(c.services); err != nil {
			// 		return err
			// 	}
		}
	}

	c.log.Debug().Interface("columns", columns).Interface("dbRef", dbRef).Str("id", c.id).
		Msg("storing connection")

	hasId := false
	for _, cl := range columns {
		if cl == models.ClientColumns.ID {
			hasId = true
			break
		}
	}
	if !hasId {
		columns = append(columns, models.ClientColumns.ID)
	}

	columnsNoId := make([]string, 0, len(columns)-1)
	for _, cl := range columns {
		if cl != models.ClientColumns.ID {
			columnsNoId = append(columnsNoId, cl)
		}
	}

	return dbRef.Upsert(ctx, c.db.Load(), true,
		[]string{models.ClientColumns.ID},
		boil.Whitelist(columnsNoId...),
		boil.Whitelist(columns...))
}

func (c *Connection) storeUnsaved(ctx context.Context) error {
	c.lock.Lock()
	unsaved := make([]string, 0, len(c.unsaved))
	for cl := range c.unsaved {
		unsaved = append(unsaved, cl)
	}
	c.lock.Unlock()

	if err := c.store(ctx, unsaved); err != nil {
		return fmt.Errorf("failed to store unsaved fields for connection %s: %w", c.id, err)
	}

	c.lock.Lock()
	c.unsaved = make(map[string]struct{})
	c.lock.Unlock()

	return nil
}

func (c *Connection) Store(ctx context.Context) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	allColumns := []string{
		models.ClientColumns.FriendlyName,
		models.ClientColumns.Primary,
		models.ClientColumns.Credentials,
		models.ClientColumns.Services,
		models.ClientColumns.Attributes,
		models.ClientColumns.Secondaries,
		models.ClientColumns.Owner,
		models.ClientColumns.ClientName,
		models.ClientColumns.Topics,
	}

	return c.store(ctx, allColumns)
}

func (c *Connection) ID() string {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.id
}

func (c *Connection) Name() string {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.friendlyName
}

func (c *Connection) Owner() string {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.owner
}

func (c *Connection) PX() (*gopxgrid.PxGridConsumer, error) {
	if pxCnsm := c.pxCnsm.Load(); pxCnsm != nil {
		return pxCnsm, nil
	}

	return c.RebuildPxGridConsumer()
}

func (c *Connection) ToProto() *pb.Connection {
	c.lock.Lock()
	defer c.lock.Unlock()

	nodes := make([]*pb.Node, 0, len(c.secondaryNodes)+1)
	nodes = append(nodes, &pb.Node{
		Fqdn:        c.primaryNode.FQDN,
		ControlPort: uint32(c.primaryNode.ControlPort),
	})

	for _, node := range c.secondaryNodes {
		nodes = append(nodes, &pb.Node{
			Fqdn:        node.FQDN,
			ControlPort: uint32(node.ControlPort),
		})
	}

	credentials := &pb.Credentials{}
	switch c.credentials.Type {
	case CredentialsTypeCertificate:
		credentials.Type = pb.CredentialsType_CREDENTIALS_TYPE_CERTIFICATE
		credentials.Kind = &pb.Credentials_Certificate{
			Certificate: &pb.CredentialsCertificate{
				Certificate:    c.credentials.Certificate,
				PrivateKey:     c.credentials.PrivateKey,
				CaCertificates: c.credentials.Chain,
			},
		}
	case CredentialsTypePassword:
		credentials.Type = pb.CredentialsType_CREDENTIALS_TYPE_PASSWORD
		credentials.Kind = &pb.Credentials_Password{
			Password: &pb.CredentialsPassword{
				Password: c.credentials.Password,
			},
		}
	}

	var dnsDetails *pb.DNSDetails
	if prsd, err := gopxgrid.ParseDNSHost(c.dns); err == nil {
		dnsDetails = &pb.DNSDetails{
			Dns: &pb.DNS{Ip: prsd.IP.String(), Port: uint32(prsd.Port)},
		}

	} else {
		c.log.Warn().Err(err).Str("dns", c.dns).Msg("failed to parse dns host")
		dnsDetails = &pb.DNSDetails{
			Dns: &pb.DNS{Ip: c.dns, Port: 53},
		}
	}
	switch c.dnsStrategy {
	case gopxgrid.IPv4:
		dnsDetails.Strategy = pb.FamilyPreference_FamilyPreference_IPv4
	case gopxgrid.IPv6:
		dnsDetails.Strategy = pb.FamilyPreference_FamilyPreference_IPv6
	case gopxgrid.IPv46:
		dnsDetails.Strategy = pb.FamilyPreference_FamilyPreference_IPv4AndIPv6
	case gopxgrid.IPv64:
		dnsDetails.Strategy = pb.FamilyPreference_FamilyPreference_IPv6AndIPv4
	}

	// svcs := make(map[string]*pb.Service, len(c.services))
	// for name, svc := range c.services {
	// 	svcs[string(name)] = svc.ToProto()
	// }

	tpcs := make(map[string]*pb.TopicMap, len(c.topics))
	for svc, topics := range c.topics {
		tpc := make(map[string]*pb.Subscription, len(topics))
		for topic, sub := range topics {
			tpc[string(topic)] = sub.ToProto()
		}
		tpcs[string(svc)] = &pb.TopicMap{Subscriptions: tpc}
	}

	return &pb.Connection{
		Id:           c.id,
		FriendlyName: c.friendlyName,
		Nodes:        nodes,
		Credentials:  credentials,
		State:        string(c.state),
		Description:  c.description,
		ClientName:   c.clientName,
		Owner:        &pb.User{Uid: c.owner},
		DnsDetails:   dnsDetails,
	}
}

func (c *Connection) CleanupSubscriptions() {
	c.lock.Lock()
	defer c.lock.Unlock()

	for svc, topics := range c.topics {
		for topic, t := range topics {
			_ = t.s.Unsubscribe()
			delete(topics, topic)
		}
		delete(c.topics, svc)
	}
}

func (c *Connection) RebuildPxGridConsumer() (*gopxgrid.PxGridConsumer, error) {
	err := c.RebuildPxGridConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to rebuild pxgrid config for connection %s: %w", c.id, err)
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	pxCfg := c.pxCfg.Load()
	pxCnsm, err := gopxgrid.NewPxGridConsumer(pxCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pxgrid consumer for connection %s: %w", c.id, err)
	}

	c.pxCnsm.Store(pxCnsm)
	return pxCnsm, nil
}

func (c *Connection) GetServiceByName(name string) (gopxgrid.PxGridService, error) {
	return c.getServiceByName(name)
}

func (c *Connection) Stop() {
	c.CleanupSubscriptions()
	c.log.Stop()
}
