package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	ipfscliwrapper "github.com/bartmika/ipfs-cli-wrapper"
)

/*
DESCRIPTION:
Here is an example of utilizing the `WithContinousOperation` option which will
cause `ipfs` binary to run contiously in the background of the operating system
regardless of how many times your program terminates / restarts. The reason for
this is that the `WithContinousOperation` causes `ipfs` to be loaded up as
process disassociated from the Go process our app and therefore `ipfs` will
run independently.

HOW TO RUN:
STEP 1: While in this project root directory, run the following:
$ go work use ./examples/continous

STEP 2: Go to this directory.
$ cd ./examples/continous

STEP 3: Run the code.
$ go run main.go
*/
func main() {
	wrapper, initErr := ipfscliwrapper.NewDaemonLauncher(ipfscliwrapper.WithContinousOperation())
	if initErr != nil {
		log.Fatalf("failed creating ipfs-cli-wrapper: %v", initErr)
	}
	if wrapper == nil {
		log.Fatal("cannot return nil wrapper")
	}

	// Create a context that will cancel after 2 minutes
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// This code will execute `ipfs` binary in `daemon` mode so it will run
	// in the background of this application and not interrupt the flow of
	// this function so the code is not blocking.
	if startErr := wrapper.StartDaemonInBackground(); startErr != nil {
		log.Fatal(startErr)
	}

	// Channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	// Register for specific signals
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either the context to be done or an OS signal to be received
	select {
	case <-ctx.Done():
		// Wait for the context to be done (which will be in 2 minutes).
		// While you are waiting, checkout the link `http://127.0.0.1:5002/webui` to
		// confirm the code is working. If you see a GUI then you have successfully
		// executed `ipfs` binary from this app.
		log.Println("Context deadline reached, terminating process...")
	case sig := <-sigChan:
		// OS signal received, terminate the process gracefully
		log.Printf("Received signal: %v, terminating process...", sig)
	}

	// After 2 minutes, kill the process. However because we are using
	// continous operation, this `ShutdownDaemon` will do nothing! The
	// `ipfs` binary will continue to operate in the background. If you re-run
	// this program you'll see it start quick with `ipfs` binary already
	// running in the background in daemon mode.
	if endErr := wrapper.ShutdownDaemon(); endErr != nil {
		log.Fatal(endErr)
	}

	// However, if you want to terminate `ipfs` binary then uncomment the
	// following code and you will see it work!
	// if endErr := wrapper.ForceShutdownDaemon(); endErr != nil {
	// 	log.Fatal(endErr)
	// }
}
