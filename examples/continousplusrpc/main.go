package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	ipfsFiles "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/kubo/client/rpc"

	ipfscliwrapper "github.com/bartmika/ipfs-cli-wrapper"
)

/*
DESCRIPTION
This is an example of using our package with the `ipfs` network to post a
test file to it while the app is running in continous mode.

HOW TO RUN:
STEP 1: While in this project root directory, run the following:
$ go work use ./examples/continousplusrpc

STEP 2: Go to this directory.
$ cd ./examples/continousplusrpc

STEP 3: Run the code.
$ go run main.go
*/

func main() {
	wrapper, initErr := ipfscliwrapper.NewDaemonLauncher(
		ipfscliwrapper.WithContinousOperation(),
	)
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

	content := strings.NewReader("Hellow world via `github.com/bartmika/ipfs-cli-wrapper/examples/continousplusrpc/main.go`!")
	p, err := httpApi.Unixfs().Add(context.Background(), ipfsFiles.NewReaderFile(content))
	if err != nil {
		log.Printf("failed adding data to ipfs: %v\n", err)
		return
	}

	fmt.Printf("Data successfully stored in IPFS with file CID: %v\n", p)

	// If you want to verify this file works for yourself, uncomment this
	// artifical delay and then go to the url `http://127.0.0.1:5001/webui`
	// in your browser to see a GUI and navigate to the files and import from IPFS and
	// enter `/ipfs/Qmc691KgD6oMSMuk6Yn3ZKZohADiRan8MmGfCymUwvm81u` and you'll
	// see the contents we set! You might need to change the delay time of
	// `60` seconds to something else depending on your machine.
	time.Sleep(60 * time.Second)
}
