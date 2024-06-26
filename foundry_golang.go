package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type FoundryArtifact struct {
	Abi      []interface{}   `json:"abi"`
	Bytecode FoundryByteCode `json:"bytecode"`
	Metadata FoundryMetadata `json:"metadata"`
}

type FoundryByteCode struct {
	Object string `json:"object"`
}

type FoundryMetadata struct {
	Settings FoundrySetting `json:"settings"`
}

type FoundrySetting struct {
	CompilationTarget map[string]string `json:"compilationTarget"`
}

type moduleInfo struct {
	contractNames []string
	abis          []string
	bytecodes     []string
}

func (m *moduleInfo) addArtifact(artifact FoundryArtifact) {
	abi, err := json.Marshal(artifact.Abi)
	if err != nil {
		log.Fatal(err)
	}
	for _, contractName := range artifact.Metadata.Settings.CompilationTarget {
		m.contractNames = append(m.contractNames, contractName)
	}
	m.abis = append(m.abis, string(abi))
	m.bytecodes = append(m.bytecodes, artifact.Bytecode.Object)
}

func (m *moduleInfo) exportABIs(dest string) {
	for i, name := range m.contractNames {
		path := filepath.Join(dest, name+".abi")
		abi := m.abis[i] + "\n"

		// #nosec G306
		err := os.WriteFile(path, []byte(abi), 0o644)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("bad path")
	}
	root := filepath.Dir(filename)
	parent := filepath.Dir(root)
	filePaths, err := filepath.Glob(filepath.Join(parent, "contracts", "out", "*", "*.json"))
	if err != nil {
		log.Fatal(err)
	}

	modules := make(map[string]*moduleInfo)

	for _, path := range filePaths {
		if strings.Contains(path, ".dbg.json") {
			continue
		}

		dir, file := filepath.Split(path)
		dir, _ = filepath.Split(dir[:len(dir)-1])
		_, module := filepath.Split(dir[:len(dir)-1])
		module = strings.ReplaceAll(module, "-", "_")
		module += "gen"

		name := file[:len(file)-5]

		//#nosec G304
		data, err := os.ReadFile(path)
		if err != nil {
			log.Fatal("could not read", path, "for contract", name, err)
		}

		artifact := FoundryArtifact{}
		if err := json.Unmarshal(data, &artifact); err != nil {
			log.Fatal("failed to parse contract", name, err)
		}
		modInfo := modules[module]
		if modInfo == nil {
			modInfo = &moduleInfo{}
			modules[module] = modInfo
		}
		modInfo.addArtifact(artifact)
	}

	for module := range modules {
		fmt.Println(module)
	}

	for module, info := range modules {

		code, err := bind.Bind(
			info.contractNames,
			info.abis,
			info.bytecodes,
			nil,
			module,
			bind.LangGo,
			nil,
			nil,
		)
		if err != nil {
			log.Fatal(err)
		}

		folder := filepath.Join(root, "go", module)

		//#nosec G301
		err = os.MkdirAll(folder, 0o755)
		if err != nil {
			log.Fatal(err)
		}

		/*
			#nosec G306
			This file contains no private information so the permissions can be lenient
		*/
		err = os.WriteFile(filepath.Join(folder, module+".go"), []byte(code), 0o644)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("successfully generated go abi files")

	blockscout := filepath.Join(parent, "blockscout", "init", "data")
	if _, err := os.Stat(blockscout); err != nil {
		fmt.Println("skipping abi export since blockscout is not present")
	} else {
		modules["precompilesgen"].exportABIs(blockscout)
		modules["node_interfacegen"].exportABIs(blockscout)
		fmt.Println("successfully exported abi files")
	}
}
