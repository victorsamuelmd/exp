package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	val "github.com/go-ozzo/ozzo-validation"

	_ "github.com/go-sql-driver/mysql"
	"github.com/satori/go.uuid"
)

// import "encoding/json"

type DbUser struct {
	Username   string    `json:"userName"`
	Password   string    `json:"password"`
	Email      string    `json:"email"`
	FirstName  string    `json:"firstName"`
	SecondName string    `json:"secondName"`
	LastDate   time.Time `json:"lastDate"`
}

type Server struct {
	DB *sql.DB
}

// NewServer creates a connextion to a mysql server
func NewServer() Server {

	password := "74bc5df74572aaa50b5f2a6bb7fa020ec43c2ecb35d2be39699095ec0a79"
	dbUsername := "root"
	dbHost := "opuslibertati"

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", dbUsername, password, dbHost))

	if err != nil {
		panic(err)
	}

	return Server{db}
}

// Credentials struct for passing out values
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Grupo    int    `json:"grupo"`
}

func (s Server) autenticarUsuario(w http.ResponseWriter, r *http.Request) {
	var decoder = json.NewDecoder(r.Body)
	var encoder = json.NewEncoder(w)
	var jsonRequest Credentials
	var err error

	err = decoder.Decode(&jsonRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h := md5.New()
	io.WriteString(h, jsonRequest.Password)
	jsonRequest.Password = fmt.Sprintf("%x", h.Sum(nil))

	sqlStmt, err := s.DB.
		Prepare("select username, passwd, id_grupo from d9_users where username=? and passwd=?")
	result := sqlStmt.QueryRow(jsonRequest.Username, jsonRequest.Password)

	err = result.Scan(&jsonRequest.Username, &jsonRequest.Password, &jsonRequest.Grupo)
	if err == sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	encoder.Encode(jsonRequest)

}

// Usuario es una estructura para pasar usuarios para crearlos
type Usuario struct {
	ID              string `json:"id"`
	Nombres         string `json:"nombres"`
	Apellidos       string `json:"apellidos"`
	Grupo           string `json:"grupo"`
	Empresa         string `json:"empresa"`
	UsuarioNombre   string `json:"usuario"`
	DocumentoNumero string `json:"documentoNumero"`
}

// Validate valida los datos del Usuario
func (u Usuario) Validate() error {
	return val.ValidateStruct(&u,
		val.Field(&u.Apellidos, val.Required, val.Length(2, 255)),
		val.Field(&u.Nombres, val.Required, val.Length(2, 255)),
		val.Field(&u.Grupo, val.Required, val.Length(2, 20)),
		val.Field(&u.Empresa, val.Required, val.Length(2, 20)),
		val.Field(&u.UsuarioNombre, val.Required, val.Length(2, 60)),
		val.Field(&u.DocumentoNumero, val.Required, val.Length(4, 15)))
}

func (s Server) crearUsuario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST request", http.StatusBadRequest)
		return
	}

	var user Usuario

	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&user)
	err := user.Validate()
	if err != nil {
		http.Error(w, fmt.Sprintf("Something went wrong: %s", err), http.StatusBadRequest)
		return
	}
	stmt := `insert into d9_users
	(id, username, id_grupo, id_empresa, p_nombre, p_apellido, documento_numero)
	values (?, ?, ?, ?, ?, ?, ?)`

	u2, err := uuid.NewV4()
	if err != nil {
		http.Error(w, fmt.Sprintf("Something went wrong: %s", err), http.StatusInternalServerError)
		return
	}

	sqlStmt, err := s.DB.Prepare(stmt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Something went wrong: %s", err), http.StatusInternalServerError)
		return
	}

	_, err = sqlStmt.Exec(u2, user.UsuarioNombre,
		user.Grupo, user.Empresa, user.Nombres, user.Apellidos, user.DocumentoNumero)
	if err != nil {
		http.Error(w, fmt.Sprintf("Something went wrong: %s", err), http.StatusBadRequest)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(u2.String())
}

func (s Server) obtenerUsuario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET request Allowed", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()
	if q.Get("id") == "" {
		http.Error(w, "Id query value missing", http.StatusBadRequest)
		return
	}

	slqStmt := `select id, username, p_nombre, p_apellido, documento_numero from d9_users where id = ?`
	row := s.DB.QueryRow(slqStmt, q.Get("id"))

	var u Usuario

	err := row.Scan(&u.ID, &u.UsuarioNombre, &u.Nombres, &u.Apellidos, &u.DocumentoNumero)

	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("User with id: %s not found", q.Get("id")), http.StatusNotFound)
		return
	}
	encoder := json.NewEncoder(w)
	if err = encoder.Encode(u); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (s Server) buscarUsuario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET request Allowed", http.StatusBadRequest)
		return
	}
	q := r.URL.Query()
	rows, err := s.DB.Query(`SELECT
		id, username, p_nombre, p_apellido, documento_numero
		FROM d9_users
		WHERE p_nombre LIKE ?
		AND p_apellido LIKE ?
		AND documento_numero LIKE ?`,

		fmt.Sprintf("%%%s%%", q.Get("nombre")),
		fmt.Sprintf("%%%s%%", q.Get("apellido")),
		fmt.Sprintf("%%%s%%", q.Get("documento")))

	defer rows.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var users []Usuario

	for rows.Next() {
		var user Usuario
		rows.Scan(&user.ID, &user.UsuarioNombre, &user.Nombres, &user.Apellidos, &user.DocumentoNumero)
		users = append(users, user)
	}
	rows.Close()

	if len(users) == 0 {
		http.Error(w, "No hay usuarios", http.StatusNotFound)
		return
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (s Server) actualizarUsuario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Only PUT allowed", http.StatusBadRequest)
		return
	}
	sqlStmt := `UPDATE d9_users
		SET %s
		WHERE id = ?`

	var canBeUpdated = func(field string) bool {
		switch field {
		case
			"id",
			"p_apellido",
			"p_nombre",
			"numero_documento",
			"empresa":
			return true
		}
		return false
	}

	var req map[string]string
	var fieldsToUpdate []string
	var toUpdate []interface{}
	err := json.NewDecoder(r.Body).Decode(&req)

	for k, v := range req {
		if canBeUpdated(k) {
			fieldsToUpdate = append(fieldsToUpdate, fmt.Sprintf(`%s=?`, k))
			toUpdate = append(toUpdate, v)
		}
	}
	// id is the last argument
	toUpdate = append(toUpdate, req["id"])

	prepStmt := fmt.Sprintf(sqlStmt, strings.Join(fieldsToUpdate, ","))
	stmt, err := s.DB.Prepare(prepStmt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res, err := stmt.Exec(toUpdate...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows, err := res.RowsAffected()
	if rows > 1 {
		panic("This should not happen")
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, "success")
}

func activateCors(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, PUT, GET")
		w.Header().Set("Access-Control-Allow-Headers", "content-type, authorization")
		w.Header().Set(`Content-Type`, `application/json`)
		if r.Method == http.MethodOptions {
			return
		}
		f.ServeHTTP(w, r)
	}
}

func main() {
	server := NewServer()
	defer server.DB.Close()

	w := http.NewServeMux()
	w.HandleFunc("/login", activateCors(server.autenticarUsuario))
	w.HandleFunc("/nuevousuario", activateCors(server.crearUsuario))
	w.HandleFunc("/obtenerusuario", activateCors(server.obtenerUsuario))
	w.HandleFunc("/buscarusuario", activateCors(server.buscarUsuario))
	w.HandleFunc("/actualizarusuario", activateCors(server.actualizarUsuario))
	http.ListenAndServe(":8070", w)
}
