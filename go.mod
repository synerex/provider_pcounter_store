module pcounter-store

go 1.13

require (
	github.com/golang/protobuf v1.5.0
	github.com/synerex/proto_pcounter v0.0.6
	github.com/synerex/synerex_api v0.3.1
	github.com/synerex/synerex_proto v0.1.6
	github.com/synerex/synerex_sxutil v0.4.9
	google.golang.org/protobuf v1.27.1 // indirect
)

replace github.com/synerex/proto_pcounter v0.0.6 => github.com/nagata-yoshiteru/proto_pcounter v0.0.11
