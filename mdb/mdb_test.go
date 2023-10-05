package mdb

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/hottguy/3go/cfg"
)

var conf = cfg.GetInstance("../conf/config.json")

func TestXxx(t *testing.T) {
	mdbx := GetInstance()
	mdbx.Open(
		"mongodb://localhost:27017/",
		time.Duration(conf.GetInt("MDB_ConnectTimeout"))*time.Millisecond,
		time.Duration(conf.GetInt("MDB_SocketTimeout"))*time.Millisecond,
		time.Duration(conf.GetInt("MDB_ServerSelectionTimeout"))*time.Millisecond,
	)
	defer mdbx.Close()

	coll := mdbx.Collection(conf.GetString("MDB_Name"), "test")
	coll.InsertOne(context.TODO(), M{"name": "강아지"})
	r := coll.FindOne(context.TODO(), M{})
	m := M{}
	r.Decode(m)
	log.Printf("%+v", m)
}
