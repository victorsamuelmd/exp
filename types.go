package main

import (
	"errors"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	val "github.com/go-ozzo/ozzo-validation"
	"golang.org/x/crypto/bcrypt"
)

type DBConn struct {
	MDB *mgo.Session
}

// Credentials struct for passing out values
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Grupo    string `json:"grupo"`
}

// Usuario es una estructura para pasar usuarios para crearlos
type Usuario struct {
	ID                bson.ObjectId `json:"id" bson:"_id"`
	UsuarioNombre     string        `json:"usuario" bson:"usuario"`
	CorreoElectronico string        `json:"correoElectronico" bson:"correoElectronico"`
	PalabraClave      string        `json:"palabraClave" bson:"palabraClave"`
	Grupo             string        `json:"grupo" bson:"grupo"`
	FechaCreacion     time.Time     `json:"fechaCreacion" bson:"fechaCreacion"`
}

// Validate valida los datos del Usuario
func (u Usuario) Validate() error {
	return val.ValidateStruct(&u,
		val.Field(&u.UsuarioNombre, val.Required, val.Length(2, 60)),
		val.Field(&u.CorreoElectronico, val.Required, val.Length(2, 60)),
		val.Field(&u.PalabraClave, val.Required, val.Length(4, 128)),
		val.Field(&u.Grupo, val.Required, val.In("ADMIN", "MED", "FACT")))
}

// Save intenta guardar el usuario
func (u Usuario) Save(db *mgo.Session) (string, error) {
	err := u.Validate()
	u.FechaCreacion = time.Now()

	if err != nil {
		return "", err
	}

	hashPalabraClave, err := bcrypt.GenerateFromPassword([]byte(u.PalabraClave), 4)
	if err != nil {
		return "", err
	}

	u.PalabraClave = string(hashPalabraClave)
	u.ID = bson.NewObjectId()

	err = db.DB(dbName).C(usuariosC).Insert(u)
	if err != nil {
		return "", err
	}

	return u.ID.Hex(), nil
}

type Perfil struct {
	ID                     bson.ObjectId `json:"id" bson:"_id"`
	UsuarioId              string        `json:"usuarioId" bson:"usuarioId"`
	Nombres                string        `json:"nombres" bson:"nombres"`
	Apellidos              string        `json:"apellidos" bson:"apellidos"`
	Genero                 string        `json:"genero" bson:"genero"`
	DocumentoNumero        string        `json:"documentoNumero" bson:"documentoNumero"`
	DocumentoTipo          string        `json:"documentoTipo" bson:"documentoTipo"`
	FechaNacimiento        time.Time     `json:"fechaNacimiento" bson:"fechaNacimiento"`
	FechaUltimoIngreso     time.Time     `json:"fechaUltimoIngreso" bson:"fechaUltimoIngreso"`
	Telefono               string        `json:"telefono" bson:"telefono"`
	ResidenciaPais         string        `json:"residenciaPais" bson:"residenciaPais"`
	ResidenciaDepartamento string        `json:"residenciaDepartamento" bson:"residenciaDepartamento"`
	ResidenciaMunicipio    string        `json:"residenciaMunicipio" bson:"residenciaMunicipio"`
	ResidenciaBarrio       string        `json:"residenciaBarrio" bson:"residenciaBarrio"`
	ResidenciaDireccion    string        `json:"residenciaDireccion" bson:"residenciaDireccion"`
}

func (p Perfil) Validate() error {
	return val.ValidateStruct(
		&p,
		val.Field(&p.Genero, val.Required, val.In("M", "F", "I")),
		val.Field(&p.Nombres, val.Required),
		val.Field(&p.Apellidos, val.Required),
		val.Field(&p.DocumentoNumero, val.Required),
		val.Field(&p.DocumentoTipo, val.Required, val.In("CC", "CE", "TI", "AS", "MS", "PS", "RC")),
	)
}

func (p Perfil) Save(db *mgo.Session) error {
	err := p.Validate()
	if err != nil {
		return err
	}

	p.ID = bson.NewObjectId()
	return db.DB("mednote").C("perfiles").Insert(p)
}

func (p *Perfil) ObtenerPorID(db *mgo.Session, ID string) error {
	if !bson.IsObjectIdHex(ID) {
		return errors.New("ID should be a valid hex")
	}

	return db.DB(dbName).C(perfilesC).FindId(bson.ObjectIdHex(ID)).One(&p)

}

// HistoriaUrgencias contiene la informacion de una consulta de urgencias.
// Tiene la relacion entre un paciente (perfil) y una medico (usuario)
type HistoriaUrgencias struct {
	ID                        bson.ObjectId `json:"id" bson:"_id"`
	Paciente                  bson.ObjectId `json:"paciente" bson:"paciente"`
	FechaInicio               time.Time     `json:"fechaInicio" bson:"fechaInicio"`
	FechaFinalizacion         time.Time     `json:"fechaFinalizacion" bson:"fechaFinalizacion"`
	AdmitidoPor               string        `json:"admitidoPor" bson:"admitidoPor"`
	Registro                  string        `json:"registro" bson:"registro"`
	OrigenAtencion            string        `json:"origenAtencion" bson:"origenAtencion"`
	MotivoConsulta            string        `json:"motivoConsulta" bson:"motivoConsulta"`
	EnfermedadActual          string        `json:"enfermedadActual" bson:"enfermedadActual"`
	AntecedentesFamiliares    string        `json:"antecedentesFamiliares" bson:"antecedentesFamiliares"`
	AntecedentesPersonales    string        `json:"antecedentesPersonales" bson:"antecedentesPersonales"`
	RespuestaOcular           int           `json:"respuestaOcular" bson:"respuestaOcular"`
	RespuestaVerbal           int           `json:"respuestaVerbal" bson:"respuestaVerbal"`
	RespuestaMotora           int           `json:"respuestaMotora" bson:"respuestaMotora"`
	FuerzaSuperiorI           int           `json:"fuerzaSuperiorI" bson:"fuerzaSuperiorI"`
	FuerzaSuperiorD           int           `json:"fuerzaSuperiorD" bson:"fuerzaSuperiorD"`
	FuerzaInferiorI           int           `json:"fuerzaInferiorI" bson:"fuerzaInferiorI"`
	FuerzaInferiorD           int           `json:"fuerzaInferiorD" bson:"fuerzaInferiorD"`
	ReflejosSuperiorI         string        `json:"reflejosSuperiorI" bson:"reflejosSuperiorI"`
	ReflejosSuperiorD         string        `json:"reflejosSuperiorD" bson:"reflejosSuperiorD"`
	ReflejosInferiorI         string        `json:"reflejosInferiorI" bson:"reflejosInferiorI"`
	ReflejosInferiorD         string        `json:"reflejosInferiorD" bson:"reflejosInferiorD"`
	EstadoGeneral             string        `json:"estadoGeneral" bson:"estadoGeneral"`
	TensionArterialSistolica  int           `json:"tensionArterialSistolica" bson:"tensionArterialSistolica"`
	TensionArterialDiastolica int           `json:"tensionArterialDiastolica" bson:"tensionArterialDiastolica"`
	FrecuenciaCardiaca        int           `json:"frecuenciaCardiaca" bson:"frecuenciaCardiaca"`
	FrecuenciaRespiratoria    int           `json:"frecuenciaRespiratoria" bson:"frecuenciaRespiratoria"`
	Temperatura               float64       `json:"temperatura" bson:"temperatura"`
	SaturacionOxigeno         int           `json:"saturacionOxigeno" bson:"saturacionOxigeno"`
	Peso                      float64       `json:"peso" bson:"peso"`
	Talla                     float64       `json:"talla" bson:"talla"`
	ExamenFisico              string        `json:"examenFisico" bson:"examenFisico"`
	AnalisisConducta          string        `json:"analisisConducta" bson:"analisisConducta"`
	DiagnosticoPrincipal      string        `json:"diagnosticoPrincipal" bson:"diagnosticoPrincipal"`
	DiagnosticoRelacionado1   string        `json:"diagnosticoRelacionado1" bson:"diagnosticoRelacionado1"`
	DiagnosticoRelacionado2   string        `json:"diagnosticoRelacionado2" bson:"diagnosticoRelacionado2"`
	DiagnosticoRelacionado3   string        `json:"diagnosticoRelacionado3" bson:"diagnosticoRelacionado3"`
	Medico                    string        `json:"medico" bson:"medico"`
}

func (hu HistoriaUrgencias) Save(db *mgo.Session) error {

	hu.ID = bson.NewObjectId()
	return db.DB("mednote").C("historiaUrgencias").Insert(hu)
}

// Obtener llena la estructura con la informacion de la base de datos, requiere
// tener primero el id
func (hu *HistoriaUrgencias) Obtener(db *mgo.Session) error {
	return db.DB(dbName).C(historiaUrgenciasC).FindId(hu.ID).One(hu)
}
