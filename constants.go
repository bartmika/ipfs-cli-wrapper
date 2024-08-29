package ipfscliwrapper

import "fmt"

// Constants related to the IPFS binary and data directory paths.
const (
	// IPFSBinaryFilePath defines the path to the IPFS binary executable
	// (commonly known as 'kubo'). This path is used when executing IPFS
	// commands via the command line interface in the application.
	IPFSBinaryFilePath = "./bin/kubo/ipfs"

	// IPFSDataDirPath defines the path to the directory where IPFS stores
	// its data, including the repository and configuration files. This path
	// is crucial for ensuring the IPFS node has access to its necessary
	// data files during operation.
	IPFSDataDirPath = "./bin/kubo/data"

	// IPFSDenylistDirPath defines the path to the denylist directory within
	// the IPFS data directory. Denylists are used to block or restrict
	// access to certain content on the IPFS network by specifying content
	// that should not be accessed or shared.
	IPFSDenylistDirPath = IPFSDataDirPath + "/denylists/"
)

// Constants representing various types of pins in IPFS.
const (
	// AllPinType represents the option to list all types of pinned objects in IPFS.
	// This type includes recursive, direct, and indirect pins.
	AllPinType = "all"

	// RecursivePinType represents recursively pinned objects in IPFS.
	// Recursively pinned objects ensure that all their linked content (descendants)
	// is also pinned, keeping the entire data tree available on the node.
	RecursivePinType = "recursive"

	// IndirectPinType represents indirectly pinned objects in IPFS.
	// Indirect pins occur when an object is pinned due to being linked by
	// a recursively pinned object, rather than being pinned directly.
	IndirectPinType = "indirect"

	// DirectPinType represents directly pinned objects in IPFS.
	// Direct pins ensure that only the specified object is kept available,
	// without recursively pinning any linked content.
	DirectPinType = "direct"
)

// getDownloadURL provides a download link for a zipped binary of the `ipfs` executable
// based on the specified operating system and architecture.
//
// The function determines the correct download URL by matching the given `os` and `arch`
// parameters to a pre-defined map of URLs. These URLs correspond to official releases
// of the IPFS Kubo binaries hosted at https://dist.ipfs.tech/#kubo.
//
// Supported operating systems include Darwin (macOS), Linux, FreeBSD, OpenBSD, and Windows,
// and supported architectures include arm, arm64, 386, and amd64. The returned URL points
// to a compressed archive (either .tar.gz or .zip, depending on the OS) that contains
// the IPFS binary for the specified platform.
//
// Parameters:
//   - os: A string representing the operating system. Expected values include "darwin", "linux",
//     "freebsd", "openbsd", and "windows".
//   - arch: A string representing the CPU architecture. Expected values include "arm", "arm64",
//     "386", and "amd64".
//
// Returns:
//   - (string, error): The function returns a string containing the download URL for the
//     requested binary. If the combination of operating system and architecture
//     is not supported, it returns an empty string and an error.
//
// Example usage:
//
//	url, err := getDownloadURL("linux", "amd64")
//	if err != nil {
//	    log.Fatalf("Failed to get download URL: %v", err)
//	}
//	fmt.Println("Download URL:", url)
//
// Errors:
//   - The function returns an error if the specified operating system and architecture combination
//     is not found in the internal map. The error message will indicate the unsupported OS and
//     architecture combination, helping developers identify unsupported platform configurations.
//
// Note:
//   - This function relies on hardcoded URLs for specific versions (e.g., v0.29.0) of the Kubo binaries.
//     To update the version or add support for additional OS/arch combinations, modify the `urlsMap`
//     in the function accordingly.
func getDownloadURL(os string, arch string) (string, error) {
	urlsMap := map[string]map[string]string{
		"darwin": map[string]string{
			"arm64": "https://dist.ipfs.tech/kubo/v0.29.0/kubo_v0.29.0_darwin-arm64.tar.gz",
			"amd64": "https://dist.ipfs.tech/kubo/v0.29.0/kubo_v0.29.0_darwin-amd64.tar.gz",
		},
		"linux": map[string]string{
			"arm":   "https://dist.ipfs.tech/kubo/v0.29.0/kubo_v0.29.0_linux-arm.tar.gz",
			"arm64": "https://dist.ipfs.tech/kubo/v0.29.0/kubo_v0.29.0_linux-arm64.tar.gz",
			"386":   "https://dist.ipfs.tech/kubo/v0.29.0/kubo_v0.29.0_linux-386.tar.gz",
			"amd64": "https://dist.ipfs.tech/kubo/v0.29.0/kubo_v0.29.0_linux-amd64.tar.gz",
		},
		"freebsd": map[string]string{
			"arm":   "https://dist.ipfs.tech/kubo/v0.29.0/kubo_v0.29.0_freebsd-arm.tar.gz",
			"386":   "https://dist.ipfs.tech/kubo/v0.29.0/kubo_v0.29.0_freebsd-386.tar.gz",
			"amd64": "https://dist.ipfs.tech/kubo/v0.29.0/kubo_v0.29.0_freebsd-amd64.tar.gz",
		},
		"openbsd": map[string]string{
			"arm":   "https://dist.ipfs.tech/kubo/v0.29.0/kubo_v0.29.0_openbsd-arm.tar.gz",
			"386":   "https://dist.ipfs.tech/kubo/v0.29.0/kubo_v0.29.0_openbsd-386.tar.gz",
			"amd64": "https://dist.ipfs.tech/kubo/v0.29.0/kubo_v0.29.0_openbsd-amd64.tar.gz",
		},
		"windows": map[string]string{
			"arm":   "https://dist.ipfs.tech/kubo/v0.29.0/kubo_v0.29.0_windows-arm64.zip",
			"386":   "https://dist.ipfs.tech/kubo/v0.29.0/kubo_v0.29.0_windows-386.zip",
			"amd64": "https://dist.ipfs.tech/kubo/v0.29.0/kubo_v0.29.0_windows-amd64.zip",
		},
	}

	val, ok := urlsMap[os][arch]
	if !ok {
		return "", fmt.Errorf("could not find downloadable link for operating system `%s` and architecture `%s`", os, arch)
	}
	return val, nil
}

// IpfsNodeInfo represents the structured data of the `id` command results.
type IpfsNodeInfo struct {
	ID              string   `json:"ID"`
	PublicKey       string   `json:"PublicKey"`
	Addresses       []string `json:"Addresses"`
	AgentVersion    string   `json:"AgentVersion"`
	ProtocolVersion string   `json:"ProtocolVersion"`
}
