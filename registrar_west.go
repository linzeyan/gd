package gd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type WestZone struct {
	Result   int    `json:"result"`
	ClientId string `json:"clientid"`
	Data     struct {
		Items []WestZoneView `json:"items"`
	} `json:"data"`
	Limit     string `json:"limit"`
	Total     string `json:"total"`
	PageNo    string `json:"pageno"`
	TotalPage string `json:"totalpage"`
}

type WestZoneView struct {
	Domain       string `json:"domain"`
	RegistryDate string `json:"regdate"`
	ExpiryDate   string `json:"expdate"`
	Dns1         string `json:"dns1"`
	Dns2         string `json:"dns2"`
	Dns3         string `json:"dns3"`
	Dns4         string `json:"dns4"`
	Dns5         string `json:"dns5"`
	Dns6         string `json:"dns6"`
	Year         string `json:"year"`
	Version      string `json:"version"`
	Hold         string `json:"clienthold"`
}

func (w *WestZoneView) CreateTableQuery() string {
	e := reflect.ValueOf(w).Elem()
	var tableFields string
	for i := 0; i < e.NumField(); i++ {
		var sqlType string
		varName := e.Type().Field(i).Name
		varType := e.Field(i).Type().String()
		if varType == "int" {
			sqlType = varType
		} else {
			sqlType = "varchar(256)"
		}
		tableFields = tableFields + varName + " " + sqlType + ","
	}
	tableFields = strings.TrimRight(tableFields, ",")
	TableWestZone := `westZone` + Now()
	createQuery := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(%s)`, TableWestZone, tableFields)
	return createQuery
}

func (w *WestZoneView) InsertData() (string, []interface{}) {
	e := reflect.ValueOf(w).Elem()
	TableWestZone := `westZone` + Now()
	insertQuery := fmt.Sprintf(`INSERT %s SET `, TableWestZone)
	var insertValue []interface{}
	for i := 0; i < e.NumField(); i++ {
		varName := e.Type().Field(i).Name
		varValue := e.Field(i).Interface()
		insertQuery = insertQuery + varName + "=?,"
		insertValue = append(insertValue, varValue)
	}
	insertQuery = strings.TrimRight(insertQuery, ",")
	return insertQuery, insertValue
}

func (w *WestZoneView) GetZones(api, account, key string) ([]string, []WestZoneView) {
	var zone WestZone
	var id []string
	content, err := DoRequest(WestRequest(api, account, key, "zone", ""))
	if err != nil {
		log.Println(err)
		return nil, nil
	}
	err = json.Unmarshal(content, &zone)
	if err != nil {
		log.Println(err)
		return nil, nil
	}
	for i := range zone.Data.Items {
		id = append(id, zone.Data.Items[i].Domain)
	}
	return id, zone.Data.Items
}

type WestRecord struct {
	Result   int    `json:"result"`
	ClientId string `json:"clientid"`
	Data     struct {
		PageNo    int              `json:"pageno"`
		Limit     int              `json:"limit"`
		Total     int              `json:"total"`
		PageCount int              `json:"pagecount"`
		Items     []WestRecordView `json:"items"`
	} `json:"data"`
	TotalPage string `json:"totalpages"`
}

type WestRecordView struct {
	Id    int    `json:"id"`
	Item  string `json:"item"`
	Value string `json:"value"`
	Type  string `json:"type"`
	Level int    `json:"level"`
	TTL   int    `json:"ttl"`
	Line  string `json:"line"`
	Pause int    `json:"pause"`
	Zone  string
}

func (w *WestRecordView) CreateTableQuery() string {
	e := reflect.ValueOf(w).Elem()
	var tableFields string
	for i := 0; i < e.NumField(); i++ {
		var sqlType string
		varName := e.Type().Field(i).Name
		varType := e.Field(i).Type().String()
		if varType == "int" {
			sqlType = varType
		} else {
			sqlType = "varchar(256)"
		}
		tableFields = tableFields + varName + " " + sqlType + ","
	}
	tableFields = strings.TrimRight(tableFields, ",")
	TableWestRecord := `westRecord` + Now()
	createQuery := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(%s)`, TableWestRecord, tableFields)
	return createQuery
}

func (w *WestRecordView) InsertData() (string, []interface{}) {
	e := reflect.ValueOf(w).Elem()
	TableWestRecord := `westRecord` + Now()
	insertQuery := fmt.Sprintf(`INSERT %s SET `, TableWestRecord)
	var insertValue []interface{}
	for i := 0; i < e.NumField(); i++ {
		varName := e.Type().Field(i).Name
		varValue := e.Field(i).Interface()
		insertQuery = insertQuery + varName + "=?,"
		insertValue = append(insertValue, varValue)
	}
	insertQuery = strings.TrimRight(insertQuery, ",")
	return insertQuery, insertValue
}

func (w *WestRecordView) GetRecords(api, account, key, zoneId string) []WestRecordView {
	var record WestRecord
	content, err := DoRequest(WestRequest(api, account, key, "record", zoneId))
	if err != nil {
		log.Println(err)
		return nil
	}
	err = json.Unmarshal(content, &record)
	if err != nil {
		log.Println(err)
		return nil
	}
	return record.Data.Items
}

func WestRequest(api, account, key, method, zoneId string) *http.Request {
	var uri string
	now := strconv.FormatInt(time.Now().Local().UnixMilli(), 10)
	hashData := account + key + now
	data := url.Values{}
	data.Set("username", account)
	data.Set("limit", "1000")
	data.Set("time", now)
	data.Set("token", Md5encode(hashData))
	switch method {
	case "zone":
		args := "/domain/?act=getdomains"
		uri = api + args
	case "record":
		args := "/domain/?act=getdnsrecord"
		uri = api + args
		data.Set("domain", zoneId)
	}
	req, err := http.NewRequest(http.MethodGet, uri, strings.NewReader(data.Encode()))
	if err != nil {
		log.Println("Resquest error.")
		log.Println(err)
		return req
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}
