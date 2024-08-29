package main

import (
	"log"
	"time"

	ipfscliwrapper "github.com/bartmika/ipfs-cli-wrapper"
)

/*
DESCRIPTION
This example shows you how to load a denylist to be applied when you startup
your `ipfs` binary so your node will be blocking certain `cid` values.

HOW TO RUN:
STEP 1: While in this project root directory, run the following:
$ go work use ./examples/simple

STEP 2: Go to this directory.
$ cd ./examples/denylist

STEP 3: Run the code.
$ go run main.go
*/

func main() {
	// Note: Link taken from official docs via https://github.com/ipfs/kubo/blob/master/docs/content-blocking.md#denylist-file-format.
	wrapper, initErr := ipfscliwrapper.NewWrapper(
		ipfscliwrapper.WithDenylist("badbits.deny", "https://badbits.dwebops.pub/badbits.deny"),
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

	// Artifically wait 10 seconds...
	log.Println("waiting for 10 seconds...")
	time.Sleep(10 * time.Second)

	if endErr := wrapper.ShutdownDaemon(); endErr != nil {
		log.Fatal(endErr)
	}
}
