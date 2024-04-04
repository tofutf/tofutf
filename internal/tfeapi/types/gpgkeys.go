package types

// ISO8601 is used to format gpg key timestamps.
const ISO8601 = "2006-01-02T15:04:05.999Z"

type GPGKey struct {
	ID             string  `jsonapi:"primary,gpg-keys"`
	ASCIIArmor     string  `jsonapi:"attribute" json:"ascii-armor"`
	CreatedAt      string  `jsonapi:"attribute" json:"created-at"`
	KeyID          string  `jsonapi:"attribute" json:"key-id"`
	Namespace      string  `jsonapi:"attribute" json:"namespace"`
	Source         string  `jsonapi:"attribute" json:"source"`
	SourceURL      *string `jsonapi:"attribute" json:"source-url"`
	TrustSignature string  `jsonapi:"attribute" json:"trust-signature"`
	UpdatedAt      string  `jsonapi:"attribute" json:"updated-at,omitempty"`
}

// GPGKeyCreateOptions represents all the available options used to create a GPG key.
type GPGKeyCreateOptions struct {
	Type       string `jsonapi:"primary,gpg-keys"`
	Namespace  string `jsonapi:"attribute" json:"namespace"`
	ASCIIArmor string `jsonapi:"attribute" json:"ascii-armor"`
}

// GPGKeyCreateOptions represents all the available options used to update a GPG key.
type GPGKeyUpdateOptions struct {
	Type      string `jsonapi:"primary,gpg-keys"`
	Namespace string `jsonapi:"attribute" json:"namespace"`
}
