package main

import (
	"database/sql"
	"time"

	val "github.com/go-ozzo/ozzo-validation"
)

type DBConn struct {
	DB *sql.DB
}

// Credentials struct for passing out values
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Grupo    string `json:"grupo"`
}

// Usuario es una estructura para pasar usuarios para crearlos
type Usuario struct {
	ID                string    `json:"id"`
	UsuarioNombre     string    `json:"usuario"`
	CorreoElectronico string    `json:"correoElectronico"`
	PalabraClave      string    `json:"palabraClave"`
	Grupo             string    `json:"grupo"`
	FechaCreacion     time.Time `json:"fechaCreacion"`
}

// Validate valida los datos del Usuario
func (u Usuario) Validate() error {
	return val.ValidateStruct(&u,
		val.Field(&u.UsuarioNombre, val.Required, val.Length(2, 60)),
		val.Field(&u.CorreoElectronico, val.Required, val.Length(2, 60)),
		val.Field(&u.PalabraClave, val.Required, val.Length(4, 128)))
}

type Perfil struct {
	ID              string    `json:"id"`
	Nombres         string    `json:"nombres"`
	Apellidos       string    `json:"apellidos"`
	DocumentoNumero string    `json:"documentoNumero"`
	FechaNacimiento time.Time `json:"fechaNacimiento"`
}
