package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type Response struct {
	Body any `json:"response,omitempty"`
}

type Table struct {
	Body any `json:"tables,omitempty"`
}

type Error struct {
	Body any `json:"error,omitempty"`
}

type Handler struct {
	DB *sql.DB
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	return &Handler{DB: db}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			h.List(w, r)
		}
		return
	}

	switch len(parts) {
	case 1:
		ctx := context.WithValue(r.Context(), "table", parts[0])
		r = r.WithContext(ctx)

		switch r.Method {
		case http.MethodPut:
			h.Add(w, r)
		case http.MethodGet:
			h.ListRecords(w, r)
		}
	case 2:
		ctx := context.WithValue(r.Context(), "table", parts[0])
		ctx = context.WithValue(ctx, "id", parts[1])
		r = r.WithContext(ctx)

		switch r.Method {
		case http.MethodGet:
			h.OneRecord(w, r)
		case http.MethodPost:
			h.Update(w, r)
		case http.MethodDelete:
			h.Delete(w, r)
		}
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

// return list of name tables
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {

	rows, err := h.DB.Query("SHOW TABLES")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var tableName string
	var tableNames []string

	for rows.Next() {
		err := rows.Scan(&tableName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tableNames = append(tableNames, tableName)
	}

	resp := Response{
		Body: Table{
			Body: tableNames,
		},
	}

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// return list of records
func (h *Handler) ListRecords(w http.ResponseWriter, r *http.Request) {
	// default settings
	limit := 5
	offset := 0

	nameTable, ok := r.Context().Value("table").(string)
	if !ok || nameTable == "" {
		http.Error(w, "table not specified", http.StatusBadRequest)
		return
	}

	var exists bool
	err := h.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?)",
		nameTable,
	).Scan(&exists)

	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	if !exists {
		resp := Error{
			Body: "unknown table",
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(resp)
		return
	}

	value := r.FormValue("limit")
	if value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
	}

	value = r.FormValue("offset")
	if value != "" {
		offset, err = strconv.Atoi(value)
		if err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
	}

	rows, err := h.DB.Query("SHOW FULL COLUMNS FROM ?", nameTable)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	columns, err := rows.Columns()
	_ = limit
	_ = offset

	resp := Response{
		Body: Table{
			Body: columns,
		},
	}

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h *Handler) OneRecord(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	// // в целям упрощения примера пропущена валидация
	// result, err := h.DB.Exec(
	// 	"INSERT INTO items (`title`, `description`) VALUES (?, ?)",
	// 	r.FormValue("title"),
	// 	r.FormValue("description"),
	// )
	// if err != nil {
	// 	panic(err)
	// }

	// affected, err := result.RowsAffected()
	// if err != nil {
	// 	panic(err)
	// }
	// lastID, err := result.LastInsertId()
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println("Insert - RowsAffected", affected, "LastInsertId: ", lastID)

	// http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// id, err := strconv.Atoi(vars["id"])
	// if err != nil {
	// 	panic(err)
	// }

	// // в целям упрощения примера пропущена валидация
	// result, err := h.DB.Exec(
	// 	"UPDATE items SET"+
	// 		"`title` = ?"+
	// 		",`description` = ?"+
	// 		",`updated` = ?"+
	// 		"WHERE id = ?",
	// 	r.FormValue("title"),
	// 	r.FormValue("description"),
	// 	"rvasily",
	// 	id,
	// )
	// if err != nil {
	// 	panic(err)
	// }

	// affected, err := result.RowsAffected()
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println("Update - RowsAffected", affected)

	// http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// id, err := strconv.Atoi(vars["id"])
	// if err != nil {
	// 	panic(err)
	// }

	// result, err := h.DB.Exec(
	// 	"DELETE FROM items WHERE id = ?",
	// 	id,
	// )
	// if err != nil {
	// 	panic(err)
	// }

	// affected, err := result.RowsAffected()
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println("Delete - RowsAffected", affected)

	// w.Header().Set("Content-type", "application/json")
	// resp := `{"affected": ` + strconv.Itoa(int(affected)) + `}`
	// w.Write([]byte(resp))
}
