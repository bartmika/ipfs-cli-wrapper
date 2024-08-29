package ipfscliwrapper

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

func (wrap *IpfsCliWrapper) AddFile(ctx context.Context, filepath string) (string, error) {
	// Prepare the command to add the file using the IPFS binary
	cmd := exec.CommandContext(ctx, IPFSBinaryFilePath, "add", filepath)

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

func (wrap *IpfsCliWrapper) AddFileContent(ctx context.Context, filename string, fileContent []byte) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("cannot have missing: %v", "filename")
	}
	if fileContent == nil {
		return "", fmt.Errorf("cannot have missing: %v", "fileContent")
	}

	// Save in the current directory this application is running.
	filepath := "./" + filename

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

	cid, err := wrap.AddFile(ctx, filename)
	if err != nil {
		wrap.logger.Error("failed adding file to ipfs",
			slog.Any("error", err))
		return "", err
	}

	if rmErr := os.Remove(filepath); rmErr != nil {
		wrap.logger.Error("failed removing from local filesystem",
			slog.Any("error", err))
		return "", err
	}

	return cid, err
}

func (wrap *IpfsCliWrapper) GetFile(ctx context.Context, cid string) error {
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

func (wrap *IpfsCliWrapper) Cat(ctx context.Context, cid string) ([]byte, error) {
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

func (wrap *IpfsCliWrapper) ListPins(ctx context.Context) ([]string, error) {
	return wrap.ListPinsByType(ctx, "all")
}

func (wrap *IpfsCliWrapper) ListPinsByType(ctx context.Context, typeID string) ([]string, error) {
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

func (wrap *IpfsCliWrapper) Pin(ctx context.Context, cid string) error {
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

func (wrap *IpfsCliWrapper) Unpin(ctx context.Context, cid string) error {
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

func (wrap *IpfsCliWrapper) GarbageCollection(ctx context.Context) error {
	// Prepare the command run garbage collection for the `ipfs` binary.
	cmd := exec.CommandContext(context.Background(), IPFSBinaryFilePath, "repo", "gc")

	// Capture the output of the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		wrap.logger.Error("error removing pinning from ipfs",
			slog.Any("error", err),
			slog.String("output", string(output)))
		return fmt.Errorf("failed to run garbage collection pin from ipfs: %v, output: %s", err, string(output))
	}

	return nil
}
