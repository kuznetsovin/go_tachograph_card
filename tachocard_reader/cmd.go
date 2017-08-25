package main

import (
	"fmt"
	"github.com/howeyc/gopass"
)

func hideInput(param *string) error {
	field, err := gopass.GetPasswd()
	if err != nil {
		return err
	}

	*param = string(field)
	return err
}

func UploadAll() error {
	var login, pwd string

	fmt.Print("Username: ")
	fmt.Scanln(&login)

	fmt.Print("Password: ")
	if err := hideInput(&pwd); err != nil {
		return err
	}

	if err := UploadFilesToQueue(login, pwd); err != nil {
		return err
	}

	return nil
}

func SaveLocal() error {
	var pin string

	if err := checkEnableReaders(); err != nil {
		return err
	}

	indexReader, err := waitCard()
	if err != nil {
		return err
	}

	fmt.Print("PIN: ")
	if err := hideInput(&pin); err != nil {
		return err
	}

	dddFile, err := ReadСard(pin, indexReader)
	if err != nil {
		return err
	}

	if err := SaveDdd(dddFile); err != nil {
		return err
	}
	return err
}

func SaveMQ() error {
	var pin, login, pwd string

	if err := checkEnableReaders(); err != nil {
		return err
	}

	indexReader, err := waitCard()
	if err != nil {
		return err
	}

	fmt.Print("PIN: ")
	if err := hideInput(&pin); err != nil {
		return err
	}

	dddFile, err := ReadСard(pin, indexReader)
	if err != nil {
		return err
	}

	fmt.Print("Username: ")
	fmt.Scanln(&login)

	fmt.Print("Password: ")
	if err := hideInput(&pwd); err != nil {
		return err
	}

	uploadFileName := createFileName(dddFile)
	if err := SendDddToQueue(login, pwd, &dddFile, uploadFileName); err != nil {
		return err
	}
	return err
}

func UnblockCard() error {
	var puk, newPin string

	if err := checkEnableReaders(); err != nil {
		return err
	}

	indexReader, err := waitCard()
	if err != nil {
		return err
	}

	fmt.Print("PUK: ")
	if err := hideInput(&puk); err != nil {
		return err
	}

	fmt.Print("New PIN: ")
	if err := hideInput(&newPin); err != nil {
		return err
	}

	if err := UnblockCardByPUK(puk, newPin, indexReader); err != nil {
		return err
	}

	return nil
}

func checkPIN(pin string) error {
	if err := checkEnableReaders(); err != nil {
		return err
	}

	indexReader, err := waitCard()
	if err != nil {
		return err
	}

	if err := VerifyPIN(pin, indexReader); err != nil {
		return err
	}

	return nil
}
