package connection

type (
	credentialsType string

	Credentials struct {
		Type        credentialsType `json:"type"`
		NodeName    string          `json:"nodename,omitempty"`
		Password    string          `json:"password,omitempty"`
		PrivateKey  string          `json:"privatekey,omitempty"`
		Certificate string          `json:"certificate,omitempty"`
		Chain       []string        `json:"chain,omitempty"`
	}
)

const (
	CredentialsTypePassword    credentialsType = "password"
	CredentialsTypeCertificate credentialsType = "certificate"
)
