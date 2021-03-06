package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/linzeyan/gd"
	"github.com/spf13/viper"
)

const (
	usage = `Get DNS Zones and Records

Usage: gd [option] {once|hourly}

Options:
`
)

var (
	configFile        = flag.String("c", "", "Specify Config file")
	domain            = flag.String("d", "", "Specify domain")
	operator          = flag.String("o", "nothing", "Fetch DNS or check ICP status.(once, hourly, icp)")
	version           = flag.Bool("v", false, "Print version info")
	mySql             = gd.NewDB()
	queryWestZoneHold = fmt.Sprintf(`select Domain from %s where Hold = "0"`, gd.TableWestZone)
)

var appVersion, appBuildTime, appCommit, appGoVersion, appPlatform string

func aws(keyId, key string, wg *sync.WaitGroup) {
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
	wg.Done()
}

func cloudflare(api, key, mail string, wg *sync.WaitGroup) {
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
	wg.Done()
}

func west(api, account, key string, wg *sync.WaitGroup) {
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
	wg.Done()
}

/* Check West newly lock domains */
func checkWestDomain(list, domains []string, token, chatId string) {
	diff := gd.CompareString(list, domains)
	if len(diff) > 0 {
		log.Println(diff)
		msg := fmt.Sprintf("%v was hold by West", diff)
		uri := gd.TelegramSendMessage(token, chatId, msg)
		req, err := http.NewRequest(http.MethodPost, uri, strings.NewReader(``))
		if err != nil {
			log.Println("Resquest error.")
			log.Println(err)
			return
		}
		gd.DoRequest(req)
	}
}

type Icp struct {
	Domain    string `json:"domain"`
	Icp       string `json:"icp"`
	IcpStatus string `json:"icpstatus"`
}

/* Check ICP status using West api */
func checkWestIcp(api, account, key, domain, token, chatId string) {
	var hash_data string = account + key + "domainname"
	sig := gd.Md5encode(hash_data)
	rawCmd := fmt.Sprintf("domainname\r\ncheck\r\nentityname:icp\r\ndomains:%s\r\n.\r\n", domain)
	strCmd := url.QueryEscape(rawCmd)
	uri := fmt.Sprintf(`%s/?userid=%s&strCmd=%s&versig=%s`, api, account, strCmd, sig)
	req, err := http.NewRequest(http.MethodPost, uri, strings.NewReader(``))
	if err != nil {
		log.Println("Resquest error.")
		log.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	content, err := gd.WestDoRequest(req)
	if err != nil {
		log.Println(err)
		return
	}
	re, _ := regexp.Compile("{.*}")
	match := fmt.Sprintln(re.FindString(string(content)))
	var icp Icp
	json.Unmarshal([]byte(match), &icp)
	msg := icp.Domain + ":" + icp.IcpStatus
	log.Println(msg)
	tgUri := gd.TelegramSendMessage(token, chatId, msg)
	tgReq, err := http.NewRequest(http.MethodPost, tgUri, strings.NewReader(``))
	if err != nil {
		log.Println("Resquest error.")
		log.Println(err)
		return
	}
	gd.DoRequest(tgReq)
}

func readConf() {
	if *configFile != "" {
		viper.SetConfigType("yaml")
		viper.SetConfigFile(*configFile)
	} else {
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME")
		viper.AddConfigPath("/etc")
		viper.SetConfigName("dns.yaml")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Println(err)
		os.Exit(2)
	}
	viper.WatchConfig()
}

func fetch() {
	log.Println("Start program")
	readConf()
	mySql.Connection()
	/* AWS */
	awsId := viper.GetString("aws.id")
	awsKey := viper.GetString("aws.key")
	/* Cloudflare */
	cfApi := viper.GetString("cloudflare.api")
	cfKey := viper.GetString("cloudflare.key")
	cfMail := viper.GetString("cloudflare.mail")
	/* West Digital */
	westApi := viper.GetString("west.api")
	westAccount := viper.GetString("west.account")
	westKey := viper.GetString("west.key")
	list := viper.GetStringSlice("west.hold")
	token := viper.GetString("telegram.token")
	chatId := viper.GetString("telegram.chatid")

	var wg sync.WaitGroup
	wg.Add(3)
	go aws(awsId, awsKey, &wg)
	go cloudflare(cfApi, cfKey, cfMail, &wg)
	go west(westApi, westAccount, westKey, &wg)
	wg.Wait()
	log.Println("Records fetch completed")
	domains, err := mySql.QueryWestDomain(queryWestZoneHold)
	if err != nil {
		log.Println(err)
		return
	}
	checkWestDomain(list, domains, token, chatId)
	log.Println("Domain check completed")
}

func cron() {
	job := gocron.NewScheduler(time.Local)
	job.Cron("0 * * * *").Do(fetch)
	job.StartAsync()
	select {}
}

func icp() {
	readConf()
	westApi := viper.GetString("west.icp")
	westAccount := viper.GetString("west.account")
	westKey := viper.GetString("west.key")
	token := viper.GetString("telegram.token")
	chatId := viper.GetString("telegram.chatid")
	checkWestIcp(westApi, westAccount, westKey, *domain, token, chatId)
}

func versionInfo() {
	fmt.Printf(`{"Version":"%s","BuildTime":"%s","GitCommit":"%s","GoVersion":"%s","Platform":"%s"}`,
		appVersion, appBuildTime, appCommit, appGoVersion, appPlatform)
	os.Exit(0)
}

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	if *version {
		versionInfo()
	}

	switch *operator {
	case "once":
		fetch()
	case "hourly":
		cron()
	case "icp":
		icp()
	default:
		fmt.Print(usage)
		flag.PrintDefaults()
	}
}
