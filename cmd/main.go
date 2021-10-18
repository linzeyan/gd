package main

import (
	"flag"
	"fmt"
	"os"
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
	QueryWestDomainHold string = `select Domain from westZone where Hold = "0"`
)

var (
	ConfigFile = flag.String("c", "", "Specify Config file")
	domain     = flag.String("d", "", "Specify domain")
	operator   = flag.String("o", "nothing", "Fetch DNS once or hourly")
	mySql      = new(gd.Sql)
)

func readConf() {
	if *ConfigFile != "" {
		viper.SetConfigType("yaml")
		viper.SetConfigFile(*ConfigFile)
	} else {
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME")
		viper.AddConfigPath("/etc")
		viper.SetConfigName("dns.yaml")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	viper.WatchConfig()
}

func fetch() {
	readConf()
	mySql.Connection()
	/* AWS */
	awsId := viper.GetString("aws.id")
	awsKey := viper.GetString("aws.key")
	aws(awsId, awsKey)
	/* Cloudflare */
	cfApi := viper.GetString("cloudflare.api")
	cfKey := viper.GetString("cloudflare.key")
	cfMail := viper.GetString("cloudflare.mail")
	cloudflare(cfApi, cfKey, cfMail)
	/* West Digital */
	westApi := viper.GetString("west.api")
	westAccount := viper.GetString("west.account")
	westKey := viper.GetString("west.key")
	list := viper.GetStringSlice("west.hold")
	token := viper.GetString("telegram.token")
	chatId := viper.GetString("telegram.chatid")
	west(westApi, westAccount, westKey)
	domains, err := mySql.QueryWestDomain(QueryWestDomainHold)
	if err != nil {
		fmt.Println(err)
		return
	}
	checkWestDomain(list, domains, token, chatId)
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

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

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
