// Package oskit provides a set of convenience functions related to operating
// system operations such as creating directories, moving files, checking if
// a program is running, and terminating processes.
package oskit

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// OSOperater defines methods related to OS operations.
type OSOperater interface {
	// CreateDirIfDoesNotExist creates a directory at the specified path if it does not already exist.
	// It sets the permissions to the default permissions for the operating system.
	//
	// Parameters:
	// - dirPath (string): The path of the directory to create.
	//
	// Returns:
	// - error: Returns an error if the directory cannot be created.
	//
	// Example:
	//
	//	err := CreateDirIfDoesNotExist("/path/to/dir")
	//	if err != nil {
	//	    log.Fatal(err)
	//	}
	CreateDirIfDoesNotExist(dirPath string) error

	// CreateDirsIfDoesNotExist creates multiple directories based on the provided slice of paths.
	// It sets the permissions for each directory to the default permissions for the operating system.
	//
	// Parameters:
	// - dirPaths ([]string): A slice of directory paths to create.
	//
	// Returns:
	// - error: Returns an error if any directory cannot be created.
	//
	// Example:
	//
	//	err := CreateDirsIfDoesNotExist([]string{"/path/to/dir1", "/path/to/dir2"})
	//	if err != nil {
	//	    log.Fatal(err)
	//	}
	CreateDirsIfDoesNotExist(dirs []string) error

	// TerminateProgram attempts to terminate all processes matching the given
	// program name by sending a SIGTERM signal. It uses the `pgrep` command to
	// find the process IDs of the running instances and iterates over each to
	// send the termination signal.
	//
	// Parameters:
	// - processName (string): The name of the process to terminate.
	//
	// Returns:
	// - error: Returns an error if the process cannot be found or terminated.
	//
	// Example:
	//
	//	err := TerminateProgram("ipfs")
	//	if err != nil {
	//	    log.Fatal(err)
	//	}
	TerminateProgram(program string) error

	// MoveFile moves a file from the source path to the destination path.
	// This function handles copying the file content, closing the file descriptors,
	// and removing the source file after successful copying.
	//
	// Parameters:
	// - sourcePath (string): The path to the source file.
	// - destPath (string): The path where the file should be moved.
	//
	// Returns:
	// - error: Returns an error if the file cannot be moved.
	//
	// Example:
	//
	//	err := MoveFile("/path/to/source.txt", "/path/to/destination.txt")
	//	if err != nil {
	//	    log.Fatal(err)
	//	}
	MoveFile(sourcePath string, destPath string) error

	// IsProgramRunning checks if a program with the given name is currently running
	// in the operating system. It uses the `pgrep` command to search for processes
	// matching the exact program name.
	//
	// Parameters:
	// - programName (string): The name of the program to check.
	//
	// Returns:
	// - bool: Returns true if the program is running, false otherwise.
	// - error: Returns an error if the check fails due to issues with executing the `pgrep` command.
	//
	// Example:
	//
	//	running, err := IsProgramRunning("ipfs")
	//	if err != nil {
	//	    log.Fatal(err)
	//	}
	//	if running {
	//	    fmt.Println("IPFS is running.")
	//	}
	IsProgramRunning(programName string) (bool, error)
}

// DefaultOSKit is the default implementation of OSOperater.
type DefaultOSKit struct{}

func (d *DefaultOSKit) CreateDirIfDoesNotExist(dirPath string) error {
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to make directory: %v", err)
	}
	return nil
}

func (d *DefaultOSKit) CreateDirsIfDoesNotExist(dirs []string) error {
	for _, dirPath := range dirs {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to make directory: %v", err)
		}
	}

	return nil
}

func (d *DefaultOSKit) TerminateProgram(processName string) error {
	// DEVELOPERS NOTE:
	// (1)
	// `pgrep` is a unix app used to lookup programs running in background and
	// it returns the process id value of the running instance.
	//
	// (2)
	// To ensure that code targets only processes with the exact name "ipfs" and
	// not those that include "ipfs" as a substring (e.g.,
	// "comicbookss_ipfs_backend"), you can refine the pgrep command by using
	// the -x flag, which matches the exact process name.

	// Use `pgrep` to get the PIDs of the processes with the given name
	cmd := exec.Command("pgrep", "-x", processName)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to find process: %v\n", err)
	}

	// Split the output to get individual PIDs
	pids := strings.Fields(out.String())

	// Iterate over each PID and terminate the process
	for _, pidStr := range pids {
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			return fmt.Errorf("Failed to parse PID: %v\n", err)
		}

		// Find the process by PID
		process, err := os.FindProcess(pid)
		if err != nil {
			return fmt.Errorf("Failed to find process with PID %d: %v\n", pid, err)
		}

		// Developers Note
		// SIGTERM (syscall.SIGTERM): This is a gentle request for the process to terminate. The process can handle this signal and clean up resources before exiting.
		// SIGKILL (syscall.SIGKILL): This forces the process to terminate immediately, and the process doesn’t get a chance to clean up.

		// Send a SIGTERM signal to the process (soft kill)
		if err := process.Signal(syscall.SIGTERM); err != nil {
			fmt.Printf("Failed to terminate process with PID %d: %v\n", pid, err)
			continue
		}

		// // Send a SIGKILL signal to the process (force kill)
		// if err := process.Signal(syscall.SIGKILL); err != nil {
		// 	fmt.Printf("Failed to kill process: %v\n", err)
		// 	return
		// }

		fmt.Printf("Process with PID %d terminated successfully.\n", pid)
	}
	return nil
}

func (d *DefaultOSKit) MoveFile(sourcePath string, destPath string) error {
	// DEVELOPERS NOTE:
	// Code was copied from: https://stackoverflow.com/a/50744122

	src, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer src.Close()
	fi, err := src.Stat()
	if err != nil {
		return err
	}
	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	perm := fi.Mode() & os.ModePerm
	dst, err := os.OpenFile(destPath, flag, perm)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil {
		dst.Close()
		os.Remove(destPath)
		return err
	}
	err = dst.Close()
	if err != nil {
		return err
	}
	err = src.Close()
	if err != nil {
		return err
	}
	err = os.Remove(sourcePath)
	if err != nil {
		return err
	}
	return nil
}

func (d *DefaultOSKit) IsProgramRunning(programName string) (bool, error) {
	// DEVELOPERS NOTE:
	// (1)
	// `pgrep` is a unix app used to lookup programs running in background and
	// it returns the process id value of the running instance.
	//
	// (2)
	// To ensure that code targets only processes with the exact name "ipfs" and
	// not those that include "ipfs" as a substring (e.g.,
	// "comicbookss_ipfs_backend"), you can refine the pgrep command by using
	// the -x flag, which matches the exact process name.

	// Execute the `pgrep` command to find processes by name
	cmd := exec.Command("pgrep", "-x", programName)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		// If `pgrep` exits with a status 1, it means no processes were found
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return false, nil
		}
		return false, err
	}

	// If the output from `pgrep` is not empty, the process is running
	return strings.TrimSpace(out.String()) != "", nil
}
