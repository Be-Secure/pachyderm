package server //import "go.pachyderm.com/pachyderm/src/pfs/server"

import (
	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pfs/drive"
	"go.pachyderm.com/pachyderm/src/pfs/route"
	"go.pedge.io/google-protobuf"
)

const (
	// InitialCommitID is the initial id before any commits are made in a repository.
	InitialCommitID = "scratch"
)

var (
	// ReservedCommitIDs are the commit ids used by the system.
	ReservedCommitIDs = map[string]bool{
		InitialCommitID: true,
	}
	emptyInstance = &google_protobuf.Empty{}
)

type APIServer interface {
	pfs.APIServer
	route.Frontend
}

type InternalAPIServer interface {
	pfs.InternalAPIServer
	route.Server
}

// NewAPIServer returns a new APIServer.
func NewAPIServer(
	sharder route.Sharder,
	router route.Router,
) APIServer {
	return newAPIServer(
		sharder,
		router,
	)
}

// NewInternalAPIServer returns a new InternalAPIServer.
func NewInternalAPIServer(
	sharder route.Sharder,
	router route.Router,
	driver drive.Driver,
) InternalAPIServer {
	return newInternalAPIServer(
		sharder,
		router,
		driver,
	)
}
