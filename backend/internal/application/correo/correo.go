package correo

import (
	"context"
	"errors"
	"log"
	"math"
	"net/url"
	"strings"
	"time"
)

// Nombres de plantilla que el Renderizador debe conocer.
const (
	PlantillaVerificacion       = "verificacion"
	PlantillaInvitacion         = "invitacion"
	PlantillaRecuperacion       = "recuperacion"
	PlantillaContrasenaCambiada = "contrasena_cambiada"
)

var ErrEnvio = errors.New("no fue posible enviar el correo")

// Mensaje es el correo ya renderizado que recibe el transporte.
type Mensaje struct {
	Para   string
	Asunto string
	HTML   string
	Texto  string
}

// Contenido es el resultado de renderizar una plantilla.
type Contenido struct {
	Asunto string
	HTML   string
	Texto  string
}

// Transporte entrega un mensaje y devuelve su identificador.
type Transporte interface {
	Enviar(ctx context.Context, mensaje Mensaje) (string, error)
}

// Renderizador convierte una plantilla y sus datos en asunto, HTML y texto.
type Renderizador interface {
	Render(plantilla string, datos any) (Contenido, error)
}

// Base contiene los datos comunes a todas las plantillas.
type Base struct {
	NombreApp string
	URLApp    string
	Anio      int
}

type DatosVerificacion struct {
	Base
	Codigo  string
	Minutos int
}

type DatosInvitacion struct {
	Base
	NombreNegocio  string
	NombreInvitado string
	Enlace         string
	Dias           int
}

type DatosRecuperacion struct {
	Base
	Enlace  string
	Minutos int
}

type DatosContrasenaCambiada struct {
	Base
	EnlaceRecuperacion string
}

type Config struct {
	NombreApp   string
	FrontendURL string
}

// Servicio expone un método por caso de uso. Cumple de forma estructural las
// interfaces declaradas por `cuenta` y `negocio`, que no importan este paquete.
type Servicio struct {
	transporte Transporte
	plantillas Renderizador
	config     Config
	now        func() time.Time
}

func NewServicio(transporte Transporte, plantillas Renderizador, cfg Config) *Servicio {
	cfg.FrontendURL = strings.TrimRight(cfg.FrontendURL, "/")
	return &Servicio{transporte: transporte, plantillas: plantillas, config: cfg, now: time.Now}
}

func (s *Servicio) SendVerificationOTP(ctx context.Context, recipient, code string, expiresAt time.Time) error {
	return s.enviar(ctx, recipient, PlantillaVerificacion, DatosVerificacion{
		Base: s.base(), Codigo: code, Minutos: s.minutosHasta(expiresAt),
	})
}

func (s *Servicio) SendPasswordReset(ctx context.Context, recipient, challengeID, token string, expiresAt time.Time) error {
	fragmento := url.Values{"desafio": {challengeID}, "token": {token}}.Encode()
	return s.enviar(ctx, recipient, PlantillaRecuperacion, DatosRecuperacion{
		Base:    s.base(),
		Enlace:  s.config.FrontendURL + "/restablecer-contrasena#" + fragmento,
		Minutos: s.minutosHasta(expiresAt),
	})
}

func (s *Servicio) SendPasswordChanged(ctx context.Context, recipient string) error {
	return s.enviar(ctx, recipient, PlantillaContrasenaCambiada, DatosContrasenaCambiada{
		Base: s.base(), EnlaceRecuperacion: s.config.FrontendURL + "/recuperar-contrasena",
	})
}

func (s *Servicio) EnviarInvitacion(ctx context.Context, para, nombreNegocio, nombreInvitado, token string, expiraEn time.Time) error {
	dias := int(math.Round(expiraEn.Sub(s.now()).Hours() / 24))
	if dias < 1 {
		dias = 1
	}
	return s.enviar(ctx, para, PlantillaInvitacion, DatosInvitacion{
		Base:           s.base(),
		NombreNegocio:  nombreNegocio,
		NombreInvitado: nombreInvitado,
		Enlace:         s.config.FrontendURL + "/invitacion/" + url.PathEscape(token),
		Dias:           dias,
	})
}

func (s *Servicio) enviar(ctx context.Context, para, plantilla string, datos any) error {
	contenido, err := s.plantillas.Render(plantilla, datos)
	if err != nil {
		log.Printf("evento=correo_fallido plantilla=%s causa=render error=%v", plantilla, err)
		return ErrEnvio
	}
	inicio := s.now()
	id, err := s.transporte.Enviar(ctx, Mensaje{Para: para, Asunto: contenido.Asunto, HTML: contenido.HTML, Texto: contenido.Texto})
	latencia := s.now().Sub(inicio).Milliseconds()
	if err != nil {
		// El error del transporte no incluye cuerpo ni secretos; el destinatario se enmascara.
		log.Printf("evento=correo_fallido plantilla=%s para=%s latencia_ms=%d error=%v", plantilla, Enmascarar(para), latencia, err)
		return ErrEnvio
	}
	log.Printf("evento=correo_enviado plantilla=%s para=%s id_mensaje=%s latencia_ms=%d", plantilla, Enmascarar(para), id, latencia)
	return nil
}

func (s *Servicio) base() Base {
	return Base{NombreApp: s.config.NombreApp, URLApp: s.config.FrontendURL, Anio: s.now().Year()}
}

func (s *Servicio) minutosHasta(t time.Time) int {
	minutos := int(t.Sub(s.now()).Round(time.Minute).Minutes())
	if minutos < 1 {
		return 1
	}
	return minutos
}

// Enmascarar deja visible la primera letra y el dominio: p***@dominio.com.
func Enmascarar(correo string) string {
	partes := strings.SplitN(correo, "@", 2)
	if len(partes) != 2 || partes[0] == "" {
		return "***"
	}
	return string([]rune(partes[0])[:1]) + "***@" + partes[1]
}
