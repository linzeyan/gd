package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/linzeyan/gd"
)

func aws(keyId, key string) {
	var zone gd.AwsZoneView
	mySql.DropTable(gd.TableAwsZone)
	mySql.CreateTable(zone.CreateTableQuery())
	id, resp := zone.GetZones(keyId, key)
	for i := range resp {
		mySql.InsertData(resp[i].InsertDataQuery())
	}

	var record gd.AwsRecordView
	mySql.DropTable(gd.TableAwsRecord)
	mySql.CreateTable(record.CreateTableQuery())
	for i := range id {
		data := record.GetRecords(keyId, key, id[i])
		for j := range data {
			mySql.InsertData(data[j].InsertDataQuery())
		}
	}
}

func cloudflare(api, key, mail string) {
	var zone gd.CloudflareZoneView
	mySql.DropTable(gd.TableCloudflareZone)
	mySql.CreateTable(zone.CreateTableQuery())
	id, resp := zone.GetZones(api, key, mail)
	for i := range resp {
		mySql.InsertData(resp[i].InsertData())
	}

	var record gd.CloudflareRecordView
	mySql.DropTable(gd.TableCloudflareRecord)
	mySql.CreateTable(record.CreateTableQuery())
	for i := range id {
		data := record.GetRecords(api, key, mail, id[i])
		for j := range data {
			mySql.InsertData(data[j].InsertData())
		}
	}
}

func west(api, account, key string) {
	var zone gd.WestZoneView
	mySql.DropTable(gd.TableWestZone)
	mySql.CreateTable(zone.CreateTableQuery())
	id, resp := zone.GetZones(api, account, key)
	for i := range resp {
		mySql.InsertData(resp[i].InsertData())
	}

	var record gd.WestRecordView
	mySql.DropTable(gd.TableWestRecord)
	mySql.CreateTable(record.CreateTableQuery())
	for i := range id {
		data := record.GetRecords(api, account, key, id[i])
		for j := range data {
			mySql.InsertData(data[j].InsertData())
		}
	}
}

func checkWestDomain(list []string, token, chatId string) {
	domains, err := mySql.QueryWestDomain(QueryWestDomainHold)
	if err != nil {
		fmt.Println(err)
		return
	}
	diff := gd.CompareString(list, domains)
	if len(diff) > 0 {
		fmt.Println(diff)
		msg := fmt.Sprintf("%v was hold by West", diff)
		uri := gd.TelegramSendMessage(token, chatId, msg)
		req, err := http.NewRequest("POST", uri, strings.NewReader(``))
		if err != nil {
			fmt.Println("Resquest error.")
			fmt.Println(err)
			return
		}
		gd.DoRequest(req)
	}
}