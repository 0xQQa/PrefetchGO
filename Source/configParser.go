package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"hash"
	"io/ioutil"
	"os"
	"time"
)

type configFile struct {
	HashNames  []string
	SleepDelay int64
}

type configInfo struct {
	HashNames  []string
	HashTypes  []func() hash.Hash
	SleepDelay time.Duration
	HashCount  int
}

var config configInfo

const configPath = "..\\Configuration\\PrefetchConfiguration.json"

func createHashTypes(hashNames []string) []func() hash.Hash {
	var hashTypes []func() hash.Hash

	for _, name := range hashNames {
		switch name {
		case "sha256":
			hashTypes = append(hashTypes, sha256.New)
		case "sha1":
			hashTypes = append(hashTypes, sha1.New)
		case "md5":
			hashTypes = append(hashTypes, md5.New)
		default:
			panic("Unknow hash type choosen! [md5/sha1/sha256]")
		}
	}

	return hashTypes
}

func parseConfig() error {
	var configFromFile configFile

	configFile, err1 := os.OpenFile(configPath, os.O_RDWR, 0644)
	if err1 != nil {
		return err1
	}

	byteValue, err3 := ioutil.ReadAll(configFile)
	if err3 != nil {
		return err3
	}

	err3 = json.Unmarshal(byteValue, &configFromFile)
	if err3 != nil {
		return err3
	}

	err3 = configFile.Close()
	if err3 != nil {
		return err3
	}

	config.HashTypes = createHashTypes(configFromFile.HashNames)
	config.HashNames = configFromFile.HashNames
	config.HashCount = len(config.HashNames)
	config.SleepDelay = time.Duration(configFromFile.SleepDelay)

	return nil
}
