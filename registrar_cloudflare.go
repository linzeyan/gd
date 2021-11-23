package gd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
)

const (
	TableCloudflareZone   = `cloudflareZone`
	TableCloudflareRecord = `cloudflareRecord`
)

type CloudflareZone struct {
	Result     []CloudflareZoneView `json:"result"`
	Errors     []string             `json:"errors"`
	Messages   []string             `json:"messages"`
	ResultInfo map[string]int       `json:"result_info"`
	Success    bool                 `json:"success"`
}

type CloudflareZoneView struct {
	Id                  string                 `json:"id"`
	Name                string                 `json:"name"`
	OriginalRegistrar   string                 `json:"original_registrar"`
	Status              string                 `json:"status"`
	Type                string                 `json:"type"`
	ActivatedTime       string                 `json:"activated_on"`
	ModifiedTime        string                 `json:"modified_on"`
	CreatedTime         string                 `json:"created_on"`
	Paused              bool                   `json:"paused"`
	NameServers         []string               `json:"name_servers"`
	OriginalNameServers []string               `json:"original_name_servers"`
	Account             map[string]string      `json:"account"`
	Owner               map[string]string      `json:"owner"`
	Plan                map[string]interface{} `json:"plan"`
	Meta                map[string]interface{} `json:"meta"`
	DevelopmentMode     int                    `json:"development_mode"`
	OriginalDnshost     string                 `json:"original_dnshost"`
	// Permissions         []string               `json:"permissions"`
}

func (cf *CloudflareZoneView) CreateTableQuery() string {
	e := reflect.ValueOf(cf).Elem()
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
	createQuery := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(%s)`, TableCloudflareZone, tableFields)
	return createQuery
}

func (cf *CloudflareZoneView) InsertData() (string, []interface{}) {
	e := reflect.ValueOf(cf).Elem()
	insertQuery := fmt.Sprintf(`INSERT %s SET `, TableCloudflareZone)
	// updateQuery := fmt.Sprintf(`UPDATE %s SET `, tableCloudflareZone)
	var insertValue []interface{}
	// var updateValue []interface{}
	// var temp interface{}
	for i := 0; i < e.NumField(); i++ {
		varName := e.Type().Field(i).Name
		varValue := e.Field(i).Interface()
		if i >= 8 && i <= 14 {
			varValue = fmt.Sprintf(`%v`, varValue)
		}
		insertQuery = insertQuery + varName + "=?,"
		insertValue = append(insertValue, varValue)
		// if varName != "Id" {
		// 	updateQuery = updateQuery + varName + "=?,"
		// 	updateValue = append(updateValue, varValue)
		// } else {
		// 	temp = varValue
		// }
	}
	insertQuery = strings.TrimRight(insertQuery, ",")
	// updateQuery = strings.TrimRight(updateQuery, ",")
	// updateQuery = updateQuery + " WHERE Id=?"
	// updateValue = append(updateValue, temp)
	// searchQuery := fmt.Sprintf(`SELECT Id FROM %s WHERE Id = "%s"`, tableCloudflareZone, temp)
	// UpsertData(searchQuery, insertQuery, updateQuery, insertValue, updateValue)
	return insertQuery, insertValue
}

func (cf *CloudflareZoneView) GetZones(api, key, mail string) ([]string, []CloudflareZoneView) {
	var zone CloudflareZone
	var id []string
	content, err := DoRequest(CloudflareRequest(api, key, mail, "zone", ""))
	if err != nil {
		log.Println(err)
		return nil, nil
	}
	err = json.Unmarshal(content, &zone)
	if err != nil {
		log.Println(err)
		return nil, nil
	}
	for i := range zone.Result {
		id = append(id, zone.Result[i].Id)
	}
	return id, zone.Result
}

type CloudflareRecord struct {
	Result     []CloudflareRecordView `json:"result"`
	Errors     []string               `json:"errors"`
	Messages   []string               `json:"messages"`
	ResultInfo map[string]int         `json:"result_info"`
	Success    bool                   `json:"success"`
}

type CloudflareRecordView struct {
	Id           string                 `json:"id"`
	ZoneName     string                 `json:"zone_name"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Content      string                 `json:"content"`
	TTL          int                    `json:"ttl"`
	Proxiable    bool                   `json:"proxiable"`
	Proxied      bool                   `json:"proxied"`
	Locked       bool                   `json:"locked"`
	Meta         map[string]interface{} `json:"meta"`
	ModifiedTime string                 `json:"modified_on"`
	CreatedTime  string                 `json:"created_on"`
	ZoneId       string                 `json:"zone_id"`
}

func (cf *CloudflareRecordView) CreateTableQuery() string {
	e := reflect.ValueOf(cf).Elem()
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
	createQuery := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(%s)`, TableCloudflareRecord, tableFields)
	return createQuery
}

func (cf *CloudflareRecordView) InsertData() (string, []interface{}) {
	e := reflect.ValueOf(cf).Elem()
	insertQuery := fmt.Sprintf(`INSERT %s SET `, TableCloudflareRecord)
	// updateQuery := fmt.Sprintf(`UPDATE %s SET `, tableCloudflareRecord)
	var insertValue []interface{}
	// var updateValue []interface{}
	// var temp interface{}
	for i := 0; i < e.NumField(); i++ {
		varName := e.Type().Field(i).Name
		varValue := e.Field(i).Interface()
		if i >= 5 && i <= 9 {
			varValue = fmt.Sprintf(`%v`, varValue)
		}
		insertQuery = insertQuery + varName + "=?,"
		insertValue = append(insertValue, varValue)
		// if varName != "Id" {
		// 	updateQuery = updateQuery + varName + "=?,"
		// 	updateValue = append(updateValue, varValue)
		// } else {
		// 	temp = varValue
		// }
	}
	insertQuery = strings.TrimRight(insertQuery, ",")
	// updateQuery = strings.TrimRight(updateQuery, ",")
	// updateQuery = updateQuery + " WHERE Id=?"
	// updateValue = append(updateValue, temp)
	// searchQuery := fmt.Sprintf(`SELECT Id FROM %s WHERE Id = "%s"`, tableCloudflareRecord, temp)
	// UpsertData(searchQuery, insertQuery, updateQuery, insertValue, updateValue)
	return insertQuery, insertValue
}

func (cf *CloudflareRecordView) GetRecords(api, key, mail, zoneId string) []CloudflareRecordView {
	var record CloudflareRecord
	content, err := DoRequest(CloudflareRequest(api, key, mail, "record", zoneId))
	if err != nil {
		log.Println(err)
		return nil
	}
	err = json.Unmarshal(content, &record)
	if err != nil {
		log.Println(err)
		return nil
	}
	return record.Result
}

func CloudflareRequest(api, key, mail, method, zoneId string) *http.Request {
	var uri string
	switch method {
	case "zone":
		args := "?per_page=200&direction=desc"
		uri = api + args
	case "record":
		args := fmt.Sprintf("/%s/dns_records?per_page=200&direction=desc", zoneId)
		uri = api + args
	}
	data := strings.NewReader(``)
	req, err := http.NewRequest("GET", uri, data)
	if err != nil {
		log.Println("Resquest error.")
		log.Println(err)
		return req
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-Auth-Key", key)
	req.Header.Set("X-Auth-Email", mail)
	return req
}
