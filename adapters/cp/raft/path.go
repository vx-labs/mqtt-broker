package raft

import (
	"github.com/vx-labs/mqtt-broker/path"
)

func dataDir() string {
	return path.DataDir()
}

func buildDataDir(id string) string {
	return path.ServiceDataDir(id, "raft")
}
