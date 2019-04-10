package config

// Config is used to handle configuration of dpinner
type Config struct {
	Discord  `json:"discord"`
	Temporal `json:"temporal"`
}

// Discord is used to configure discord both authentication
type Discord struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Token        string `json:"token"`
}

// Temporal is used to configure our IPFS and Temporal connection
type Temporal struct {

	// WarpURL is the URL of our warp ipfs http api reverse proxy
	WarpURL string `json:"warp_url"`
	// User is the username of an account with Temporal
	User string `json:"user"`
	// Pass is the password of an account with Temporal
	Pass string `json:"pass"`
}
