/*
몽고드라이버 이외에 외부 패키지 사용 금지
데이터베이스 연결/해제 그리고 컬렉션 인스턴스만 관리하고
실제 데이터조작(CRUD)은 이 패키지를 사용하는 곳에서 할 것.
이유는 에러에 대한 처리가 편리하기 때문.
*/
package mdb

import (
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type M map[string]any //Map
type D primitive.D    //Document
type A primitive.A    //Array

// 몽고DB 유틸리티 구조체
type DB struct {
	clientOption *options.ClientOptions
	client       *mongo.Client
	database     map[string]*mongo.Database
	collection   map[string]*mongo.Collection
	dmu          *sync.Mutex //database map 보호용 뮤텍스
	cmu          *sync.Mutex //colleciton map 보호용 뮤텍스
}

// DBX 생성
func GetInstance() *DB {
	return &DB{
		dmu: &sync.Mutex{},
		cmu: &sync.Mutex{},
	}
}

// 데이터베이스 접속
// 실패시 프로세스 종료
func (dbx *DB) Open(addr string, cto, sto, ssto time.Duration) {

	//연결
	dbx.clientOption = options.Client().ApplyURI(addr)
	dbx.clientOption.SetConnectTimeout(cto)
	dbx.clientOption.SetSocketTimeout(sto)
	dbx.clientOption.SetServerSelectionTimeout(ssto)

	c, err := mongo.Connect(context.TODO(), dbx.clientOption)
	if err != nil {
		panic(err)
	}

	//테스트
	err = c.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}

	//멤버초기화
	dbx.client = c
	dbx.database = make(map[string]*mongo.Database)
	dbx.collection = make(map[string]*mongo.Collection)

}

// DB 접속종료. 서버 셧다운 시도이므로 발생하는 에러는 무시함.
func (dbx *DB) Close() {
	if dbx.client == nil {
		return
	}
	dbx.client.Disconnect(context.TODO())
}

// 클라이언트
func (dbx *DB) Client() *mongo.Client {
	return dbx.client
}

// 클라이언트 옵션 정보
func (dbx *DB) GetClientOption() *options.ClientOptions {
	return dbx.clientOption
}

// dname 은 데이터베이스 이름
func (dbx *DB) Database(dname string) *mongo.Database {
	dbx.dmu.Lock()
	defer dbx.dmu.Unlock()

	if dbx.database[dname] == nil {
		dbx.database[dname] = dbx.client.Database(dname)
	}
	return dbx.database[dname]
}

// dname 은 데이터베이스 이름, cname 은 컬렉션 이름
func (dbx *DB) Collection(dname, cname string) *mongo.Collection {
	db := dbx.Database(dname)
	cstr := dname + "." + cname

	dbx.cmu.Lock()
	defer dbx.cmu.Unlock()

	if dbx.collection[cstr] == nil {
		dbx.collection[cstr] = db.Collection(cname)
	}
	return dbx.collection[cstr]
}
