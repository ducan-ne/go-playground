module clgt.io/go-playground

go 1.16

require (
	github.com/certifi/gocertifi v0.0.0-20200922220541-2c3bb06c6054 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/hashicorp/raft v1.3.1
	github.com/hashicorp/raft-boltdb v0.0.0-20210422161416-485fa74b0b01
	go.etcd.io/etcd/client/v3 v3.5.0-alpha.0 // indirect
	go.etcd.io/etcd/raft/v3 v3.5.0-alpha.0
	go.etcd.io/etcd/server/v3 v3.0.0-20210320072418-e51c697ec6e8
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb
)

replace (
	go.etcd.io/etcd/client/v3 => go.etcd.io/etcd/client/v3 v3.0.0-20210320072418-e51c697ec6e8
	go.etcd.io/etcd/raft/v3 => go.etcd.io/etcd/raft/v3 v3.0.0-20210320072418-e51c697ec6e8
)
