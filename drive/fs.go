// drive/fs.go
// Filesystem commands and functions for Lily drive objects.

package drive

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/cubeflix/lily/fs"
	"github.com/cubeflix/lily/network"
	"github.com/cubeflix/lily/security/access"
)

var ErrEmptyPath = errors.New("lily.drive: Empty path")
var ErrNotAChildOf = errors.New("lily.drive: Path is not a child of parent")
var ErrAlreadyExists = errors.New("lily.drive: Path already exists")
var ErrInvalidDirectoryTree = errors.New("lily.drive: Invalid directory tree")
var ErrInvalidName = errors.New("lily.drive: Invalid name")
var ErrInvalidLength = errors.New("lily.drive: Invalid length of array")
var ErrInvalidChunks = errors.New("lily.drive: Invalid chunks")
var ErrInvalidStartEnd = errors.New("lily.drive: Invalid start and end values")

var IllegalNames = "\"*/:<>?\\|"

// Path status.
type PathStatus struct {
	Exists bool
	Name   string
	IsFile bool
}

// Get the full host system path for a local path, given a drive.
func (d *Drive) getHostPath(path string) string {
	return filepath.Join(d.path, path)
}

// Mutex optimization consists of grouping the directories by parent
// directory and then holding the parent mutex while adding each new
// subdir. This prevents a common slowdown where the mutex is acquired
// then immediately released, then acquired again for each subsequent
// operation. This function organizes directories into a map of parent
// directories and directories. The function also takes a list of access
// settings and can organize those as well.
func groupPathsByParentDir(dirs []string,
	settings []*access.AccessSettings) (map[string][]string, map[string][]*access.AccessSettings, error) {
	groups := map[string][]string{}
	accessGroups := map[string][]*access.AccessSettings{}
	trackAccess := len(settings) != 0

	// Loop through the directories.
	for i := range dirs {
		// Split the directory.
		split, err := fs.SplitPath(dirs[i])
		if err != nil {
			return map[string][]string{}, map[string][]*access.AccessSettings{}, err
		}

		// If the parent directory already exists within the map, add the
		// directory to the list.
		parent := strings.Join(split[:len(split)-1], "/")
		if _, ok := groups[parent]; ok {
			groups[parent] = append(groups[parent], dirs[i])
			if trackAccess {
				accessGroups[parent] = append(accessGroups[parent], settings[i])
			}
		} else {
			// If not, add the new directory and create the new slice.
			groups[parent] = []string{dirs[i]}
			if trackAccess {
				accessGroups[parent] = []*access.AccessSettings{settings[i]}
			}
		}
	}

	// Return the grouped directories and access objects.
	return groups, accessGroups, nil
}

// Create directories.
func (d *Drive) CreateDirs(dirs []string, settings []*access.AccessSettings, useParentAccessSettings, performMutexOptimization bool) error {
	var err error

	// Check that the lengths of the slices are correct.
	if !useParentAccessSettings {
		if len(dirs) != len(settings) {
			return ErrInvalidLength
		}
	}

	// Clean the directories.
	for i := range dirs {
		dirs[i], err = fs.CleanPath(dirs[i])
		if err != nil {
			return err
		}
	}

	// Perform mutex optimization.
	if performMutexOptimization {
		// Group the directories by parent directory.
		var groups map[string][]string
		var accessGroups map[string][]*access.AccessSettings
		if useParentAccessSettings {
			groups, accessGroups, err = groupPathsByParentDir(dirs, []*access.AccessSettings{})
		} else {
			groups, accessGroups, err = groupPathsByParentDir(dirs, settings)
		}
		if err != nil {
			return err
		}

		// Create the directories in groups.
		for key := range groups {
			// Grab the lock on the parent.
			parent, err := d.GetDirectoryByPath(key)
			if err != nil {
				return err
			}
			parent.AcquireLock()

			// Create the directories.
			directories := map[string]*fs.Directory{}
			for dir := range groups[key] {
				// Check for an empty directory.
				if groups[key][dir] == "" {
					parent.ReleaseLock()
					return ErrEmptyPath
				}

				// Split the directory.
				split, err := fs.SplitPath(groups[key][dir])
				if err != nil {
					parent.ReleaseLock()
					return err
				}

				// Check that the name is valid.
				if strings.ContainsAny(split[len(split)-1], IllegalNames) {
					parent.ReleaseLock()
					return ErrInvalidName
				}

				// Check if the directory already exists.
				_, err = parent.GetSubdirsByName([]string{split[len(split)-1]})
				if err == nil {
					// Already exists.
					parent.ReleaseLock()
					return ErrAlreadyExists
				}
				_, err = parent.GetFilesByName([]string{split[len(split)-1]})
				if err == nil {
					// Already exists.
					parent.ReleaseLock()
					return ErrAlreadyExists
				}

				// Create the directory object.
				var newdir *fs.Directory
				if useParentAccessSettings {
					parentSettings := *parent.Settings
					newdir, err = fs.NewDirectory(split[len(split)-1], false,
						parent, &parentSettings)
				} else {
					newdir, err = fs.NewDirectory(split[len(split)-1], false,
						parent, accessGroups[key][dir])
				}
				if err != nil {
					parent.ReleaseLock()
					return err
				}

				// Add the new directory object to the map.
				directories[split[len(split)-1]] = newdir
			}

			// Add the directories to the parent.
			parent.SetSubdirsByName(directories)

			// Create the new directories in the host filesystem.
			for i := range groups[key] {
				err := os.Mkdir(d.getHostPath(groups[key][i]), os.ModeDir)
				if err != nil {
					parent.ReleaseLock()
					return err
				}
			}

			// Release the lock on the parent.
			parent.ReleaseLock()
		}
	} else {
		// Do not perform mutex optimization.
		for i := range dirs {
			// Check for an empty directory.
			if dirs[i] == "" {
				return ErrEmptyPath
			}

			// Split the directory.
			split, err := fs.SplitPath(dirs[i])
			if err != nil {
				return err
			}
			// Grab the lock on the parent.
			parent, err := d.GetDirectoryByPath(strings.Join(split[:len(split)-1], "/"))
			if err != nil {
				return err
			}
			parent.AcquireLock()

			// Check that the name is valid.
			if strings.ContainsAny(split[len(split)-1], IllegalNames) {
				parent.ReleaseLock()
				return ErrInvalidName
			}

			// Check if the directory already exists.
			_, err = parent.GetSubdirsByName([]string{split[len(split)-1]})
			if err == nil {
				// Already exists.
				parent.ReleaseLock()
				return ErrAlreadyExists
			}
			_, err = parent.GetFilesByName([]string{split[len(split)-1]})
			if err == nil {
				// Already exists.
				parent.ReleaseLock()
				return ErrAlreadyExists
			}

			// Create the directory object.
			var newdir *fs.Directory
			if useParentAccessSettings {
				parentSettings := *parent.Settings
				newdir, err = fs.NewDirectory(split[len(split)-1], false,
					parent, &parentSettings)
			} else {
				newdir, err = fs.NewDirectory(split[len(split)-1], false,
					parent, settings[i])
			}
			if err != nil {
				parent.ReleaseLock()
				return err
			}

			// Set the new directory object in the parent.
			parent.SetSubdirsByName(map[string]*fs.Directory{split[len(split)-1]: newdir})

			// Create the new directory in the host filesystem.
			err = os.Mkdir(d.getHostPath(dirs[i]), os.ModeDir)
			if err != nil {
				parent.ReleaseLock()
				return err
			}

			// Release the lock on the parent.
			parent.ReleaseLock()
		}
	}

	// Return.
	return nil
}

// Create a directory tree (all new directories fall under a newly-created parent dir).
// The list of directories should be in order from the first to add to the last. These
// directories should be local within the parent.
func (d *Drive) CreateDirsTree(parent string, dirs []string, parentSettings *access.AccessSettings,
	settings []*access.AccessSettings, useParentAccessSettings bool) error {
	var err error

	// Clean the directories and check that all the directories fall under the same parent.
	for i := range dirs {
		dirs[i], err = fs.CleanPath(dirs[i])
		if err != nil {
			return err
		}

		if _, err = filepath.Rel(parent, dirs[i]); err != nil {
			return ErrNotAChildOf
		}
	}

	// Check for an empty path.
	cleanParent, err := fs.CleanPath(parent)
	if err != nil {
		return err
	}
	if cleanParent == "" {
		return ErrEmptyPath
	}

	// Get the parent's parent dir.
	split, err := fs.SplitPath(parent)
	if err != nil {
		return err
	}
	parentParent, err := d.GetDirectoryByPath(strings.Join(split[:len(split)-1], "/"))
	if err != nil {
		return err
	}

	// Grab the lock on the parent's parent.
	parentParent.AcquireLock()

	// Check if the directory already exists.
	_, err = parentParent.GetSubdirsByName([]string{split[len(split)-1]})
	if err == nil {
		// Already exists.
		return ErrAlreadyExists
	}
	_, err = parentParent.GetFilesByName([]string{split[len(split)-1]})
	if err == nil {
		// Already exists.
		return ErrAlreadyExists
	}

	// Create the parent directory object.
	var root *fs.Directory
	if useParentAccessSettings {
		parentSettings := *parentParent.Settings
		root, err = fs.NewDirectory(split[len(split)-1], false, parentParent, &parentSettings)
		if err != nil {
			parentParent.ReleaseLock()
			return err
		}
	} else {
		root, err = fs.NewDirectory(split[len(split)-1], false, parentParent, parentSettings)
		if err != nil {
			parentParent.ReleaseLock()
			return err
		}
	}

	// Add the parent directory object.
	parentParent.SetSubdirsByName(map[string]*fs.Directory{split[len(split)-1]: root})

	// Grab the lock on the parent.
	root.AcquireLock()

	// Create the directory in the root filesystem.
	err = os.Mkdir(d.getHostPath(parent), os.ModeDir)
	if err != nil {
		parentParent.ReleaseLock()
		return err
	}

	// Release the lock on the parent's parent.
	parentParent.ReleaseLock()

	// Add the subdirectories.
	for i := range dirs {
		// Check for an empty directory.
		if dirs[i] == "" {
			root.ReleaseLock()
			return ErrEmptyPath
		}

		// Split the path.
		splitPath, err := fs.SplitPath(dirs[i])
		if err != nil {
			root.ReleaseLock()
			return err
		}

		// Check that the name is valid.
		if strings.ContainsAny(split[len(split)-1], IllegalNames) {
			root.ReleaseLock()
			return ErrInvalidName
		}

		// Traverse the root to find the new directory's parent.
		current := root
		for i := range splitPath[:len(splitPath)-1] {
			// We don't need to get any locks since we immediately got the
			// write lock on the parent so no other clients can find these'
			// subdirectories.

			// Get the subdirectory for the current directory.
			subdirs, err := current.GetSubdirsByName([]string{splitPath[i]})
			if err != nil {
				if err == fs.ErrItemNotFound {
					// Replace the item not found error with a more useful
					// invalid tree error.
					root.ReleaseLock()
					return ErrInvalidDirectoryTree
				}
				root.ReleaseLock()
				return err
			}
			current = subdirs[0]
		}

		// Create the directory object.
		var newdir *fs.Directory
		if useParentAccessSettings {
			newdir, err = fs.NewDirectory(splitPath[len(splitPath)-1], false, current, parentSettings)
			if err != nil {
				root.ReleaseLock()
				return err
			}
		} else {
			newdir, err = fs.NewDirectory(splitPath[len(splitPath)-1], false, current, settings[i])
			if err != nil {
				root.ReleaseLock()
				return err
			}
		}

		// Add the new directory to the parent.
		current.SetSubdirsByName(map[string]*fs.Directory{splitPath[len(splitPath)-1]: newdir})

		// Create the directory on the host's filesystem.
		err = os.Mkdir(d.getHostPath(filepath.Join(parent, dirs[i])), os.ModeDir)
		if err != nil {
			root.ReleaseLock()
			return err
		}

	}

	// Now that we've added all the new subdirectories we can release the lock
	// on the root.
	root.ReleaseLock()

	// Return.
	return nil
}

// List directory.
func (d *Drive) ListDir(dir string) ([]fs.ListDirObj, error) {
	// Get the directory object.
	dirobj, err := d.GetDirectoryByPath(dir)
	if err != nil {
		return nil, err
	}

	// Return the listed directory.
	return dirobj.ListDir(), nil
}

// Rename directories.
func (d *Drive) RenameDirs(dirs []string, newNames []string, performMatrixOptimization bool) error {
	var err error

	// Check that the lengths of the slices are correct.
	if len(dirs) != len(newNames) {
		return ErrInvalidLength
	}

	// Clean the directories and check that all the new names are valid. Also
	// create a map of dirs to new names.
	dirsToNames := map[string]string{}
	for i := range dirs {
		dirs[i], err = fs.CleanPath(dirs[i])
		if err != nil {
			return err
		}

		// Make sure that the name is correct.
		if strings.ContainsAny(newNames[i], IllegalNames) || newNames[i] == "" {
			return ErrInvalidName
		}

		// Add it to the map.
		dirsToNames[dirs[i]] = newNames[i]
	}

	if performMatrixOptimization {
		// Perform matrix optimization. Group the directories by parent
		// directory.
		var groups map[string][]string
		groups, _, err = groupPathsByParentDir(dirs, []*access.AccessSettings{})
		if err != nil {
			return err
		}

		// Rename the directories in groups.
		for key := range groups {
			// Grab the lock on the parent.
			parent, err := d.GetDirectoryByPath(key)
			if err != nil {
				if err == fs.ErrItemNotFound {
					return ErrPathNotFound
				} else {
					return err
				}
			}
			parent.AcquireLock()

			// Rename the directories.
			directories := map[string]*fs.Directory{}
			for dir := range groups[key] {
				// Check for an empty path.
				if groups[key][dir] == "" {
					parent.ReleaseLock()
					return ErrEmptyPath
				}

				// Split the path.
				split, err := fs.SplitPath(groups[key][dir])
				if err != nil {
					parent.ReleaseLock()
					return err
				}

				// Check that the new name doesn't already exist.
				_, err = parent.GetSubdirsByName([]string{dirsToNames[groups[key][dir]]})
				if err == nil {
					// Already exists.
					parent.ReleaseLock()
					return ErrAlreadyExists
				}
				_, err = parent.GetFilesByName([]string{dirsToNames[groups[key][dir]]})
				if err == nil {
					// Already exists.
					parent.ReleaseLock()
					return ErrAlreadyExists
				}

				// Delete the old object but save the access settings.
				oldSubdir, err := parent.GetSubdirsByName([]string{split[len(split)-1]})
				if err != nil {
					parent.ReleaseLock()
					if err == fs.ErrItemNotFound {
						return ErrPathNotFound
					} else {
						return err
					}
				}
				accessSettings := oldSubdir[0].Settings

				oldSubdirs := parent.GetSubdirs()
				delete(oldSubdirs, split[len(split)-1])
				parent.SetSubdirs(oldSubdirs)

				// Create the new directory object.
				dirobj, err := fs.NewDirectory(dirsToNames[groups[key][dir]], false, parent, accessSettings)
				if err != nil {
					parent.ReleaseLock()
					return err
				}

				// Add the new object.
				directories[dirsToNames[groups[key][dir]]] = dirobj
			}
			// Add the directories to the parent.
			parent.SetSubdirsByName(directories)

			// Rename the new directories in the host filesystem.
			for i := range groups[key] {
				err := os.Rename(d.getHostPath(groups[key][i]), d.getHostPath(filepath.Join(key, dirsToNames[groups[key][i]])))
				if err != nil {
					parent.ReleaseLock()
					return err
				}
			}

			// Release the lock on the parent.
			parent.ReleaseLock()
		}
	} else {
		// Do not perform mutex optimization.
		for i := range dirs {
			// Split the directory.
			split, err := fs.SplitPath(dirs[i])
			if err != nil {
				return err
			}
			// Grab the lock on the parent.
			parent, err := d.GetDirectoryByPath(strings.Join(split[:len(split)-1], "/"))
			if err != nil {
				return err
			}
			parent.AcquireLock()

			// Check if the directory already exists.
			_, err = parent.GetSubdirsByName([]string{dirsToNames[dirs[i]]})
			if err == nil {
				// Already exists.
				parent.ReleaseLock()
				return ErrAlreadyExists
			}
			_, err = parent.GetFilesByName([]string{dirsToNames[dirs[i]]})
			if err == nil {
				// Already exists.
				parent.ReleaseLock()
				return ErrAlreadyExists
			}

			// Delete the old object.
			oldSubdir, err := parent.GetSubdirsByName([]string{split[len(split)-1]})
			if err != nil {
				parent.ReleaseLock()
				if err == fs.ErrItemNotFound {
					return ErrPathNotFound
				} else {
					return err
				}
			}
			accessSettings := oldSubdir[0].Settings

			oldSubdirs := parent.GetSubdirs()
			delete(oldSubdirs, split[len(split)-1])
			parent.SetSubdirs(oldSubdirs)

			// Create the directory object.
			var newdir *fs.Directory
			newdir, err = fs.NewDirectory(dirsToNames[dirs[i]], false,
				parent, accessSettings)
			if err != nil {
				parent.ReleaseLock()
				return err
			}

			// Set the new directory object in the parent.
			parent.SetSubdirsByName(map[string]*fs.Directory{dirsToNames[dirs[i]]: newdir})

			// Rename the directory in the host filesystem.
			err = os.Rename(d.getHostPath(dirs[i]),
				d.getHostPath(filepath.Join(strings.Join(split[:len(split)-1], "/"), dirsToNames[dirs[i]])))
			if err != nil {
				parent.ReleaseLock()
				return err
			}

			// Release the lock on the parent.
			parent.ReleaseLock()
		}
	}

	// Return.
	return nil
}

// Move directories.
func (d *Drive) MoveDirs(dirs, dests []string) error {
	var err error

	// Check that the lengths of the slices are correct.
	if len(dirs) != len(dests) {
		return ErrInvalidLength
	}

	// Clean the directories and check that all the new destinations are valid.
	for i := range dirs {
		dirs[i], err = fs.CleanPath(dirs[i])
		if err != nil {
			return err
		}

		// Clean the destination paths.
		dests[i], err = fs.CleanPath(dests[i])
		if err != nil {
			return err
		}

		// Split the destination path.
		split, err := fs.SplitPath(dests[i])
		if err != nil {
			return err
		}

		// Make sure that the name is correct.
		if strings.ContainsAny(split[len(split)-1], IllegalNames) || split[len(split)-1] == "" {
			return ErrInvalidName
		}
	}

	// Move each directory.
	for i := range dirs {
		// Split the directory and the destination.
		splitDir, err := fs.SplitPath(dirs[i])
		if err != nil {
			return err
		}
		splitDest, err := fs.SplitPath(dests[i])
		if err != nil {
			return err
		}

		// Grab the lock on the parents of the directory and destination.
		parentDir, err := d.GetDirectoryByPath(strings.Join(splitDir[:len(splitDir)-1], "/"))
		if err != nil {
			return err
		}
		parentDir.AcquireLock()
		// If the directory and destination parents are the same, we shouldn't get the destination
		// again.
		var parentDest *fs.Directory
		if strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/") {
			parentDest = parentDir
		} else {
			parentDest, err = d.GetDirectoryByPath(strings.Join(splitDest[:len(splitDest)-1], "/"))
			if err != nil {
				parentDir.ReleaseLock()
				return err
			}
			parentDest.AcquireLock()
		}

		// Check if the directory exists.
		_, err = parentDest.GetSubdirsByName([]string{splitDest[len(splitDest)-1]})
		if err == nil {
			// Already exists.
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return ErrAlreadyExists
		}
		_, err = parentDest.GetFilesByName([]string{splitDest[len(splitDest)-1]})
		if err == nil {
			// Already exists.
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return ErrAlreadyExists
		}

		// Get the old subdir object.
		oldSubdir, err := parentDir.GetSubdirsByName([]string{splitDir[len(splitDir)-1]})
		if err != nil {
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return err
		}

		accessSettings := oldSubdir[0].Settings

		oldSubdirs := parentDir.GetSubdirs()
		delete(oldSubdirs, splitDir[len(splitDir)-1])
		parentDir.SetSubdirs(oldSubdirs)

		// Create the directory object.
		var newdir *fs.Directory
		newdir, err = fs.NewDirectory(splitDest[len(splitDest)-1], false,
			parentDest, accessSettings)
		if err != nil {
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return err
		}

		// Set the new directory object in the parent.
		parentDest.SetSubdirsByName(map[string]*fs.Directory{splitDest[len(splitDest)-1]: newdir})

		// Move the directory on the host filesystem.
		err = os.Rename(d.getHostPath(dirs[i]), d.getHostPath(dests[i]))
		if err != nil {
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return err
		}

		// Release the locks from the parents.
		parentDir.ReleaseLock()
		if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
			parentDest.ReleaseLock()
		}
	}

	// Return.
	return nil
}

// Delete directories.
func (d *Drive) DeleteDirs(dirs []string, performMutexOptimization bool) error {
	var err error

	// Clean the directories.
	for i := range dirs {
		dirs[i], err = fs.CleanPath(dirs[i])
		if err != nil {
			return err
		}
	}

	// Perform mutex optimization.
	if performMutexOptimization {
		// Group the directories by parent directory.
		var groups map[string][]string
		groups, _, err = groupPathsByParentDir(dirs, []*access.AccessSettings{})
		if err != nil {
			return err
		}

		// Delete the directories in groups.
		for key := range groups {
			// Grab the lock on the parent.
			parent, err := d.GetDirectoryByPath(key)
			if err != nil {
				return err
			}
			parent.AcquireLock()

			// Delete the directories.
			directories := parent.GetSubdirs()
			for dir := range groups[key] {
				// Check for an empty directory.
				if groups[key][dir] == "" {
					parent.ReleaseLock()
					return ErrEmptyPath
				}

				// Split the directory.
				split, err := fs.SplitPath(groups[key][dir])
				if err != nil {
					parent.ReleaseLock()
					return err
				}

				// Check to make sure the directory already exists.
				_, err = parent.GetSubdirsByName([]string{split[len(split)-1]})
				if err != nil {
					// Does not exist.
					parent.ReleaseLock()
					if err == fs.ErrItemNotFound {
						return ErrPathNotFound
					}
				}

				// Delete the directory object.
				delete(directories, split[len(split)-1])
			}

			// Set the directories of the parent.
			parent.SetSubdirs(directories)

			// Delete the new directories in the host filesystem.
			for i := range groups[key] {
				err := os.RemoveAll(d.getHostPath(groups[key][i]))
				if err != nil {
					parent.ReleaseLock()
					return err
				}
			}

			// Release the lock on the parent.
			parent.ReleaseLock()
		}
	} else {
		// Do not perform mutex optimization.
		for i := range dirs {
			// Check for an empty directory.
			if dirs[i] == "" {
				return ErrEmptyPath
			}

			// Split the directory.
			split, err := fs.SplitPath(dirs[i])
			if err != nil {
				return err
			}

			// Grab the lock on the parent.
			parent, err := d.GetDirectoryByPath(strings.Join(split[:len(split)-1], "/"))
			if err != nil {
				return err
			}
			parent.AcquireLock()

			// Check if the directory already exists.
			_, err = parent.GetSubdirsByName([]string{split[len(split)-1]})
			if err != nil {
				// Does not exist.
				parent.ReleaseLock()
				if err == fs.ErrItemNotFound {
					return ErrPathNotFound
				}
			}

			// Delete the directory object.
			subdirs := parent.GetSubdirs()
			delete(subdirs, split[len(split)-1])
			parent.SetSubdirs(subdirs)

			// Delete the new directory in the host filesystem.
			err = os.RemoveAll(d.getHostPath(dirs[i]))
			if err != nil {
				parent.ReleaseLock()
				return err
			}

			// Release the lock on the parent.
			parent.ReleaseLock()
		}
	}

	// Return.
	return nil
}

// Create files.
func (d *Drive) CreateFiles(files []string, settings []*access.AccessSettings, useParentAccessSettings, performMutexOptimization bool) error {
	var err error

	// Check that the lengths of the slices are correct.
	if !useParentAccessSettings {
		if len(files) != len(settings) {
			return ErrInvalidLength
		}
	}

	// Clean the paths.
	for i := range files {
		files[i], err = fs.CleanPath(files[i])
		if err != nil {
			return err
		}
	}

	// Perform mutex optimization.
	if performMutexOptimization {
		// Group the files by parent directory.
		var groups map[string][]string
		var accessGroups map[string][]*access.AccessSettings
		if useParentAccessSettings {
			groups, accessGroups, err = groupPathsByParentDir(files, []*access.AccessSettings{})
		} else {
			groups, accessGroups, err = groupPathsByParentDir(files, settings)
		}
		if err != nil {
			return err
		}

		// Create the files in groups.
		for key := range groups {
			// Grab the lock on the parent.
			parent, err := d.GetDirectoryByPath(key)
			if err != nil {
				return err
			}
			parent.AcquireLock()

			// Create the files.
			newFiles := map[string]*fs.File{}
			for dir := range groups[key] {
				// Check for an empty file.
				if groups[key][dir] == "" {
					parent.ReleaseLock()
					return ErrEmptyPath
				}

				// Split the file.
				split, err := fs.SplitPath(groups[key][dir])
				if err != nil {
					parent.ReleaseLock()
					return err
				}

				// Check that the name is valid.
				if strings.ContainsAny(split[len(split)-1], IllegalNames) {
					parent.ReleaseLock()
					return ErrInvalidName
				}

				// Check if the file already exists.
				_, err = parent.GetSubdirsByName([]string{split[len(split)-1]})
				if err == nil {
					// Already exists.
					parent.ReleaseLock()
					return ErrAlreadyExists
				}
				_, err = parent.GetFilesByName([]string{split[len(split)-1]})
				if err == nil {
					// Already exists.
					parent.ReleaseLock()
					return ErrAlreadyExists
				}

				// Create the file object.
				var newfile *fs.File
				if useParentAccessSettings {
					parentSettings := *parent.Settings
					newfile, err = fs.NewFile(split[len(split)-1], &parentSettings)
				} else {
					newfile, err = fs.NewFile(split[len(split)-1], accessGroups[key][dir])
				}
				if err != nil {
					parent.ReleaseLock()
					return err
				}

				// Add the new file object to the map.
				newFiles[split[len(split)-1]] = newfile
			}

			// Add the files to the parent.
			parent.SetFilesByName(newFiles)

			// Create the new files in the host filesystem.
			for i := range groups[key] {
				file, err := os.Create(d.getHostPath(groups[key][i]))
				if err != nil {
					parent.ReleaseLock()
					return err
				}
				err = file.Close()
				if err != nil {
					parent.ReleaseLock()
					return err
				}
			}

			// Release the lock on the parent.
			parent.ReleaseLock()
		}
	} else {
		// Do not perform mutex optimization.
		for i := range files {
			// Check for an empty file.
			if files[i] == "" {
				return ErrEmptyPath
			}

			// Split the file.
			split, err := fs.SplitPath(files[i])
			if err != nil {
				return err
			}
			// Grab the lock on the parent.
			parent, err := d.GetDirectoryByPath(strings.Join(split[:len(split)-1], "/"))
			if err != nil {
				return err
			}
			parent.AcquireLock()

			// Check that the name is valid.
			if strings.ContainsAny(split[len(split)-1], IllegalNames) {
				parent.ReleaseLock()
				return ErrInvalidName
			}

			// Check if the file already exists.
			_, err = parent.GetSubdirsByName([]string{split[len(split)-1]})
			if err == nil {
				// Already exists.
				parent.ReleaseLock()
				return ErrAlreadyExists
			}
			_, err = parent.GetFilesByName([]string{split[len(split)-1]})
			if err == nil {
				// Already exists.
				parent.ReleaseLock()
				return ErrAlreadyExists
			}

			// Create the file object.
			var newfile *fs.File
			if useParentAccessSettings {
				parentSettings := *parent.Settings
				newfile, err = fs.NewFile(split[len(split)-1], &parentSettings)
			} else {
				newfile, err = fs.NewFile(split[len(split)-1], settings[i])
			}
			if err != nil {
				parent.ReleaseLock()
				return err
			}

			// Set the new file object in the parent.
			parent.SetFilesByName(map[string]*fs.File{split[len(split)-1]: newfile})

			// Create the new directory in the host filesystem.
			file, err := os.Create(d.getHostPath(files[i]))
			if err != nil {
				parent.ReleaseLock()
				return err
			}
			err = file.Close()
			if err != nil {
				parent.ReleaseLock()
				return err
			}

			// Release the lock on the parent.
			parent.ReleaseLock()
		}
	}

	// Return.
	return nil
}

// Read files.
func (d *Drive) ReadFiles(files []string, start []int64, end []int64, handler network.ChunkHandler, chunkSize int64) error {
	// Check that the length of starts and ends are correct.
	if len(files) != len(start) || len(files) != len(end) {
		return ErrInvalidStartEnd
	}

	// Get the sizes of each file.
	chunks := []network.ChunkInfo{}
	for i := range files {
		// Check for an empty file.
		clean, err := fs.CleanPath(files[i])
		if err != nil {
			return err
		}
		if clean == "" {
			return ErrEmptyPath
		}

		// Get the file lock.
		file, err := d.GetFileByPath(clean)
		if err != nil {
			return err
		}
		file.AcquireRLock()

		// Get the file size.
		info, err := os.Stat(d.getHostPath(clean))
		if err != nil {
			file.ReleaseRLock()
			return err
		}

		if end[i] == -1 {
			if !(start[i] < info.Size()) {
				file.ReleaseRLock()
				return ErrInvalidStartEnd
			}
			chunks = append(chunks, network.ChunkInfo{
				Name:      files[i],
				NumChunks: int(math.Ceil(float64(info.Size()-start[i]) / float64(chunkSize)))})
		} else {
			if !(start[i] < info.Size() && start[i] >= 0) || !(end[i] <= info.Size() && end[i] > 0) || !(start[i] <= end[i]) {
				file.ReleaseRLock()
				return ErrInvalidStartEnd
			}
			chunks = append(chunks, network.ChunkInfo{
				Name:      files[i],
				NumChunks: int(math.Ceil(float64((end[i] - start[i])) / float64(chunkSize)))})
		}
	}

	// Write the chunks to the handler.
	handler.WriteChunkResponseInfo(chunks)

	for i := range files {
		// We don't have to check again.
		clean, err := fs.CleanPath(files[i])
		if err != nil {
			return err
		}

		// Get the file lock.
		file, err := d.GetFileByPath(clean)
		if err != nil {
			return err
		}
		file.AcquireRLock()

		// Read the file into the chunk handler.
		numChunks := chunks[i].NumChunks
		err = fs.ReadFileChunks(files[i], d.getHostPath(clean), numChunks, chunkSize, start[i], end[i], handler)
		if err != nil {
			file.ReleaseRLock()
			return err
		}

		// Release the lock.
		file.ReleaseRLock()
	}

	// Return.
	return nil
}

// Write files.
func (d *Drive) WriteFiles(files []string, start []int64, handler network.ChunkHandler) error {
	var err error

	// Check that the length of starts and ends are correct.
	if len(files) != len(start) {
		return ErrInvalidStartEnd
	}

	// Read the chunks from the handler.
	chunks, err := handler.GetChunkRequestInfo()
	if err != nil {
		return err
	}

	// Ensure the chunks are correct.
	if len(files) != len(chunks) {
		return ErrInvalidChunks
	}
	for i := range chunks {
		if chunks[i].Name != files[i] {
			return ErrInvalidChunks
		}
	}

	for i := range files {
		// Check for an empty path.
		clean, err := fs.CleanPath(files[i])
		if err != nil {
			return err
		}
		if clean == "" {
			return ErrEmptyPath
		}

		// Get the file lock.
		file, err := d.GetFileByPath(clean)
		if err != nil {
			return err
		}
		file.AcquireLock()

		// Check that the start is correct.
		stat, err := os.Stat(d.getHostPath(clean))
		if err != nil {
			file.ReleaseLock()
			return err
		}
		if !(start[i] <= stat.Size() && start[i] >= 0) {
			return ErrInvalidStartEnd
		}

		// Write to the file from the chunk handler.
		err = fs.WriteFileChunks(files[i], d.getHostPath(clean), int(chunks[i].NumChunks), start[i], handler)
		if err != nil {
			file.ReleaseLock()
			return err
		}

		// Release the lock.
		file.ReleaseLock()
	}

	// Return.
	return nil
}

// Rename files.
func (d *Drive) RenameFiles(files []string, newNames []string, performMatrixOptimization bool) error {
	var err error

	// Check that the lengths of the slices are correct.
	if len(files) != len(newNames) {
		return ErrInvalidLength
	}

	// Clean the paths and check that all the new names are valid. Also
	// create a map of files to new names.
	dirsToNames := map[string]string{}
	for i := range files {
		files[i], err = fs.CleanPath(files[i])
		if err != nil {
			return err
		}

		// Make sure that the name is correct.
		if strings.ContainsAny(newNames[i], IllegalNames) || newNames[i] == "" {
			return ErrInvalidName
		}

		// Add it to the map.
		dirsToNames[files[i]] = newNames[i]
	}

	if performMatrixOptimization {
		// Perform matrix optimization. Group the directories by parent
		// directory.
		var groups map[string][]string
		groups, _, err = groupPathsByParentDir(files, []*access.AccessSettings{})
		if err != nil {
			return err
		}

		// Rename the files in groups.
		for key := range groups {
			// Grab the lock on the parent.
			parent, err := d.GetDirectoryByPath(key)
			if err != nil {
				if err == fs.ErrItemNotFound {
					return ErrPathNotFound
				} else {
					return err
				}
			}
			parent.AcquireLock()

			// Rename the files.
			newFiles := map[string]*fs.File{}
			for dir := range groups[key] {
				// Check for an empty path.
				if groups[key][dir] == "" {
					parent.ReleaseLock()
					return ErrEmptyPath
				}

				// Split the path.
				split, err := fs.SplitPath(groups[key][dir])
				if err != nil {
					parent.ReleaseLock()
					return err
				}

				// Check that the new name doesn't already exist.
				_, err = parent.GetSubdirsByName([]string{dirsToNames[groups[key][dir]]})
				if err == nil {
					// Already exists.
					parent.ReleaseLock()
					return ErrAlreadyExists
				}
				_, err = parent.GetFilesByName([]string{dirsToNames[groups[key][dir]]})
				if err == nil {
					// Already exists.
					parent.ReleaseLock()
					return ErrAlreadyExists
				}

				// Delete the old object but save the access settings.
				oldFile, err := parent.GetFilesByName([]string{split[len(split)-1]})
				if err != nil {
					fmt.Printf(split[len(split)-1])
					parent.ReleaseLock()
					if err == fs.ErrItemNotFound {
						return ErrPathNotFound
					} else {
						return err
					}
				}
				accessSettings := oldFile[0].Settings

				oldFiles := parent.GetFiles()
				delete(oldFiles, split[len(split)-1])
				parent.SetFiles(oldFiles)

				// Create the new file object.
				fileobj, err := fs.NewFile(dirsToNames[groups[key][dir]], accessSettings)
				if err != nil {
					parent.ReleaseLock()
					return err
				}

				// Add the new object.
				newFiles[dirsToNames[groups[key][dir]]] = fileobj
			}
			// Add the files to the parent.
			parent.SetFilesByName(newFiles)

			// Create the new f in the host filesystem.
			for i := range groups[key] {
				err := os.Rename(d.getHostPath(groups[key][i]), d.getHostPath(filepath.Join(key, dirsToNames[groups[key][i]])))
				if err != nil {
					parent.ReleaseLock()
					return err
				}
			}

			// Release the lock on the parent.
			parent.ReleaseLock()
		}
	} else {
		// Do not perform mutex optimization.
		for i := range files {
			// Split the file.
			split, err := fs.SplitPath(files[i])
			if err != nil {
				return err
			}
			// Grab the lock on the parent.
			parent, err := d.GetDirectoryByPath(strings.Join(split[:len(split)-1], "/"))
			if err != nil {
				return err
			}
			parent.AcquireLock()

			// Check if the directory already exists.
			_, err = parent.GetSubdirsByName([]string{dirsToNames[files[i]]})
			if err == nil {
				// Already exists.
				parent.ReleaseLock()
				return ErrAlreadyExists
			}
			_, err = parent.GetFilesByName([]string{dirsToNames[files[i]]})
			if err == nil {
				// Already exists.
				parent.ReleaseLock()
				return ErrAlreadyExists
			}

			// Delete the old object.
			oldFile, err := parent.GetFilesByName([]string{split[len(split)-1]})
			if err != nil {
				parent.ReleaseLock()
				if err == fs.ErrItemNotFound {
					return ErrPathNotFound
				} else {
					return err
				}
			}
			accessSettings := oldFile[0].Settings

			oldFiles := parent.GetFiles()
			delete(oldFiles, split[len(split)-1])
			parent.SetFiles(oldFiles)

			// Create the file object.
			var newfile *fs.File
			newfile, err = fs.NewFile(dirsToNames[files[i]], accessSettings)
			if err != nil {
				parent.ReleaseLock()
				return err
			}

			// Set the new file object in the parent.
			parent.SetFilesByName(map[string]*fs.File{dirsToNames[files[i]]: newfile})

			// Rename the file in the host filesystem.
			err = os.Rename(d.getHostPath(files[i]),
				d.getHostPath(filepath.Join(strings.Join(split[:len(split)-1], "/"), dirsToNames[files[i]])))
			if err != nil {
				parent.ReleaseLock()
				return err
			}

			// Release the lock on the parent.
			parent.ReleaseLock()
		}
	}

	// Return.
	return nil
}

// Move files.
func (d *Drive) MoveFiles(files, dests []string) error {
	var err error

	// Check that the lengths of the slices are correct.
	if len(files) != len(dests) {
		return ErrInvalidLength
	}

	// Clean the files and check that all the new destinations are valid.
	for i := range files {
		files[i], err = fs.CleanPath(files[i])
		if err != nil {
			return err
		}

		// Clean the destination paths.
		dests[i], err = fs.CleanPath(dests[i])
		if err != nil {
			return err
		}

		// Split the destination path.
		split, err := fs.SplitPath(dests[i])
		if err != nil {
			return err
		}

		// Make sure that the name is correct.
		if strings.ContainsAny(split[len(split)-1], IllegalNames) || split[len(split)-1] == "" {
			return ErrInvalidName
		}
	}

	// Move each file.
	for i := range files {
		// Split the directory and the destination.
		splitDir, err := fs.SplitPath(files[i])
		if err != nil {
			return err
		}
		splitDest, err := fs.SplitPath(dests[i])
		if err != nil {
			return err
		}

		// Grab the lock on the parents of the directory and destination.
		parentDir, err := d.GetDirectoryByPath(strings.Join(splitDir[:len(splitDir)-1], "/"))
		if err != nil {
			return err
		}
		parentDir.AcquireLock()
		// If the directory and destination parents are the same, we shouldn't get the destination
		// again.
		var parentDest *fs.Directory
		if strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/") {
			parentDest = parentDir
		} else {
			parentDest, err = d.GetDirectoryByPath(strings.Join(splitDest[:len(splitDest)-1], "/"))
			if err != nil {
				parentDir.ReleaseLock()
				return err
			}
			parentDest.AcquireLock()
		}

		// Check if the directory exists.
		_, err = parentDest.GetSubdirsByName([]string{splitDest[len(splitDest)-1]})
		if err == nil {
			// Already exists.
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return ErrAlreadyExists
		}
		_, err = parentDest.GetFilesByName([]string{splitDest[len(splitDest)-1]})
		if err == nil {
			// Already exists.
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return ErrAlreadyExists
		}

		// Get the old file object.
		oldFile, err := parentDir.GetFilesByName([]string{splitDir[len(splitDir)-1]})
		if err != nil {
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return err
		}

		accessSettings := oldFile[0].Settings

		oldFiles := parentDir.GetFiles()
		delete(oldFiles, splitDir[len(splitDir)-1])
		parentDir.SetFiles(oldFiles)

		// Create the file object.
		var newfile *fs.File
		newfile, err = fs.NewFile(splitDest[len(splitDest)-1], accessSettings)
		if err != nil {
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return err
		}

		// Set the new file object in the parent.
		parentDest.SetFilesByName(map[string]*fs.File{splitDest[len(splitDest)-1]: newfile})

		// Move the file on the host filesystem.
		err = os.Rename(d.getHostPath(files[i]), d.getHostPath(dests[i]))
		if err != nil {
			parentDir.ReleaseLock()
			if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
				parentDest.ReleaseLock()
			}
			return err
		}

		// Release the locks from the parents.
		parentDir.ReleaseLock()
		if !(strings.Join(splitDest[:len(splitDest)-1], "/") == strings.Join(splitDir[:len(splitDir)-1], "/")) {
			parentDest.ReleaseLock()
		}
	}

	// Return.
	return nil
}

// Delete files.
func (d *Drive) DeleteFiles(files []string, performMutexOptimization bool) error {
	var err error

	// Clean the files.
	for i := range files {
		files[i], err = fs.CleanPath(files[i])
		if err != nil {
			return err
		}
	}

	// Perform mutex optimization.
	if performMutexOptimization {
		// Group the files by parent directory.
		var groups map[string][]string
		groups, _, err = groupPathsByParentDir(files, []*access.AccessSettings{})
		if err != nil {
			return err
		}

		// Delete the files in groups.
		for key := range groups {
			// Grab the lock on the parent.
			parent, err := d.GetDirectoryByPath(key)
			if err != nil {
				return err
			}
			parent.AcquireLock()

			// Delete the files.
			dirFiles := parent.GetFiles()
			for dir := range groups[key] {
				// Check for an empty path.
				if groups[key][dir] == "" {
					parent.ReleaseLock()
					return ErrEmptyPath
				}

				// Split the file.
				split, err := fs.SplitPath(groups[key][dir])
				if err != nil {
					parent.ReleaseLock()
					return err
				}

				// Check to make sure the file already exists.
				_, err = parent.GetFilesByName([]string{split[len(split)-1]})
				if err != nil {
					// Does not exist.
					parent.ReleaseLock()
					if err == fs.ErrItemNotFound {
						return ErrPathNotFound
					}
				}

				// Delete the file object.
				delete(dirFiles, split[len(split)-1])
			}

			// Set the files of the parent.
			parent.SetFiles(dirFiles)

			// Delete the new files in the host filesystem.
			for i := range groups[key] {
				err := os.Remove(d.getHostPath(groups[key][i]))
				if err != nil {
					parent.ReleaseLock()
					return err
				}
			}

			// Release the lock on the parent.
			parent.ReleaseLock()
		}
	} else {
		// Do not perform mutex optimization.
		for i := range files {
			// Check for an empty path.
			if files[i] == "" {
				return ErrEmptyPath
			}

			// Split the path.
			split, err := fs.SplitPath(files[i])
			if err != nil {
				return err
			}

			// Grab the lock on the parent.
			parent, err := d.GetDirectoryByPath(strings.Join(split[:len(split)-1], "/"))
			if err != nil {
				return err
			}
			parent.AcquireLock()

			// Check if the file already exists.
			_, err = parent.GetFilesByName([]string{split[len(split)-1]})
			if err != nil {
				// Does not exist.
				parent.ReleaseLock()
				if err == fs.ErrItemNotFound {
					return ErrPathNotFound
				}
			}

			// Delete the file object.
			dirFiles := parent.GetFiles()
			delete(dirFiles, split[len(split)-1])
			parent.SetFiles(dirFiles)

			// Delete the new file in the host filesystem.
			err = os.Remove(d.getHostPath(files[i]))
			if err != nil {
				parent.ReleaseLock()
				return err
			}

			// Release the lock on the parent.
			parent.ReleaseLock()
		}
	}

	// Return.
	return nil
}

// Get the status for paths.
func (d *Drive) Stat(paths []string) ([]PathStatus, error) {
	// Loop over each path.
	outputs := []PathStatus{}
	for i := range paths {
		// Split the path.
		split, err := fs.SplitPath(paths[i])
		if err != nil {
			return []PathStatus{}, err
		}

		// If the path is empty, return the stat for the root.
		if len(split) == 0 {
			outputs = append(outputs, PathStatus{
				Exists: true,
				Name:   paths[i],
				IsFile: false,
			})
			continue
		}

		// List the parent directory.
		parent := strings.Join(split[:len(split)-1], "/")
		listdir, err := d.ListDir(parent)
		if err != nil {
			return []PathStatus{}, err
		}

		// Try to find our item.
		found := false
		for j := range listdir {
			if listdir[j].Name == split[len(split)-1] {
				// Found it.
				outputs = append(outputs, PathStatus{
					Exists: true,
					Name:   paths[i],
					IsFile: listdir[j].File,
				})
				found = true
			}
		}
		if found {
			continue
		}
		// Did not find it.
		outputs = append(outputs, PathStatus{
			Exists: false,
			Name:   paths[i],
			IsFile: false,
		})
	}

	// Return.
	return outputs, nil
}

// Recalculate hashes for files.
func (d *Drive) ReHash(files []string, performMutexOptimization bool) error {
	var err error

	// Clean the files.
	for i := range files {
		files[i], err = fs.CleanPath(files[i])
		if err != nil {
			return err
		}
	}

	hasher := sha256.New()
	for i := range files {
		// Check for an empty path.
		if files[i] == "" {
			return ErrEmptyPath
		}

		// Grab the lock on the file.
		file, err := d.GetFileByPath(files[i])
		if err != nil {
			return err
		}
		file.AcquireLock()

		// Calculate the hash.
		f, err := os.Open(d.getHostPath(files[i]))
		if err != nil {
			file.ReleaseLock()
			return err
		}
		if _, err := io.Copy(hasher, f); err != nil {
			file.ReleaseLock()
			f.Close()
			return err
		}
		hash := []byte{}
		hash = hasher.Sum(hash)
		// Release the lock on the file.
		file.ReleaseLock()

		file.SetHash(hash)
		hasher.Reset()

		f.Close()
	}

	// Return.
	return nil
}

// Verify hashes for files.
func (d *Drive) VerifyHashes(files []string) (map[string]bool, error) {
	hasher := sha256.New()
	outputs := map[string]bool{}
	for i := range files {
		// Check for an empty path.
		clean, err := fs.CleanPath(files[i])
		if err != nil {
			return map[string]bool{}, err
		}
		if clean == "" {
			return map[string]bool{}, ErrEmptyPath
		}

		// Grab the lock on the file.
		file, err := d.GetFileByPath(clean)
		if err != nil {
			return map[string]bool{}, err
		}
		file.AcquireLock()

		// Calculate the hash.
		f, err := os.Open(d.getHostPath(clean))
		if err != nil {
			file.ReleaseLock()
			return map[string]bool{}, err
		}
		if _, err := io.Copy(hasher, f); err != nil {
			file.ReleaseLock()
			f.Close()
			return map[string]bool{}, err
		}
		hash := []byte{}
		hash = hasher.Sum(hash)
		hasher.Reset()

		f.Close()

		// Release the lock on the file.
		file.ReleaseLock()

		// Compare the hashes.
		if bytes.Equal(hash, file.GetHash()) {
			outputs[files[i]] = true
		} else {
			outputs[files[i]] = false
		}
	}

	// Return.
	return outputs, nil
}
