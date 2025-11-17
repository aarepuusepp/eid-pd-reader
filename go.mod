module ee.aarepuusepp.eid-pd-reader

go 1.25.4

require (
	github.com/ebfe/scard v0.0.0-20241214075232-7af069cabc25
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/aarepuusepp/libopenpace-go v0.0.0-00010101000000-000000000000
	go.gnd.pw/crypto v0.0.0-20231118094619-86ae7742a3a2 // indirect
)

replace github.com/aarepuusepp/libopenpace-go => ../libopenpace-go
