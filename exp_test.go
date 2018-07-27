package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUsuarios(t *testing.T) {
	db := NewDBConn()
	defer db.DB.Close()

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

	db.DB.Exec("delete from usuarios where id = $1", id)
}

func TestBuscarUsuario(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, `/buscarusuario?nombre=victor%20samuel`, nil)
	if err != nil {
		t.Fatal(err)
		return
	}

	s := NewDBConn()
	defer s.DB.Close()
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.buscarUsuario)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v: %v",
			status, http.StatusOK, rr.Body)
	}
}
