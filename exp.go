package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/globalsign/mgo"
	"golang.org/x/crypto/bcrypt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
)

var secretKey, _ = rsa.GenerateKey(rand.Reader, 1024)

const dbName = "mednote"
const perfilesC = "perfiles"
const usuariosC = "usuarios"
const historiaUrgenciasC = "historiaUrgencias"

// NewDBConn creates a connextion to a mysql server
func NewDBConn() DBConn {

	mdb, err := mgo.Dial(`root:example@localhost`)

	if err != nil {
		panic(err)
	}

	return DBConn{mdb}
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
	var usr Usuario

	ss := s.MDB.Copy()
	defer ss.Close()

	err = ss.DB(dbName).C(usuariosC).
		Find(bson.M{"usuario": jsonRequest.Username}).One(&usr)

	if err != nil {
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(usr.PalabraClave), []byte(jsonRequest.Password))
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusNotFound)
		return
	}

	tokenString, err := createToken(usr.UsuarioNombre, usr.Grupo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	encoder.Encode(map[string]interface{}{
		"username":      usr.UsuarioNombre,
		"authorization": tokenString,
		"grupo":         usr.Grupo})
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
		js, _ := json.Marshal(map[string]interface{}{"errors": err.Error()})
		http.Error(w, string(js), http.StatusBadRequest)
		return
	}

	ss := s.MDB.Copy()
	defer ss.Close()

	id, err := usr.Save(ss)
	if err != nil {
		js, _ := json.Marshal(map[string]interface{}{"errors": err.Error()})
		http.Error(w, string(js), http.StatusBadRequest)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(id)
}

func (s DBConn) obtenerUsuario(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	var usr Usuario
	ss := s.MDB.Clone()
	defer ss.Close()

	if !bson.IsObjectIdHex(vars["id"]) {
		http.Error(w, "Id should be a valid Hex", http.StatusBadRequest)
		return
	}

	err := ss.DB(dbName).C(usuariosC).FindId(bson.ObjectIdHex(vars["id"])).One(&usr)

	if err == mgo.ErrNotFound {
		http.Error(w, "Not Found", http.StatusNotFound)
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
	var users []Usuario

	ss := s.MDB.Copy()
	defer ss.Close()

	err := ss.DB(dbName).C(usuariosC).Find(bson.M{}).All(&users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(users)
}

func (s DBConn) buscarPerfiles(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	ss := s.MDB.Copy()
	defer ss.Close()

	var prf []Perfil

	nombres := bson.M{"$regex": q.Get("nombres"), "$options": "i"}
	apellidos := bson.M{"$regex": q.Get("apellidos"), "$options": "i"}
	documento := bson.M{"$regex": q.Get("documento"), "$options": "i"}
	query := bson.M{"nombres": nombres, "apellidos": apellidos, "documentoNumero": documento}

	err := ss.DB("mednote").C("perfiles").Find(query).All(&prf)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if len(prf) == 0 {
		w.WriteHeader(http.StatusNotFound)
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(prf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (s DBConn) actualizarUsuario(w http.ResponseWriter, r *http.Request) {
	ss := s.MDB.Copy()
	defer ss.Close()

	vars := mux.Vars(r)
	var usr Usuario
	err := json.NewDecoder(r.Body).Decode(&usr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !bson.IsObjectIdHex(vars["id"]) {
		http.Error(w, "Id should be a valid Hex", http.StatusBadRequest)
		return
	}

	err = ss.DB(dbName).C(usuariosC).
		UpdateId(bson.ObjectIdHex(vars["id"]), bson.M{"correoElectronico": usr.CorreoElectronico})

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
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
		"username": username,
		"grupo":    authLevel,
	})

	return token.SignedString(secretKey)
}

func (db DBConn) historias(w http.ResponseWriter, r *http.Request) {
	var hus []HistoriaUrgencias

	ss := db.MDB.Copy()
	defer ss.Close()

	err := ss.DB("mednote").C("historiaUrgencias").Find(bson.M{}).All(&hus)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(hus)
}

func (db DBConn) crearHistoria(w http.ResponseWriter, r *http.Request) {

	var hus HistoriaUrgencias

	err := json.NewDecoder(r.Body).Decode(&hus)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ss := db.MDB.Copy()
	defer ss.Close()

	hus.FechaFinalizacion = time.Now()
	err = hus.Save(ss)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(hus.ID)
}

func (db DBConn) crearPerfil(w http.ResponseWriter, r *http.Request) {
	var p Perfil
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&p)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = p.Save(db.MDB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (db DBConn) obtenerPerfil(w http.ResponseWriter, r *http.Request) {
	var p Perfil
	ss := db.MDB.Copy()
	defer ss.Close()
	vars := mux.Vars(r)

	if len(vars["id"]) != 24 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	err := p.ObtenerPorID(ss, vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(&p)
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

	w.HandleFunc("/perfiles/{id}", activateCors(dbConn.obtenerPerfil)).
		Methods(http.MethodGet, http.MethodOptions)

	w.HandleFunc("/perfiles", activateCors(dbConn.buscarPerfiles)).
		Methods(http.MethodGet, http.MethodOptions)

	w.HandleFunc("/perfiles", activateCors(dbConn.crearPerfil)).
		Methods(http.MethodPost, http.MethodOptions)

	w.HandleFunc("/historia", activateCors(dbConn.historias)).
		Methods(http.MethodGet, http.MethodOptions)

	w.HandleFunc("/historia", activateCors(dbConn.crearHistoria)).
		Methods(http.MethodPost, http.MethodOptions)
	return w
}

func main() {

	dbConn := NewDBConn()
	defer dbConn.MDB.Close()
	r := Router(dbConn)
	http.ListenAndServe(":8070", r)

}
