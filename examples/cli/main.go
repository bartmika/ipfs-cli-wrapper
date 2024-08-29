package main

import (
	"context"
	"log"
	"os"
	"time"

	ipfscliwrapper "github.com/bartmika/ipfs-cli-wrapper"
)

/*
DESCRIPTION
This is an example of how to use the wrapper functions over the `ipfs` binary
command line interface.

HOW TO RUN:
STEP 1: While in this project root directory, run the following:
$ go work use ./examples/cli

STEP 2: Go to this directory.
$ cd ./examples/cli

STEP 3: Run the code.
$ go run main.go
*/

func main() {
	// Check to make sure we have a sample file we can use in our demo...
	if _, err := os.Stat("./sample.txt"); err != nil {
		log.Fatalf("Our sample file does not exist or error reading from file: %v", err)
	}

	// Step 1: Start IPFS.

	wrapper, initErr := ipfscliwrapper.NewWrapper(
		ipfscliwrapper.WithForcedShutdownDaemonOnStartup(),
	)
	if initErr != nil {
		log.Fatalf("failed creating ipfs-cli-wrapper: %v", initErr)
	}
	if wrapper == nil {
		log.Fatal("cannot return nil wrapper")
	}
	defer func(wrap ipfscliwrapper.IpfsCliWrapper) {
		// Step X: Close our IPFS.
		if endErr := wrap.ShutdownDaemon(); endErr != nil {
			log.Fatal(endErr)
		}
	}(wrapper)

	if startErr := wrapper.StartDaemonInBackground(); startErr != nil {
		log.Fatal(startErr)
	}

	// Get our daemon's id values:
	nodeInfo, err := wrapper.Id(context.Background())
	if err != nil {
		log.Fatal("failed adding file")
	}
	log.Println("id:", nodeInfo)

	// Step 2: Add our file.
	cid, addFileErr := wrapper.AddFile(context.Background(), "./sample.txt")
	if addFileErr != nil {
		log.Fatal("failed adding file")
	}

	log.Println("File added successfully to IPFS and returned content ID:", cid)

	// Step 3: Get our file.
	getFileErr := wrapper.GetFile(context.Background(), cid)
	if getFileErr != nil {
		log.Fatal("failed getting file")
	}
	log.Println("Successfully retrieved file from IPFS and saved locally")

	// Step 5: Print the contents of the file.
	fileContent, catErr := wrapper.Cat(context.Background(), cid)
	if catErr != nil {
		log.Fatal("failed getting file")
	}
	log.Println("Successfully retrieved file from IPFS with output:", string(fileContent))

	// Step 6: Pin
	if pinErr := wrapper.Pin(context.Background(), cid); pinErr != nil {
		log.Fatalf("failed pinning: %v", pinErr)
	}
	log.Println("Successfully pinned file on IPFS")

	// Step 7: List pins
	pins, pinErr := wrapper.ListPins(context.Background())
	if pinErr != nil {
		log.Fatalf("failed pinning: %v", pinErr)
	}
	log.Println("Successfully listed pins from IPFS:", pins)

	// Step 7: Unpin
	if unpinErr := wrapper.Unpin(context.Background(), cid); unpinErr != nil {
		log.Fatalf("failed unpinning: %v", unpinErr)
	}
	log.Println("Successfully uninned file on IPFS")

	// Give artifical delay
	log.Println("closing in 5 seconds...")
	time.Sleep(5 * time.Second)
}
