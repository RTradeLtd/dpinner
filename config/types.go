package config

// Config is used to handle configuration of dpinner
type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Token        string `json:"token"`
	IPFSURL      string `json:"ipfs_url"`
}
