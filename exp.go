package main

import (
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"
)

var secretKey, _ = rsa.GenerateKey(rand.Reader, 1024)

// import "encoding/json"

// NewDBConn creates a connextion to a mysql server
func NewDBConn() DBConn {

	password := "NataliaVictor12122801"
	dbUsername := "victorsamuelmd"
	dbName := "node"
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		dbUsername, password, dbName)

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	return DBConn{db}
}

func (s DBConn) autenticarUsuario(w http.ResponseWriter, r *http.Request) {
	var decoder = json.NewDecoder(r.Body)
	defer r.Body.Close()
	var encoder = json.NewEncoder(w)
	var jsonRequest Credentials
	var err error

	err = decoder.Decode(&jsonRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sqlStmt, err := s.DB.
		Prepare(`SELECT usuario, palabra_clave, grupo
			FROM usuarios
			WHERE usuario = $1`)
	result := sqlStmt.QueryRow(jsonRequest.Username)

	var pass []byte
	var grupo string

	err = result.Scan(&jsonRequest.Username, &pass, &grupo)
	if err == sql.ErrNoRows {
		http.Error(w, "Bad Credentials", http.StatusNotFound)
		return
	}

	err = bcrypt.CompareHashAndPassword(pass, []byte(jsonRequest.Password))
	if err != nil {
		http.Error(w, "Bad Credentials", http.StatusNotFound)
		return
	}

	tokenString, err := createToken(jsonRequest.Username, jsonRequest.Grupo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	encoder.Encode(map[string]interface{}{
		"username":      jsonRequest.Username,
		"authorization": tokenString,
		"grupo":         grupo})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (s DBConn) crearUsuario(w http.ResponseWriter, r *http.Request) {

	var usr Usuario

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	decoder.Decode(&usr)
	err := usr.Validate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	hashPalabraClave, err := bcrypt.GenerateFromPassword([]byte(usr.PalabraClave), 4)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	usr.PalabraClave = string(hashPalabraClave)
	stmt := `insert into usuarios
		(id, usuario, palabra_clave, correo_electronico, grupo)
		values ($1, $2, $3, $4, $5)`

	u2, err := uuid.NewV4()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sqlStmt, err := s.DB.Prepare(stmt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = sqlStmt.Exec(u2, usr.UsuarioNombre, usr.PalabraClave,
		usr.CorreoElectronico, usr.Grupo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(u2.String())
}

func (s DBConn) obtenerUsuario(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	slqStmt := `SELECT id, usuario, correo_electronico, fecha_creacion FROM usuarios WHERE id = $1`
	row := s.DB.QueryRow(slqStmt, vars["id"])

	var usr Usuario

	err := row.Scan(&usr.ID, &usr.UsuarioNombre, &usr.CorreoElectronico, &usr.FechaCreacion)

	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("User with id: %s not found", vars["id"]), http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	if err = encoder.Encode(usr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (s DBConn) buscarUsuario(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	rows, err := s.DB.Query(`SELECT
		id, usuario, correo_electronico
		FROM usuarios
		WHERE usuario LIKE $1
		AND correo_electronico LIKE $2`,

		fmt.Sprintf("%%%s%%", q.Get("nombres")),
		fmt.Sprintf("%%%s%%", q.Get("correo_electronico")))

	defer rows.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var users []Usuario

	for rows.Next() {
		var user Usuario
		rows.Scan(&user.ID, &user.UsuarioNombre, &user.CorreoElectronico)
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

func (s DBConn) buscarPerfiles(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	rows, err := s.DB.Query(`SELECT
		id, nombres, apellidos, documento_numero, fecha_nacimiento
		FROM perfiles
		WHERE nombres LIKE $1
		AND apellidos LIKE $2`,

		fmt.Sprintf("%%%s%%", q.Get("nombres")),
		fmt.Sprintf("%%%s%%", q.Get("apellidos")))

	defer rows.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var prf []Perfil

	for rows.Next() {
		var p Perfil
		rows.Scan(&p.ID, &p.Nombres, &p.Apellidos, &p.DocumentoNumero, &p.FechaNacimiento)
		prf = append(prf, p)
	}
	rows.Close()

	if len(prf) == 0 {
		http.Error(w, "No hay usuarios", http.StatusNotFound)
		return
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(prf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (s DBConn) actualizarUsuario(w http.ResponseWriter, r *http.Request) {
	sqlStmt := `UPDATE usuarios
		SET correo_electronico = $1
		WHERE id = $2`

	var req map[string]string
	err := json.NewDecoder(r.Body).Decode(&req)

	vars := mux.Vars(r)

	stmt, err := s.DB.Prepare(sqlStmt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res, err := stmt.Exec(req["correoElectronico"], vars["id"])
	fmt.Print(vars["id"])
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
	fmt.Fprint(w, "success", rows)
}

func activateCors(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, PUT, GET")
		w.Header().Set("Access-Control-Allow-Headers", "content-type, authorization")
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodOptions {
			return
		}
		f.ServeHTTP(w, r)
	}
}

func protect(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if ok := validateToken(token); !ok {
			http.Error(w, "Unautho", http.StatusUnauthorized)
			return
		}
		f.ServeHTTP(w, r)
	}
}

func validateToken(token string) bool {
	if _, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return secretKey.Public(), nil
	}); err != nil {
		return false
	}
	return true
}

func createToken(username, authLevel string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"usuario": username,
		"grupo":   authLevel,
	})

	return token.SignedString(secretKey)
}

func main() {

	dbConn := NewDBConn()
	defer dbConn.DB.Close()
	r := Router(dbConn)
	http.ListenAndServe(":8070", r)

}

// Router es la funcion que crea un manejador, necesario para las pruebas
func Router(dbConn DBConn) *mux.Router {

	w := mux.NewRouter()

	w.HandleFunc("/login", activateCors(dbConn.autenticarUsuario)).
		Methods(http.MethodPost, http.MethodOptions)

	w.HandleFunc("/usuarios", activateCors(dbConn.crearUsuario)).
		Methods(http.MethodPost, http.MethodOptions)

	w.HandleFunc("/usuarios", activateCors(dbConn.buscarUsuario)).
		Methods(http.MethodGet, http.MethodOptions)

	w.HandleFunc("/usuarios/{id}", activateCors(dbConn.obtenerUsuario)).
		Methods(http.MethodGet, http.MethodOptions)

	w.HandleFunc("/usuarios/{id}", activateCors(dbConn.actualizarUsuario)).
		Methods(http.MethodPut, http.MethodOptions)

	w.HandleFunc("/perfiles", activateCors(dbConn.buscarPerfiles)).
		Methods(http.MethodGet, http.MethodOptions)
	return w
}
