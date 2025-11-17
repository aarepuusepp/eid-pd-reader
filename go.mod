module ee.aarepuusepp.eid-pd-reader

go 1.25.4

require (
	github.com/ebfe/scard v0.0.0-20241214075232-7af069cabc25
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/aarepuusepp/libopenpace-go v1.0.0
	go.gnd.pw/crypto v0.0.0-20231118094619-86ae7742a3a2 // indirect
)

// Local development: use local libopenpace-go instead of remote version
replace github.com/aarepuusepp/libopenpace-go v1.0.0 => ../libopenpace-go
