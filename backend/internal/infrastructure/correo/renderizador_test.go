package correo

import (
	"strings"
	"testing"

	application "tienda/backend/internal/application/correo"
)

func base() application.Base {
	return application.Base{NombreApp: "Tienda", URLApp: "https://tienda.mergemakers.com", Anio: 2026}
}

func TestAllTemplatesRenderSubjectHTMLAndText(t *testing.T) {
	r, err := NewRenderizador()
	if err != nil {
		t.Fatal(err)
	}
	casos := map[string]struct {
		datos    any
		asunto   string
		contiene []string
	}{
		application.PlantillaVerificacion: {
			application.DatosVerificacion{Base: base(), Codigo: "042917", Minutos: 10},
			"Tu código de verificación de Tienda", []string{"042917", "10 minutos"},
		},
		application.PlantillaInvitacion: {
			application.DatosInvitacion{Base: base(), NombreNegocio: "Abarrotes Luna", NombreInvitado: "Luis", Enlace: "https://tienda.mergemakers.com/invitacion/abc", Dias: 7},
			"Abarrotes Luna te invitó a Tienda", []string{"https://tienda.mergemakers.com/invitacion/abc", "7 días"},
		},
		application.PlantillaRecuperacion: {
			application.DatosRecuperacion{Base: base(), Enlace: "https://tienda.mergemakers.com/restablecer-contrasena#desafio=1&token=x", Minutos: 30},
			"Restablece tu contraseña de Tienda", []string{"restablecer-contrasena#desafio=1", "30 minutos"},
		},
		application.PlantillaContrasenaCambiada: {
			application.DatosContrasenaCambiada{Base: base(), EnlaceRecuperacion: "https://tienda.mergemakers.com/recuperar-contrasena"},
			"Tu contraseña de Tienda cambió", []string{"https://tienda.mergemakers.com/recuperar-contrasena"},
		},
	}
	for nombre, caso := range casos {
		t.Run(nombre, func(t *testing.T) {
			contenido, err := r.Render(nombre, caso.datos)
			if err != nil {
				t.Fatal(err)
			}
			if contenido.Asunto != caso.asunto {
				t.Fatalf("asunto = %q, se esperaba %q", contenido.Asunto, caso.asunto)
			}
			if !strings.Contains(contenido.HTML, `src="cid:`+LogoCID+`"`) || !strings.Contains(contenido.HTML, `<html lang="es">`) {
				t.Fatalf("el HTML no incluye el layout con logo embebido")
			}
			for _, texto := range caso.contiene {
				// El HTML escapa `&` en atributos y texto; se compara la versión escapada.
				if !strings.Contains(contenido.HTML, strings.ReplaceAll(texto, "&", "&amp;")) || !strings.Contains(contenido.Texto, texto) {
					t.Fatalf("falta %q en HTML o texto", texto)
				}
			}
			if strings.Contains(contenido.Texto, "<") {
				t.Fatalf("el texto plano contiene HTML: %q", contenido.Texto)
			}
		})
	}
}

func TestHTMLEscapesUserDataAndSubjectHasNoNewlines(t *testing.T) {
	r, err := NewRenderizador()
	if err != nil {
		t.Fatal(err)
	}
	contenido, err := r.Render(application.PlantillaInvitacion, application.DatosInvitacion{
		Base: base(), NombreNegocio: "<script>alert(1)</script>\r\nBcc: x@y.com", NombreInvitado: `"><img src=x>`,
		Enlace: "javascript:alert(1)", Dias: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(contenido.HTML, "<script>") || strings.Contains(contenido.HTML, `"><img src=x>`) {
		t.Fatal("el HTML no escapó datos del usuario")
	}
	if strings.Contains(contenido.HTML, `href="javascript:`) {
		t.Fatal("el HTML permitió un enlace javascript:")
	}
	if strings.ContainsAny(contenido.Asunto, "\r\n") {
		t.Fatalf("el asunto contiene saltos de línea: %q", contenido.Asunto)
	}
	if !strings.Contains(contenido.Texto, "1 día.") {
		t.Fatalf("singular de días incorrecto: %q", contenido.Texto)
	}
}

func TestUnknownTemplateFails(t *testing.T) {
	r, err := NewRenderizador()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Render("inexistente", nil); err == nil {
		t.Fatal("se esperaba error para una plantilla desconocida")
	}
}

func TestLogoIsEmbeddedPNG(t *testing.T) {
	logo, err := Logo()
	if err != nil || len(logo) < 8 || string(logo[1:4]) != "PNG" {
		t.Fatalf("logo inválido: len=%d err=%v", len(logo), err)
	}
}
