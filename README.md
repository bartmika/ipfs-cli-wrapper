# IPFS CLI Wrapper

[![GoDoc](https://godoc.org/github.com/gomarkdown/markdown?status.svg)](https://pkg.go.dev/github.com/bartmika/ipfs-cli-wrapper)
[![Go Report Card](https://goreportcard.com/badge/github.com/bartmika/ipfs-cli-wrapper)](https://goreportcard.com/report/github.com/bartmika/ipfs-cli-wrapper)
[![License](https://img.shields.io/github/license/bartmika/ipfs-cli-wrapper)](https://github.com/bartmika/ipfs-cli-wrapper/blob/master/LICENSE)
![Go version](https://img.shields.io/github/go-mod/go-version/bartmika/ipfs-cli-wrapper)

`ipfs-cli-wrapper` is a Go package that provides a convenient way to manage [IPFS](https://github.com/ipfs/kubo) nodes by running the IPFS binary as a separate process alongside your Go application. This package allows you to easily start, control, and interact with an IPFS node using the [HTTP Kubo RPC API](https://docs.ipfs.tech/reference/kubo/rpc/).

## Key Features

- **Automatic IPFS Binary Download**: Automatically downloads the appropriate IPFS binary for your machine's architecture on first use.
- **Run IPFS as a Background Process**: Start the IPFS daemon in the background, allowing your application to continue running seamlessly.
- **HTTP API Interaction**: Use the [HTTP Kubo RPC API](https://docs.ipfs.tech/reference/kubo/rpc/) to interact with your IPFS node.

## Getting Started

Follow these steps to start using `ipfs-cli-wrapper` in your Go project:

### Installation

1. **Install the package:**

   ```shell
   go get github.com/bartmika/ipfs-cli-wrapper

2. **Update your .gitignore to exclude the IPFS binary:**

   Add the following to your .gitignore to prevent the binary from being tracked in your version control:

   ```shell
   bin
   ./bin
   ./bin/*
   ```

3. **Open port 4001 on your server:**

   Make sure port 4001 (the Swarm port used by IPFS) is open on your server to allow connections to the IPFS network.

### Basic Usage
Here's a simple example to get started with `ipfs-cli-wrapper`:

```go
package main

import (
    "log"
    ipfswrap "github.com/bartmika/ipfs-cli-wrapper"
)

func main() {
    // Initialize the IPFS CLI wrapper.
    wrapper, err := ipfswrap.NewWrapper()
    if err != nil {
        log.Fatalf("Failed to create IPFS wrapper: %v", err)
    }

    // Start the IPFS daemon in the background.
    if err := wrapper.StartDaemonInBackground(); err != nil {
        log.Fatalf("Failed to start IPFS daemon: %v", err)
    }

    // Ensure the IPFS daemon shuts down gracefully on application exit.
    defer func() {
        if err := wrapper.ShutdownDaemon(); err != nil {
            log.Fatalf("Failed to shut down IPFS daemon: %v", err)
        }
    }()

    // Continue with your application logic...
}
```

### Advanced Example: Interacting with the IPFS Node
Once the IPFS daemon is running, you can interact with it using the HTTP Kubo RPC API:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "strings"

    ipfsFiles "github.com/ipfs/go-ipfs-files"
    "github.com/ipfs/kubo/client/rpc"
    ipfscliwrapper "github.com/bartmika/ipfs-cli-wrapper"
)

func main() {
    wrapper, err := ipfscliwrapper.NewWrapper()
    if err != nil {
        log.Fatalf("Failed to create IPFS wrapper: %v", err)
    }

    if err := wrapper.StartDaemonInBackground(); err != nil {
        log.Fatalf("Failed to start IPFS daemon: %v", err)
    }
    defer func() {
        if err := wrapper.ShutdownDaemon(); err != nil {
            log.Fatalf("Failed to shut down IPFS daemon: %v", err)
        }
    }()

    // Create an IPFS HTTP client.
    httpClient := &http.Client{}
    httpApi, err := rpc.NewURLApiWithClient("http://127.0.0.1:5001", httpClient)
    if err != nil {
        log.Fatalf("Failed to create IPFS HTTP API client: %v", err)
    }

    // Add data to IPFS.
    content := strings.NewReader("Hello world from IPFS CLI Wrapper!")
    p, err := httpApi.Unixfs().Add(context.Background(), ipfsFiles.NewReaderFile(content))
    if err != nil {
        log.Fatalf("Failed to add data to IPFS: %v", err)
    }

    fmt.Printf("Data stored in IPFS with CID: %v\n", p)
}
```

For the full example, see the [RPC example code](https://github.com/bartmika/ipfs-cli-wrapper/blob/main/examples/rpc/main.go).

### Documentation
Detailed documentation can be found on [pkg.go.dev](https://pkg.go.dev/github.com/bartmika/ipfs-cli-wrapper).

### Examples
See the [examples folder](https://github.com/bartmika/ipfs-cli-wrapper/tree/main/examples) for more code samples and use cases.

### Contributing
Found a bug or have a feature request? Please open an [issue](https://github.com/bartmika/ipfs-cli-wrapper/issues). Contributions are welcome!

### License
Made with ❤️ by [Bartlomiej Mika](https://bartlomiejmika.com).   
The project is licensed under the [ISC License](LICENSE).
