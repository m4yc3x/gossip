module gossip-server

go 1.22.1

replace gossip_common => ../gossip-common

require gossip_common v0.0.0-00010101000000-000000000000

require (
	github.com/ProtonMail/go-crypto v1.0.0 // indirect
	github.com/cloudflare/circl v1.3.3 // indirect
	golang.org/x/crypto v0.22.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/term v0.19.0 // indirect
)
