# IPFS CLI Wrapper
[![GoDoc](https://godoc.org/github.com/gomarkdown/markdown?status.svg)](https://pkg.go.dev/github.com/bartmika/ipfs-cli-wrapper)
[![Go Report Card](https://goreportcard.com/badge/github.com/bartmika/ipfs-cli-wrapper)](https://goreportcard.com/report/github.com/bartmika/ipfs-cli-wrapper)
[![License](https://img.shields.io/github/license/bartmika/ipfs-cli-wrapper)](https://github.com/bartmika/ipfs-cli-wrapper/blob/master/LICENSE)
![Go version](https://img.shields.io/github/go-mod/go-version/bartmika/ipfs-cli-wrapper)

**IPFS CLI Wrapper** is a Go package that manages [IPFS](https://github.com/ipfs/kubo) by running the `ipfs` binary as a separate process alongside your Go application. This package allows your Go application to start an IPFS node, connect it to the IPFS network, and interact with the node via the [HTTP Kubo RPC API](https://docs.ipfs.tech/reference/kubo/rpc/) or commands. Your Go application can also gracefully shut down the IPFS node when needed.

## Key Features

1. **Automatic IPFS Binary Download**  
   Upon the first usage in your Go application, the package automatically downloads the appropriate IPFS binary from the [official IPFS distribution site](https://dist.ipfs.tech/#kubo), based on your machine's chipset (e.g., `amd64`, `riscv`, `arm`, etc.). The download process is blocking, but once the binary is downloaded, subsequent invocations are quick.

    Example:
    ```go
    package main

    import (
        "log"

        ipfswrap "github.com/bartmika/ipfs-cli-wrapper"
    )

    func main() {
        // This step will download the IPFS binary if not already present.
        // Note: This is a blocking operation during the initial download.
        wrapper, initErr := ipfswrap.NewWrapper()
        if initErr != nil {
            log.Fatalf("Failed to create IPFS wrapper: %v", initErr)
        }
        if wrapper == nil {
            log.Fatal("Failed: wrapper cannot be nil")
        }

        // Continue with the rest of your application...
    }
    ```

2. **Starting the IPFS Daemon in Background**  
   Your Go application can start the IPFS daemon concurrently, allowing it to run in the background. This operation is non-blocking, so your application continues running while IPFS is managed by the package.

    Example:
    ```go
    // Continued from the previous example...

    // Start the IPFS daemon in the background (non-blocking).
    if startErr := wrapper.StartDaemonInBackground(); startErr != nil {
        log.Fatal(startErr)
    }

    // Ensure IPFS daemon is gracefully shut down when your application exits.
    defer func() {
        if endErr := wrapper.ShutdownDaemon(); endErr != nil {
            log.Fatal(endErr)
        }
    }()

    // Continue with the rest of your application...
    ```

3. **Interacting with the IPFS Node**  
   Once the IPFS daemon is running, your Go application can interact with the node using the [HTTP Kubo RPC API](https://docs.ipfs.tech/reference/kubo/rpc/).

    Example:
    ```go
    package main

    import (
        "context"
        "fmt"
        "log"
        "net/http"
        "os"
        "strings"
        "time"

        ipfsFiles "github.com/ipfs/go-ipfs-files"
        "github.com/ipfs/kubo/client/rpc"

        ipfscliwrapper "github.com/bartmika/ipfs-cli-wrapper"
    )

    func main() {
        wrapper, initErr := ipfscliwrapper.NewWrapper()
        if initErr != nil {
            log.Fatalf("failed creating ipfs-cli-wrapper: %v", initErr)
        }
        if wrapper == nil {
            log.Fatal("cannot return nil wrapper")
        }

        if startErr := wrapper.StartDaemonInBackground(); startErr != nil {
            log.Fatal(startErr)
        }
        defer func() {
            if endErr := wrapper.ShutdownDaemon(); endErr != nil {
                log.Fatal(endErr)
            }
        }()

        httpClient := &http.Client{}
        httpApi, err := rpc.NewURLApiWithClient("http://127.0.0.1:5001", httpClient)
        if err != nil {
            log.Printf("failed loading ipfs daemon: %v\n", err)
            return
        }

        content := strings.NewReader("Hellow world via `github.com/bartmika/ipfs-cli-wrapper/examples/rpc/main.go`!")
        p, err := httpApi.Unixfs().Add(context.Background(), ipfsFiles.NewReaderFile(content))
        if err != nil {
            log.Printf("failed adding data to ipfs: %v\n", err)
            return
        }

        fmt.Printf("Data successfully stored in IPFS with file CID: %v\n", p)
    }
    ```
*Note: You can see the full code of above in [rpc example code](https://github.com/bartmika/ipfs-cli-wrapper/blob/main/examples/rpc/main.go) folder.*

## Installation

**STEP 1 - Install our package**

```shell
go get github.com/bartmika/ipfs-cli-wrapper
```

**STEP 2 - Update your `.gitgnore` to not include binary**

Inside your projects git repository, update your `.gitgnore` to include the following so your repo doesn't save the `ipfs` binary!

```text
bin
./bin
./bin/*
```

**STEP 3 - Open port `4001` on your machine**

On your server, open port `4001` as this is called the called the **Swarm port** that `ipfs` uses to connect to the IPFS network.

## Documentation

All [documentation](https://pkg.go.dev/github.com/bartmika/ipfs-cli-wrapper) can be found here.

## Usage

See [examples](https://github.com/bartmika/ipfs-cli-wrapper/tree/main/examples) folder.

## Contributing

Found a bug? Want a feature to improve your developer experience when finding the difference? Please create an [issue](https://github.com/bartmika/ipfs-cli-wrapper/issues).

## License
Made with ❤️  by [Bartlomiej Mika](https://bartlomiejmika.com).   
The project is licensed under the [ISC License](LICENSE).
