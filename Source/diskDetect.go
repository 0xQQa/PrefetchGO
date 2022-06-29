//https://gist.github.com/kostix/398fa1c96911240a402f7801e5404323
package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

var allDisks []DiskInfo

type DiskInfo struct {
	letter string
	serial string
}

/*
 * Funkcja pozsykujaca numer seryjn z zadanego dysku
 */

func getHexSerialForDisck(diskLetter rune) string {
	var nargs uintptr = 8
	var RootPathName = string(diskLetter) + `:\`
	var VolumeSerialNumber uint32

	kernel32, _ := syscall.LoadLibrary("kernel32.dll")
	getVolume, _ := syscall.GetProcAddress(kernel32, "GetVolumeInformationW")
	ret, _, _ := syscall.Syscall9(uintptr(getVolume), nargs, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(RootPathName))),
		uintptr(unsafe.Pointer(nil)), uintptr(unsafe.Pointer(nil)), uintptr(unsafe.Pointer(&VolumeSerialNumber)),
		uintptr(unsafe.Pointer(nil)), uintptr(unsafe.Pointer(nil)), uintptr(unsafe.Pointer(nil)), uintptr(unsafe.Pointer(nil)), 0)

	if ret == 0 {
		return ""
	}

	sn := fmt.Sprintf("%X", VolumeSerialNumber)
	return sn
}

/*
 * Wrapper pozsykujacy numery seryjne z dostepnych
 * dyskow na stacji
 */

func getAllDisksHexSerials() {
	for diskLetter := 'A'; diskLetter <= 'Z'; diskLetter++ {
		foundedSerial := getHexSerialForDisck(diskLetter)
		if foundedSerial != "" {
			allDisks = append(allDisks, DiskInfo{letter: string(diskLetter), serial: foundedSerial})
		}
	}
}
