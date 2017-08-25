# Web cеврис для парсинга DDD файлов. 

## Сборка

Для сборки необходимо ввести команду:

```
go build ddd_parsing_service
```

## Параметры запуска 

```port``` - порт для запуска сервиса (По умолчанию: _":8000"_)

```log``` - имя лог файла (По умолчанию: _"ddd_parsing_service.log"_)

Пример команды запуска

```
ddd_parsing_service -port ":8000" -log "./ddd_parsing_service.log"
```

### API
Для разбора данных необходимо отправить GET запрос с параметром ``ddd`` на адрес сервиса.

```
GET /?ddd=<строка base64 из ddd файла>
```


Пример на Python:

```python
import requests
import base64

f = open("test.ddd", "rb")
encoded_ddd = base64.b64encode(f.read())

r = requests.get('http://localhost:8000/', params={'ddd': encoded_ddd})
f.close()

result = r.json()
print result["CardNumber"]
```

## Входящие данные
На вход подается строка **base64** c содержимым DDD файла с карты водителя.

## Выходные данных 
В ответ на запрос сервис возвращает json, следующей структуры:
```
 {
    "IcSerialNumber": "hexadecimal",
    "IcManufacturingReferences": "hexadecimal",
    "CardExtendedSerialNumber": "string",
    "CardApprovalNumber": "string",
    "CardPersonalizerId": "int",
    "EmbeddericAssemblerId": "string",
    "IcIdentifier": "int",
    "CardNumber": "string", 
    "CardIssuingMemberState": "int",
    "CardIssuingAuthorityName": "string",
    "CardIssueDate": "date",
    "CardValidityBegin": "date",
    "CardExpiryDate": "date",
    "LastCardDownload": "date", 
    "TypeOfTachographCardId": "int",
    "CardStructureVersion": "hexadecimal",
    "CardCertificateGost":  "hexadecimal",
    "CACertificateGost":  "hexadecimal",
    "CardCertificateESTR":  "hexadecimal",
    "CACertificateESTR":  "hexadecimal",
    "HolderSurname": "string",
    "HolderFirstNames": "string",
    "CardHolderBirthDate": "birthday",
    "CardHolderPreferredLanguage": "string",
    "DrivingLicenceIssuingAuthority": "string",
    "DrivingLicenceIssuingNation":"int",
    "DrivingLicenceNumber":"string",
    "SessionOpenTime": "date",
    "SessionOpenVehicleNumber": "string",
    "SessionOpenVehicleNation": "int",
    "CardVehicleRecords": [
        {
            "VFU": "VehicleFirstUse date",
            "VLU": "VehicleLastUse date",
            "VOB": "VehicleOdometerBegin int",
            "VOE": "VehicleOdometerEnd int",
            "VRN": "VehicleRegistrationNumber string",
            "VRNat": "VehicleRegistrationNation int"
        }
    ],
    "ActivityDailyRecords": [
        {
            "ARD": "ActivityRecordDate date",
            "ADPC": "ActivityDailyPresenceCounter daylicounter",
            "ADD": "ActivityDayDistance int",
            "AS": "ActivitiesS activites",
            "ACI": [
                {    
                    "TCRId": "TachographCardReaderId bool",
                    "SDId": "StateDrivingId bool",
                    "CPId": "CardPositionId bool",
                    "AKId": "ActivityKindId int",
                    "ACIT": "ActivityChangeInfoT int",
                    "CT": "CalculatedTime date",
                }
            ]
        }
    ],
    "PlaceRecords": [
        {
            "ET": "EntryTime date",
            "TPId": "TypePeriodId int",
            "DWPC": "DailyWorkPeriodCountry int",
            "DWPR": "DailyWorkPeriodRegion int",
            "VOV":  "VehicleOdometerValue int"
        }
    ],
    "CardEventRecords": [
        {
            "ETId": "EventTypeId int",
            "EBT": "EventBeginTime date",
            "EET": "EventEndTime date",
            "VRN": "VehicleRegistrationNumber string",
            "VRNat": "VehicleRegistrationNation int"
        }
    ],
    "CardFaultRecords": [
        {
            "FTId": "FaultTypeId int",
            "FBT": " FaultBeginTime date",
            "FET": "FaultEndTime date",
            "VRN": "VehicleRegistrationNumber string",
            "VRNat": "VehicleRegistrationNation int"
        }
    ],
    "CardControlActivityDataRecord": [
        {
            "ControlTypeId": "int",
            "ControlTime": "date",
            "CardTypeId": "int",
            "CardIssuingMemberState": "int",
            "ControlCardNumber": "string",
            "ControlDownloadPeriodBegin": "date",
            "ControlDownloadPeriodEnd": "date",
            "VehicleRegistrationNumber": "string",
            "VehicleRegistrationNation": "int"
        }
    ],
    "SpecificConditionRecord": [
        {
            "SpecificConditionTypeId":"int",
            "EntryTime": "date"
        }
    ]    
}
```
