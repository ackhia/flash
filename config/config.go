package config

import (
	"bufio"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func ReadBootstrapPeers(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var peers []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			peers = append(peers, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return peers, nil
}

func ReadGenesis(filename string) (map[string]float64, error) {
	// Read the YAML file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the YAML file into a map
	var genMap map[string]float64
	err = yaml.Unmarshal(data, &genMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return genMap, nil
}
