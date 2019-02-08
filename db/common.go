package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr"
	"github.com/gocraft/dbr/dialect"
)

const (
	recorsPerPage = 50
)

const (
	//TableHighlight defines the highlights entities database table
	TableHighlight = "highlight"
	//TableHighlightImage defines the highlights images entities database table
	TableHighlightImage = "highlight_image"
	//TableTrip defines the trips entities database table
	TableTrip = "trip"
	//TableTripParticipant defines the trip participants entities database table
	TableTripParticipant = "trip_participant"
	//TableTripInvite defines the trip invites entities database table
	TableTripInvite = "trip_invite"
	//TableTripItinerary defines the trip itineraries entities database table
	TableTripItinerary = "trip_itinerary"
	//TableTripItineraryEvent defines the trip itinerary events entities database table
	TableTripItineraryEvent = "trip_itinerary_event"
	//TableEvent defines the events entities database table
	TableEvent = "event"
	//TableEventSchedule defines the events schedule entities database table
	TableEventSchedule = "event_schedule"
	//TableCategory defines the category entities database table
	TableCategory = "category"
	//TableLocation defines the location entities database table
	TableLocation = "location"
	//TableUser defines the users entities database table
	TableUser = "user"
	//TableTranslation defines the translation entities database table
	TableTranslation = "translation"
)

type dbResult struct {
	Metadata metadata                 `json:"metadata"`
	Data     []map[string]interface{} `json:"data"`
	Errors   []error                  `json:"errors"`
}

type metadata struct {
	Page          int    `json:"page"`
	Total         int    `json:"total" db:"total"`
	TotalFiltered int    `json:"total_filtered" db:"total_filtered"`
	RecorsPerPage int    `json:"records_per_page"`
	Source        string `json:"source"`
}

type joinConfig struct {
	table string
	alias string
	on    string
}

func getTagFromInterface(object interface{}, tag, alias string) []string {
	results := []string{}
	tagAlias := alias
	if tagAlias != "" {
		tagAlias += "."
	}
	t := reflect.TypeOf(object)
	for i := 0; i < t.NumField(); i++ {
		dbTag := t.Field(i).Tag.Get(tag)
		if dbTag != "" {
			results = append(results, tagAlias+dbTag)
		}
	}
	return results
}

func checkFiltersInParams(params map[string]string) bool {
	if _, ok := params["filter"]; ok {
		return true
	}
	for k := range params {
		if k != "page" && k != "results" {
			return true
		}
	}
	return false
}

func loadTableMetadata(session *dbr.Session, table string, params map[string]string, meta objectMetadata) (metadata, error) {
	var m metadata
	_, err := session.Select("count(id) total").From(table).Load(&m)
	m.TotalFiltered = m.Total

	if checkFiltersInParams(params) {
		stmt := session.Select("count(" + table + ".id) total_filtered").From(table)
		if len(meta.joins) > 0 {
			for _, j := range meta.joins {
				if j.alias != "" {
					stmt.LeftJoin(dbr.I(j.table).As(j.alias), j.on)
				} else {
					stmt.LeftJoin(j.table, j.on)
				}
			}
		}

		//TODO: extract this filter code to a function
		var filterOr dbr.Builder
		hasFilter := false
		if val, ok := params["filter"]; ok {
			whereFilters := []dbr.Builder{}
			for _, f := range meta.filters {
				whereFilters = append(whereFilters, dbr.Expr("LOWER("+f+") like LOWER('%"+val+"%')"))
			}
			filterOr = dbr.Or(whereFilters...)
			hasFilter = true
		}

		otherFilters := []dbr.Builder{}
		for k, v := range params {
			if k != "filter" && k != "page" && k != "results" && k != "id" && k != "order" && k != "sort" {
				column := table + "." + k
				filter := dbr.Eq(column, v)
				if v == "is_not_null" {
					filter = dbr.And(
						dbr.Neq(column, nil),
						dbr.Neq(column, ""),
					)
				} else if v == "is_null" {
					filter = dbr.Or(
						dbr.Eq(column, nil),
						dbr.Eq(column, ""),
					)
				}
				otherFilters = append(otherFilters, filter)
			}
		}

		if hasFilter && len(otherFilters) == 0 {
			stmt.Where(filterOr)
		} else if !hasFilter && len(otherFilters) > 0 {
			stmt.Where(dbr.And(otherFilters...))
		} else if hasFilter && len(otherFilters) > 0 {
			stmt.Where(dbr.And(filterOr, dbr.And(otherFilters...)))
		}

		fm := metadata{}
		_, err := stmt.Load(&fm)
		if err != nil {
			fmt.Println(err.Error())
			return m, err
		}

		m.TotalFiltered = fm.TotalFiltered
	}
	m.Source = table
	m.RecorsPerPage = recorsPerPage
	if val, ok := params["results"]; ok {
		i, _ := strconv.Atoi(val)
		m.RecorsPerPage = i
	}
	m.Page = 1
	if val, ok := params["page"]; ok {
		i, _ := strconv.Atoi(val)
		m.Page = i
	}
	return m, err
}

func loadGeneric(sess *dbr.Session, table string, params map[string]string, object interface{}, meta objectMetadata, ids []string) ([]map[string]interface{}, error) {

	stmt := sess.Select(meta.columns...).From(table)
	if len(meta.joins) > 0 {
		for _, j := range meta.joins {
			if j.alias != "" {
				stmt.LeftJoin(dbr.I(j.table).As(j.alias), j.on)
			} else {
				stmt.LeftJoin(j.table, j.on)
			}
		}
	}
	if len(meta.groupByColumns) > 0 {
		stmt.GroupBy(meta.groupByColumns...)
	}

	if val, ok := params["id"]; ok {
		stmt.Where(dbr.Eq(table+".id", val))
	}

	//TODO: extract this filter code to a function
	var filterOr dbr.Builder
	hasFilter := false
	if val, ok := params["filter"]; ok {
		whereFilters := []dbr.Builder{}
		for _, f := range meta.filters {
			whereFilters = append(whereFilters, dbr.Expr("LOWER("+f+") like LOWER('%"+val+"%')"))
		}
		filterOr = dbr.Or(whereFilters...)
		hasFilter = true
	}

	otherFilters := []dbr.Builder{}
	for k, v := range params {
		if k != "filter" && k != "page" && k != "results" && k != "id" && k != "order" && k != "sort" {
			column := table + "." + k
			filter := dbr.Eq(column, v)
			if v == "is_not_null" {
				filter = dbr.And(
					dbr.Neq(column, nil),
					dbr.Neq(column, ""),
				)
			} else if v == "is_null" {
				filter = dbr.Or(
					dbr.Eq(column, nil),
					dbr.Eq(column, ""),
				)
			}
			otherFilters = append(otherFilters, filter)
		}
	}

	if hasFilter && len(otherFilters) == 0 {
		stmt.Where(filterOr)
	} else if !hasFilter && len(otherFilters) > 0 {
		stmt.Where(dbr.And(otherFilters...))
	} else if hasFilter && len(otherFilters) > 0 {
		stmt.Where(dbr.And(filterOr, dbr.And(otherFilters...)))
	}

	if len(ids) > 0 {
		idsInterface := make([]interface{}, len(ids))
		for i, v := range ids {
			idsInterface[i] = v
		}
		stmt.Where(table+".id IN ?", idsInterface...)
	}

	if col, ok := params["sort"]; ok {
		column := table + "." + col
		if val, ok := params["order"]; ok {
			if val == "asc" {
				stmt.OrderAsc(column)
			} else {
				stmt.OrderDesc(column)
			}
		} else {
			stmt.OrderDesc(column)
		}
	}

	if val, ok := params["page"]; ok {
		page, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			page = 1
		}
		rpp := recorsPerPage
		if val, ok := params["results"]; ok {
			i, err := strconv.Atoi(val)
			if err != nil {
				rpp = recorsPerPage
			} else {
				rpp = i
			}
		}
		rppUint64, _ := strconv.ParseUint(strconv.Itoa(rpp), 10, 64)
		stmt.Paginate(page, rppUint64)
	}

	rows, err := stmt.Rows()
	if err != nil {
		return nil, err
	}

	tt, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	types := make([]reflect.Type, len(tt))
	for i, tp := range tt {
		st := tp.ScanType()
		if st == nil {
			//TODO: log -> fmt.Printf("scantype is null for column %q", tp.Name())
			continue
		}
		types[i] = st
	}

	result := make([]interface{}, len(meta.fields))
	for i := range result {
		result[i] = reflect.New(types[i]).Interface()
	}

	results := []map[string]interface{}{}

	for rows.Next() {
		err := rows.Scan(result...)
		if err != nil {
			return nil, err
		}

		if getResultValue(result[0], types[0].String()) == nil {
			return results, nil
		}

		mapJSON := make(map[string]interface{})
		for i := 0; i < len(meta.fields); i++ {
			f := meta.fields[i]
			if strings.Contains(f, ".") {
				parentField := f[0:strings.Index(f, ".")]
				embeddedResult := make(map[string]interface{})
				index := i
				for index < len(meta.fields) && strings.Contains(meta.fields[index], ".") && parentField == meta.fields[index][0:strings.Index(meta.fields[index], ".")] {
					key := meta.fields[index][strings.Index(meta.fields[index], ".")+1:]
					embeddedResult[key] = getResultValue(result[index], types[index].String())
					if embeddedResult[key] == nil {
						fmt.Println("Field: " + key)
					}
					index++
				}
				mapJSON[parentField] = embeddedResult
				i = index - 1
			} else {
				mapJSON[f] = getResultValue(result[i], types[i].String())
				if mapJSON[f] == nil {
					fmt.Println("Field: " + f)
				}
			}
		}
		results = append(results, mapJSON)
	}
	rows.Close()

	return results, err
}

func getResultValue(result interface{}, interfaceType string) interface{} {
	switch interfaceType {
	case "sql.RawBytes":
		return string(*result.(*sql.RawBytes))
	case "int8":
		return *result.(*int8)
	case "int16":
		return *result.(*int16)
	case "int32":
		return *result.(*int32)
	case "int64":
		return *result.(*int64)
	case "mysql.NullTime":
		nt := *result.(*mysql.NullTime)
		return nt.Time
	default:
		fmt.Println("Type: " + interfaceType)
		return nil
	}
}

type objectMetadata struct {
	columns        []string
	groupByColumns []string
	fields         []string
	filters        []string
	joins          []joinConfig
	aggregation    bool
}

func parseObjectTagsRecursively(alias, table string, object interface{}) objectMetadata {
	data := objectMetadata{}

	t := reflect.TypeOf(object)
	v := reflect.ValueOf(object)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("embedded") != "" {
			if field.Tag.Get("table") != "" {
				config := joinConfig{
					table: field.Tag.Get("table"),
					alias: field.Tag.Get("alias"),
					on:    field.Tag.Get("on"),
				}
				data.joins = append(data.joins, config)
			}
			embeddedObjectMetadata := parseObjectTagsRecursively(field.Tag.Get("alias"), field.Tag.Get("table"), v.Field(i).Interface())
			data.columns = append(data.columns, embeddedObjectMetadata.columns...)
			data.groupByColumns = append(data.groupByColumns, embeddedObjectMetadata.groupByColumns...)
			data.fields = append(data.fields, embeddedObjectMetadata.fields...)
			data.joins = append(data.joins, embeddedObjectMetadata.joins...)
			data.filters = append(data.filters, embeddedObjectMetadata.filters...)
			data.aggregation = embeddedObjectMetadata.aggregation
		} else {
			if field.Tag.Get("db") != "" {
				col := table + "." + field.Tag.Get("db")
				if alias != "" {
					col = alias + "." + field.Tag.Get("db")
				}
				data.columns = append(data.columns, col)
				data.groupByColumns = append(data.groupByColumns, col)
				if field.Tag.Get("filter") != "" {
					data.filters = append(data.filters, col)
				}
			}
			if field.Tag.Get("json") != "" && field.Tag.Get("embedded") == "" {
				val := field.Tag.Get("json")
				if alias != "" {
					val = alias + "." + val
				}
				data.fields = append(data.fields, val)
			}
			if field.Tag.Get("aggr") != "" {
				data.columns = append(data.columns, field.Tag.Get("aggr"))
				data.aggregation = true
			}
			if field.Tag.Get("table") != "" {
				config := joinConfig{
					table: field.Tag.Get("table"),
					alias: field.Tag.Get("alias"),
					on:    field.Tag.Get("on"),
				}
				data.joins = append(data.joins, config)
			}
		}
	}
	if !data.aggregation {
		data.groupByColumns = []string{}
	}
	return data
}

func parseObjectFieldsToUpdatableMap(alias string, object interface{}, values map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{}

	t := reflect.TypeOf(object)
	v := reflect.ValueOf(object)

	if alias != "" {
		alias = alias + "."
	}

	for i := 0; i < v.NumField(); i++ {
		tagLock := t.Field(i).Tag.Get("lock")
		tagDB := t.Field(i).Tag.Get("db")
		tagJSON := t.Field(i).Tag.Get("json")
		value, ok := values[alias+tagJSON]
		if ok && tagDB != "" && tagLock == "" {
			data[tagDB] = value
		}
	}

	return data
}

func debugSQLQuery(builder dbr.Builder) string {
	buf := dbr.NewBuffer()
	builder.Build(dialect.MySQL, buf)
	return buf.String()
}
