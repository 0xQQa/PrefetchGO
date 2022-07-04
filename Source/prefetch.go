//https://github.com/EricZimmerman/Prefetch
package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const Win10OrWin11 = 30

type SysTimesPF struct {
	crTime int64
	laTime int64
	lwTime int64
}

type SysTimesAsTimePF struct {
	CrTime time.Time
	LaTime time.Time
	LwTime time.Time
}

type FileInfoPF struct {
	FullName    string
	ShortName   string
	HashName    string
	HashValue   string
	SizeInBytes int
}

type RunInfoPF struct {
	Times   uint32
	RunList []time.Time
}

type VolumeInfoPF struct {
	Name              string
	Serial            string
	Created           time.Time
	Letter            string
	DirectoriesAmount uint32
	Directories       []string
}

type CountedHash struct {
	Name  string
	Value string
}

type FilesPF struct {
	Name string
	Hash []CountedHash
}

type InfoPF struct {
	Times            SysTimesAsTimePF
	FileInfo         FileInfoPF
	RunInfo          RunInfoPF
	FilesAmount      int
	Files            []FilesPF
	VolumeInfoAmount uint32
	VolumeInfo       []VolumeInfoPF
}

/*
 * Funkcja zpozyskujaca date stworzenia, ostatnie dostepu
 * i zapisu z pliku o zdanej sciezce
 */

func getSysTimes(path string) (SysTimesPF, error) {
	statInfo, err := os.Stat(path)
	if err != nil {
		return SysTimesPF{}, err
	}

	fileInfo := statInfo.Sys().(*syscall.Win32FileAttributeData)
	sysTimes := SysTimesPF{
		crTime: fileInfo.CreationTime.Nanoseconds(),
		laTime: fileInfo.LastAccessTime.Nanoseconds(),
		lwTime: fileInfo.LastWriteTime.Nanoseconds(),
	}

	return sysTimes, nil
}

/*
 * Funkcja zwracajaca tablice wartosci binrnych, ktore
 * reprezentuja zawartosc pliku
 */

func getBytesFromPF(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	stats, err := file.Stat()
	if err != nil {
		return nil, err
	}

	size := stats.Size()
	bytes := make([]byte, size)
	bufr := bufio.NewReader(file)
	_, err = bufr.Read(bytes)
	if err != nil {
		return nil, err
	}

	err = file.Close()
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

/*
 * Funkcja wyknujaca dekompresje pliku jesli
 * posiada on znaczik kompresji
 */

func prepareBytesPF(rawData []byte) ([]byte, error) {
	patternWindows1x := []byte{'M', 'A', 'M'}
	foundWindows1x := bytes.Compare(rawData[:3], patternWindows1x)
	if foundWindows1x == 0 {
		decompressedSize := binary.LittleEndian.Uint32(rawData[4:])
		compressedBytes := rawData[8:]
		return DecompresionPF(compressedBytes, decompressedSize)
	}

	return rawData, nil
}

/*
 * Funkcja zwracaja rozmiar oraz hash
 * bajtow zawartych w badanym pliku
 */

func getHashAndSize(rawData []byte) (int, string) {
	hashRawData := md5.New()
	_, err := hashRawData.Write(rawData)
	if err != nil {
		return 0, "ERROR"
	}

	rawDataLen := len(rawData)
	strData := hex.EncodeToString(hashRawData.Sum(nil))
	return rawDataLen, strData
}

/*
 * Prosta funkcja rozlaczajaca typowe nazwy plikow PF na czlony
 * oddzielone znakie '-' bez rozszrzenia
 */

func parsePathForFilePF(path string) (string, string, string, error) {
	filename := filepath.Base(path)
	tmpSplittedFilename := strings.Split(filename, "-")
	tmpSplittedHash := strings.Split(tmpSplittedFilename[1], ".")
	if len(tmpSplittedFilename) != 2 && len(tmpSplittedHash) != 2 {
		return filename, "", "", errors.New("Invalid filename syntax")
	}

	filetype, hashPf := tmpSplittedFilename[0], tmpSplittedHash[0]
	return filename, filetype, hashPf, nil
}

/*
 * Funkcja sprawdzajaca czy srodowisko uruchomieniowe to
 * system z rodziny windows
 */

func canRunOnThisHost() {
	if runtime.GOOS != "windows" {
		panic("Lib support only Windows hosts")
	}
}

/*
 * Funkcja sprawdzajaca czy badany plik posida
 * sygnature plików Pf oraz jest przezaczony na windowsy 1x
 */

func parseBytesPF(rawData []byte) (bool, error) {
	signaturePF := []byte{'S', 'C', 'C', 'A'}
	foundSignaturePF := bytes.Compare(rawData[4:8], signaturePF)
	if foundSignaturePF != 0 {
		return false, errors.New("SCCA signature not found")
	}

	fileVersion := binary.LittleEndian.Uint32(rawData[:4])
	if fileVersion != Win10OrWin11 {
		panic("Only Windows 1x implemented")
	}

	return true, nil
}

/*
 * Funkcja uzupelniajaca strukture badawcza, odtwarza mozliwosci
 * progrmau prefetch,a wyniki skladuje jako json
 */

func fillInfoPFtoJSON(path string) (string, error) {
	var infoPF InfoPF

	timesNS, err := getSysTimes(path)
	if err != nil {
		return "", err
	}

	infoPF.Times = SysTimesAsTimePF{
		CrTime: time.Unix(0, timesNS.crTime),
		LaTime: time.Unix(0, timesNS.laTime),
		LwTime: time.Unix(0, timesNS.lwTime),
	}

	infoPF.FileInfo.FullName, infoPF.FileInfo.ShortName, infoPF.FileInfo.HashName, err = parsePathForFilePF(path)
	if err != nil {
		return "", err
	}

	rawData, _ := getBytesFromPF(path)
	rawData, _ = prepareBytesPF(rawData)
	infoPF.FileInfo.SizeInBytes, infoPF.FileInfo.HashValue = getHashAndSize(rawData)
	parseBytesPF(rawData)
	getPFInfo(rawData, &infoPF)

	jsonByte, err := json.MarshalIndent(infoPF, "", " ")
	if err != nil {
		return "", err
	}

	//testwo zapis do pliku
	lastInd := strings.LastIndex(infoPF.FileInfo.FullName, ".")
	pfJsonName := infoPF.FileInfo.FullName[:lastInd] + ".json"
	ioutil.WriteFile("tmp\\"+pfJsonName, jsonByte, os.ModePerm)

	return pfJsonName, nil
}

/*
 * Funkcja wejscia do programu
 */

func main() {
	canRunOnThisHost()
	parseConfig()
	getAllDisksHexSerials()
	handleCTRL_C()
	restoreListFromFile()
	startDetection()
	storeListToFile()
}
