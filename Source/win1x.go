package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const SecToUnixEpoch = 11644473600000
const WindowsTick = 10000000

/*
 *	Funkcja zwracajaca czas w fromacie Unix'a
 */

func getAsDateTime(windowsTicks int64) time.Time {
	return time.Unix((windowsTicks-SecToUnixEpoch*10000)/WindowsTick, 0)
}

/*
 *	Wrapper do pozyskania hash'u z zadnaego po sciezce pliku
 */

func getFileHash(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return "ERROR"
	}

	defer f.Close()
	h := choosedHash.Type()
	_, err = io.Copy(h, f)
	if err != nil {
		return "ERROR"
	}

	return hex.EncodeToString(h.Sum(nil))
}

/*
 *	Funkcja obcinajaca pierwszy folder w siezce
 */

func parseBranchname(branchString string) string {
	branchArray := strings.SplitN(branchString, "\\", 3)
	branchname := branchArray[2]
	return branchname
}

/*
 *	Funkcja odzyskujaca prawdziwa sciezke na bazie numeru
 *	seryjnego dysku i wczesniej obliczonych wartosci
 */

func resolvePathPF(volumeInfo []VolumeInfoPF, pathPF string) string {
	for _, volume := range volumeInfo {
		if strings.HasPrefix(pathPF, volume.Name) {
			for _, disk := range allDisks {
				if disk.serial == volume.Serial {
					chunkPath := parseBranchname(pathPF)
					return disk.letter + ":\\" + chunkPath
				}
			}
		}
	}

	return ""
}

/*
 *	Funkcja uzupelniaja strukture infoPF o informacje
 *	prefetcha, zgodne z proejktem oryginalnym
 */

func getPFInfo(rawData []byte, infoPF *InfoPF) {
	var lastRunTimes []time.Time
	var runCount uint32
	var filenames []FilesPF
	var dirStringsList []string
	var tmpString strings.Builder

	fileInfoBytes := rawData[84 : 84+224]
	filenameStringsOffset := binary.LittleEndian.Uint32(fileInfoBytes[16:20])
	filenameStringsSize := binary.LittleEndian.Uint32(fileInfoBytes[20:24])
	volumesInfoOffset := binary.LittleEndian.Uint32(fileInfoBytes[24:28])
	volumeCount := binary.LittleEndian.Uint32(fileInfoBytes[28:32])

	runtimeBytes := fileInfoBytes[44 : 44+64]
	runCountPre := binary.LittleEndian.Uint32(fileInfoBytes[120:124])
	if runCountPre == 0 {
		runCount = binary.LittleEndian.Uint32(fileInfoBytes[124:128])
	} else {
		runCount = binary.LittleEndian.Uint32(fileInfoBytes[116:120])
	}

	for index := 0; index < 8; index++ {
		tmpRawData := int64(binary.LittleEndian.Uint64(runtimeBytes[index*8 : index*8+8]))
		if tmpRawData == 0 {
			break
		}

		datetimeData := getAsDateTime(tmpRawData)
		lastRunTimes = append(lastRunTimes, datetimeData)
	}

	infoPF.UsedHash = choosedHash.Name
	infoPF.VolumeInfoAmount = volumeCount
	infoPF.RunInfo.Times = runCount
	infoPF.RunInfo.RunList = lastRunTimes

	for index := 0; index < int(volumeCount); index += 1 {
		volumeInfoBytes := rawData[volumesInfoOffset+uint32(index)*96 : volumesInfoOffset+uint32(index)*96+96]
		volDevOffset := binary.LittleEndian.Uint32(volumeInfoBytes[0:4])
		volDevNumChar := binary.LittleEndian.Uint32(volumeInfoBytes[4:8])
		ct := binary.LittleEndian.Uint64(volumeInfoBytes[8:16])
		devNameBytes := rawData[volumesInfoOffset+volDevOffset : volumesInfoOffset+volDevOffset+volDevNumChar*2]
		tmpString.Reset()
		for index := 0; index < int(volDevNumChar*2); index += 2 {
			tmpString.WriteByte(devNameBytes[index])
		}

		devName := tmpString.String()
		sn := fmt.Sprintf("%X", binary.LittleEndian.Uint32(volumeInfoBytes[16:20]))
		dirStringsOffset := binary.LittleEndian.Uint32(volumeInfoBytes[28:32])
		numDirectoryStrings := int(binary.LittleEndian.Uint32(volumeInfoBytes[32:36]))
		dirStringsIndex := volumesInfoOffset + dirStringsOffset + 2
		dirStringsBytes := rawData[dirStringsIndex:]

		tmpString.Reset()
		for index := 0; index < len(dirStringsBytes); index += 2 {
			if dirStringsBytes[index] == 0 {
				dirStringsList = append(dirStringsList, tmpString.String())
				if len(dirStringsList) == numDirectoryStrings {
					break
				}

				tmpString.Reset()
				index += 2
			} else {
				tmpString.WriteByte(dirStringsBytes[index])
			}
		}

		volumeInfoPF := VolumeInfoPF{Name: devName, Created: getAsDateTime(int64(ct)), Serial: sn, Directories: dirStringsList, DirectoriesAmount: uint32(numDirectoryStrings)}
		infoPF.VolumeInfo = append(infoPF.VolumeInfo, volumeInfoPF)
	}

	tmpString.Reset()
	filenameStringsBytes := rawData[filenameStringsOffset : filenameStringsOffset+filenameStringsSize]
	for index := 0; index < int(filenameStringsSize); index += 2 {
		if filenameStringsBytes[index] == 0 {
			realPatfhPf := resolvePathPF(infoPF.VolumeInfo, tmpString.String())
			filenames = append(filenames, FilesPF{Name: tmpString.String(), Hash: getFileHash(realPatfhPf)})
			tmpString.Reset()
		} else {
			tmpString.WriteByte(filenameStringsBytes[index])
		}
	}

	infoPF.Files = filenames
	infoPF.FilesAmount = len(filenames)
}
