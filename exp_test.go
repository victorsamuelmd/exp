package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGuardarUsuario(t *testing.T) {
	js := `{"nombres": "Victor Samuel",
		"apellidos": "Mosquera Artamonov",
		"grupo": "medico", "empresa": "ninguna",
		"usuario": "victorsamuelmd",
		"documentoNumero": "1087998004"}`

	req, err := http.NewRequest(http.MethodPost, "/crearusuario", strings.NewReader(js))
	if err != nil {
		t.Fatal(err)
	}
	s := NewServer()
	defer s.DB.Close()
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.crearUsuario)
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v: %v",
			status, http.StatusOK, rr.Body)
	}
	fmt.Printf("%v", rr.Body)
}

func TestObtenerUsuario(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/obtenerusuario?id=274978", nil)
	if err != nil {
		t.Fatal(err)
	}

	s := NewServer()
	defer s.DB.Close()
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.obtenerUsuario)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v: %v",
			status, http.StatusOK, rr.Body)
	}
}

func TestBuscarUsuario(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, `/buscarusuario?nombres=victor%20samuel`, nil)
	if err != nil {
		t.Fatal(err)
		return
	}

	s := NewServer()
	defer s.DB.Close()
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.buscarUsuario)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v: %v",
			status, http.StatusOK, rr.Body)
	}
}

func TestActualizarUsuario(t *testing.T) {
	js := `{"p_nombre": "Fulanito", "id":"274978",
		"p_apellido": "de Tal",
		"grupo": "medico", "empresa": "ninguna",
		"usuario": "victorsamuelmd",
		"documentoNumero": "1087998004"}`

	req, err := http.NewRequest(http.MethodPut, "/actualizarusuario", strings.NewReader(js))
	if err != nil {
		t.Fatal(err)
	}
	s := NewServer()
	defer s.DB.Close()
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.actualizarUsuario)
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v: %v",
			status, http.StatusOK, rr.Body)
	}
	fmt.Printf("%v", rr.Body)
}
