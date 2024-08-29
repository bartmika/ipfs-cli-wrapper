// ipfscliwrapper is a package that manages running an IPFS node in the background
// and offers a user-friendly interface, enabling you to build IPFS-embedded
// Golang applications more easily.
package ipfscliwrapper

import "context"

// IpfsCliWrapper interface for performing operations on the `ipfs` bunary in the system.
type IpfsCliWrapper interface {
	StartDaemonInBackground() error
	ShutdownDaemon() error
	ForceShutdownDaemon() error
	AddFile(ctx context.Context, filepath string) (string, error)
	AddFileContent(ctx context.Context, filename string, fileContent []byte) (string, error)
	GetFile(ctx context.Context, cid string) error
	Cat(ctx context.Context, cid string) ([]byte, error)
	ListPins(ctx context.Context) ([]string, error)
	ListPinsByType(ctx context.Context, typeID string) ([]string, error)
	Pin(ctx context.Context, cid string) error
	Unpin(ctx context.Context, cid string) error
	GarbageCollection(ctx context.Context) error
}
