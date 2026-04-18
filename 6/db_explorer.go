package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type Column struct {
	Field      string
	Type       string
	Collation  sql.Null[string]
	Null       string
	Key        string
	Default    sql.Null[string]
	Extra      string
	Privileges string
	Comment    string
}

type ctxKey string

type ctxKeys struct {
	table ctxKey
	id    ctxKey
}

type Handler struct {
	DB           *sql.DB
	Keys         *ctxKeys
	tableColumns map[string][]string
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {

	return &Handler{
		DB: db,
		Keys: &ctxKeys{
			table: "table",
			id:    "id",
		}}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			h.ListTables(w, r)
		}
		return
	}

	switch len(parts) {
	case 1:
		ctx := context.WithValue(r.Context(), h.Keys.table, parts[0])
		r = r.WithContext(ctx)

		switch r.Method {
		case http.MethodPut:
			h.ValidateTableMiddleware(h.Add)(w, r)
		case http.MethodGet:
			h.ValidateTableMiddleware(h.ListRecords)(w, r)
		}
	case 2:
		ctx := context.WithValue(r.Context(), h.Keys.table, parts[0])
		ctx = context.WithValue(ctx, h.Keys.id, parts[1])
		r = r.WithContext(ctx)

		switch r.Method {
		case http.MethodGet:
			h.ValidateTableMiddleware(h.OneRecord)(w, r)
		case http.MethodPost:
			h.ValidateTableMiddleware(h.Update)(w, r)
		case http.MethodDelete:
			h.ValidateTableMiddleware(h.Delete)(w, r)
		}
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func sanitizeTableName(name string) (string, error) {
	if matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", name); !matched {
		return "", fmt.Errorf("invalid table name")
	}
	return name, nil
}

func (h *Handler) getListTables() ([]string, error) {
	rows, err := h.DB.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tableName string
	var tableNames []string

	for rows.Next() {
		err := rows.Scan(&tableName)
		if err != nil {
			return nil, err
		}
		tableNames = append(tableNames, tableName)
	}
	return tableNames, nil
}

func (h *Handler) getListColumns(nameTable string) (map[string]Column, error) {

	nameTable, err := sanitizeTableName(nameTable)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SHOW FULL COLUMNS FROM `%s`", nameTable)

	rows, err := h.DB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	columns := make(map[string]Column)

	for rows.Next() {
		var c Column
		if err := rows.Scan(
			&c.Field,
			&c.Type,
			&c.Collation,
			&c.Null,
			&c.Key,
			&c.Default,
			&c.Extra,
			&c.Privileges,
			&c.Comment,
		); err != nil {
			return nil, err
		}
		columns[c.Field] = c
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}

// ListTables return list of name tables
func (h *Handler) ListTables(w http.ResponseWriter, r *http.Request) {

	tableNames, err := h.getListTables()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"response": map[string]any{
			"tables": tableNames,
		},
	}

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) ValidateTableMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		nameTable, ok := r.Context().Value(h.Keys.table).(string)
		if !ok || nameTable == "" {
			http.Error(w, "table not specified", http.StatusBadRequest)
			return
		}

		nameTable, err = sanitizeTableName(nameTable)
		if err != nil {
			http.Error(w, "name table incorrect", http.StatusInternalServerError)
			return
		}

		tableNames, err := h.getListTables()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var found bool
		for _, name := range tableNames {
			if nameTable == name {
				found = true
				break
			}
		}

		if !found {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]any{"error": "unknown table"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ListRecords return list of records
func (h *Handler) ListRecords(w http.ResponseWriter, r *http.Request) {
	var err error

	// default settings
	limit := 5
	offset := 0

	if r.FormValue("limit") != "" {
		limit, err = strconv.Atoi(r.FormValue("limit"))
		if err != nil {
			limit = 5
		}
	}

	if r.FormValue("offset") != "" {
		offset, err = strconv.Atoi(r.FormValue("offset"))
		if err != nil {
			offset = 0
		}
	}

	nameTable := r.Context().Value(h.Keys.table).(string)
	query := fmt.Sprintf("SELECT * FROM `%s` LIMIT ? OFFSET ?", nameTable)

	rows, err := h.DB.Query(
		query,
		limit,
		offset,
	)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var resultRows []map[string]any

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		oneRow := make(map[string]any)
		for i := 0; i < len(columns); i++ {
			if b, ok := values[i].([]byte); ok {
				s := string(b)
				if num, err := strconv.Atoi(s); err == nil {
					oneRow[columns[i]] = num
				} else {
					oneRow[columns[i]] = s
				}
			} else {
				oneRow[columns[i]] = values[i]
			}
		}
		resultRows = append(resultRows, oneRow)
	}

	resp := map[string]any{
		"response": map[string]any{
			"records": resultRows,
		},
	}

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// OneRecord return one record by id
func (h *Handler) OneRecord(w http.ResponseWriter, r *http.Request) {
	var err error

	id, ok := r.Context().Value(h.Keys.id).(string)
	if !ok || id == "" {
		http.Error(w, "id not specified", http.StatusBadRequest)
		return
	}

	nameTable := r.Context().Value(h.Keys.table).(string)

	columnsType, err := h.getListColumns(nameTable)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var primaryKey string
	for col := range columnsType {
		if columnsType[col].Extra == "auto_increment" {
			primaryKey = columnsType[col].Field
		}
	}

	query := fmt.Sprintf("SELECT * FROM `%s` WHERE %s = ?", nameTable, primaryKey)

	rows, err := h.DB.Query(
		query,
		id,
	)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resultRow := make(map[string]any)

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for i := 0; i < len(columns); i++ {
			if b, ok := values[i].([]byte); ok {
				s := string(b)
				if num, err := strconv.Atoi(s); err == nil {
					resultRow[columns[i]] = num
				} else {
					resultRow[columns[i]] = s
				}
			} else {
				resultRow[columns[i]] = values[i]
			}
		}
	}

	if len(resultRow) == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{"error": "record not found"})
		return
	}

	resp := map[string]any{
		"response": map[string]any{
			"record": resultRow,
		},
	}

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Add create new record
func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	var err error

	nameTable := r.Context().Value(h.Keys.table).(string)

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	columns := make([]string, 0, len(body))
	placeholders := make([]string, 0, len(body))
	values := make([]any, 0, len(body))

	columnsType, err := h.getListColumns(nameTable)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var primaryKey string
	for col := range columnsType {
		if columnsType[col].Extra == "auto_increment" { // auto increment primary key игнорируется при вставке
			primaryKey = columnsType[col].Field
		} else {
			if columnsType[col].Null == "NO" && !columnsType[col].Default.Valid {
				if _, ok := body[col]; !ok {
					switch {
					case strings.Contains(columnsType[col].Type, "int"):
						body[col] = 0
					case strings.Contains(columnsType[col].Type, "varchar"),
						strings.Contains(columnsType[col].Type, "text"):
						body[col] = ""
					}
				}
			}
		}
	}

	for col, val := range body {
		if _, ok := columnsType[col]; ok {
			if columnsType[col].Extra != "auto_increment" { // auto increment primary key игнорируется при вставке
				columns = append(columns, fmt.Sprintf("`%s`", col))
				placeholders = append(placeholders, "?")
				values = append(values, val)
			}
		}
	}

	insert := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)",
		nameTable,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	result, err := h.DB.Exec(
		insert,
		values...,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"response": map[string]any{
			primaryKey: lastID,
		},
	}

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var err error

	nameTable := r.Context().Value(h.Keys.table).(string)

	id, ok := r.Context().Value(h.Keys.id).(string)
	if !ok || id == "" {
		http.Error(w, "id not specified", http.StatusBadRequest)
		return
	}

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	columns := make([]string, 0, len(body))
	values := make([]any, 0, len(body))

	columnsType, err := h.getListColumns(nameTable)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var primaryKey string
	for col := range columnsType {
		if columnsType[col].Extra == "auto_increment" {
			primaryKey = columnsType[col].Field
		}
	}

	for col, val := range body {
		colInfo, ok := columnsType[col]
		if !ok {
			continue
		}
		if columnsType[col].Extra != "auto_increment" { // auto increment primary key игнорируется при вставке
			switch v := val.(type) {
			case string:
				if !strings.Contains(colInfo.Type, "varchar") &&
					!strings.Contains(colInfo.Type, "text") {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]any{"error": fmt.Sprintf("field %s have invalid type", col)})
					return
				}
			case float64:
				if !strings.Contains(colInfo.Type, "int") &&
					!strings.Contains(colInfo.Type, "float") {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]any{"error": fmt.Sprintf("field %s have invalid type", col)})
					return
				}
				val = int(v)
			case nil:
				if colInfo.Null == "NO" {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]any{"error": fmt.Sprintf("field %s have invalid type", col)})
					return
				}
			}

			columns = append(columns, fmt.Sprintf("`%s` = ?", col))
			values = append(values, val)

		} else {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{"error": fmt.Sprintf("field %s have invalid type", col)})
			return
		}
	}

	query := fmt.Sprintf(
		"UPDATE `%s` SET %s WHERE %s = ?",
		nameTable,
		strings.Join(columns, ", "),
		primaryKey,
	)

	values = append(values, id)
	result, err := h.DB.Exec(
		query,
		values...,
	)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"response": map[string]any{
			"updated": affected,
		},
	}

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	nameTable := r.Context().Value(h.Keys.table).(string)

	id, ok := r.Context().Value(h.Keys.id).(string)
	if !ok || id == "" {
		http.Error(w, "id not specified", http.StatusBadRequest)
		return
	}

	columnsType, err := h.getListColumns(nameTable)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var primaryKey string
	for col := range columnsType {
		if columnsType[col].Extra == "auto_increment" {
			primaryKey = columnsType[col].Field
		}
	}

	delete := fmt.Sprintf("DELETE FROM `%s` WHERE %s = ?",
		nameTable,
		primaryKey,
	)

	result, err := h.DB.Exec(
		delete,
		id,
	)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"response": map[string]any{
			"deleted": affected,
		},
	}

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
