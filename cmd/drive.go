// cmd/drive.go
// Drive commands.

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cubeflix/lily/drive"
	"github.com/cubeflix/lily/fs"
	"github.com/cubeflix/lily/security/access"
	"github.com/spf13/cobra"
)

func DriveInit(cmd *cobra.Command, args []string) {
	ac := access.Clearance(accessClearance)
	err := ac.Validate()
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}
	mc := access.Clearance(modifyClearance)
	err = mc.Validate()
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}
	accessSettings, err := access.NewAccessSettings(ac, mc)
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}

	if !filepath.IsAbs(args[1]) {
		fmt.Println("drive: path must be absolute")
		return
	}
	if _, err := os.Stat(args[1]); err != nil {
		fmt.Println("drive:", err.Error())
		return
	}
	root, err := fs.NewDirectory("", true, nil, accessSettings)
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}

	d := drive.NewDrive(args[0], args[1], accessSettings, root)
	driveFile = strings.ReplaceAll(driveFile, "%name%", args[0])
	f, err := os.Create(driveFile)
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}
	err = d.Marshal(f)
	if err != nil {
		fmt.Println("drive:", err.Error())
		f.Close()
		return
	}
	f.Close()
}

func DriveSetPath(cmd *cobra.Command, args []string) {
	if !filepath.IsAbs(args[0]) {
		fmt.Println("drive: path must be absolute")
		return
	}
	if _, err := os.Stat(args[0]); err != nil {
		fmt.Println("drive:", err.Error())
		return
	}

	f, err := os.Open(driveFile)
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}
	d, err := drive.Unmarshal(f)
	if err != nil {
		f.Close()
		fmt.Println("drive:", err.Error())
		return
	}
	f.Close()

	d.SetPath(args[0])

	f, err = os.Create(driveFile)
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}
	err = d.Marshal(f)
	if err != nil {
		fmt.Println("drive:", err.Error())
		f.Close()
		return
	}
	f.Close()
}

func fsFromDir(path, name string, isRoot bool, parent *fs.Directory, ac, mc access.Clearance, depth int) (*fs.Directory, error) {
	fmt.Println(strings.Repeat("  ", depth)+"dir:", name)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	accessSettings, err := access.NewAccessSettings(ac, mc)
	if err != nil {
		return nil, err
	}

	newDir, err := fs.NewDirectory(name, isRoot, parent, accessSettings)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if f.IsDir() {
			newSubdir, err := fsFromDir(filepath.Join(path, f.Name()), f.Name(), false, newDir, ac, mc, depth+1)
			if err != nil {
				return nil, err
			}
			newDir.SetSubdirsByName(map[string]*fs.Directory{f.Name(): newSubdir})
		} else {
			fmt.Println(strings.Repeat("  ", depth+1)+"file:", f.Name())
			newFile, err := fs.NewFile(f.Name(), accessSettings)
			if err != nil {
				return nil, err
			}
			newDir.SetFilesByName(map[string]*fs.File{f.Name(): newFile})
		}
	}
	return newDir, nil
}

func DriveReimport(cmd *cobra.Command, args []string) {
	ac := access.Clearance(accessClearance)
	err := ac.Validate()
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}
	mc := access.Clearance(modifyClearance)
	err = mc.Validate()
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}

	f, err := os.Open(driveFile)
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}
	d, err := drive.Unmarshal(f)
	if err != nil {
		f.Close()
		fmt.Println("drive:", err.Error())
		return
	}
	f.Close()

	root, err := fsFromDir(d.GetPath(), "", true, nil, ac, mc, 0)
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}
	d.SetRoot(root)

	f, err = os.Create(driveFile)
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}
	err = d.Marshal(f)
	if err != nil {
		fmt.Println("drive:", err.Error())
		f.Close()
		return
	}
	f.Close()
}

func DriveSettings(cmd *cobra.Command, args []string) {
	f, err := os.Open(driveFile)
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}
	d, err := drive.Unmarshal(f)
	if err != nil {
		f.Close()
		fmt.Println("drive:", err.Error())
		return
	}
	f.Close()

	fmt.Println("name:", d.GetName())
	fmt.Println("path:", d.GetPath())
}

func listDriveDir(dir *fs.Directory, depth int) {
	fmt.Println(strings.Repeat("  ", depth)+"dir:", dir.GetPath())
	for i := range dir.GetFiles() {
		fmt.Println(strings.Repeat("  ", depth+1)+"file:", dir.GetFiles()[i].GetPath())
	}
	for i := range dir.GetSubdirs() {
		listDriveDir(dir.GetSubdirs()[i], depth+1)
	}
}

func DriveList(cmd *cobra.Command, args []string) {
	f, err := os.Open(driveFile)
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}
	d, err := drive.Unmarshal(f)
	if err != nil {
		f.Close()
		fmt.Println("drive:", err.Error())
		return
	}
	f.Close()

	listDriveDir(d.GetRoot(), 0)
}

func fixDir(path string, dir *fs.Directory, ac, mc access.Clearance, depth int) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	// Iterate over the OS files.
	for _, f := range files {
		if f.IsDir() {
			// Check that this directory exists in the drive.
			subdirs, err := dir.GetSubdirsByName([]string{f.Name()})
			if err != nil {
				// Does not exist.
				fmt.Println(strings.Repeat("  ", depth+1)+"discrepancy: dir", f.Name(), "found in filesystem, not in drive")
				var answer string
				for {
					fmt.Println(strings.Repeat("  ", depth+1) + "you may [r]eimport the directory, or [i]gnore:")

					fmt.Scanln(&answer)
					if len(answer) != 1 {
						continue
					}
					if answer[0] == 'r' {
						newDir, err := fsFromDir(filepath.Join(path, f.Name()), f.Name(), false, dir, ac, mc, depth+1)
						if err != nil {
							return err
						}
						dir.SetSubdirsByName(map[string]*fs.Directory{f.Name(): newDir})
						break
					} else if answer[1] == 'i' {
						break
					}
				}
			} else {
				err = fixDir(filepath.Join(path, f.Name()), subdirs[0], ac, mc, depth+1)
				if err != nil {
					return err
				}
			}
		} else {
			// Check that this file exists in the drive.
			_, err := dir.GetFilesByName([]string{f.Name()})
			if err != nil {
				// Does not exist.
				fmt.Println(strings.Repeat("  ", depth+1)+"discrepancy: file", f.Name(), "found in filesystem, not in drive")
				var answer string
				for {
					fmt.Println(strings.Repeat("  ", depth+1) + "you may [r]eimport the file, or [i]gnore:")
					fmt.Scanln(&answer)
					if len(answer) != 1 {
						continue
					}
					if answer[0] == 'r' {
						accessSettings, err := access.NewAccessSettings(ac, mc)
						if err != nil {
							return err
						}
						newFile, err := fs.NewFile(f.Name(), accessSettings)
						if err != nil {
							return err
						}
						dir.SetFilesByName(map[string]*fs.File{f.Name(): newFile})
						break
					} else if answer[1] == 'i' {
						break
					}
				}
			}
		}
	}

	// Iterate over drive subdirs and files.
	for i := range dir.GetSubdirs() {
		// Make sure the dir exists.
		stat, err := os.Stat(filepath.Join(path, i))
		if err != nil || (err == nil && !stat.IsDir()) {
			// Directory does not exist.
			fmt.Println(strings.Repeat("  ", depth+1)+"discrepancy: dir", i, "found in drive, not in filesystem")
			var answer string
			for {
				fmt.Println(strings.Repeat("  ", depth+1) + "you may [c]reate, or [i]gnore:")
				fmt.Scanln(&answer)
				if len(answer) != 1 {
					continue
				}
				if answer[0] == 'c' {
					if err := os.Mkdir(filepath.Join(path, i), 0666); err != nil {
						return err
					}
					break
				} else if answer[1] == 'i' {
					break
				}
			}
		}
		err = fixDir(filepath.Join(path, i), dir.GetSubdirs()[i], ac, mc, depth+1)
		if err != nil {
			return err
		}
	}
	for i := range dir.GetFiles() {
		// Make sure the dir exists.
		stat, err := os.Stat(filepath.Join(path, i))
		if err != nil || (err == nil && stat.IsDir()) {
			// Directory does not exist.
			fmt.Println(strings.Repeat("  ", depth+1)+"discrepancy: file", i, "found in drive, not in filesystem")
			var answer string
			for {
				fmt.Println(strings.Repeat("  ", depth+1) + "you may [c]reate, or [i]gnore:")
				fmt.Scanln(&answer)
				if len(answer) != 1 {
					continue
				}
				if answer[0] == 'c' {
					f, err := os.Create(filepath.Join(path, i))
					if err != err {
						return err
					}
					f.Close()
					break
				} else if answer[1] == 'i' {
					break
				}
			}
		}
	}

	return nil
}

func DriveFix(cmd *cobra.Command, args []string) {
	ac := access.Clearance(accessClearance)
	err := ac.Validate()
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}
	mc := access.Clearance(modifyClearance)
	err = mc.Validate()
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}

	f, err := os.Open(driveFile)
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}
	d, err := drive.Unmarshal(f)
	if err != nil {
		f.Close()
		fmt.Println("drive:", err.Error())
		return
	}
	f.Close()

	err = fixDir(d.GetPath(), d.GetRoot(), ac, mc, 0)
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}

	f, err = os.Create(driveFile)
	if err != nil {
		fmt.Println("drive:", err.Error())
		return
	}
	err = d.Marshal(f)
	if err != nil {
		fmt.Println("drive:", err.Error())
		f.Close()
		return
	}
	f.Close()
}
