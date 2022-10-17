// drive/fs.go
// Filesystem commands and functions for Lily drive objects.

package drive

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/cubeflix/lily/fs"
	"github.com/cubeflix/lily/security/access"
)

var ErrEmptyPath = errors.New("lily.drive: Empty path")
var ErrNotAChildOf = errors.New("lily.drive: Path is not a child of parent")
var ErrAlreadyExists = errors.New("lily.drive: Path already exists")
var ErrInvalidDirectoryTree = errors.New("lily.drive: Invalid directory tree")
var ErrInvalidName = errors.New("lily.drive: Invalid name")

var IllegalNames = "\"*/:<>?\\|"

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
					return err
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
				return err
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

	// Create the parent directory object.
	var root *fs.Directory
	if useParentAccessSettings {
		parentSettings := *parentParent.Settings
		root, err = fs.NewDirectory(parent, false, parentParent, &parentSettings)
		if err != nil {
			parentParent.ReleaseLock()
			return err
		}
	} else {
		root, err = fs.NewDirectory(parent, false, parentParent, parentSettings)
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
		newdir, err := fs.NewDirectory(parent, false, current, parentSettings)
		if err != nil {
			root.ReleaseLock()
			return err
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
