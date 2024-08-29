// ipfscliwrapper is a package that manages running an IPFS node in the background
// and offers a user-friendly interface, enabling you to build IPFS-embedded
// Golang applications more easily.
package ipfscliwrapper

import "context"

// IpfsCliWrapper interface represents a wrapper around the `ipfs` executable binary
// in the operating system, providing methods to control the IPFS daemon and perform
// various operations such as adding files, retrieving content, pinning, and garbage collection.
type IpfsCliWrapper interface {
	// StartDaemonInBackground starts the IPFS daemon process in the background,
	// making it ready to accept API requests. It should ensure that the daemon
	// runs independently of the calling application.
	//
	// Returns an error if the daemon fails to start.
	StartDaemonInBackground() error

	// ShutdownDaemon gracefully shuts down the running IPFS daemon.
	// It sends a termination signal to the daemon process, allowing it
	// to perform cleanup tasks before shutting down.
	//
	// Returns an error if the daemon could not be shut down.
	ShutdownDaemon() error

	// ForceShutdownDaemon immediately terminates the IPFS daemon process,
	// without allowing it to perform any cleanup. This is a forceful operation
	// that should be used when the daemon does not respond to a graceful shutdown.
	//
	// Returns an error if the daemon could not be forcefully terminated.
	ForceShutdownDaemon() error

	// AddFile adds a file to the IPFS network using its file path. The function
	// executes the `ipfs add` command to store the file in the IPFS node.
	//
	// Parameters:
	//   ctx - Context for controlling cancellation and deadlines.
	//   filepath - The path to the file to be added to IPFS.
	//
	// Returns:
	//   The CID (Content Identifier) of the added file on success.
	//   An error if the file could not be added.
	AddFile(ctx context.Context, filepath string) (string, error)

	// AddFileContent adds a file to the IPFS network from a byte slice containing
	// the file content, rather than a file path. The function handles the creation
	// and storage of the file directly in the IPFS node.
	//
	// Parameters:
	//   ctx - Context for controlling cancellation and deadlines.
	//   filename - The name of the file to be used when adding to IPFS.
	//   fileContent - The byte slice containing the content of the file.
	//
	// Returns:
	//   The CID (Content Identifier) of the added file on success.
	//   An error if the file could not be added.
	AddFileContent(ctx context.Context, filename string, fileContent []byte) (string, error)

	// GetFile retrieves a file from the IPFS network using its CID (Content Identifier).
	// The function executes the `ipfs get` command, which downloads the file from the
	// IPFS network to the local machine.
	//
	// Parameters:
	//   ctx - Context for controlling cancellation and deadlines.
	//   cid - The CID of the file to be retrieved from IPFS.
	//
	// Returns an error if the file could not be retrieved.
	GetFile(ctx context.Context, cid string) error

	// Cat retrieves the content of a file from the IPFS network using its CID and returns it as a byte slice.
	// The function executes the `ipfs cat` command, which outputs the file content directly.
	//
	// Parameters:
	//   ctx - Context for controlling cancellation and deadlines.
	//   cid - The CID of the file whose content is to be retrieved from IPFS.
	//
	// Returns:
	//   A byte slice containing the file content on success.
	//   An error if the file content could not be retrieved.
	Cat(ctx context.Context, cid string) ([]byte, error)

	// ListPins retrieves a list of all pinned objects' CIDs from the IPFS node.
	// The function executes the `ipfs pin ls` command to fetch the list of pins.
	//
	// Parameters:
	//   ctx - Context for controlling cancellation and deadlines.
	//
	// Returns:
	//   A slice of strings, each representing a CID of a pinned object.
	//   An error if the pins could not be listed.
	ListPins(ctx context.Context) ([]string, error)

	// ListPinsByType retrieves a list of pinned objects' CIDs from the IPFS node
	// filtered by a specific type (e.g., recursive, direct).
	//
	// Parameters:
	//   ctx - Context for controlling cancellation and deadlines.
	//   typeID - The type of pins to list (e.g., "all", "recursive", "direct", "indirect").
	//
	// Returns:
	//   A slice of strings, each representing a CID of a pinned object of the specified type.
	//   An error if the pins could not be listed.
	ListPinsByType(ctx context.Context, typeID string) ([]string, error)

	// Pin pins an object in the IPFS node using its CID, ensuring the object
	// remains available locally on the IPFS node and is not removed during
	// garbage collection.
	//
	// Parameters:
	//   ctx - Context for controlling cancellation and deadlines.
	//   cid - The CID of the object to pin in IPFS.
	//
	// Returns an error if the object could not be pinned.
	Pin(ctx context.Context, cid string) error

	// Unpin removes a pinned object from the IPFS node, making it eligible
	// for removal during garbage collection if it is no longer needed.
	//
	// Parameters:
	//   ctx - Context for controlling cancellation and deadlines.
	//   cid - The CID of the object to unpin from IPFS.
	//
	// Returns an error if the object could not be unpinned.
	Unpin(ctx context.Context, cid string) error

	// GarbageCollection runs the garbage collection process on the IPFS node,
	// removing any unpinned objects that are no longer needed, freeing up space.
	//
	// Parameters:
	//   ctx - Context for controlling cancellation and deadlines.
	//
	// Returns an error if the garbage collection process failed.
	GarbageCollection(ctx context.Context) error
}

// Option is a functional option type that allows us to configure the IpfsCliWrapper.
type Option func(*ipfsCliWrapper)
