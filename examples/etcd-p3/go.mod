module github.com/jamesstocktonj1/componentize-sdk/examples/etcd-p3

go 1.26

replace github.com/jamesstocktonj1/componentize-sdk/p3 => ../../p3

replace go.etcd.io/etcd/client/pkg/v3 => /home/james/Programming/OpenSource/etcd/client/pkg

require (
	github.com/jamesstocktonj1/componentize-sdk/p3 v0.0.0-00010101000000-000000000000 // indirect
	go.etcd.io/etcd/client/v3 v3.6.13
)

require (
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.7.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	go.bytecodealliance.org/pkg v0.2.3 // indirect
	go.etcd.io/etcd/api/v3 v3.6.13 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.6.13 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/net v0.54.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/grpc v1.79.3 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace github.com/jamesstocktonj1/componentize-sdk/p3/exports/http => ../../p3/exports/http

require github.com/jamesstocktonj1/componentize-sdk/p3/exports/http v0.0.0-00010101000000-000000000000
