package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

const PrefetchDB = "PrefetchInfo.json"

var listPF []nodePF

type nodePF struct {
	Name   string
	LwTime int64
	Hash   string
}

/*
 * Funkcja zapisujaca ostatni stan badanych plikow do
 * tymczasowej bazy danych jako json
 */

func storeListToFile() error {
	jsonByte, err := json.MarshalIndent(&listPF, "", " ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(PrefetchDB, jsonByte, 0644)
	if err != nil {
		return err
	}

	log.Println("Storing PF info list into: " + PrefetchDB)
	return nil
}

/*
 * Funkcja odzyskujaca ostatni stan badanych plikow do
 * tymczasowej bazy danych jako json
 */

func restoreListFromFile() error {
	tmpDB, err1 := os.OpenFile(PrefetchDB, os.O_RDWR, 0644)
	if err1 != nil {
		log.Println("Creating new PF info list")
		return err1
	}

	byteValue, err3 := ioutil.ReadAll(tmpDB)
	if err3 != nil {
		return err3
	}

	err3 = json.Unmarshal(byteValue, &listPF)
	if err3 != nil {
		return err3
	}

	err3 = tmpDB.Close()
	if err3 != nil {
		return err3
	}

	log.Println("Using previous PF info list: " + PrefetchDB)
	return nil
}

/*
 * Funkcja sprawdzajaca obecnosc pliku PF we wlasnej ich liscie
 */

func namePFinList(namePF string) (*nodePF, int) {
	for index, filePF := range listPF {
		if filePF.Name == namePF {
			return &filePF, index
		}
	}

	return nil, -1
}

/*
 * Funkcja usuwajaca plik PF z wlasnej ich listy
 */

func removePFfromList(index int) {
	lenListPF := len(listPF) - 1
	listPF[index] = listPF[lenListPF]
	listPF = listPF[:lenListPF]
}

/*
 * Wrapper odowiadajacy za zworcenie z listy oraz usuneicie
 * z niej poszukiwanego elemntu
 */

func getOldPFfromList(namePF string) *nodePF {
	filePF, index := namePFinList(namePF)
	if filePF != nil {
		removePFfromList(index)
	}

	return filePF
}
