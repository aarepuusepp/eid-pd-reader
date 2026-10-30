module ee.aarepuusepp.eid-pd-reader

go 1.27.3

require (
	github.com/aarepuusepp/libcdoc-go/v2 v2.0.4
	github.com/aarepuusepp/libopenpace-go v1.1.1
	github.com/ebfe/scard v0.0.0-20241214075232-7af069cabc25
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/google/flatbuffers v25.12.19+incompatible // indirect
	github.com/miekg/pkcs11 v1.1.2 // indirect
	go.gnd.pw/crypto v0.0.0-20231118094619-86ae7742a3a2 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	software.sslmate.com/src/go-pkcs12 v0.7.4 // indirect
)

// Local development: use local libopenpace-go instead of remote version
replace github.com/aarepuusepp/libopenpace-go v1.1.1 => ../libopenpace-go

// Local development: use workspace libcdoc-go for CDOC/PKCS#11 decryption.
replace github.com/aarepuusepp/libcdoc-go/v2 v2.0.4 => ../../libcdoc-go
