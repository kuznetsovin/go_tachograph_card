package main

import (
	"log"
	"net/http"
	"encoding/base64"
	"flag"
	"os"
)

// обработчик парсинга
func parseDDDHandler(w http.ResponseWriter, r *http.Request) {
	// получаем строку base64 с ddd файлом
	b64_ddd := r.FormValue("ddd")
	ddd, err := base64.StdEncoding.DecodeString(b64_ddd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	card := &Card{}

	//разбираем пришедший ddd файл
	if err = card.ParseFromDDD(ddd); err != nil {
		log.Printf("DDD pasre error: %v", err)
	}

	ddd_json, err := card.ExportToJson()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(ddd_json)
}

func main() {
	// разбираем параметры запуска
	defaultLogFile := os.Args[0] + ".log"

	port := flag.String("port", ":8000", "service port")
	logfile := flag.String("log", defaultLogFile, "log file")
	flag.Parse()

	// настраиваем логгер
	f, err := os.OpenFile(*logfile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// запускаем сервис
	http.HandleFunc("/", parseDDDHandler)
	http.ListenAndServe(*port, nil)
}