package db

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	//MySQL Driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr"
)

//Connect return a connection to the database
func Connect() (*dbr.Connection, error) {
	dbUser := os.Getenv("FMT_DBUSER")
	dbPass := os.Getenv("FMT_DBPASS")
	dbHost := os.Getenv("FMT_DBHOST")
	dbName := os.Getenv("FMT_DBNAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbName)

	conn, err := dbr.Open("mysql", dsn, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

type valid struct {
	Total int `json:"total" db:"total"`
}

//Validate executes a select query and return a integer Total
func Validate(session *dbr.Session, query string, values ...interface{}) (int, error) {
	v := valid{Total: 0}
	_, err := session.SelectBySql(query, values).Load(v)
	if err != nil {
		return -1, err
	}
	return v.Total, nil
}

//QueryOne load one record from the database
func QueryOne(session *dbr.Session, table string, id string, object interface{}) (interface{}, error) {
	objectMetadata := parseObjectTagsRecursively("", table, object)

	params := map[string]string{
		"id": id,
	}
	result, err := loadGeneric(session, table, params, object, objectMetadata)
	if err != nil {
		return nil, err
	}

	if len(result) > 0 {
		return result[0], nil
	}
	return nil, errors.New("invalid id, record not found")
}

//Select load records from the database
func Select(session *dbr.Session, table string, params map[string]string, object interface{}) (interface{}, error) {
	objectMetadata := parseObjectTagsRecursively("", table, object)

	var dbresult dbResult
	tableMetadata, err := loadTableMetadata(session, table, params, objectMetadata)
	if err != nil {
		dbresult.Errors = append(dbresult.Errors, err)
	}
	dbresult.Metadata = tableMetadata

	if tableMetadata.TotalFiltered == 0 {
		dbresult.Data = []map[string]interface{}{}
		return dbresult, nil
	}

	result, err := loadGeneric(session, table, params, object, objectMetadata)
	if err != nil {
		dbresult.Errors = append(dbresult.Errors, err)
	}
	dbresult.Data = result

	return dbresult, nil
}

//Insert insert an new object in the database
func Insert(tx *dbr.Tx, table string, object interface{}) error {
	_, err := tx.InsertInto(table).Columns(getTagFromInterface(object, "db", "")...).Record(object).Exec()
	if err != nil {
		return err
	}

	t := reflect.TypeOf(object)
	v := reflect.ValueOf(object)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Type().Name() == "Translation" && t.Field(i).Tag.Get("persist") != "" {
			translationObject := field.Interface()
			_, err = tx.InsertInto(TableTranslation).Columns(getTagFromInterface(translationObject, "db", "")...).Record(translationObject).Exec()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//Update change record attributes in the database
func Update(tx *dbr.Tx, table, id string, object interface{}, values map[string]interface{}) error {
	objectMap := parseObjectFieldsToUpdatableMap("", object, values)
	if len(objectMap) > 0 {
		_, err := tx.Update(table).SetMap(objectMap).Where(dbr.Eq(table+".id", id)).Exec()
		if err != nil {
			return err
		}
	}

	t := reflect.TypeOf(object)
	v := reflect.ValueOf(object)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		tagPersist := t.Field(i).Tag.Get("persist")
		tagAlias := t.Field(i).Tag.Get("alias")

		if field.Type().Name() == "Translation" && tagPersist != "" {
			translationObject := field.Interface()
			objectMap := parseObjectFieldsToUpdatableMap(tagAlias, translationObject, values)
			if len(objectMap) > 0 {
				_, err := tx.Update(TableTranslation).SetMap(objectMap).Where(dbr.And(
					dbr.Eq("parent_id", id),
					dbr.Eq("field", tagAlias),
				)).Exec()
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

//Delete a record from the database
func Delete(session *dbr.Session, table string, ids ...string) error {
	tx, err := session.Begin()
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	_, err = tx.DeleteFrom(table).Where("id IN ?", ids).Exec()
	if err != nil {
		return err
	}
	_, err = tx.DeleteFrom(TableTranslation).Where("parent_id IN ?", ids).Exec()
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}
