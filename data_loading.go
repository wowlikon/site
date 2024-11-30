package main

import (
	"encoding/json"
	"os"
)

type Question struct {
	Question string
	Choices  []string
}

type Repository struct {
	User string `json:"user"`
	Repo string `json:"repo"`
}

type Certificates map[string][]string

type MainPageData struct {
	Certificates Certificates
	Repos        []Repository
}

func loadRepositories(filename string) ([]Repository, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	var repositories []Repository
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&repositories)
	if err != nil {
		return nil, err
	}
	return repositories, nil
}

func loadCertificates(filename string) (Certificates, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var certificates Certificates
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&certificates)

	if err != nil {
		return nil, err
	}

	return certificates, nil
}
