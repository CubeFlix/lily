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
