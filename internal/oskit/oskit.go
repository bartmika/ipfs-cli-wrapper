// Package oskit provides number of useful convinence functions related to
// OS operations like `mkdir`, `mv`, etc.
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

func CreateDirIfDoesNotExist(dirPath string) error {
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to make directory: %v", err)
	}
	return nil
}

func CreateDirsIfDoesNotExist(dirPaths []string) error {
	for _, dirPath := range dirPaths {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to make directory: %v", err)
		}
	}

	return nil
}

func MoveFile(sourcePath string, destPath string) error {
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

// IsProgramRunning function checks to see if the program name exist in the
// operating system running tasks and returns true or false if the program is
// running.
func IsProgramRunning(programName string) (bool, error) {
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

func TerminateProgram(processName string) error {
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
