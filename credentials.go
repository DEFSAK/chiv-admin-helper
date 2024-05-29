package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	confNamespace = "chiv-admin-helper"
)

func setupCredentials() (credentialPath string, err error) {
	confDir, err := os.UserConfigDir()
	if err != nil {
		err = fmt.Errorf("could not determine user config dir: %w", err)
	}
	confDir = filepath.Join(confDir, confNamespace)
	err = os.Mkdir(confDir, 0700)
	if errors.Is(err, os.ErrExist) {
		// Config dir exists and might contain credentials
		var files []fs.DirEntry
		files, err = os.ReadDir(confDir)
		if err != nil {
			err = fmt.Errorf("could not read config dir: %w", err)
			return
		}
		for _, file := range files {
			if filepath.Ext(file.Name()) != ".json" {
				continue
			}
			credentialPath = filepath.Join(confDir, file.Name())
			return credentialPath, nil
		}
		err = errors.New("could not find json credential file")
		return
	} else if err == nil {
		// Config dir did not exist yet but was created
		// Ask user for their credential file
		fmt.Print("Credentials are required to use this tool. Enter the path to your credentials file and press enter: ")
		inputPath, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		inputPath = strings.Trim(inputPath, "\r\n\" ")
		inputPath = filepath.Clean(inputPath)
		// Copy credentials to a safe location
		var originalCredentials, savedCredentials *os.File
		originalCredentials, err = os.Open(inputPath)
		if err != nil {
			err = fmt.Errorf("could not open credentials file: %w", err)
			return
		}
		defer originalCredentials.Close()
		credentialPath = filepath.Join(confDir, filepath.Base(inputPath))
		savedCredentials, err = os.OpenFile(credentialPath, os.O_CREATE, 0600)
		if err != nil {
			err = fmt.Errorf("could not save credentials to config: %w", err)
		}
		defer savedCredentials.Close()
		_, _ = io.Copy(savedCredentials, originalCredentials)
		// Return the new credential file
		return
	}
	// Mkdir failed
	err = fmt.Errorf("failed to setup user config dir: %w", err)
	return
}
