// +build windows

package chezmoi

import (
	"errors"
	"path/filepath"
	"syscall"
	"os"
	"unsafe"

	"github.com/google/renameio"
	vfs "github.com/twpayne/go-vfs"
)

// WriteFile implements Mutator.WriteFile.
func (a *FSMutator) WriteFile(name string, data []byte, perm os.FileMode, currData []byte) error {
	// Special case: if writing to the real filesystem, use github.com/google/renameio
	if a.FS == vfs.OSFS {
		dir := filepath.Dir(name)
		dev, ok := a.devCache[dir]
		if !ok {

			// info, err := a.Stat(dir)
			// if err != nil {
			// 	return err
			// }
			// statT, ok := info.Sys().(*syscall.Stat_t)
			// if !ok {
			// 	return errors.New("os.FileInfo.Sys() cannot be converted to a *syscall.Stat_t")
			// }
			// dev = uint(statT.Dev)
			volumeID, err := getVolumeSerialNumber(name)
			if err != nil {
				return err
			}

			dev = volumeID
			a.devCache[dir] = dev
		}
		tempDir, ok := a.tempDirCache[dev]
		if !ok {
			tempDir = renameio.TempDir(dir)
			a.tempDirCache[dev] = tempDir
		}
		t, err := renameio.TempFile(tempDir, name)
		if err != nil {
			return err
		}
		defer func() {
			_ = t.Cleanup()
		}()
		if err := t.Chmod(perm); err != nil {
			return err
		}
		if _, err := t.Write(data); err != nil {
			return err
		}
		return t.CloseAtomicallyReplace()
	}
	return a.FS.WriteFile(name, data, perm)
}

func getVolumeSerialNumber(Path string) (uint, error) {
	fp, err := filepath.Abs(Path)
	if err != nil {
		return 0, err
	}

	// Input rootpath
	var RootPathName = filepath.VolumeName(fp) + "\\"
	
	// Output volume info
	var lpVolumeNameBuffer = make([]uint16, syscall.MAX_PATH+1)
	var nVolumeNameSize = uint32(len(lpVolumeNameBuffer))
	var lpVolumeSerialNumber uint32
	var lpMaximumComponentLength uint32
	var lpFileSystemFlags uint32
	var lpFileSystemNameBuffer = make([]uint16, 255)
	var nFileSystemNameSize uint32 = syscall.MAX_PATH + 1

	kernel32, _ := syscall.LoadLibrary("kernel32.dll")
	GetVolumeInformationW, _ := syscall.GetProcAddress(kernel32, "GetVolumeInformationW")

	var nargs uintptr = 8
	ret, _, callErr := syscall.Syscall9(uintptr(GetVolumeInformationW),
		nargs,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(RootPathName))),
		uintptr(unsafe.Pointer(&lpVolumeNameBuffer[0])),
		uintptr(nVolumeNameSize),
		uintptr(unsafe.Pointer(&lpVolumeSerialNumber)),
		uintptr(unsafe.Pointer(&lpMaximumComponentLength)),
		uintptr(unsafe.Pointer(&lpFileSystemFlags)),
		uintptr(unsafe.Pointer(&lpFileSystemNameBuffer[0])),
		uintptr(nFileSystemNameSize),
		0)

	if ret != 1 {
		return 0, errors.New(string(callErr))
	}

	return uint(lpVolumeSerialNumber), nil
}