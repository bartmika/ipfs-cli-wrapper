package ipfscliwrapper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"golift.io/xtractr"

	"github.com/bartmika/ipfs-cli-wrapper/internal/logger"
	"github.com/bartmika/ipfs-cli-wrapper/internal/oskit"
	"github.com/bartmika/ipfs-cli-wrapper/internal/randomkit"
	"github.com/bartmika/ipfs-cli-wrapper/internal/urlkit"
)

// ipfsCliWrapper represents a wrapper around the `ipfs` executable binary,
// providing methods to control the operation of an IPFS node running as a daemon
// in the operating system. This struct abstracts direct interaction with the IPFS
// binary, allowing for simplified management of IPFS processes within applications.
type ipfsCliWrapper struct {
	// logger is used to log various actions and errors occurring within the wrapper.
	logger *slog.Logger

	// ipfsDaemonCmd is a pointer to the command shell instance of the `ipfs` binary
	// running in daemon mode. It allows control over the background process that
	// runs the IPFS daemon.
	ipfsDaemonCmd *exec.Cmd

	// stdout captures the output stream from the console of the `ipfs` daemon,
	// enabling the wrapper to process or log real-time output from the IPFS node.
	stdout io.ReadCloser

	// isDaemonRunning indicates whether the IPFS binary is currently running in daemon mode.
	// This boolean flag is used internally to track the state of the IPFS daemon.
	isDaemonRunning bool

	// isDaemonRunningContinously controls whether the IPFS daemon should run indefinitely.
	// When set to true, the wrapper will prevent the daemon from shutting down unless
	// explicitly instructed to do so via the `ForceShutdown()` method.
	isDaemonRunningContinously bool

	// daemonInitialWarmupDuration specifies an artificial delay period in the `StartDaemonInBackground`
	// function before allowing other operations to proceed. This delay is intended to give the IPFS
	// daemon time to initialize fully, and should be adjusted based on the performance of the
	// underlying machine.
	daemonInitialWarmupDuration time.Duration

	// os stores the operating system on which the wrapper is running. This information may be
	// used for platform-specific adjustments or logging purposes.
	os string

	// arch stores the CPU architecture of the machine on which the wrapper is running. This
	// information is useful for ensuring compatibility with the IPFS binary and for logging.
	arch string
}

// NewWrapper creates a new instance of IpfsCliWrapper with the specified options.
// This function provides a flexible way to initialize the wrapper, allowing customization
// through a set of functional options that modify the default configuration.
//
// Parameters:
//   - options: A variadic list of Option functions that customize the behavior of the wrapper.
//     Each Option function applies a specific modification to the configuration.
//
// Returns:
//   - (IpfsCliWrapper, error): Returns an initialized IpfsCliWrapper instance or an error if
//     the configuration is invalid or if initialization fails.
//
// Example usage:
//
//	wrapper, err := NewWrapper(
//	    WithLogger(myLogger),
//	    WithDaemonWarmupDuration(5 * time.Second),
//	    WithContinuousDaemonRunning(true),
//	)
//	if err != nil {
//	    log.Fatalf("Failed to initialize IPFS CLI wrapper: %v", err)
//	}
//
// Notes:
//   - The wrapper is designed to abstract the complexities of managing the IPFS daemon,
//     providing a simple interface for starting, stopping, and interacting with the daemon.
//   - It is crucial to configure the daemon warmup duration appropriately based on the expected
//     startup time of the IPFS daemon on the host machine. Insufficient warmup time can lead
//     to unexpected errors or failures in subsequent operations.
//   - For long-running IPFS nodes that should not be interrupted, set `isDaemonRunningContinously`
//     to true to ensure the daemon persists until explicitly shut down using `ForceShutdown()`.
func NewWrapper(options ...Option) (IpfsCliWrapper, error) {
	// STEP 1: Create the needed directories in the applications root directory
	// so we can save our binary data into there.
	dirs := []string{
		"./bin", // The root folder which holds all our data we are managing.
		IPFSDataDirPath,
		IPFSDenylistDirPath,
	}
	if err := oskit.CreateDirsIfDoesNotExist(dirs); err != nil {
		log.Fatalf("failed to make directory: %v", err)
	}

	// STEP 2. Get the OS and chip architecture to use so we will know what
	// binary to utilize in our wrapper.

	// Get the architecture of the machine
	archName := runtime.GOARCH

	// Get the operating system
	osName := runtime.GOOS

	// STEP 3: Apply our option conditions.

	wrapper := &ipfsCliWrapper{
		logger:                      logger.NewProvider(),
		isDaemonRunning:             false,
		isDaemonRunningContinously:  false,
		daemonInitialWarmupDuration: time.Duration(5) * time.Second,
		os:                          osName,
		arch:                        archName,
	}

	// Apply all the functional options to configure the client.
	for _, opt := range options {
		opt(wrapper)
	}

	// STEP 4: Check to see if we have our `ipfs` binary ready to execute and if
	// not then we will need to download it and get it ready for execution.
	if _, err := os.Stat(IPFSBinaryFilePath); err != nil {
		if err := downloadAndUnzip(wrapper.logger, wrapper.os, wrapper.arch); err != nil {
			log.Fatalf("failed to get ipfs binary from url: %v", err)
		}

		// STEP 5: Execute our `ipfs` binary `init` command so the application gets
		// setup; however, we will also set the environment variable before
		// executing the command, therefore pointing to a different location for
		// saving data. Please note, ignore error and output here. We do this
		// because if we run `init` again after this app was already called then
		// `ipfs` will return error so we don't care.
		initCmd := exec.Command(IPFSBinaryFilePath, "init")
		initCmd.Env = append(os.Environ(), "IPFS_PATH="+IPFSDataDirPath)

		// Execute the command and check for errors
		if output, err := initCmd.CombinedOutput(); err != nil {
			wrapper.logger.Error("failed to initialize IPFS",
				slog.Any("error", err),
				slog.String("output", string(output)))
			// Log or handle the error appropriately, if needed
		} else {
			wrapper.logger.Debug("IPFS initialization completed successfully",
				slog.String("output", string(output)))
		}

		// Set an artificial delay to give time for the `ipfs` binary to load up.
		// Another perspective is this is the `warmup time`.
		time.Sleep(wrapper.daemonInitialWarmupDuration)
	}

	// Setup the command we will execute in our shell.
	app := IPFSBinaryFilePath
	arg0 := "daemon"
	arg1 := "--enable-gc" // Enable automatic garbage collection in runtime.
	daemonCmd := exec.Command(app, arg0, arg1)

	// Set the environment variable before executing the command
	daemonCmd.Env = append(os.Environ(), "IPFS_PATH="+IPFSDataDirPath)

	// Create a pipe to read the output of the command
	stdout, err := daemonCmd.StdoutPipe()
	if err != nil {
		wrapper.logger.Error("error creating stdout pipe", slog.Any("error", err))
		return nil, fmt.Errorf("Error creating stdout pipe: %v\n", err)
	}

	wrapper.ipfsDaemonCmd = daemonCmd
	wrapper.stdout = stdout

	wrapper.logger.Debug("ipfs daemon wrapper initialized",
		slog.String("os", wrapper.os),
		slog.String("arch", wrapper.arch),
		slog.String("ipfs_bin_path", IPFSBinaryFilePath),
		slog.String("ipfs_data_path", IPFSDataDirPath))

	return wrapper, nil
}

func (wrap *ipfsCliWrapper) StartDaemonInBackground() error {
	// Before we begin our code, let's check if the `ipfs` binary is already
	// running in the background, for whatever reason.
	if isRunningAlready, err := oskit.IsProgramRunning("ipfs"); isRunningAlready || err != nil {
		if isRunningAlready {
			wrap.isDaemonRunning = true
			wrap.logger.Debug("ipfs daemon is already running and waiting for api call from your app")
			return nil
		}

		wrap.logger.Error("is program running err", slog.Any("error", err))
		return fmt.Errorf("is program running error: %v", err)
	}
	wrap.logger.Debug("ipfs daemon is starting...")

	// If `isDaemonRunningContinously` is true then
	if wrap.isDaemonRunningContinously {
		wrap.logger.Debug("continous operation mode detected, ipfs daemon will run independently of this app")

		// Ensure that the process is disassociated from the Go process and will run independently
		wrap.ipfsDaemonCmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true, // Create a new session, which makes the process independent
		}

		// Redirect stdout and stderr to /dev/null to detach from the terminal
		devNull, err := os.Open(os.DevNull)
		if err != nil {
			return err
		}
		defer devNull.Close()
		wrap.ipfsDaemonCmd.Stdout = devNull
		wrap.ipfsDaemonCmd.Stderr = devNull

		// // Redirect stdout and stderr to files or `/dev/null` to detach from terminal
		// wrap.ipfsDaemonCmd.Stdout = os.Stdout // or you can redirect to a file with os.Create("/path/to/output.log")
		// wrap.ipfsDaemonCmd.Stderr = os.Stderr // or os.Create("/path/to/error.log")
	}

	// Start the command
	if err := wrap.ipfsDaemonCmd.Start(); err != nil {
		wrap.logger.Error("error starting command", slog.Any("error", err))
		return fmt.Errorf("Error starting command: %v\n", err)
	}

	wrap.isDaemonRunning = true

	// Set an artificial delay to give time for the `ipfs` binary to load up.
	// Another perspective is this is the `warmup time`.
	time.Sleep(wrap.daemonInitialWarmupDuration)
	wrap.logger.Debug("ipfs daemon is running and waiting for api call from your app")
	return nil
}

// ForceShutdownDaemon function will send KILL signal to the operating system
// for the `ipfs` running daemon in background to force that binary to shutdown.
func (wrap *ipfsCliWrapper) ForceShutdownDaemon() error {
	if wrap.isDaemonRunningContinously {
		wrap.isDaemonRunning = false

		// This code is special because we need to lookup the `ipfs` running
		// process in the operating system and send a `SIGTERM` signal via
		// the operating system to cause that app to shutdown.
		return oskit.TerminateProgram("ipfs")
	}
	return wrap.ShutdownDaemon()
}

func (wrap *ipfsCliWrapper) ShutdownDaemon() error {
	if wrap.isDaemonRunningContinously {
		wrap.logger.Debug("Ignoring daemon shutdown as wrapper is running in continous operation mode")
		return nil
	}
	wrap.isDaemonRunning = false

	// Send the process kill signal to our running application in the shell and
	// return any errors if anything fails in this operation.
	if killErr := wrap.ipfsDaemonCmd.Process.Kill(); killErr != nil {
		wrap.logger.Error("error killing process", slog.Any("error", killErr))
		return fmt.Errorf("Error killing process: %v\n", killErr)
	}

	// Wait for the command to exit.
	if waitErr := wrap.ipfsDaemonCmd.Wait(); waitErr != nil {
		if exitError, ok := waitErr.(*exec.ExitError); ok && exitError.ProcessState.ExitCode() == -1 {
			// This is the expected behavior, the command was killed.
			// log.Println("Process was killed as expected.")
			wrap.logger.Debug("ipfs daemon has exited")
		} else {
			// Handle other errors.
			wrap.logger.Error("command exited with error", slog.Any("error", waitErr))
			return fmt.Errorf("Command exited with error: %v\n", waitErr)
		}
	}
	return nil
}

// downloadAndUnzip function will download the `ipfs` binary based on your
// machine operating system and CPU architecture; afterwords, unzip the binary
// and have it ready for execution.
func downloadAndUnzip(logger *slog.Logger, osName, archName string) error {
	logger.Debug("ipfs binary does not exist, need to fetch now...")

	binaryDirName := "bin"
	zippedBinaryFilePath := "./bin/ipfs.tar.gz"
	unzippedDirPath := "./bin/kubo"

	// Download the file if it wasn't downloaded before.
	if _, err := os.Stat(zippedBinaryFilePath); err != nil {
		// Lookup the binary to download based on what OS and architecture you are
		// using so the correct binary gets downloaded that will work on your
		// machine.
		url, err := getDownloadURL(osName, archName)
		if err != nil {
			logger.Error("failed finding download link",
				slog.Any("error", err),
				slog.String("os", osName),
				slog.String("arch", archName))
			return fmt.Errorf("failed finding download link: %v", err)
		}

		logger.Debug("fetching zip file",
			slog.String("os", osName),
			slog.String("arch", archName),
			slog.String("url", url))

		if downloadErr := urlkit.DownloadFile(url, zippedBinaryFilePath); downloadErr != nil {
			logger.Error("failed downloading the binary",
				slog.Any("error", err),
				slog.String("url", url),
				slog.String("os", osName),
				slog.String("arch", archName))
			return fmt.Errorf("failed downloading the binary: %v", downloadErr)
		}
	}

	logger.Debug("ipfs binary unzipping...")

	if err := oskit.CreateDirIfDoesNotExist(unzippedDirPath); err != nil {
		logger.Error("failed to make directory",
			slog.Any("error", err),
			slog.String("os", osName),
			slog.String("arch", archName))
		log.Fatalf("failed to make directory: %v", err)
	}
	if err := oskit.CreateDirIfDoesNotExist(IPFSDataDirPath); err != nil {
		logger.Error("failed to make directory",
			slog.Any("error", err),
			slog.String("os", osName),
			slog.String("arch", archName))
		log.Fatalf("failed to make directory: %v", err)
	}

	// Developers Note:
	// Permission value of `777` is a permission in Unix based system with full
	// read/write/execute permission to owner, group and everyone.

	// Special thanks to: https://github.com/golift/xtractr?tab=readme-ov-file
	x := &xtractr.XFile{
		FilePath:  zippedBinaryFilePath,
		OutputDir: binaryDirName,
		FileMode:  os.FileMode(int(0777)), // Note: https://stackoverflow.com/a/28969523
		DirMode:   os.FileMode(int(0777)),
	}

	// size is how many bytes were written.
	// files may be nil, but will contain any files written (even with an error).
	size, files, err := xtractr.ExtractTarGzip(x)
	if err != nil || files == nil {
		logger.Error("failed extracting tar gzip",
			slog.Int64("bytes written", size),
			slog.Any("files extracted", files),
			slog.Any("error", err),
			slog.String("os", osName),
			slog.String("arch", archName))
		log.Fatal(size, files, err)
	}

	logger.Debug("ipfs binary unzipped: Bytes written:",
		slog.Int64("bytes written", size),
		slog.String("files extracted", strings.Join(files, "\n -")),
	)

	// Remove our compressed file we downloaded in this function previously above.
	if err := os.Remove(zippedBinaryFilePath); err != nil {
		logger.Error("failed deleting zip",
			slog.String("path", zippedBinaryFilePath),
			slog.Any("error", err),
			slog.String("os", osName),
			slog.String("arch", archName))
		return fmt.Errorf("failed deleting zip: %v", err)
	}

	// Set the permission of the file to be readable. Do this in case the above
	// `ExtractTarGzip` library failed in any of the different operating system.
	// This code is essentially a `just-in-case` sort of thing to run.
	os.Chmod(IPFSBinaryFilePath, 0777)

	return nil
}

func (wrap *ipfsCliWrapper) AddFile(ctx context.Context, filepath string) (string, error) {
	// Prepare the command to add the file using the IPFS binary and utilize
	// the latest cid implementation.
	cmd := exec.CommandContext(ctx, IPFSBinaryFilePath, "add", filepath, "--cid-version=1")

	// Capture the output of the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		wrap.logger.Error("error adding file to ipfs",
			slog.String("filepath", filepath),
			slog.Any("error", err),
			slog.String("output", string(output)))
		return "", fmt.Errorf("failed to add file to ipfs: %v, output: %s", err, string(output))
	}

	// ALGORITHM

	parts := strings.Fields(string(output))

	// Uncomment for debugging purposes only to see what's going on.
	// wrap.logger.Debug("command executed",
	// 	slog.String("filepath", filepath),
	// 	slog.Int("parts len", len(parts)),
	// 	slog.Any("parts", parts))

	var filename string
	var cid string
	var foundAddedText bool = false

	for _, part := range parts {
		// wrap.logger.Debug(part) // Uncomment for debugging purposes only to see what's going on.
		if cid != "" {
			filename = part
			break
		}
		if foundAddedText {
			cid = part
			continue
		}
		if strings.Contains(part, "added") {
			foundAddedText = true
			continue
		}
	}

	wrap.logger.Debug("file added to ipfs successfully",
		slog.String("filepath", filepath),
		slog.String("filename", filename),
		slog.String("cid", cid))

	return cid, nil
}

func (wrap *ipfsCliWrapper) AddFileContent(ctx context.Context, fileContent []byte) (string, error) {
	if fileContent == nil {
		return "", fmt.Errorf("cannot have missing: %v", "fileContent")
	}

	// Save in the current directory this application is running; however,
	// generate a random filename to be used to store the content locally and
	// then we will delete.
	filepath := fmt.Sprintf("./ipfscliwrapper_tempfile_%v", randomkit.String(5))

	// open output file
	fo, err := os.Create(filepath)
	if err != nil {
		wrap.logger.Error("failed creating file in local filesystem",
			slog.Any("error", err))
		return "", err
	}

	if _, err := fo.Write(fileContent); err != nil {
		wrap.logger.Error("failed writing file to local filesystem",
			slog.Any("error", err))
		return "", err
	}

	// close fo on exit and check for its returned error
	if err := fo.Close(); err != nil {
		wrap.logger.Error("failed closing file in local filesystem",
			slog.Any("error", err))
		return "", err
	}

	// Delete our tempfile after we finished submitting
	defer func() {
		if rmErr := os.Remove(filepath); rmErr != nil {
			wrap.logger.Error("failed removing from local filesystem",
				slog.Any("error", err))
			return
		}
	}()

	cid, err := wrap.AddFile(ctx, filepath)
	if err != nil {
		wrap.logger.Error("failed adding file to ipfs",
			slog.Any("error", err))
		return "", err
	}

	return cid, err
}

func (wrap *ipfsCliWrapper) GetFile(ctx context.Context, cid string) error {
	// Prepare the command to get the file using the IPFS binary
	cmd := exec.CommandContext(ctx, IPFSBinaryFilePath, "get", cid)

	// Capture the output of the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		wrap.logger.Error("error getting file from ipfs",
			slog.String("cid", cid),
			slog.Any("error", err),
			slog.String("output", string(output)))
		return fmt.Errorf("failed to get file from ipfs: %v, output: %s", err, string(output))
	}

	return nil
}

func (wrap *ipfsCliWrapper) Cat(ctx context.Context, cid string) ([]byte, error) {
	// Prepare the command to retrieve the file contents using the IPFS binary
	cmd := exec.CommandContext(ctx, IPFSBinaryFilePath, "cat", cid)

	// Capture the output of the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		wrap.logger.Error("error catting file from ipfs",
			slog.String("cid", cid),
			slog.Any("error", err),
			slog.String("output", string(output)))
		return []byte{}, fmt.Errorf("failed to cat file from ipfs: %v, output: %s", err, string(output))
	}

	// Log successful retrieval of the file contents
	wrap.logger.Debug("file content retrieved from ipfs successfully",
		slog.String("cid", cid),
		slog.String("output", string(output)))

	// Return the file content as a string
	return output, nil
}

func (wrap *ipfsCliWrapper) ListPins(ctx context.Context) ([]string, error) {
	return wrap.ListPinsByType(ctx, "all")
}

func (wrap *ipfsCliWrapper) ListPinsByType(ctx context.Context, typeID string) ([]string, error) {
	// Prepare the command to list all local pins using the IPFS binary
	//
	// Notes:
	// (1)
	// `--type=all` <-- Filter to apply on what sort of cid's to return.
	// There are three types of pins in the ipfs world:
	// * "direct": pin that specific object.
	// * "recursive": pin that specific object, and indirectly pin all its descendants
	// * "indirect": pinned indirectly by an ancestor (like a refcount)
	// * "all"
	//
	// (2)
	// `--stream=true` <-- if you get such an error because of large list, you can make use of the streaming option
	// https://stackoverflow.com/questions/60926526/how-can-one-list-all-of-the-currently-pinned-files-for-an-ipfs-instance

	cmd := exec.CommandContext(ctx, IPFSBinaryFilePath, "pin", "ls", "--type="+typeID, "--stream=true")

	// Capture the output of the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		wrap.logger.Error("error pinning file content on ipfs",
			slog.Any("error", err),
			slog.String("output", string(output)))
		return nil, fmt.Errorf("failed to pin file content on ipfs: %v, output: %s", err, string(output))
	}

	parts := strings.Fields(string(output))

	// // Uncomment for debugging purposes only to see what's going on.
	// wrap.logger.Debug("command executed",
	// 	slog.Int("parts len", len(parts)),
	// 	slog.Any("parts", parts))

	cids := make([]string, 0)
	ignorePartArr := []string{"recursive", "indirect", "direct"}
	var ignoreFound bool = false

	for _, part := range parts {
		// wrap.logger.Debug(part) // Uncomment for debugging purposes only to see what's going on.

		for _, ignorePart := range ignorePartArr {
			if part == ignorePart {
				ignoreFound = true
				continue // Skip to the next root.
			}
		}

		// Record our content ID if it's not a reserved word.
		if !ignoreFound {
			cids = append(cids, part)
		}

		ignoreFound = false // Reset the checker since it's the end of the loop.
	}

	return cids, nil
}

func (wrap *ipfsCliWrapper) Pin(ctx context.Context, cid string) error {
	// Prepare the command to pin the file contents using the IPFS binary
	cmd := exec.CommandContext(ctx, IPFSBinaryFilePath, "pin", "add", cid)

	// Capture the output of the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		wrap.logger.Error("error pinning file content on ipfs",
			slog.String("cid", cid),
			slog.Any("error", err),
			slog.String("output", string(output)))
		return fmt.Errorf("failed to pin file content on ipfs: %v, output: %s", err, string(output))
	}
	return nil
}

func (wrap *ipfsCliWrapper) Unpin(ctx context.Context, cid string) error {
	// Prepare the command to remove the pin using the IPFS binary
	cmd := exec.CommandContext(ctx, IPFSBinaryFilePath, "pin", "rm", cid)

	// Capture the output of the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		wrap.logger.Error("error removing pinning from ipfs",
			slog.String("cid", cid),
			slog.Any("error", err),
			slog.String("output", string(output)))
		return fmt.Errorf("failed to remove pin from ipfs: %v, output: %s", err, string(output))
	}

	return nil
}

func (wrap *ipfsCliWrapper) GarbageCollection(ctx context.Context) error {
	// Prepare the command run garbage collection for the `ipfs` binary.
	cmd := exec.CommandContext(context.Background(), IPFSBinaryFilePath, "repo", "gc")

	// Capture the output of the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		wrap.logger.Error("error garbage collecting in ipfs",
			slog.Any("error", err),
			slog.String("output", string(output)))
		return fmt.Errorf("failed to run garbage collection pin from ipfs: %v, output: %s", err, string(output))
	}

	return nil
}

func (wrap *ipfsCliWrapper) Id(ctx context.Context) (*IpfsNodeInfo, error) {
	// Special thanks:
	// https://github.com/ipfs-shipyard/ipfs-primer/blob/12d7298f436fa83e8395ade6969d2a4df298b334/going-online/lessons/connect-your-node.md

	// Prepare the command run garbage collection for the `ipfs` binary.
	cmd := exec.CommandContext(context.Background(), IPFSBinaryFilePath, "id")

	// Capture the output of the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		wrap.logger.Error("error getting ipfs id",
			slog.Any("error", err),
			slog.String("output", string(output)))
		return nil, fmt.Errorf("failed to run `id` in ipfs: %v, output: %s", err, string(output))
	}

	// Create an instance of IPFSInfo.
	var info IpfsNodeInfo

	// Parse the JSON string into the struct.
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	return &info, nil
}
