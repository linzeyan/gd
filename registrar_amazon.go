package gd

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

const (
	TableAwsZone   = `awsZone`
	TableAwsRecord = `awsRecord`
)

type AwsZone struct {
	IsTruncated bool          `json:"IsTruncated"`
	MaxItems    string        `json:"MaxItems"`
	NextMarker  string        `json:"NextMarker"`
	HostedZones []AwsZoneView `json:"HostedZones"`
}

type AwsZoneView struct {
	CallerReference        string                 `json:"CallerReference"`
	Id                     string                 `json:"Id"`
	Name                   string                 `json:"Name"`
	ResourceRecordSetCount int                    `json:"ResourceRecordSetCount"`
	Config                 map[string]interface{} `json:"Config"`
}

func (a *AwsZoneView) CreateTableQuery() string {
	e := reflect.ValueOf(a).Elem()
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
	createQuery := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(%s)`, TableAwsZone, tableFields)
	return createQuery
}

func (a *AwsZoneView) InsertDataQuery() (string, []interface{}) {
	e := reflect.ValueOf(a).Elem()
	insertQuery := fmt.Sprintf(`INSERT %s SET `, TableAwsZone)
	var insertValue []interface{}
	for i := 0; i < e.NumField(); i++ {
		varName := e.Type().Field(i).Name
		varValue := e.Field(i).Interface()
		if i >= 4 {
			varValue = fmt.Sprintf(`%v`, varValue)
		}
		insertQuery = insertQuery + varName + "=?,"
		insertValue = append(insertValue, varValue)
	}
	insertQuery = strings.TrimRight(insertQuery, ",")
	return insertQuery, insertValue
}

func (a *AwsZoneView) GetZones(keyId, key string) ([]string, []AwsZoneView) {
	var zone AwsZone
	var id []string
	req := AwsRequest(keyId, key)
	param := &route53.ListHostedZonesInput{
		MaxItems: aws.String("1000"),
	}
	for zone.IsTruncated || zone.NextMarker == "" {
		if zone.IsTruncated {
			param = &route53.ListHostedZonesInput{
				Marker: aws.String(zone.NextMarker),
			}
		}
		resp, err := req.ListHostedZones(param)
		if err != nil {
			log.Println(err)
			return nil, nil
		}
		content, err := json.Marshal(resp)
		if err != nil {
			log.Println(err)
			return nil, nil
		}
		err = json.Unmarshal(content, &zone)
		if err != nil {
			log.Println(err)
			return nil, nil
		}
		for i := range zone.HostedZones {
			id = append(id, zone.HostedZones[i].Id)
		}
	}
	return id, zone.HostedZones
}

type AwsRecord struct {
	ResourceRecordSets   []AwsRecordView `json:"ResourceRecordSets"`
	IsTruncated          bool            `json:"IsTruncated"`
	MaxItems             string          `json:"MaxItems"`
	NextRecordIdentifier string          `json:"NextRecordIdentifier"`
	NextRecordName       string          `json:"NextRecordName"`
	NextRecordType       string          `json:"NextRecordType"`
}

type AwsRecordView struct {
	Name                    string              `json:"Name"`
	TTL                     int                 `json:"TTL"`
	Type                    string              `json:"Type"`
	Weight                  int                 `json:"Weight"`
	ResourceRecords         []map[string]string `json:"ResourceRecords"`
	GeoLocation             map[string]string   `json:"GeoLocation"`
	AliasTarget             string              `json:"AliasTarget"`
	Failover                string              `json:"Failover"`
	HealthCheckId           string              `json:"HealthCheckId"`
	MultiValueAnswer        string              `json:"MultiValueAnswer"`
	Region                  string              `json:"Region"`
	SetIdentifier           string              `json:"SetIdentifier"`
	TrafficPolicyInstanceId string              `json:"TrafficPolicyInstanceId"`
}

func (a *AwsRecordView) CreateTableQuery() string {
	e := reflect.ValueOf(a).Elem()
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
	createQuery := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(%s)`, TableAwsRecord, tableFields)
	return createQuery
}

func (a *AwsRecordView) InsertDataQuery() (string, []interface{}) {
	e := reflect.ValueOf(a).Elem()
	insertQuery := fmt.Sprintf(`INSERT %s SET `, TableAwsRecord)
	var insertValue []interface{}
	for i := 0; i < e.NumField(); i++ {
		varName := e.Type().Field(i).Name
		varValue := e.Field(i).Interface()
		if i >= 4 && i <= 5 {
			varValue = fmt.Sprintf(`%v`, varValue)
		}
		insertQuery = insertQuery + varName + "=?,"
		insertValue = append(insertValue, varValue)
	}
	insertQuery = strings.TrimRight(insertQuery, ",")
	return insertQuery, insertValue
}

func (a *AwsRecordView) GetRecords(keyId, key, zoneId string) []AwsRecordView {
	req := AwsRequest(keyId, key)
	param := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneId),
		// MaxItems:              aws.String("100"),
		// StartRecordIdentifier: aws.String("update"),
	}
	var record AwsRecord
	resp, err := req.ListResourceRecordSets(param)
	if err != nil {
		log.Println(err)
		return nil
	}
	content, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
		return nil
	}
	err = json.Unmarshal(content, &record)
	if err != nil {
		log.Println(err)
		return nil
	}
	return record.ResourceRecordSets
	/* Replace \052 with "*" */
	// if strings.Contains(view.Name, `\052`) {
	// 	m := regexp.MustCompile(`\\052`)
	// 	str := "*"
	// 	m.ReplaceAllString(view.Name, str)
	// }
}

func AwsRequest(keyId, key string) (client *route53.Route53) {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(keyId, key, ""),
	})
	if err != nil {
		log.Println("Session error,")
		log.Println(err)
		return
	}
	client = route53.New(sess)
	return
}
