// Package filesys provides a collection of utility functions for common file system operations.
// It includes functions for creating, deleting, copying, reading, and searching files and directories,
// as well as checking file existence and managing the current working directory.
package filesys

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrIsNotDir = errors.New("path isn't a directory")
)

// CreateDir creates a directory at the specified path with the given permissions.
//
// If the directory already exists:
//   - If 'force' is true, it proceeds without error.
//   - If 'force' is false, it returns an error.
//
// It also returns an error if the existing path is a file (not a directory).
func CreateDir(dirPath string, permission os.FileMode, force bool) error {
	// Get file information for the given path.
	stat, err := os.Stat(dirPath)
	// If 'force' is false and the path exists
	// return the error (indicating the directory already exists).
	if !force && !os.IsNotExist(err) {
		return err
	}

	// If the path exists and it's not a directory, return an error.
	if stat != nil && !stat.IsDir() {
		return ErrIsNotDir
	}

	// Create all necessary parent directories if they don't exist, with the specified permissions.
	if err := os.MkdirAll(dirPath, permission); err != nil {
		return err
	}

	// Change the permissions of the newly created directory to 0755 (rwxr-xr-x).
	return os.Chmod(dirPath, 0755)
}

// DeleteDir deletes a directory and all its contents recursively.
// It returns any error encountered during the removal.
func DeleteDir(path string) error {
	return os.RemoveAll(path)
}

// CopyDir copies the entire contents of a source directory to a destination directory.
// It preserves the file modes of the source directory and files.
// It returns an error if the source is not a directory or if any other I/O operation fails.
func CopyDir(src, dest string) error {
	// Get file information for the source path.
	srcStat, err := os.Stat(src)
	if err != nil {
		return err
	}
	// If the source is not a directory, return an error.
	if !srcStat.IsDir() {
		return ErrIsNotDir
	}

	// Create the destination directory with the same permissions as the source directory.
	if err := os.MkdirAll(dest, srcStat.Mode()); err != nil {
		return err
	}

	// Walk through the source directory recursively.
	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		// If an error occurred during walking, return it.
		if err != nil {
			return err
		}

		// If the current item is not a regular file (e.g., a directory, symlink), skip it.
		if !info.Mode().IsRegular() {
			return nil
		}

		// Construct the destination path for the current file.
		// `path[len(src)+1:]` gets the relative path from the source directory.
		destPath := filepath.Join(dest, path[len(src)+1:])
		// Create any necessary parent directories for the destination file with default permissions.
		if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
			return err
		}

		// Open the source file for reading.
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close() // Ensure the source file is closed.

		// Create the destination file for writing.
		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer destFile.Close() // Ensure the destination file is closed.

		// Copy the contents from the source file to the destination file.
		if _, err := io.Copy(destFile, srcFile); err != nil {
			return err
		}

		return nil
	})
	// If an error occurred during the walk, return it.
	if err != nil {
		return err
	}

	return nil
}

// ReadDir reads the directory specified by `dirName` and returns a list of matching file paths.
// It uses `filepath.Glob` which means `dirName` can contain glob patterns (e.g., "mydir/*.txt").
func ReadDir(dirName string) ([]string, error) {
	files, err := filepath.Glob(dirName)
	return files, err
}

// CreateFile creates a new file at the specified `filePath`.
//
// If the file already exists:
//   - If 'force' is true, it overwrites the existing file.
//   - If 'force' is false, it returns an error.
func CreateFile(filePath string, force bool) (*os.File, error) {
	// Check if the file exists.
	_, err := os.Stat(filePath)
	// If 'force' is false and the file exists, return an error.
	if !force && os.IsExist(err) {
		return nil, fmt.Errorf("error in getting file stat %s because of %v", filePath, err)
	}
	// Create the file. If it exists and 'force' is true, it will be truncated.
	return os.Create(filePath)
}

// WriteFile writes the provided `contents` to the file at `filePath` with the given `permission`.
// If the file does not exist, it will be created. If it exists, it will be truncated.
func WriteFile(filePath string, permission os.FileMode, contents []byte) error {
	return os.WriteFile(filePath, contents, permission)
}

// DeleteFile deletes the file at the specified `filePath`.
// It returns an error if the file cannot be removed.
func DeleteFile(filePath string) error {
	return os.Remove(filePath)
}

// CopyFile copies a single file from `sourcePath` to `destPath`.
// It reads the entire content of the source file into memory and then writes it to the destination.
// The destination file will have default permissions (0644).
func CopyFile(sourcePath, destPath string) error {
	// Read the entire content of the source file.
	input, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	// Write the content to the destination file with permissions 0644 (rw-r--r--).
	return os.WriteFile(destPath, input, 0644)
}

// ReadFile reads the entire content of the file at `filePath` into a byte slice.
// It returns the file content and any error encountered.
func ReadFile(filePath string) ([]byte, error) {
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return contents, err
}

// SearchFiles searches for files with a specific `searchFile` name within `sourceDir`.
// It excludes directories listed in `excludeDirs` from the search.
// It returns a slice of full paths to the found files.
func SearchFiles(sourceDir string, excludeDirs []string, searchFile string) ([]string, error) {
	files := make([]string, 0) // Initialize an empty slice to store found file paths.

	// Walk the directory tree rooted at `sourceDir`.
	if err := filepath.WalkDir(sourceDir, fs.WalkDirFunc(func(path string, ds fs.DirEntry, err error) error {
		// If an error occurred during walking, return it.
		if err != nil {
			return err
		}

		// Check if the current entry is a regular file, not within an excluded directory,
		// and its base name matches `searchFile`.
		if !ds.IsDir() && !isAncestor(excludeDirs, path) && filepath.Base(path) == searchFile {
			files = append(files, path) // Add the file path to the results.
		}
		return nil
	})); err != nil {
		return nil, err
	}

	return files, nil
}

// SearchFileExtensions searches for files with a specific `extension` within `sourceDir`.
// It excludes directories listed in `excludeDirs` from the search.
// It returns a slice of full paths to the found files.
func SearchFileExtensions(sourceDir string, excludeDirs []string, extension string) ([]string, error) {
	files := make([]string, 0) // Initialize an empty slice to store found file paths.

	// Walk the directory tree rooted at `sourceDir`.
	if err := filepath.WalkDir(sourceDir, fs.WalkDirFunc(func(path string, ds fs.DirEntry, err error) error {
		// If an error occurred during walking, return it.
		if err != nil {
			return err
		}

		// Check if the current entry is a regular file, not within an excluded directory,
		// and its extension matches the `extension`.
		if !ds.IsDir() && !isAncestor(excludeDirs, path) && filepath.Ext(path) == extension {
			files = append(files, path) // Add the file path to the results.
		}
		return nil
	})); err != nil {
		return nil, err
	}

	return files, nil
}

// Pwd returns the present working directory (current directory).
func Pwd() (string, error) {
	return os.Getwd()
}

// Exists checks if a file or directory at the given `file` path exists.
// It returns true if the file/directory exists, false if it does not,
// and an error if there's any other issue checking its status.
func Exists(file string) (bool, error) {
	_, err := os.Stat(file)
	if err == nil {
		return true, nil // Path exists.
	}
	// If the error indicates that the file does not exist, return false.
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// Cd changes the current working directory to `dir`.
// It returns any error encountered during the change.
func Cd(dir string) error {
	return os.Chdir(dir)
}

// isAncestor checks if any of the `excludeDirs` are ancestors (or part of the path) of `path`.
// It returns true if `path` contains any of the `excludeDirs` as a substring, false otherwise.
func isAncestor(excludeDirs []string, path string) bool {
	for _, excludeDir := range excludeDirs {
		if strings.Contains(path, excludeDir) {
			return true
		}
	}
	return false
}
