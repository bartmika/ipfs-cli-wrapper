package oskit_test

import (
	"errors"
	"testing"
)

// MockOSOperator is a mock implementation of the OSOperater interface for testing.
type MockOSOperator struct {
	CreateDirFunc        func(string) error
	CreateDirsFunc       func([]string) error
	TerminateProgramFunc func(string) error
	MoveFileFunc         func(string, string) error
	IsProgramRunningFunc func(string) (bool, error)
}

func (m *MockOSOperator) CreateDirIfDoesNotExist(dirPath string) error {
	return m.CreateDirFunc(dirPath)
}

func (m *MockOSOperator) CreateDirsIfDoesNotExist(dirs []string) error {
	return m.CreateDirsFunc(dirs)
}

func (m *MockOSOperator) TerminateProgram(program string) error {
	return m.TerminateProgramFunc(program)
}

func (m *MockOSOperator) MoveFile(sourcePath, destPath string) error {
	return m.MoveFileFunc(sourcePath, destPath)
}

func (m *MockOSOperator) IsProgramRunning(programName string) (bool, error) {
	return m.IsProgramRunningFunc(programName)
}

// Test for CreateDirIfDoesNotExist
func TestCreateDirIfDoesNotExist(t *testing.T) {
	mock := &MockOSOperator{
		CreateDirFunc: func(dirPath string) error {
			if dirPath == "/valid/path" {
				return nil
			}
			return errors.New("invalid path")
		},
	}

	err := mock.CreateDirIfDoesNotExist("/valid/path")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	err = mock.CreateDirIfDoesNotExist("/invalid/path")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

// Test for CreateDirsIfDoesNotExist
func TestCreateDirsIfDoesNotExist(t *testing.T) {
	mock := &MockOSOperator{
		CreateDirsFunc: func(dirs []string) error {
			if len(dirs) == 0 {
				return errors.New("no directories specified")
			}
			for _, dir := range dirs {
				if dir == "" {
					return errors.New("empty directory path")
				}
			}
			return nil
		},
	}

	err := mock.CreateDirsIfDoesNotExist([]string{"/valid/path1", "/valid/path2"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	err = mock.CreateDirsIfDoesNotExist([]string{})
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

// Test for TerminateProgram
func TestTerminateProgram(t *testing.T) {
	mock := &MockOSOperator{
		TerminateProgramFunc: func(program string) error {
			if program == "existing_program" {
				return nil
			}
			return errors.New("program not found")
		},
	}

	err := mock.TerminateProgram("existing_program")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	err = mock.TerminateProgram("nonexistent_program")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

// Test for MoveFile
func TestMoveFile(t *testing.T) {
	mock := &MockOSOperator{
		MoveFileFunc: func(sourcePath, destPath string) error {
			if sourcePath == "/valid/source" && destPath == "/valid/dest" {
				return nil
			}
			return errors.New("invalid source or destination path")
		},
	}

	err := mock.MoveFile("/valid/source", "/valid/dest")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	err = mock.MoveFile("/invalid/source", "/invalid/dest")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

// Test for IsProgramRunning
func TestIsProgramRunning(t *testing.T) {
	mock := &MockOSOperator{
		IsProgramRunningFunc: func(programName string) (bool, error) {
			if programName == "running_program" {
				return true, nil
			}
			return false, errors.New("program not running")
		},
	}

	running, err := mock.IsProgramRunning("running_program")
	if err != nil || !running {
		t.Errorf("expected program to be running, got %v, %v", running, err)
	}

	running, err = mock.IsProgramRunning("not_running_program")
	if err == nil || running {
		t.Errorf("expected program to not be running, got %v, %v", running, err)
	}
}
