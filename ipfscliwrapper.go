// IPFS-DaemonLauncher is a package that manages running an IPFS node in the background
// and offers a user-friendly interface, enabling you to build IPFS-embedded
// Golang applications more efficiently.
package ipfscliwrapper

import (
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
	"github.com/bartmika/ipfs-cli-wrapper/internal/urlkit"
)

// IpfsCliWrapper represents the a wrapper over a `ipfs` executable binary in
// the operating system that we use to control the operation of.
type IpfsCliWrapper struct {
	logger *slog.Logger
	// ipfsDaemonCmd variable is the pointer to the command shell instance of
	// the `ipfs` binary running as daemon mode in the current operating system
	// as background process.
	ipfsDaemonCmd *exec.Cmd

	// stdout variable is the output stream from the console of the running
	// `ipfs` binary running in daemon mode.
	stdout io.ReadCloser

	// isDaemonRunning variable controls the state in our wrapper if we see the
	// `ipfs` binary running in daemon mode.
	isDaemonRunning bool

	// isDaemonRunningContinously variable controls our wrapper to never shutdown our
	// `ipfs` binary running in daemon mode unless you use the `ForceShutdown()`
	// function.
	isDaemonRunningContinously bool

	// daemonInitialWarmupDuration variable used to artificially delay the
	// `StartDaemonInBackground` function before releasing execution flow to
	// the program.
	// Set an artificial delay to give time for the `ipfs` binary to load up.
	// This is dependent on your machine.
	daemonInitialWarmupDuration time.Duration

	// os variable used to track the operating system our wrapper is running on.
	os string

	// arch variable used to track the CPU chip architecture that our wrapper.
	// is running on.
	arch string
}

// Option is a functional option type that allows us to configure the IpfsCliWrapper.
type Option func(*IpfsCliWrapper)

func NewDaemonLauncher(options ...Option) (*IpfsCliWrapper, error) {
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

	wrapper := &IpfsCliWrapper{
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
	arg1 := "--enable-gc" // Enable garbage collection
	daemonCmd := exec.Command(app, arg0)

	// Set the environment variable before executing the command
	daemonCmd.Env = append(os.Environ(), "IPFS_PATH="+IPFSDataDirPath, arg1)

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

func (wrap *IpfsCliWrapper) StartDaemonInBackground() error {
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
func (wrap *IpfsCliWrapper) ForceShutdownDaemon() error {
	if wrap.isDaemonRunningContinously {
		wrap.isDaemonRunning = false

		// This code is special because we need to lookup the `ipfs` running
		// process in the operating system and send a `SIGTERM` signal via
		// the operating system to cause that app to shutdown.
		return oskit.TerminateProgram("ipfs")
	}
	return wrap.ShutdownDaemon()
}

func (wrap *IpfsCliWrapper) ShutdownDaemon() error {
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
