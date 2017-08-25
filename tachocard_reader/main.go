package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var config = Settings{}

func die(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func main() {
	if err := config.init(); err != nil {
		die(err)
	}

	logFile, err := os.OpenFile(config.Log, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		die(err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	fist_arg := ""
	second_arg := ""
	msg := ""

	switch len(os.Args) {
	case 1:
		msg = main_help()
		message_print(msg)
	case 2:
		fist_arg = os.Args[1]
	case 3:
		fist_arg = os.Args[1]
		second_arg = os.Args[2]
	}

	switch fist_arg {
	case "save":
		if second_arg == "help" {
			msg = save_help()
		} else {
			if strings.HasPrefix(second_arg, "mq") {
				if err := SaveMQ(); err != nil {
					die(err)
				}
			} else {
				if err := SaveLocal(); err != nil {
					die(err)
				}
			}
			msg = "Карта успешно сохранена"
		}
	case "uploadall":
		if second_arg == "help" {
			msg = uploadall_help()
		} else {
			if err := UploadAll(); err != nil {
				die(err)
			}
		}
		msg = "Данные успешно отправлены"
	case "serve":
		if second_arg == "help" {
			msg = serve_help()
		} else {
			if err := RunService(); err != nil {
				die(err)
			}
		}
	case "verify":
		if second_arg == "help" || second_arg == "" {
			msg = verify_help()
		} else {
			if err := checkPIN(second_arg); err != nil {
				die(err)
			}
			msg = "OK"
		}
	case "unblock":
		if second_arg == "help" {
			msg = unblock_help()
		} else {
			if err := UnblockCard(); err != nil {
				die(err)
			}
			msg = "Карта разблокирована."
		}
	case "help":
		msg = main_help()
	default:
		msg = "Не корректная операция наберите help для справки."
	}

	message_print(msg)
}

func message_print(text string) {
	fmt.Printf("\n%s\n", text)
	os.Exit(0)
}

func main_help() string {
	return `tcreader <операция> [параметры]

Список доступных операций:
       save - записывает данные карты на диск или на сервер
       uploadall - команда выполняет отправку всех файлов лежащих в локльной папке для выгрузки
       serve - запускает сервис чтения карт на локальной машине
       verify - выполняет проверку PIN
       unblock - выполняет разблокировку карты вводом PUK или с сервера
       help - выводит данную справку

Дополнительную информацю по операции можно получить

tcreader <операция> help

например

tcreader save help
`
}

func save_help() string {
	return `tcreader save [URL]
Комадна сохраняет данные карты на сервер или локально. При этом для чтения карты нужен PIN.

Параметры:
    URL - путь для сохранения данных (поддерживаются протоколы stomp и file)

например

tcreader save
PIN:

tcreader save mq
PIN:
username:
password:

в случае сохранения в mq интерактивно спрашивается username: и password:
`
}

func uploadall_help() string {
	return `tcreader uploadall
Команда производит загрузку всех хранящихся в локалной папке ddd файлов на сервер.

tcreader uploadall
username:
password:

в при сохрании интерактивно спрашивается username: и password:
`
}

func serve_help() string {
	return `tcreader serve
Запускает локальный web сервис, для работы через браузер.
`
}

func verify_help() string {
	return `tcreader verify <PIN>
Команда проверят PIN код карты

Параметры:
    PIN - PIN код карты
`
}

func unblock_help() string {
	return `tcreader unblock ?[URL]
Команда производит разблокировку карты вводом PUK кода.

например

tcreader unblock
PUK:
New PIN:
`
}
