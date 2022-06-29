package main

import (
	"syscall"
	"unsafe"
)

var (
	CompressionFormatXpressHuff    = 4
	ntdllDLL                       = syscall.NewLazyDLL("ntdll.dll")
	RtlDecompressBufferEx          = ntdllDLL.NewProc("RtlDecompressBufferEx")
	RtlGetCompressionWorkSpaceSize = ntdllDLL.NewProc("RtlGetCompressionWorkSpaceSize")
)

/*
 * Funkcja wywolujaca dekompresje lzx na pliku PF dla windows 1x
 */

func DecompresionPF(rawData []byte, decompressedSize uint32) ([]byte, error) {
	var compressBufferWorkSpaceSize, compressFragmentWorkSpaceSize uint64

	ret, _, err := RtlGetCompressionWorkSpaceSize.Call(uintptr(CompressionFormatXpressHuff), uintptr(unsafe.Pointer(&compressBufferWorkSpaceSize)), uintptr(unsafe.Pointer(&compressFragmentWorkSpaceSize)))
	if ret != 0 {
		return nil, err
	}

	outBuffer := make([]byte, decompressedSize)
	workSpace := make([]byte, compressFragmentWorkSpaceSize)
	destinationSize := 0
	ret, _, err = RtlDecompressBufferEx.Call(uintptr(CompressionFormatXpressHuff), uintptr(unsafe.Pointer(&outBuffer[0])), uintptr(decompressedSize), uintptr(unsafe.Pointer(&rawData[0])), uintptr(len(rawData)), uintptr(unsafe.Pointer(&destinationSize)), uintptr(unsafe.Pointer(&workSpace[0])))
	if ret != 0 {
		return nil, err
	}

	return outBuffer, nil
}
