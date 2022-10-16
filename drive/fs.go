// drive/fs.go
// Filesystem commands and functions for Lily drive objects.

package drive

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/cubeflix/lily/fs"
	"github.com/cubeflix/lily/security/access"
)

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
				// Split the directory.
				split, err := fs.SplitPath(groups[key][dir])
				if err != nil {
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
