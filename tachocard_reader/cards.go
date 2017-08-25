package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var cardMF = []CardFile{
	CardFile{"EF_ICC", [2]byte{0x00, 0x02}, 25, 25, false},
	CardFile{"EF_IC", [2]byte{0x00, 0x05}, 8, 8, false},
}

var cardGDriver = []CardFile{
	CardFile{"EF_Application_Identification", [2]byte{0x05, 0x01}, 10, 10, true},
	CardFile{"EF_Card_Certificate_GOST", [2]byte{0xC2, 0x00}, 1000, 1000, false},
	CardFile{"EF_Key_Identificators", [2]byte{0xC2, 0x01}, 16, 16, true},
	CardFile{"EF_CA_Certificate_GOST", [2]byte{0xC2, 0x08}, 1000, 1000, false},
	CardFile{"EF_Identification", [2]byte{0x05, 0x20}, 143, 143, true},
	CardFile{"EF_Card_Download", [2]byte{0x05, 0x0E}, 4, 4, true},
	CardFile{"EF_Driver_Licence_Info", [2]byte{0x05, 0x21}, 53, 53, true},
	CardFile{"EF_Events_Data", [2]byte{0x05, 0x02}, 864, 1728, true},
	CardFile{"EF_Faults_Data", [2]byte{0x05, 0x03}, 576, 1152, true},
	CardFile{"EF_Driver_Activity_Data", [2]byte{0x05, 0x04}, 5548, 13780, true},
	CardFile{"EF_Vehicles_Used", [2]byte{0x05, 0x05}, 2606, 6202, true},
	CardFile{"EF_Places", [2]byte{0x05, 0x06}, 841, 1121, true},
	CardFile{"EF_Current_Usage", [2]byte{0x05, 0x07}, 19, 19, true},
	CardFile{"EF_Control_Activity_Data", [2]byte{0x05, 0x08}, 46, 46, true},
	CardFile{"EF_Specific_Conditions", [2]byte{0x05, 0x22}, 280, 280, true},
}

var cardGWorkshop = []CardFile{
	CardFile{"EF_Application_Identification", [2]byte{0x05, 0x01}, 11, 11, true},
	CardFile{"EF_Card_Certificate_GOST", [2]byte{0xC2, 0x00}, 1000, 1000, false},
	CardFile{"EF_Key_Identificators", [2]byte{0xC2, 0x01}, 16, 16, true},
	CardFile{"EF_Temporary_Cert", [2]byte{0xC2, 0x03}, 1024, 1024, true},
	CardFile{"EF_Cert_Request", [2]byte{0xC2, 0x04}, 1024, 1024, true},
	CardFile{"EF_VU_Cert", [2]byte{0xC2, 0x05}, 2048, 2048, true},
	CardFile{"EF_Archive_Request", [2]byte{0xC2, 0x06}, 14, 14, true},
	CardFile{"EF_VU_Cert_Request", [2]byte{0xC2, 0x07}, 1024, 1024, true},
	CardFile{"EF_CA_Certificate_GOST", [2]byte{0xC2, 0x08}, 1000, 1000, false},
	CardFile{"EF_Identification", [2]byte{0x05, 0x20}, 211, 211, true},
	CardFile{"EF_Card_Download", [2]byte{0x05, 0x09}, 2, 2, true},
	CardFile{"EF_Calibration", [2]byte{0x05, 0x0A}, 9243, 26778, true},
	CardFile{"EF_Sensor_Installation_Data", [2]byte{0x05, 0x0B}, 16, 16, true},
	CardFile{"EF_Events_Data", [2]byte{0x05, 0x02}, 432, 432, true},
	CardFile{"EF_Faults_Data", [2]byte{0x05, 0x03}, 288, 288, true},
	CardFile{"EF_Driver_Activity_Data", [2]byte{0x05, 0x04}, 202, 496, true},
	CardFile{"EF_Vehicles_Used", [2]byte{0x05, 0x05}, 126, 250, true},
	CardFile{"EF_Places", [2]byte{0x05, 0x06}, 61, 81, true},
	CardFile{"EF_Current_Usage", [2]byte{0x05, 0x07}, 19, 19, true},
	CardFile{"EF_Control_Activity_Data", [2]byte{0x05, 0x08}, 46, 46, true},
	CardFile{"EF_Specific_Conditions", [2]byte{0x05, 0x22}, 10, 10, true},
}

var cardGControl = []CardFile{
	CardFile{"EF_Application_Identification", [2]byte{0x05, 0x01}, 5, 5, true},
	CardFile{"EF_Card_Certificate_GOST", [2]byte{0xC2, 0x00}, 1000, 1000, false},
	CardFile{"EF_Key_Identificators", [2]byte{0xC2, 0x01}, 16, 16, true},
	CardFile{"EF_CA_Certificate_GOST", [2]byte{0xC2, 0x08}, 1000, 1000, false},
	CardFile{"EF_Identification", [2]byte{0x05, 0x20}, 211, 211, true},
	CardFile{"EF_Controllor_Activity_Data", [2]byte{0x05, 0x0C}, 10582, 23922, true},
	CardFile{"EF_Archive_Requests", [2]byte{0xC2, 0x06}, 14, 14, true},
}

var cardGCompany = []CardFile{
	CardFile{"EF_Application_Identification", [2]byte{0x05, 0x01}, 5, 5, true},
	CardFile{"EF_Card_Certificate_GOST", [2]byte{0xC2, 0x00}, 1000, 1000, false},
	CardFile{"EF_Key_Identificators", [2]byte{0xC2, 0x01}, 16, 16, true},
	CardFile{"EF_CA_Certificate_GOST", [2]byte{0xC2, 0x08}, 1000, 1000, false},
	CardFile{"EF_Identification", [2]byte{0x05, 0x20}, 211, 211, true},
	CardFile{"EF_Company_Activity_Data", [2]byte{0x05, 0x0C}, 10582, 23922, true},
	CardFile{"EF_Archive_Requests", [2]byte{0xC2, 0x06}, 14, 14, true},
}

var typeCard = map[byte][]CardFile{
	0x01: cardGDriver,
	0x02: cardGWorkshop,
	0x03: cardGControl,
	0x04: cardGCompany,
}

func ReadСard(pin string, indexReader int) ([]byte, error) {
	var result_buf bytes.Buffer

	ctx, d, err := CardConnect(indexReader)

	if err != nil {
		return nil, err
	}

	if _, err = verify(pin, d); err != nil {
		ctx.Release()
		return nil, err
	}

	for _, fileSign := range cardMF {
		rf, err := readFile(&fileSign, d)
		if err != nil {
			ctx.Release()
			return nil, err
		}

		result_buf.Write(rf)
		fmt.Printf("Файл %s успешно прочитан.\n", fileSign.Name)
	}

	_, err = selectFile([]byte{0xFF, 0x54, 0x41, 0x43, 0x48, 0x4F}, d)
	if err != nil {
		ctx.Release()
		return nil, err
	}

	cardType, err := initCardType(d, &countFieldRec)
	if err != nil {
		return nil, err
	}

	for _, fileSign := range typeCard[cardType] {
		if fileSign.Name == "EF_Card_Download" {
			if err := UpdateUploadDate(fileSign.Tag, d); err != nil {
				return nil, err
			}
		}
		rf, err := readFile(&fileSign, d)
		if err != nil {
			ctx.Release()
			return nil, err
		}

		result_buf.Write(rf)
		fmt.Printf("Файл %s успешно прочитан.\n", fileSign.Name)
	}

	result := result_buf.Bytes()

	ctx.Release()
	return result, err
}

func createFileName(ddd []byte) string {
	var cardView, result string
	var GOST_CERT_SECTION = []byte{0xc2, 0x00}

	cardName := ddd[2233:2249]
	uploadFileDateTime := time.Now()

	if bytes.Equal(ddd[127:129], GOST_CERT_SECTION) {
		cardView = "RFSKZI"
	} else {
		cardView = "ESTR"
	}

	result = fmt.Sprintf("%s-%s-%d%02d%02d_%02d%02d.ddd", cardView, cardName,
		uploadFileDateTime.Year(), uploadFileDateTime.Month(),
		uploadFileDateTime.Day(), uploadFileDateTime.Hour(),
		uploadFileDateTime.Minute())

	return result
}

func SaveDdd(card_ddd []byte) error {
	uploadFileName := createFileName(card_ddd)
	path_to_save := filepath.Join(config.UploadDir, uploadFileName)

	f, err := os.Create(path_to_save)
	if err != nil {
		return err
	}
	defer f.Close()

	f.Write(card_ddd)

	return nil
}

func VerifyPIN(pin string, indexReader int) error {
	ctx, d, err := CardConnect(indexReader)

	if err != nil {
		return err
	}

	if _, err = verify(pin, d); err != nil {
		ctx.Release()
		return err
	}

	ctx.Release()
	return err
}

func UnblockCardByPUK(puk, newPin string, indexReader int) error {
	ctx, d, err := CardConnect(indexReader)

	if err != nil {
		return err
	}

	if _, err = unblock(puk, newPin, d); err != nil {
		ctx.Release()
		return err
	}

	ctx.Release()
	return err
}
