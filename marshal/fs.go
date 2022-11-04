// marshal/fs.go
// Marshaling functions for internal filesystem objects.

package marshal

import (
	"encoding/binary"
	"io"
	"time"

	"github.com/cubeflix/lily/fs"
)

// Marshal a directory object.
func MarshalDirectory(d *fs.Directory, w io.Writer) error {
	// Write the local name.
	err := MarshalString(d.GetPath(), w)
	if err != nil {
		return err
	}

	// Write the last edit information.
	if MarshalString(d.GetLastEditor(), w) != nil {
		return err
	}
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(d.GetLastEditTime().Unix()))
	if _, err := w.Write(data); err != nil {
		return err
	}

	// Acquire the read lock.
	d.AcquireRLock()

	// Write the directory access settings.
	if MarshalAccess(d.Settings, w) != nil {
		d.ReleaseRLock()
		return err
	}

	// Get the subdirectories and files.
	subdirs := d.GetSubdirs()
	files := d.GetFiles()
	data = make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(len(subdirs)))
	if _, err := w.Write(data); err != nil {
		d.ReleaseRLock()
		return err
	}
	binary.LittleEndian.PutUint32(data, uint32(len(files)))
	if _, err := w.Write(data); err != nil {
		d.ReleaseRLock()
		return err
	}

	// Release the lock, as we now have the information we need.
	d.ReleaseRLock()

	// Write the subdirectories.
	for dir := range subdirs {
		// Write the subdirectory.
		if MarshalDirectory(subdirs[dir], w) != nil {
			return err
		}
	}

	// Write the files.
	for file := range files {
		// Write the file.
		if MarshalFile(files[file], w) != nil {
			return err
		}
	}

	// Return.
	return nil
}

// Unmarshal a directory.
func UnmarshalDirectory(r io.Reader, isRoot bool, parent *fs.Directory) (*fs.Directory, error) {
	// Get the local name.
	name, err := UnmarshalString(r)
	if err != nil {
		return nil, err
	}

	// Get the last edit information
	lastEditor, err := UnmarshalString(r)
	if err != nil {
		return nil, err
	}
	data := make([]byte, 8)
	if _, err := r.Read(data); err != nil {
		return nil, err
	}
	lastEditUnix := binary.LittleEndian.Uint64(data)
	lastEdit := time.Unix(int64(lastEditUnix), 0)
	if err != nil {
		return nil, err
	}

	// Get the directory access settings.
	aobj, err := UnmarshalAccess(r)
	if err != nil {
		return nil, err
	}

	// Create the new directory.
	dirobj, err := fs.NewDirectory(name, isRoot, parent, aobj)
	if err != nil {
		return nil, err
	}
	dirobj.SetLastEditor(lastEditor)
	dirobj.SetLastEditTime(lastEdit)

	// Get the subdirectories and files.
	data = make([]byte, 4)
	if _, err := r.Read(data); err != nil {
		return nil, err
	}
	numSubdirs := int(binary.LittleEndian.Uint32(data))
	if _, err := r.Read(data); err != nil {
		return nil, err
	}
	numFiles := int(binary.LittleEndian.Uint32(data))
	subdirs := map[string]*fs.Directory{}
	for i := 0; i < numSubdirs; i++ {
		newdir, err := UnmarshalDirectory(r, false, dirobj)
		if err != nil {
			return nil, err
		}
		subdirs[newdir.GetPath()] = newdir
	}
	files := map[string]*fs.File{}
	for i := 0; i < numFiles; i++ {
		newfile, err := UnmarshalFile(r)
		if err != nil {
			return nil, err
		}
		files[newfile.GetPath()] = newfile
	}

	// Set the children.
	dirobj.SetSubdirs(subdirs)
	dirobj.SetFiles(files)

	// Return.
	return dirobj, nil
}

// Marshal a file object.
func MarshalFile(f *fs.File, w io.Writer) error {
	// Write the local name.
	err := MarshalString(f.GetPath(), w)
	if err != nil {
		return err
	}

	// Write the last edit information.
	if MarshalString(f.GetLastEditor(), w) != nil {
		return err
	}
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(f.GetLastEditTime().Unix()))
	if _, err := w.Write(data); err != nil {
		return err
	}

	// Write the hash.
	hash := f.GetHash()
	if len(hash) != 32 {
		hash = make([]byte, 32)
	}
	if _, err := w.Write(hash); err != nil {
		return err
	}

	// Acquire the read lock.
	f.AcquireRLock()

	// Write the file access settings.
	if MarshalAccess(f.Settings, w) != nil {
		f.ReleaseRLock()
		return err
	}

	// Release the read lock.
	f.ReleaseRLock()

	// Return.
	return nil
}

// Unmarshal a file.
func UnmarshalFile(r io.Reader) (*fs.File, error) {
	// Get the local name.
	name, err := UnmarshalString(r)
	if err != nil {
		return nil, err
	}

	// Get the last edit information
	lastEditor, err := UnmarshalString(r)
	if err != nil {
		return nil, err
	}
	data := make([]byte, 8)
	if _, err := r.Read(data); err != nil {
		return nil, err
	}
	lastEditUnix := binary.LittleEndian.Uint64(data)
	lastEdit := time.Unix(int64(lastEditUnix), 0)
	if err != nil {
		return nil, err
	}

	// Get the hash.
	hash := make([]byte, 32)
	if _, err := r.Read(hash); err != nil {
		return nil, err
	}

	// Get the file access settings.
	aobj, err := UnmarshalAccess(r)
	if err != nil {
		return nil, err
	}

	// Create the new file.
	fileobj, err := fs.NewFile(name, aobj)
	if err != nil {
		return nil, err
	}
	fileobj.SetLastEditor(lastEditor)
	fileobj.SetLastEditTime(lastEdit)
	fileobj.SetHash(hash)

	// Return.
	return fileobj, nil
}
