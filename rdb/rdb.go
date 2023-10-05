/*
MariaDB(MySQL) 에서 데이터베이스를 생성할 때 아래와 같이 인코딩타입을 선언해야
이후 테이블이나 프로시저에서 모든 컬럼마다 utf8을 명시하는 불편함이 사라진다.

CREATE DATABASE heyid CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
*/
package rdb

import (
	"database/sql"
	"fmt"
	"reflect"

	_ "github.com/go-sql-driver/mysql"
)

type Result [][]map[string]any

type RDB struct {
	db *sql.DB
}

func GetInstance() *RDB {
	return &RDB{}
}

func (rdbx *RDB) Open(dbName string, dataSourceName string) {
	db, err := sql.Open(dbName, dataSourceName)
	if err != nil {
		panic(err)
	}
	rdbx.db = db
}

func (rdbx *RDB) Close() {
	if rdbx.db != nil {
		rdbx.db.Close()
	}
}

/*
sql.DB.Query() 함수의 args 파라메터가 []any 이므로 리턴 타입을 맞춤
*/
func Convert(param ...string) (args []any, questions string) {

	comma := ""
	for _, s := range param {
		args = append(args, s)
		questions = questions + comma + "?"
		comma = ","
	}
	return
}

/*
프로시져명과 파라메터 목록을 전달하면
result set 배열을 리턴
즉, 하나의 프로시져에서 여러개의 select 문장을 수행할 수 있음.
*/
func (rdbx *RDB) Call(proc string, param ...string) (Result, error) {

	slice := Result{}

	// 프로시져 실행
	args, questions := Convert(param...)
	SQL := fmt.Sprintf("CALL %s(%s)", proc, questions)

	// 트랜잭션 시작
	tx, err := rdbx.db.Begin()
	if err != nil {
		return slice, err
	}

	//rows, err := rdbx.db.Query(SQL, args...)
	rows, err := tx.Query(SQL, args...)
	if err != nil {
		tx.Rollback()
		return slice, err
	}
	defer rows.Close()

	/*
		호출 결과 result set 배열을 완성하고 반환
		실행된 프로시져내에 select 문이 없다면 빈 배열을 반환
	*/
	for true {
		result := []map[string]any{}
		for rows.Next() {
			/*
				각 컬럼의 타입정보 슬라이스
			*/
			types, err := rows.ColumnTypes()
			if err != nil {
				tx.Rollback()
				return slice, err
			}

			/*
				sql.Rows.Scan() 함수의 파라메터는 포인터 변수이어야 한다.
				컬럼 갯수 만큼의 포인터를 담을 ptrs 슬라이스를 만들고
				실제 값을 담을 columns 슬라이스를 만든 후
				ptrs 엘리먼트들은 columns 엘리먼트의 포인터를 갖도록 한다.
			*/
			columns := make([]any, len(types))
			ptrs := make([]any, len(types))
			for i := 0; i < len(types); i++ {
				ptrs[i] = &columns[i]
			}

			/*
				준비된 포인터 슬라이스(ptrs)에 레코드의 각 필드 값이 Scan 되도록 함.
			*/
			if err := rows.Scan(ptrs...); err != nil {
				tx.Rollback()
				return slice, err
			}

			/*
				레코드의 각 필드명에 맞는 값으로 kv map을 채움.
				kv map 하나는 하나의 레코드와 대응.
			*/
			kv := map[string]any{}
			for i := 0; i < len(types); i++ {
				x := columns[i]

				if x == nil {
					kv[types[i].Name()] = nil
				} else if reflect.TypeOf(x).Kind() == reflect.Int64 {
					kv[types[i].Name()] = x
				} else {
					kv[types[i].Name()] = string(x.([]uint8))
				}
			}

			/*
				채워진 kv를 결과 배열에 붙임.
			*/
			result = append(result, kv)
		}

		/*
			result set을 상위 배열에 붙임.
		*/
		slice = append(slice, result)

		// 다음 result set 없으면 루프 종료
		if !rows.NextResultSet() {
			break
		}
	}
	// 트랜잭션 커밋
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return slice, err
	}
	return slice, nil
}
