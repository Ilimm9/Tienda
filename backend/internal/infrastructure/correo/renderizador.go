package correo

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"

	application "tienda/backend/internal/application/correo"
)

// LogoCID identifica el logo embebido; la plantilla base lo referencia como `cid:<LogoCID>`.
const LogoCID = "logo@tienda"

//go:embed plantillas/*.tmpl plantillas/logo.png
var plantillasFS embed.FS

type plantilla struct {
	html  *htmltemplate.Template
	texto *texttemplate.Template
}

// Renderizador compila todas las plantillas al iniciar; una plantilla rota impide arrancar.
type Renderizador struct {
	plantillas map[string]plantilla
}

func NewRenderizador() (*Renderizador, error) {
	nombres := []string{
		application.PlantillaVerificacion,
		application.PlantillaInvitacion,
		application.PlantillaRecuperacion,
		application.PlantillaContrasenaCambiada,
	}
	// html/template bloquea esquemas no web; `cid:` es seguro porque el valor es constante.
	funciones := htmltemplate.FuncMap{"logoSrc": func() htmltemplate.URL { return htmltemplate.URL("cid:" + LogoCID) }}
	r := &Renderizador{plantillas: make(map[string]plantilla, len(nombres))}
	for _, nombre := range nombres {
		html, err := htmltemplate.New(nombre).Funcs(funciones).ParseFS(plantillasFS, "plantillas/base.html.tmpl", "plantillas/"+nombre+".html.tmpl")
		if err != nil {
			return nil, fmt.Errorf("plantilla HTML %s: %w", nombre, err)
		}
		texto, err := texttemplate.New(nombre).ParseFS(plantillasFS, "plantillas/"+nombre+".txt.tmpl")
		if err != nil {
			return nil, fmt.Errorf("plantilla de texto %s: %w", nombre, err)
		}
		r.plantillas[nombre] = plantilla{html: html, texto: texto}
	}
	return r, nil
}

func (r *Renderizador) Render(nombre string, datos any) (application.Contenido, error) {
	p, ok := r.plantillas[nombre]
	if !ok {
		return application.Contenido{}, fmt.Errorf("plantilla desconocida: %s", nombre)
	}
	var asunto, html, texto bytes.Buffer
	if err := p.texto.ExecuteTemplate(&asunto, "asunto", datos); err != nil {
		return application.Contenido{}, err
	}
	if err := p.html.ExecuteTemplate(&html, "base", datos); err != nil {
		return application.Contenido{}, err
	}
	if err := p.texto.ExecuteTemplate(&texto, "texto", datos); err != nil {
		return application.Contenido{}, err
	}
	return application.Contenido{
		// Un asunto nunca debe contener saltos de línea: evita inyección de cabeceras.
		Asunto: strings.Join(strings.Fields(asunto.String()), " "),
		HTML:   html.String(),
		Texto:  strings.TrimSpace(texto.String()) + "\n",
	}, nil
}

// Logo devuelve el PNG embebido que acompaña a cada correo HTML.
func Logo() ([]byte, error) {
	return plantillasFS.ReadFile("plantillas/logo.png")
}
