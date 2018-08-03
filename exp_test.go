package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/globalsign/mgo/bson"
)

var db = NewDBConn()

func TestUsuarios(t *testing.T) {

	srv := httptest.NewServer(Router(db))
	defer srv.Close()

	client := &http.Client{}
	var id string

	t.Run("Create user", func(st *testing.T) {

		js := `{"grupo": "MED", 
			"correoElectronico": "medico@gmail.com",
			"usuario": "medico",
			"palabraClave": "natalia1988"}`

		req, err := http.NewRequest(
			http.MethodPost,
			fmt.Sprintf("%s/usuarios", srv.URL),
			strings.NewReader(js))

		if err != nil {
			st.Fatal(err)
		}

		res, err := client.Do(req)
		defer res.Body.Close()
		if err != nil {
			st.Fatal(err)
		}
		err = json.NewDecoder(res.Body).Decode(&id)
		fmt.Println("ID ", id)
		if err != nil {
			st.Fatal(err)
		}

		if status := res.StatusCode; status != http.StatusOK {
			st.Errorf("handler returned wrong status code: got %v want %v: %s",
				status, http.StatusOK, id)
		}

		if conType := res.Header.Get("Content-Type"); conType != "application/json" {
			st.Errorf("Expected header %s, got %s", "application/json", conType)
		}

	})

	t.Run("Update User", func(st *testing.T) {

		js := `{"correoElectronico": "natalia@gmail.com"}`

		req, err := http.NewRequest(http.MethodPut,
			fmt.Sprintf("%s/usuarios/%s", srv.URL, id),
			strings.NewReader(js))

		if err != nil {
			t.Fatal(err)
		}
		res, err := client.Do(req)
		defer res.Body.Close()
		if err != nil {
			st.Fatal(err)
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			st.Fatal(err)
		}
		if status := res.StatusCode; status != http.StatusOK {
			st.Errorf("handler returned wrong status code: got %v want %v: %s",
				status, http.StatusOK, body)
		}
	})

	t.Run("Get User", func(st *testing.T) {
		res, err := http.Get(fmt.Sprintf("%s/usuarios/%s", srv.URL, id))
		defer res.Body.Close()

		if err != nil {
			t.Fatal(err)
		}

		if res.StatusCode != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v: %v",
				res.StatusCode, http.StatusOK, res.Body)
		}
	})

	db.MDB.DB(dbName).C(perfilesC).RemoveId(id)
}

func TestBuscarUsuario(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, `/buscarusuario?nombre=victor%20samuel`, nil)
	if err != nil {
		t.Fatal(err)
		return
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(db.buscarUsuario)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v: %v",
			status, http.StatusOK, rr.Body)
	}
}

func TestSaveHistoriaUrgencias(t *testing.T) {

	var hu = HistoriaUrgencias{
		Medico:            "b867de22-f938-4211-9754-a47a84ea2bf6",
		Paciente:          bson.NewObjectId(),
		Peso:              65.5,
		Talla:             1.65,
		FechaInicio:       time.Now().Add(-10 * time.Minute),
		FechaFinalizacion: time.Now(),
	}

	err := hu.Save(db.MDB)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestSavePerfil(t *testing.T) {

	var p = Perfil{
		Nombres:         "Natalia",
		Apellidos:       "Ramirez Orozco",
		DocumentoNumero: "1087998476",
		DocumentoTipo:   "CC",
		Genero:          "F",
	}
	err := p.Save(db.MDB.Copy())
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestObtenerPerfilPorID(t *testing.T) {

	var p Perfil
	err := p.ObtenerPorID(db.MDB.Copy(), "5b63656adb437400095a7068")
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestObtenerHistoriaUrgencias(t *testing.T) {

}
