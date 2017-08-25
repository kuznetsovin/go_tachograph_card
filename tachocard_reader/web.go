package main

import (
	"archive/zip"
	"encoding/json"
	"github.com/skratchdot/open-golang/open"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type OperationRespState struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
}

var dddFile []byte
var stateRun chan bool
var TO int = 60

func unzipResFile(file *zip.File) error {
	path := file.Name
	if file.FileInfo().IsDir() {
		os.MkdirAll(path, file.Mode())
		return nil
	}

	fileReader, err := file.Open()
	if err != nil {
		return err
	}
	defer fileReader.Close()

	resourceFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer resourceFile.Close()

	if _, err := io.Copy(resourceFile, fileReader); err != nil {
		return err
	}
	return err
}
func extractResources() error {
	resources, err := zip.OpenReader("resources")
	if err != nil {
		return err
	}
	defer resources.Close()

	for _, file := range resources.File {
		if err := unzipResFile(file); err != nil {
			return err
		}
	}
	return err
}

func cleanResources() error {
	err := os.RemoveAll("gui")
	if err == nil {
		log.Println("Cleaning resources...")
	}
	return err
}

func runningChecker() {
	timer := TO
	for {
		select {
		case msg := <-stateRun:
			if msg {
				timer = TO
			}
		case <-time.After(time.Second):
			timer = timer - 1
			if timer == 0 {
				StopService()
			}
		}
	}
}

func checkPinHandler(w http.ResponseWriter, r *http.Request) {
	status := OperationRespState{
		Success: true,
		Msg:     "",
	}

	pin := r.FormValue("pin")
	indexReader, _ := strconv.Atoi(r.FormValue("indexReader"))

	if err := VerifyPIN(pin, indexReader); err != nil {
		log.Println("Validate pin error:", err)
		status.Success = false
		status.Msg = "Невозможно считать карту. Проверьте PIN или положение карты в ридере"
	}

	req, err := json.Marshal(status)
	if err != nil {
		log.Println("Json marshaling error at read card:", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(req)
}

func readCardHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	status := OperationRespState{
		Success: true,
		Msg:     "",
	}

	pin := r.FormValue("pin")
	indexReader, _ := strconv.Atoi(r.FormValue("indexReader"))

	if dddFile, err = ReadСard(pin, indexReader); err != nil {
		log.Println("Read card error:", err)
		status.Success = false
		status.Msg = "Ошибка при чтении карты."
	} else {
		log.Println("Card read success.")
	}

	req, err := json.Marshal(status)
	if err != nil {
		log.Println("Json marshaling error at read card:", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(req)
}

func localSaveHandler(w http.ResponseWriter, r *http.Request) {
	status := OperationRespState{
		Success: true,
		Msg:     "Файл сохранен в локальном хранилище.",
	}

	if err := SaveDdd(dddFile); err != nil {
		log.Println("Save card error:", err)
		status.Success = false
		status.Msg = "Ошибка сохранения файла."
	} else {
		log.Println("Card save success.")
	}

	req, err := json.Marshal(status)
	if err != nil {
		log.Println("Json marshaling error at read card:", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(req)
}

func queueSaveHandler(w http.ResponseWriter, r *http.Request) {
	status := OperationRespState{
		Success: true,
		Msg:     "Файл успешно отправлен.",
	}
	login := r.FormValue("login")
	password := r.FormValue("password")

	uploadFileName := createFileName(dddFile)
	if err := SendDddToQueue(login, password, &dddFile, uploadFileName); err != nil {
		log.Println("Load card error:", err)
		status.Success = false
		status.Msg = "Ошибка при отправке файла. Проверьте корретность логина и пароля."
	} else {
		log.Println("Card send success.")
	}

	req, err := json.Marshal(status)
	if err != nil {
		log.Println("Json marshaling error at read card:", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(req)
}

func uploadAllHandler(w http.ResponseWriter, r *http.Request) {
	status := OperationRespState{
		Success: true,
		Msg:     "Все файлы успешно отправлены",
	}
	login := r.FormValue("login")
	password := r.FormValue("password")

	if err := UploadFilesToQueue(login, password); err != nil {
		log.Println("Load card error:", err)
		status.Success = false
		status.Msg = "Ошибка при отправке файлов. Проверьте корретность логина и пароля."
	} else {
		log.Println("All card send success")
	}

	req, err := json.Marshal(status)
	if err != nil {
		log.Println("Json marshaling error at read card:", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(req)
}

func unblockHandler(w http.ResponseWriter, r *http.Request) {
	status := OperationRespState{
		Success: true,
		Msg:     "Разблокировка выполена успешно.",
	}
	puk := r.FormValue("puk")
	newPin := r.FormValue("newPin")
	indexReader, _ := strconv.Atoi(r.FormValue("indexReader"))

	if err := UnblockCardByPUK(puk, newPin, indexReader); err != nil {
		log.Println("Unblock card error:", err)
		status.Success = false
		status.Msg = "Ошибка при разблокировке."
	} else {
		log.Println("Card unblock success")
	}

	req, err := json.Marshal(status)
	if err != nil {
		log.Println("Json marshaling error at read card:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(req)
}

func checkerHandler(w http.ResponseWriter, r *http.Request) {
	stateRun <- true
}

func readerCheckerHandler(w http.ResponseWriter, r *http.Request) {
	status := OperationRespState{
		Success: true,
		Msg:     "Ожидаем карту",
	}

	if err := checkEnableReaders(); err != nil {
		log.Println("Find cardreader error:", err)
		status.Success = false
		status.Msg = "Не обнаружено ни одного ридера"
	} else {
		log.Println("Find reader success")
	}

	req, err := json.Marshal(status)
	if err != nil {
		log.Println("Json marshaling error at read card:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(req)

}

func waitCardHandler(w http.ResponseWriter, r *http.Request) {
	status := OperationRespState{
		Success: true,
		Msg:     "0",
	}

	indexReader, err := waitCard()
	if err != nil {
		log.Println("Card read error:", err)
		status.Success = false
		status.Msg = "Невозможно считать карту"

	}
	status.Msg = strconv.Itoa(indexReader)

	req, err := json.Marshal(status)
	if err != nil {
		log.Println("Json marshaling error at read card:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(req)

}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	mainPage, err := ioutil.ReadFile("gui/index.html")
	if err != nil {
		log.Fatal(err)
	}

	w.Write(mainPage)
}

func RunService() error {
	if err := extractResources(); err != nil {
		return err
	}
	log.Println("Resources unpacked...")

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChannel
		StopService()
	}()

	fs := http.FileServer(http.Dir("gui/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/readcard", readCardHandler)
	http.HandleFunc("/checkpin", checkPinHandler)
	http.HandleFunc("/localsavecard", localSaveHandler)
	http.HandleFunc("/queuesavecard", queueSaveHandler)
	http.HandleFunc("/uploadall", uploadAllHandler)
	http.HandleFunc("/unblockcard", unblockHandler)
	http.HandleFunc("/checker", checkerHandler)
	http.HandleFunc("/checkreader", readerCheckerHandler)
	http.HandleFunc("/waitcard", waitCardHandler)

	log.Println("Starting service...")
	l, err := net.Listen("tcp", "localhost:8081")
	if err != nil {
		return err
	}

	log.Println("Starting browser...")
	if err := open.Run("http://localhost:8081"); err != nil {
		return err
	}

	stateRun = make(chan bool)
	go runningChecker()

	if err := http.Serve(l, nil); err != nil {
		if err := cleanResources(); err != nil {
			return err
		}
		return err
	}

	return nil
}

func StopService() {
	if err := cleanResources(); err != nil {
		log.Fatal(err)
	}
	log.Println("Stopping service...")
	os.Exit(0)
}
