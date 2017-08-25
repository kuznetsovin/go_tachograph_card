package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Преобразует байтовую матрицу в формат TLV записей. Нужно для циклических файлов
func recTblToTlvs(tag string, recTbl [][]byte) []map[string][]byte {
	result := []map[string][]byte{}

	for _, rec := range recTbl {
		tmp := map[string][]byte{tag: rec}
		result = append(result, tmp)
	}
	return result
}

// Функция для разбивки байтового массива в матрицу. Нужна для обработки
// циклических записей в tlv выгрузке.
func readFileRecords(bytes []byte, recordLen int, offsetBeginRecord int) [][]byte {

	allFileRecords := bytes[offsetBeginRecord:]
	var result [][]byte

	beginIdx := 0
	endIdx := recordLen
	for beginIdx < len(allFileRecords) {
		result = append(result, allFileRecords[beginIdx:endIdx])
		beginIdx = endIdx
		endIdx = endIdx + recordLen
	}

	return result
}

// Функция для разбивки поля ActivityDalyRecords по записям. Нужна для обработки
// циклических записей в tlv выгрузке.
// Фукция совершает 2 прохода прямой и обратный. Прямой проходит от указателя самой
// старой записи (oldestPointer) к концу файла, а обратный -  от самой свежей записи (newestPointer)
// к началу файла.
func readActivityDailyRecs(bytesRec []byte, offset int) [][]byte {
	var result, resultBackward [][]byte
	sizeLenBytes := 2

	// Первые 2 байта содержат размер предыдущей записи
	// следующие 2 байта, размер текущей записи, поэтому смещение 4, чтобы удобнее получать
	// непосредственное значение записи
	prefixArdLensSize := 4
	oldestPointer, _ := hexToInt(bytesRec[:2])
	newestPointer, _ := hexToInt(bytesRec[2:prefixArdLensSize])

	if oldestPointer != 0 {
		//если задан указатель на самую старую запись, то чтение начентся с нее
		offset = offset + oldestPointer
	}

	recordBegin := offset
	activityADRs := bytes.TrimRight(bytesRec, string(0x00))
	prevBlockLen := 0
	activityDailyRecsLen := len(activityADRs)
	// "прямой" проход
	for recordBegin < activityDailyRecsLen {
		lenBytesDelimiterIdx := recordBegin + sizeLenBytes
		prevRecLen := activityADRs[recordBegin:lenBytesDelimiterIdx]
		curRecLen := activityADRs[lenBytesDelimiterIdx : lenBytesDelimiterIdx + sizeLenBytes]

		curBlockLen, _ := bytesToInt(curRecLen)
		prevLenCurrentBlock, _ := bytesToInt(prevRecLen)

		if prevLenCurrentBlock != prevBlockLen {
			//нестандартная ситуация, скорее всего неисправность карты
			break
		}

		ardValueBeginOffset := recordBegin + prefixArdLensSize
		// Т. к. записи начинаются с 4 байтов длин (см. prefixArdLensSize), то размер значения,
		// текущего Ard, равен {длина записи} - 4 байта
		ardValueEndOffset := ardValueBeginOffset + curBlockLen - prefixArdLensSize

		activity := []byte{}
		// обработка если запись выходит за пределы длины файла активностей
		if ardValueEndOffset > activityDailyRecsLen {
			//обрабатываем перенос в начало файла окночания записи об активности
			activity = activityADRs[ardValueBeginOffset:]
			transferByteCount := ardValueEndOffset - activityDailyRecsLen

			// перенесенные байты читаются от начала файла, за вычетом 4 байт
			// (указатели на самую страрую и самую новую записи)
			transferBytesOffset := prefixArdLensSize + transferByteCount
			activity = append(activity, activityADRs[prefixArdLensSize:transferBytesOffset]...)
		} else {
			//если запись не выходит за пределы секции, то просто читаем активность
			activity = activityADRs[ardValueBeginOffset:ardValueEndOffset]
		}

		result = append(result, activity)

		recordBegin = ardValueEndOffset
		prevBlockLen = curBlockLen
	}

	// "обратный" проход, таким образом данные читаются задом наперед
	if newestPointer < oldestPointer {
		recordBegin = newestPointer + prefixArdLensSize
		for {
			lenRecsSep := recordBegin + sizeLenBytes
			prevRecLen := activityADRs[recordBegin:lenRecsSep]
			curRecLen := activityADRs[lenRecsSep : lenRecsSep + sizeLenBytes]

			curBlockLen, _ := bytesToInt(curRecLen)
			prevBlockLen, _ := bytesToInt(prevRecLen)

			ardValueBeginOffset := recordBegin + prefixArdLensSize
			ardValueEndOffset := ardValueBeginOffset + curBlockLen - prefixArdLensSize

			//записи записываются в обратной последовательности
			resultBackward = append(resultBackward, activityADRs[ardValueBeginOffset:ardValueEndOffset])

			recordBegin = recordBegin - prevBlockLen
			lenRecsSep = recordBegin + sizeLenBytes

			if recordBegin < 0 || lenRecsSep < 0 {
				break
			}

			nextBlockLen := activityADRs[lenRecsSep : lenRecsSep + sizeLenBytes]
			prevLenCurrentBlock, _ := hexToInt(nextBlockLen)

			if prevLenCurrentBlock != prevBlockLen {
				break
			}

		}
		// для правильной последовательности записей на карте, необходимо перевернуть
		// результат "обратного" прохода
		result = append(result, reverseBytes(resultBackward)...)
	}

	return result
}

// Функция перерабатывает ddd файл в словарь вида {Тэг: Значение}
// Например { '0002': []byte{0x00, 0x01} }
func extractFieldVals(ddd []byte) (map[string][]byte, error) {
	/*
	 */
	var result = map[string][]byte{}

	offset := 0
	tagCountBytes := 3
	lenCountBytes := 2

	for offset < len(ddd) {
		tag := readBytes(ddd, tagCountBytes, offset)

		offset = offset + tagCountBytes

		// 81 - флаг подписи для СКЗИ, 01 - для ЕСТР
		if tag[2] == 0x81 || tag[2] == 0x01 {
			digitLenBytes := readBytes(ddd, lenCountBytes, offset)
			digitLen, err := bytesToInt(digitLenBytes)
			if err != nil {
				return result, err
			}
			offset = offset + lenCountBytes + digitLen
			continue
		}

		hex_len := readBytes(ddd, lenCountBytes, offset)
		intLen, err := bytesToInt(hex_len)
		if err != nil {
			return result, err
		}
		offset = offset + lenCountBytes

		val := readBytes(ddd, intLen, offset)
		offset = offset + intLen

		fieldName := fmt.Sprintf("%X", tag[:2])
		result[fieldName] = val
	}
	return result, nil
}

// Функция для чтения среза из байтового массива.
func readBytes(bytes []byte, count int, offset int) []byte {
	if count < 0 {
		return bytes[offset:]
	}
	return bytes[offset : offset + count]
}

// Функция для преобразования 2-х байтного числа в int.
func bytesToInt(hexLen []byte) (int, error) {
	size_array := len(hexLen)
	if size_array > 2 {
		return 0, errors.New("Hex len > 2 byte")
	}

	max_byte_num := size_array - 1
	buf := new(bytes.Buffer)
	if max_byte_num == 0 {
		buf.WriteByte(hexLen[max_byte_num])
		buf.WriteByte(0x00)
	} else {
		for i := max_byte_num; i >= 0; i-- {
			buf.WriteByte(hexLen[i])
		}
	}
	hex_val := buf.Bytes()

	result := binary.LittleEndian.Uint16(hex_val)
	return int(result), nil
}

// Функция для преобразования массива байт в int64
func bytesToInt64(hexLen []byte) (int64, error) {
	size_array := len(hexLen)
	if size_array < 3 {
		return 0, errors.New("Invalid length []byte use func BytesToInt()")
	}

	max_byte_num := size_array - 1
	buf := new(bytes.Buffer)

	for i := max_byte_num; i >= 0; i-- {
		buf.WriteByte(hexLen[i])
	}

	hex_val := buf.Bytes()
	size_result_array := len(hex_val)

	for size_result_array < 4 {
		hex_val = append(hex_val, 0x00)
		size_result_array++
	}

	result := binary.LittleEndian.Uint32(hex_val)
	return int64(result), nil
}

// Функция перевода десятичной строки в набор байтов.
func hexdemicalToBytes(hexdem string) ([]byte, error) {
	var result []byte
	var err error
	i := 0
	for i < len(hexdem) {
		curByte := hexdem[i : i + 2]
		z, _ := strconv.ParseUint(curByte, 16, 8)
		result = append(result, uint8(z))
		i = i + 2
	}

	return result, err
}

// Функция преобразования битов в число
func bitsToInt(Bits string, base int, bitLen int) (int, error) {
	s, err := strconv.ParseInt(Bits, base, bitLen)
	if err != nil {
		return int(s), err
	}
	return int(s), err

}

// Функция преобразования байтов и биты
func bytesToBits(bytes []byte) string {
	return fmt.Sprintf("%08b%08b", bytes[0], bytes[1])
}

func byteToBool(b byte) (bool, error) {
	return strconv.ParseBool(string(b))
}

func hexToDate(hex []byte) (time.Time, error) {
	result := time.Unix(0, 0)
	unix_sec, err := bytesToInt64(hex)
	if err != nil {
		return result, err
	}

	result = time.Unix(unix_sec, 0).UTC()

	return result, err
}

func decodeString(str string, decoder *encoding.Decoder) (string, error) {
	sr := strings.NewReader(str)
	tr := transform.NewReader(sr, decoder)
	buf, err := ioutil.ReadAll(tr)
	if err != err {
		return "", err
	}

	return string(buf), err
}

// Функция для перекодировки hex(latin) в строку utf8.
func hexStringToUtf8(bytes_arr []byte) (string, error) {
	var decoder *encoding.Decoder
	var err error

	result := ""
	encode_byte := bytes_arr[0]
	switch encode_byte {
	case 0x05:
		decoder = charmap.ISO8859_5.NewDecoder()
		result = string(bytes_arr[1:])
	case 0x01:
		decoder = charmap.ISO8859_1.NewDecoder()
		result = string(bytes_arr[1:])
	default:
		if !utf8.Valid(bytes_arr) {
			decoder = charmap.Windows1251.NewDecoder()
		}
		result = string(bytes_arr)
	}

	if decoder != nil {
		result, err = decodeString(result, decoder)
		if err != nil {
			return result, err
		}
	}

	result = strings.Trim(result, " ")
	result = strings.Replace(result, string(0x00), "", -1)
	result = strings.Replace(result, string(0x05), " ", -1)

	re := regexp.MustCompile("\\s+")
	result = re.ReplaceAllString(result, " ")

	return result, err
}

// Метод перевода байтов и int32.
func hexToInt(hexVal []byte) (int, error) {
	var result int
	var err error

	lenVal := len(hexVal)

	if lenVal > 2 {
		int64Val, err := bytesToInt64(hexVal)
		if err != nil {
			return result, err
		}
		result = int(int64Val)
	} else {
		return bytesToInt(hexVal)
	}

	return result, err
}

func reverseBytes(numbers [][]byte) [][]byte {
	for i := 0; i < len(numbers) / 2; i++ {
		j := len(numbers) - i - 1
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}
	return numbers
}

// функция загрузки информации об активности за день по водителю
func parseChangeInfoData(recDate time.Time, currentACI string, aci *activityChangeInfo) error {

	r, err := hexdemicalToBytes(currentACI)
	if err != nil {
		return fmt.Errorf("Can't decode activity change info: %v", err)
	}

	aciBytes := bytesToBits(r)

	flag, err := byteToBool(aciBytes[0])
	if err != nil {
		return fmt.Errorf("Can't get TachographCardReaderId from activity change info: %v", err)
	}
	if flag {
		aci.TachographCardReaderId = 1
	} else {
		aci.TachographCardReaderId = 0
	}

	flag, err = byteToBool(aciBytes[1])
	if err != nil {
		return fmt.Errorf("Can't get StateDrivingId from activity change info: %v", err)
	}
	if flag {
		aci.StateDrivingId = 1
	} else {
		aci.StateDrivingId = 0
	}

	flag, err = byteToBool(aciBytes[2])
	if err != nil {
		return fmt.Errorf("Can't get CardPositionId from activity change info: %v", err)
	}
	if flag {
		aci.CardPositionId = 1
	} else {
		aci.CardPositionId = 0
	}

	kindId, err := bitsToInt(aciBytes[3:5], 2, 8)
	if err != nil {
		return fmt.Errorf("Can't get ActivityKindId from activity change info: %v", err)
	}

	aci.ActivityKindId = kindId
	minutes, err := bitsToInt(aciBytes[5:], 2, 12)
	if err != nil {
		return fmt.Errorf("Can't get ActivityChangeInfoT from activity change info: %v", err)
	}
	aci.ActivityChangeInfoT = minutes

	aci.CalculatedTime = recDate.Add(time.Duration(minutes) * time.Minute)

	return err
}
