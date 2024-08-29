package ipfscliwrapper

import "fmt"

const (
	IPFSBinaryFilePath  = "./bin/kubo/ipfs"
	IPFSDataDirPath     = "./bin/kubo/data"
	IPFSDenylistDirPath = IPFSDataDirPath + "/denylists/"
)

const (
	AllPinType       = "all"
	RecursivePinType = "recursive"
	IndirectPinType  = "indirect"
	DirectPinType    = "direct"
)

// getDownloadURL provides a link where to download a zipped up binary of
// the `ipfs` executable based on the operating system and architecture of the
// parameter inputs. The URLs are taken from the official IPFS distribution:
// https://dist.ipfs.tech/#kubo
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
