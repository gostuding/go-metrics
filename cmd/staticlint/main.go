package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"

	"github.com/gostuding/go-metrics/staticlint/errcheckanalyzer"

	"honnef.co/go/tools/staticcheck"
)

// Config — имя файла конфигурации.
const Config = `config.json`

// ConfigData описывает структуру файла конфигурации.
type ConfigData struct {
	Staticcheck []string
}

func main() {
	mychecksStaticChecks, err := readConfigFile()
	if err != nil {
		log.Printf("file 'config.json' read error: %v\n", err)
	}
	// определяем map подключаемых правил
	checks := map[string]bool{
		"SA5000": true,
		"SA6000": true,
		"SA9004": true,
	}
	for _, v := range staticcheck.Analyzers {
		if checks[v.Analyzer.Name] {
			mychecksStaticChecks = append(mychecksStaticChecks, v.Analyzer)
		}
	}
	mychecksStaticChecks = append(mychecksStaticChecks, errcheckanalyzer.ErrCheckAnalyzer)
	mychecksStaticChecks = append(mychecksStaticChecks, printf.Analyzer)
	mychecksStaticChecks = append(mychecksStaticChecks, shadow.Analyzer)
	mychecksStaticChecks = append(mychecksStaticChecks, structtag.Analyzer)
	mychecksStaticChecks = append(mychecksStaticChecks, shift.Analyzer)

	multichecker.Main(
		mychecksStaticChecks...,
	)
}

func readConfigFile() ([]*analysis.Analyzer, error) {
	configChecks := make([]*analysis.Analyzer, 0)

	appfile, err := os.Executable()
	if err != nil {
		return configChecks, fmt.Errorf("get file name error: %w", err)
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), Config))
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("config.json is not exist")
			return configChecks, nil
		}
		return configChecks, fmt.Errorf("read config file error: %w", err)
	}
	var cfg ConfigData
	if err = json.Unmarshal(data, &cfg); err != nil {
		return configChecks, fmt.Errorf("config file json incorrect: %w", err)
	}

	checks := make(map[string]bool)
	for _, v := range cfg.Staticcheck {
		checks[v] = true
	}
	// добавляем анализаторы из staticcheck, которые указаны в файле конфигурации
	for _, v := range staticcheck.Analyzers {
		if checks[v.Analyzer.Name] {
			configChecks = append(configChecks, v.Analyzer)
		}
	}
	return configChecks, nil
}
