package gd

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func DoRequest(req *http.Request) (content []byte, err error) {
	var client = &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Response error.")
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		content, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	return
}

func TelegramSendMessage(token, chatId, msg string) string {
	return fmt.Sprintf(`https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s`, token, chatId, msg)
}

/* For debug West use */
func WestDoRequest(req *http.Request) (content []byte, err error) {
	var client = &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Response error.")
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		/* Convert GBK to UTF-8 */
		reader := simplifiedchinese.GB18030.NewDecoder().Reader(resp.Body)
		content, err = ioutil.ReadAll(reader)
		if err != nil {
			fmt.Println("Content error.")
			fmt.Println(err)
			return
		}
	}
	return
}

func Md5encode(v string) string {
	m := md5.Sum([]byte(v))
	return hex.EncodeToString(m[:])
}

/* Use a as Baseline*/
func CompareString(a, b []string) []string {
	for _, av := range a {
		for i, bv := range b {
			if av == bv {
				RemoveString(b, i)
				b[len(b)-1] = ""
				b = b[:len(b)-1]
				break
			}
		}
	}
	return b
}

func RemoveString(s []string, i int) []string {
	return append(s[:i], s[i+1:]...)
}

type Sql struct {
	Db        *sql.DB
	Err       error
	dbName    string
	dsnNormal string
	dsnRoot   string
}

func (my *Sql) getDSN() {
	user := viper.GetString("mysql.user")
	pass := viper.GetString("mysql.pass")
	host := viper.GetString("mysql.host")
	port := viper.GetString("mysql.port")
	my.dbName = viper.GetString("mysql.schema")
	my.dsnRoot = fmt.Sprintf("%s:%s@tcp(%s:%s)/",
		user, pass, host, port)
	my.dsnNormal = fmt.Sprintf("%s%s",
		my.dsnRoot, my.dbName)
}

func (my *Sql) Connection() *sql.DB {
	/* Create connection pool */
	my.getDSN()
	var db *sql.DB
	db, my.Err = sql.Open("mysql", my.dsnRoot)
	if my.Err != nil {
		fmt.Println(my.Err)
		return nil
	}
	/* Create database */
	_, my.Err = db.Exec("CREATE DATABASE IF NOT EXISTS " + my.dbName)
	if my.Err != nil {
		fmt.Println(my.Err)
		return nil
	}
	db.Close()
	/* Re-Connect to MySQL */
	my.Db, my.Err = sql.Open("mysql", my.dsnNormal)
	if my.Err != nil {
		fmt.Println(my.Err)
		return nil
	}
	// defer my.Db.Close()
	return my.Db
}

func (my *Sql) CreateTable(query string) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelfunc()
	/* Create Table */
	_, err := my.Db.ExecContext(ctx, query)
	if err != nil {
		fmt.Println(err)
	}
	defer my.Db.Close()
}

func (my *Sql) DropTable(tableName string) {
	query := "DROP TABLE IF EXISTS " + tableName
	my.CreateTable(query)
}

func (my *Sql) QueryData(query string) (err error) {
	_, err = my.Db.Query(query)
	if err != nil {
		fmt.Println(err)
	}
	defer my.Db.Close()
	return
}

func (my *Sql) InsertData(query string, val []interface{}) (err error) {
	/* Insert data */
	ins, err := my.Db.Prepare(query)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = ins.Exec(val...)
	if err != nil {
		fmt.Println(err)
	}
	defer ins.Close()
	return
}

func (my *Sql) QueryWestDomain(query string) (result []string, err error) {
	rows, err := my.Db.Query(query)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			fmt.Println(err)
			return nil, err
		}
		result = append(result, domain)
	}
	return
}
