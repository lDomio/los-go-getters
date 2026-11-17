package config

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

func IniciarConfiguracion[T any](filePath string) *T {
	var config *T
	configFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	return config
}

func ConfigurarLogger(nombre string) { // Habria que ver de usar slog
	logFile, err := os.OpenFile(nombre, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	//log.Println("Soy un log")
	log.SetOutput(mw)
}
