package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/gmallard/stompngo"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
    "errors"
)

func SendDddToQueue(login string, pwd string, ddd *[]byte, uploadFileName string) error {

	netConn, err := net.Dial("tcp", net.JoinHostPort(config.Server, config.Port))
	if err != nil {
		return err
	}

	rabbitHeader := stompngo.Headers{"login", login, "passcode", pwd, "accept-version", "1.0", "host", "/"}
	rabbitConn, err := stompngo.Connect(netConn, rabbitHeader)
	if err != nil {
		return err
	}
	
	msgId := stompngo.Uuid()
	msgHeader := stompngo.Headers{
		"destination", config.Queue,
		"login", login,
		"hash", LoginHash(login),
		"filename", uploadFileName,
		"receipt", msgId,
		"persistent", "true",
	}
	err = rabbitConn.SendBytes(msgHeader, *ddd)
	if err != nil {
		return err
	}
	
	r := <-rabbitConn.MessageData
	receiptId := r.Message.Headers.Value("receipt-id")
	if msgId != receiptId {
	    return errors.New("Receipt and message id not equals.")
	}


	err = netConn.Close()
	if err != nil {
		return err
	}

	return nil
}

func UploadFilesToQueue(login string, pwd string) error {
	err := filepath.Walk(config.UploadDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			currentDddFiles, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			_, uploadFileName := filepath.Split(path)
			err = SendDddToQueue(login, pwd, &currentDddFiles, uploadFileName)
			if err != nil {
				return err
			}
			os.Remove(path)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func LoginHash(login string) string {
	/*
		Функция для вычисления хэша.
	*/

	h := sha1.New()
	salt := "tahogram"
	io.WriteString(h, login)
	io.WriteString(h, salt)
	return fmt.Sprintf("%x", h.Sum(nil))
}
