package main

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"log"
)

type fieldTag struct {
	Name       string
	ValueLen   int
	Offset     int
	OutputType string
	Required   bool
}

// Функция производит заполнение структур даанными из tlv файла,
// с помощью маппинга, сделанного из тэгов подсказок `tlv` у соответствующего поля.
func loadFields(customStruct interface{}, tlvRecords map[string][]byte) error {
	var structType reflect.Type
	var structValRef reflect.Value
	var err error

	switch t := customStruct.(type) {
	case *cardEventRecords:
		tag := "0502"
		recordsOfCircleFile := readFileRecords(tlvRecords[tag], 24, 0)
		tlvs := recTblToTlvs(tag, recordsOfCircleFile)
		c := cardEventRecord{}
		for _, tlv := range tlvs {
			if err := loadFields(&c, tlv); err != nil {
				return err
			}

			if !EventRecordIsEmpty(&c) {
				*t = append(*t, c)
			}
		}
		return nil
	case *сardVehicleRecords:
		tag := "0505"
		recordsOfCircleFile := readFileRecords(tlvRecords[tag], 31, 2)
		tlvs := recTblToTlvs(tag, recordsOfCircleFile)
		c := сardVehicleRecord{}
		for _, tlv := range tlvs {
			if err := loadFields(&c, tlv); err != nil {
				return err
			}
			if !VehicleRecordIsEmpty(&c) {
				*t = append(*t, c)
			}
		}
		return nil
	case *activityDailyRecords:
		tag := "0504"
		recordsOfCircleFile := readActivityDailyRecs(tlvRecords[tag], 4)
		tlvs := recTblToTlvs(tag, recordsOfCircleFile)
		c := activityDailyRecord{}
		for _, tlv := range tlvs {
			if err := loadFields(&c, tlv); err != nil {
				return err
			}

			if err := c.ParseChangeInfo(); err != nil {
				return err
			}

			*t = append(*t, c)
		}
		return nil
	case *placeRecords:
		tag := "0506"
		recordsOfCircleFile := readFileRecords(tlvRecords[tag], 10, 1)
		tlvs := recTblToTlvs(tag, recordsOfCircleFile)
		c := placeRecord{}
		for _, tlv := range tlvs {
			if err := loadFields(&c, tlv); err != nil {
				return err
			}
			if !PlaceRecordIsEmpty(&c) {
				*t = append(*t, c)
			}
		}
		return nil
	case *cardFaultRecords:
		tag := "0503"
		recordsOfCircleFile := readFileRecords(tlvRecords[tag], 24, 0)
		tlvs := recTblToTlvs(tag, recordsOfCircleFile)
		c := cardFaultRecord{}
		for _, tlv := range tlvs {
			if err := loadFields(&c, tlv); err != nil {
				return err
			}
			if !FaultRecordIsEmpty(&c) {
				*t = append(*t, c)
			}
		}
		return nil
	case *cardControlActivityDataRecords:
		tag := "0508"
		recordsOfCircleFile := readFileRecords(tlvRecords[tag], 46, 0)
		tlvs := recTblToTlvs(tag, recordsOfCircleFile)
		c := cardControlActivityDataRecord{}
		for _, tlv := range tlvs {
			if err := loadFields(&c, tlv); err != nil {
				return err
			}
			if !ControlActivityDataIsEmpty(&c) {
				*t = append(*t, c)
			}
		}
		return nil
	case *specificConditionRecords:
		tag := "0522"
		recordsOfCircleFile := readFileRecords(tlvRecords[tag], 5, 0)
		tlvs := recTblToTlvs(tag, recordsOfCircleFile)
		c := specificConditionRecord{}
		for _, tlv := range tlvs {
			if err := loadFields(&c, tlv); err != nil {
				return err
			}
			if !SpecificConditionIsEmpty(&c) {
				*t = append(*t, c)
			}
		}
		return nil
	case *card:
		structType = reflect.TypeOf(*t)
		structValRef = reflect.ValueOf(t)
	case *driver:
		structType = reflect.TypeOf(*t)
		structValRef = reflect.ValueOf(t)
	case *dlicense:
		structType = reflect.TypeOf(*t)
		structValRef = reflect.ValueOf(t)
	case *sessionOpen:
		structType = reflect.TypeOf(*t)
		structValRef = reflect.ValueOf(t)
	case *cardInfo:
		structType = reflect.TypeOf(*t)
		structValRef = reflect.ValueOf(t)
	case *cardEventRecord:
		structType = reflect.TypeOf(*t)
		structValRef = reflect.ValueOf(t)
	case *сardVehicleRecord:
		structType = reflect.TypeOf(*t)
		structValRef = reflect.ValueOf(t)
	case *activityDailyRecord:
		structType = reflect.TypeOf(*t)
		structValRef = reflect.ValueOf(t)
	case *placeRecord:
		structType = reflect.TypeOf(*t)
		structValRef = reflect.ValueOf(t)
	case *cardFaultRecord:
		structType = reflect.TypeOf(*t)
		structValRef = reflect.ValueOf(t)
	case *cardControlActivityDataRecord:
		structType = reflect.TypeOf(*t)
		structValRef = reflect.ValueOf(t)
	case *specificConditionRecord:
		structType = reflect.TypeOf(*t)
		structValRef = reflect.ValueOf(t)
	default:
	}

	for i := 0; i < structType.NumField(); i++ {
		current_field := structType.Field(i)

		tlv_config, err := parseFieldTag(&current_field, "tlv")
		if err != nil {
			continue
		}

		var tlvVal = []byte{}
		var hasIndex bool
		if tlvVal, hasIndex = tlvRecords[tlv_config.Name]; !hasIndex {
			if !tlv_config.Required {
				continue
			}
			log.Printf("Can't find section [name:%s]", tlv_config.Name)
			return errors.New("Not valid input file")
		}

		hexVal := readBytes(tlvVal, tlv_config.ValueLen, tlv_config.Offset)

		field_val, err := decodeValue(tlv_config.OutputType, hexVal)
		if err != nil {
			return err
		}
		structValRef.Elem().Field(i).Set(field_val)
	}
	return err
}

// Функция для парсинга подсказок к полям, через которые будет осуществляться
// маппинг из байтовых значений.
// Формат подсказок следующий:
// 	`tlv:<имя тэга в ddd> <длинна значения> <смещение от начала tlv значения> <выходной тип данных><обязательность>`
// Например:
// 	`tlv:"0005 4 0 hexadecimal"`
// Означает что из файла 0005 будет взято 4 байта, начиная с 1 и будут преобразованы в
// HEX строку(функция decodeValue). При этом поле обязательно должно быть в ddd файле и его отсутствие
// вызовет ошибку Not valid input file.
// Чтобы не произошло этой ошибки надо установить флаг необязательности поля:
// 	`tlv:"0005 4 0 hexadecimal 0"`
// Необязательность нужна, например, при обработке файлов сетрификатов, т. к. для СКЗИ и ЕСТР они находятся
// в разных тегах. По умолчанию он равер 1.
func parseFieldTag(current_field *reflect.StructField, tag string) (fieldTag, error) {
	/*

	 */
	result := fieldTag{}
	var err error = nil
	var intVal int

	tlv_config := current_field.Tag.Get(tag)
	if tlv_config == "" {
		return result, errors.New("Field not have current tag")
	}

	field_config := strings.Split(tlv_config, " ")

	result.Name = field_config[0]

	intVal, err = strconv.Atoi(field_config[1])
	if err != nil {
		return result, err
	}
	result.ValueLen = intVal

	intVal, err = strconv.Atoi(field_config[2])
	if err != nil {
		return result, err
	}
	result.Offset = intVal

	result.OutputType = field_config[3]

	// проверяем флаг обязательности в тэге поля
	// если его нет, считаем поле обязательным
	if len(field_config) == 5 {
		if required, err := strconv.ParseBool(field_config[4]); err != nil {
			return result, err
		} else {
			result.Required = required
		}
	} else {
		result.Required = true
	}

	return result, err

}

// Функция декодирует байтовый массив в соответствующий тип для записи в БД.
func decodeValue(field_type string, hexVal []byte) (reflect.Value, error) {
	var result reflect.Value
	var err error = nil

	switch field_type {
	case "string":
		utf_string, err := hexStringToUtf8(hexVal)
		if err != nil {
			return result, err
		}
		result = reflect.ValueOf(utf_string)
	case "hexadecimal":
		trimmingHex := bytes.Trim(hexVal, string(0x00))
		trimming_str := fmt.Sprintf("%x", trimmingHex)
		result = reflect.ValueOf(trimming_str)
	case "daylicounter":
		trimmingHex := bytes.TrimLeft(hexVal, string(0x00))
		trimming_str := fmt.Sprintf("%x", trimmingHex)
		result = reflect.ValueOf(trimming_str)
	case "activites":
		trimming_str := fmt.Sprintf("%x", hexVal)
		result = reflect.ValueOf(trimming_str)
	case "int":
		intVal, err := hexToInt(hexVal)
		if err != nil {
			return result, err
		}
		result = reflect.ValueOf(intVal)
	case "birthday":
		cur_date := time.Unix(0, 0)

		date_str := fmt.Sprintf("%x-%02x-%02xT00:00:00Z", hexVal[0:2],
			hexVal[2], hexVal[3])

		if err := cur_date.UnmarshalText([]byte(date_str)); err != nil {
			cur_date, err = hexToDate(hexVal)
			if err != nil {
				return result, err
			}
		}
		result = reflect.ValueOf(cur_date)
	case "date":
		cur_date, err := hexToDate(hexVal)
		if err != nil {
			return result, err
		}

		result = reflect.ValueOf(cur_date)
	}

	return result, err
}
