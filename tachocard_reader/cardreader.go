package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/ebfe/scard"
	"log"
	"time"
)

func checkEnableReaders() error {
	ctx, err := scard.EstablishContext()
	if err != nil {
		return err
	}
	defer ctx.Release()

	readers, err := ctx.ListReaders()
	if err != nil {
		return err
	}

	if len(readers) == 0 {
		err = errors.New("Не обнаружено ни одного ридера")
	}

	return err
}

func waitCard() (int, error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		return -1, err
	}
	defer ctx.Release()

	readers, err := ctx.ListReaders()
	if err != nil {
		return -1, err
	}

	fmt.Println("Ожидаем ввод карты")
	return waitUntilCardPresent(ctx, readers)
}

func CardConnect(indexReader int) (*scard.Context, *scard.Card, error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		return nil, nil, err
	}

	readers, err := ctx.ListReaders()
	if err != nil {
		return nil, nil, err
	}
	currentReader := readers[indexReader]

	log.Println("Соединение с картой в ридере ", currentReader)
	card, err := ctx.Connect(currentReader, scard.ShareExclusive, scard.ProtocolAny)
	if err != nil {
		log.Println(err)
		return nil, nil, errors.New("Невозможно считать карту")
	}

	log.Println("Статус карты:")
	if status, err := card.Status(); err != nil {
		return nil, nil, err
	} else {
		log.Println("Ридер: ", status.Reader)
		log.Println("Статус: ", status.State)
		log.Println("Активный протокол: ", status.ActiveProtocol)
		log.Println("Atr: ", status.Atr)
	}

	return ctx, card, err
}

func waitUntilCardPresent(ctx *scard.Context, readers []string) (int, error) {
	rs := make([]scard.ReaderState, len(readers))
	for i := range rs {
		rs[i].Reader = readers[i]
		rs[i].CurrentState = scard.StateUnaware
	}

	for {
		for i := range rs {
			if rs[i].EventState&scard.StatePresent != 0 {
				return i, nil
			}
			rs[i].CurrentState = rs[i].EventState
		}
		err := ctx.GetStatusChange(rs, -1)
		if err != nil {
			return -1, err
		}
	}
}

func sendApdu(cmd []byte, card *scard.Card) ([]byte, error) {
	var sw_idx, resp_len int
	var result, sw_byte []byte

	card_resp, err := card.Transmit(cmd)

	if err != nil {
		return nil, err
	}

	resp_len = len(card_resp)
	sw_idx = resp_len - 2

	sw_byte = card_resp[sw_idx:]
	result = card_resp[:sw_idx]

	if !(sw_byte[0] == 0x90 && sw_byte[1] == 0x00) {
		error_msg := fmt.Sprintf("Ошибка выполнения команды % x. Код ошибки: %x.\n", cmd, sw_byte)
		return nil, errors.New(error_msg)
	}

	return result, err
}

func selectFile(fid []byte, card *scard.Card) ([]byte, error) {
	var cmd []byte

	if fid[0] == 0xFF {
		cmd = []byte{0x00, 0xA4, 0x04, 0x0C, 0x06}
	} else {
		cmd = []byte{0x00, 0xA4, 0x02, 0x0C, 0x02}
	}

	cmd = append(cmd, fid...)
	return sendApdu(cmd, card)
}

func readBinary(size int, card *scard.Card) ([]byte, error) {
	READ_BLOCK_SIZE := 200
	pos := 0
	cmd := []byte{}
	val := []byte{}
	tmp_val := []byte{}
	var expected, begin_byte, end_byte byte
	var err error

	for pos < size {
		if size-pos < READ_BLOCK_SIZE {
			expected = int8ToByte(size - pos)
		} else {
			expected = int8ToByte(READ_BLOCK_SIZE)
		}

		begin_byte = int8ToByte(pos >> 8 & 0xFF)
		end_byte = int8ToByte(pos & 0xFF)
		cmd = []byte{0x00, 0xB0, begin_byte, end_byte, expected}
		tmp_val, err = sendApdu(cmd, card)

		if err != nil {
			return nil, err
		}

		val = append(val, tmp_val...)
		pos = pos + len(tmp_val)
	}

	return val, err
}

func updateBynary(recBytes []byte, size int, card *scard.Card) ([]byte, error) {
	begin_byte := int8ToByte(size >> 8 & 0xFF)
	end_byte := begin_byte
	len_byte := int8ToByte(size)

	cmd := []byte{0x00, 0xD6, begin_byte, end_byte, len_byte}
	cmd = append(cmd, recBytes...)

	return sendApdu(cmd, card)
}

func performHash(card *scard.Card) ([]byte, error) {
	cmd := []byte{0x80, 0x2A, 0x90, 0x00}
	return sendApdu(cmd, card)
}

func computeDigitalSignature(card *scard.Card) ([]byte, error) {
	cmd := []byte{0x00, 0x2A, 0x9E, 0x9A, 0x40}
	return sendApdu(cmd, card)
}

func codeAlign(code []byte) []byte {
	var MAX_PIN_LEN int = 8
	var DEFAULT_BYTE_ALIGNMENT byte = 0xFF
	result := code

	for len(result) < MAX_PIN_LEN {
		result = append(result, DEFAULT_BYTE_ALIGNMENT)
	}
	return result
}

func verify(pin string, card *scard.Card) ([]byte, error) {
	pin_for_send := codeAlign([]byte(pin))

	cmd := []byte{0x00, 0x20, 0x00, 0x00, 0x08}
	cmd = append(cmd, pin_for_send...)

	return sendApdu(cmd, card)
}

func unblock(puk, newPin string, card *scard.Card) ([]byte, error) {
	puk_for_send := codeAlign([]byte(puk))
	pin_for_send := codeAlign([]byte(newPin))

	cmd := []byte{0x00, 0x2C, 0x00, 0x00, 0x10}
	cmd = append(cmd, puk_for_send...)
	cmd = append(cmd, pin_for_send...)

	return sendApdu(cmd, card)
}

func int8ToByte(int_val int) byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, uint8(int_val))
	if err != nil {
		log.Fatal("binary.Write failed:", err)
	}
	return buf.Bytes()[0]
}

func lenToByte(int_val int) []byte {
	result := make([]byte, 2)
	binary.LittleEndian.PutUint16(result, uint16(int_val))

	return reverseBytes(result)
}

func hexToDec(bytes []byte) int {
	hex_digit := make([]byte, len(bytes))
	copy(hex_digit, bytes)

	if len(hex_digit) < 2 {
		hex_digit = append(hex_digit, 0x00)
	} else {
		hex_digit = reverseBytes(hex_digit)
	}

	result := binary.LittleEndian.Uint16(hex_digit)
	return int(result)
}

func uploadTime() []byte {
	now := time.Now().UTC()

	result := make([]byte, 4)
	binary.LittleEndian.PutUint32(result, uint32(now.Unix()))
	return reverseBytes(result)
}

func reverseBytes(numbers []byte) []byte {
	for i := 0; i < len(numbers)/2; i++ {
		j := len(numbers) - i - 1
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}
	return numbers
}

var countFieldRec map[string]int

func initCardType(card *scard.Card, countFieldRec *map[string]int) (byte, error) {
	var param_bytes []byte
	var card_type byte
	var err error

	_, err = selectFile([]byte{0x05, 0x01}, card)
	if err != nil {
		return 0x00, err
	}

	param_bytes, err = readBinary(1, card)
	if err != nil {
		return 0x00, err
	}
	card_type = param_bytes[0]

	switch card_type {
	case 0x01:
		param_bytes, err = readBinary(10, card)
		if err != nil {
			return 0x00, err
		}

		NoOfEventsPerType := param_bytes[3:4]
		NoOfFaultsPerType := param_bytes[4:5]
		NoOfCardVehicleRecords := param_bytes[7:9]
		NoOfCardPlaceRecords := param_bytes[9:]
		CardActivityLenghtRange := param_bytes[5:7]

		*countFieldRec = map[string]int{
			"EF_Events_Data":          hexToDec(NoOfEventsPerType) * 24 * 6,
			"EF_Faults_Data":          hexToDec(NoOfFaultsPerType) * 24 * 2,
			"EF_Vehicles_Used":        hexToDec(NoOfCardVehicleRecords)*31 + 2,
			"EF_Places":               hexToDec(NoOfCardPlaceRecords)*10 + 1,
			"EF_Driver_Activity_Data": hexToDec(CardActivityLenghtRange) + 4,
		}
	case 0x02:
		param_bytes, err = readBinary(11, card)
		if err != nil {
			return 0x00, err
		}

		NoOfEventsPerType := param_bytes[3:4]
		NoOfFaultsPerType := param_bytes[4:5]
		NoOfCardVehicleRecords := param_bytes[7:9]
		NoOfCardPlaceRecords := param_bytes[9:10]
		CardActivityLenghtRange := param_bytes[5:7]
		NoOfCalibrationRecords := param_bytes[10:]

		*countFieldRec = map[string]int{
			"EF_Events_Data":          hexToDec(NoOfEventsPerType) * 6 * 24,
			"EF_Faults_Data":          hexToDec(NoOfFaultsPerType) * 2 * 24,
			"EF_Vehicles_Used":        hexToDec(NoOfCardVehicleRecords)*31 + 2,
			"EF_Places":               hexToDec(NoOfCardPlaceRecords)*10 + 1,
			"EF_Calibration":          hexToDec(NoOfCalibrationRecords)*105 + 3,
			"EF_Driver_Activity_Data": hexToDec(CardActivityLenghtRange) + 4,
		}
	case 0x03:
		param_bytes, err = readBinary(5, card)
		if err != nil {
			return 0x00, err
		}

		NoOfControllActivityRecords := param_bytes[3:5]

		*countFieldRec = map[string]int{
			"EF_Controller_Activity_Data": hexToDec(NoOfControllActivityRecords)*46 + 2,
		}
	case 0x04:
		param_bytes, err = readBinary(5, card)
		if err != nil {
			return 0x00, err
		}

		NoOfCompanyActivityRecords := param_bytes[3:5]

		*countFieldRec = map[string]int{
			"EF_Company_Activity_Data": hexToDec(NoOfCompanyActivityRecords)*46 + 2,
		}
	}

	return card_type, err
}

type CardFile struct {
	Name    string
	Tag     [2]byte
	Min_len int
	Max_len int
	Signed  bool
}

func (cardfile *CardFile) size() int {
	if cardfile.Max_len == cardfile.Min_len {
		return cardfile.Min_len
	} else {
		return countFieldRec[cardfile.Name]
	}
}

func readFile(cf *CardFile, card *scard.Card) ([]byte, error) {
	var fid, result_len, result []byte
	var resp_len int

	fid = cf.Tag[:]

	resp, err := selectFile(fid, card)
	if err != nil {
		return nil, err
	}

	resp_len = cf.size()
	resp, err = readBinary(resp_len, card)
	if err != nil {
		return nil, err
	}

	result_len = lenToByte(resp_len)
	result = append(fid, 0x00)
	result = append(result, result_len...)
	result = append(result, resp...)

	if cf.Signed {
		_, err = performHash(card)
		if err != nil {
			return nil, err
		}

		signature, err := computeDigitalSignature(card)
		if err != nil {
			return nil, err
		}

		result = append(result, fid...)
		result = append(result, []byte{0x81, 0x00, 0x40}...)
		result = append(result, signature...)
	}

	return result, err
}

func UpdateUploadDate(fid [2]byte, card *scard.Card) error {
	if _, err := selectFile(fid[:], card); err != nil {
		return err
	}

	uploadDateBytes := uploadTime()
	if _, err := updateBynary(uploadDateBytes, len(uploadDateBytes), card); err != nil {
		return err
	}

	return nil
}
