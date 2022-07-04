package main

import (
	"crypto/md5"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const prefetchPath = "C:\\Windows\\Prefetch\\"

var isWorking bool

/*
 * Funkcja zakanczajaca dzialanie programu
 */

func finishDetection() {
	isWorking = false
}

/*
 * Funkcja przeznaczona dla sygnalu ctrl+c w celu zakoczenia dzialania programu
 */

func handleCTRL_C() {
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel, os.Interrupt, os.Kill, syscall.SIGTERM)

	go func() {
		<-signalChanel
		signal.Stop(signalChanel)
		finishDetection()
	}()
}

/*
 * Funkcja obliczajaca hash md5 na potrzeby detekcji plików pf w ich liscie
 */

func getInfoPFFileHash(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return "ERROR"
	}

	defer f.Close()
	return getFileHash(f, md5.New)
}

/*
 * Funkcja aktualizujaca wewnetrzna liste plików pf o te
 * istniejace w lokalizacji prefetch, pozbywa sie rowniez
 * tych ktore zostaly usuniete
 */

func checkPFinPathState(files []fs.FileInfo) error {
	var newListPF []nodePF

	for _, file := range files {
		fileExtension := filepath.Ext(file.Name())
		if fileExtension != ".pf" {
			continue
		}

		newFound := getOldPFfromList(file.Name())
		if newFound != nil {
			newListPF = append(newListPF, *newFound)
		} else {
			log.Println("Found new PF file: " + file.Name())
			fullPathPF := prefetchPath + file.Name()
			fileTimes, _ := getSysTimes(fullPathPF)
			hash := getInfoPFFileHash(fullPathPF)
			newNodePF := nodePF{Name: file.Name(), LwTime: fileTimes.lwTime, Hash: hash}
			newListPF = append(newListPF, newNodePF)
			jsonName, err := fillInfoPFtoJSON(fullPathPF)
			if err == nil {
				log.Println("Parsed: " + file.Name() + " as " + jsonName)
			} else {
				log.Println("Got error while parsing: " + file.Name())
			}
		}
	}

	listPF = newListPF
	return nil
}

/*
 * Funkcja sprawdzajaca zmiany w pliku pierwotnie poprzez zmiane lwt
 * nastepnie przez hash pliku, w przypadku wykrycia roznica aktualizuje
 * stworozny plik json
 */

func checkChangeInFilesPF(files []fs.FileInfo) {
	for index, file := range listPF {
		fileTimes, _ := getSysTimes(prefetchPath + file.Name)
		if fileTimes.lwTime != listPF[index].LwTime {
			listPF[index].LwTime = fileTimes.lwTime
			fullPathPF := prefetchPath + file.Name
			hash := getInfoPFFileHash(fullPathPF)

			if hash == "ERROR" && file.Hash != "ERROR" {
				log.Println("Got error while updating hash for: " + file.Name)
				continue
			}

			if hash != file.Hash {
				jsonName, err := fillInfoPFtoJSON(fullPathPF)
				if err == nil {
					listPF[index].Hash = hash
					log.Println("Upadted PF file: " + file.Name + " as " + jsonName)
				} else {
					log.Println("Got error while updating json for: " + file.Name)
				}
			} else {
				log.Println("Upadte last write time for PF file: " + file.Name)
			}
		}
	}
}

/*
 * Glwona funkcja programu, ktora odpowiada za cykliczne sprawdzanie
 * zawartosci folderu prefetch
 */

func startDetection() (bool, error) {
	isWorking = true

	for isWorking {
		files, err := ioutil.ReadDir(prefetchPath)
		if err != nil {
			return false, err
		}

		checkPFinPathState(files)
		checkChangeInFilesPF(files)
		time.Sleep(config.SleepDelay * time.Millisecond)
	}

	return true, nil
}
