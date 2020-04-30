package main

import (
	"flag"
	"liu5140/redis-select-ip/geo"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var DSN string
var REDISDSN string

func main() {
	flag.StringVar(&DSN, "dsn", "", "data source name")
	flag.StringVar(&REDISDSN, "redisdsn", "", "redis source name")

	flag.Parse()

	log.Print("dsn:" + DSN)
	log.Print("redisdsn:" + REDISDSN)

	dataLoad := &geo.DataLoad{}

	dataLoad.Load(DSN, REDISDSN)

	log.Println("Finished")
}
