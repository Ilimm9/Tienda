package venta

import (
	"errors"
	"testing"

	"tienda/backend/internal/domain/venta"

	"github.com/google/uuid"
)

type repoFalso struct {
	porCodigo   []ProductoVendible
	porNombre   []ProductoVendible
	palabras    []string
	codigos     int
	nombres     int
	registrado  *RegistrarComando
	cancelacion *CancelarComando
}

func (r *repoFalso) BuscarPorCodigo(_, _ uuid.UUID, _ string) ([]ProductoVendible, error) {
	r.codigos++
	return append([]ProductoVendible(nil), r.porCodigo...), nil
}
func (r *repoFalso) BuscarPorNombre(_, _ uuid.UUID, palabras []string, _ int) ([]ProductoVendible, error) {
	r.nombres++
	r.palabras = palabras
	return append([]ProductoVendible(nil), r.porNombre...), nil
}
func (r *repoFalso) Registrar(c RegistrarComando) (venta.Venta, error) {
	r.registrado = &c
	return venta.Venta{}, nil
}
func (r *repoFalso) Cancelar(c CancelarComando) (venta.Venta, error) {
	r.cancelacion = &c
	return venta.Venta{}, nil
}
func (r *repoFalso) Listar(uuid.UUID, Filtro) (ListadoVentas, error)   { return ListadoVentas{}, nil }
func (r *repoFalso) Obtener(uuid.UUID, uuid.UUID) (venta.Venta, error) { return venta.Venta{}, nil }

func TestBuscarCodigoPrimeroYNombreSinAcentos(t *testing.T) {
	costo := venta.Centavos(1500)
	producto := ProductoVendible{Nombre: "Coca-Cola 600 ml", Precio: 2000, Costo: &costo}

	repo := &repoFalso{porCodigo: []ProductoVendible{producto}}
	resultado, _ := NewService(repo).Buscar(uuid.New(), uuid.New(), "123", true)
	if !resultado.PorCodigo || len(resultado.Items) != 1 || repo.nombres != 0 {
		t.Fatalf("un código corto exacto debe encontrarse sin buscar por nombre: %+v", resultado)
	}
	if resultado.Items[0].Margen == nil || *resultado.Items[0].Margen != 0.25 || *resultado.Items[0].NivelMargen != venta.NivelMargenOK {
		t.Fatalf("con permiso de costos se calcula el margen: %+v", resultado.Items[0])
	}

	repo = &repoFalso{porNombre: []ProductoVendible{producto}}
	resultado, _ = NewService(repo).Buscar(uuid.New(), uuid.New(), "7501055300075", false)
	if !resultado.PorCodigo || len(resultado.Items) != 0 || repo.nombres != 0 {
		t.Fatalf("un escaneo que no existe no cae a búsqueda por nombre: %+v", resultado)
	}

	repo = &repoFalso{porNombre: []ProductoVendible{producto}}
	resultado, _ = NewService(repo).Buscar(uuid.New(), uuid.New(), "  GALLETAS  Marías ", false)
	if resultado.PorCodigo || repo.nombres != 1 || len(repo.palabras) != 2 || repo.palabras[0] != "galletas" || repo.palabras[1] != "marias" {
		t.Fatalf("la búsqueda por nombre normaliza acentos y mayúsculas: %v", repo.palabras)
	}
	if resultado.Items[0].Costo != nil || resultado.Items[0].Margen != nil {
		t.Fatalf("sin permiso de costos no se envía costo: %+v", resultado.Items[0])
	}

	repo = &repoFalso{}
	if resultado, _ := NewService(repo).Buscar(uuid.New(), uuid.New(), "   ", false); len(resultado.Items) != 0 || repo.codigos != 0 {
		t.Fatal("un texto vacío no consulta la base")
	}
}

func TestValidarRegistro(t *testing.T) {
	negocio, usuario, sucursal, producto := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	otro := uuid.New()
	referencia := "  AUT-123 "
	comando, err := ValidarRegistro(negocio, usuario, RegistrarInput{SucursalID: sucursal.String(), FormaPago: venta.PagoTarjeta, Referencia: &referencia,
		Lineas: []LineaInput{{producto.String(), 1}, {otro.String(), 2}, {producto.String(), 1}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(comando.Lineas) != 2 || comando.Lineas[0].ProductoNegocioID != producto || comando.Lineas[0].Cantidad != 2 || comando.Lineas[1].Cantidad != 2 {
		t.Fatalf("las líneas repetidas se suman y conservan el orden: %+v", comando.Lineas)
	}
	if comando.Referencia == nil || *comando.Referencia != "AUT-123" {
		t.Fatalf("la referencia se recorta: %v", comando.Referencia)
	}
	comando, _ = ValidarRegistro(negocio, usuario, RegistrarInput{SucursalID: sucursal.String(), FormaPago: venta.PagoEfectivo, Referencia: &referencia, Lineas: []LineaInput{{producto.String(), 1}}})
	if comando.Referencia != nil {
		t.Fatal("efectivo no guarda referencia")
	}
	negativo := venta.Centavos(-1)
	for nombre, input := range map[string]RegistrarInput{
		"sin líneas":        {SucursalID: sucursal.String(), FormaPago: venta.PagoEfectivo},
		"cantidad cero":     {SucursalID: sucursal.String(), FormaPago: venta.PagoEfectivo, Lineas: []LineaInput{{producto.String(), 0}}},
		"cantidad enorme":   {SucursalID: sucursal.String(), FormaPago: venta.PagoEfectivo, Lineas: []LineaInput{{producto.String(), MaxCantidadLinea}, {producto.String(), 1}}},
		"producto inválido": {SucursalID: sucursal.String(), FormaPago: venta.PagoEfectivo, Lineas: []LineaInput{{"x", 1}}},
		"forma de pago":     {SucursalID: sucursal.String(), FormaPago: "CREDITO", Lineas: []LineaInput{{producto.String(), 1}}},
		"recibido negativo": {SucursalID: sucursal.String(), FormaPago: venta.PagoEfectivo, Recibido: &negativo, Lineas: []LineaInput{{producto.String(), 1}}},
		"sucursal":          {SucursalID: "x", FormaPago: venta.PagoEfectivo, Lineas: []LineaInput{{producto.String(), 1}}},
	} {
		if _, err := ValidarRegistro(negocio, usuario, input); err == nil {
			t.Fatalf("%s debe rechazarse", nombre)
		}
	}
}

func TestCancelarExigeMotivoDeLista(t *testing.T) {
	repo := &repoFalso{}
	for _, comando := range []CancelarComando{{MotivoCodigo: ""}, {MotivoCodigo: "ME_ARREPENTI"}, {MotivoCodigo: venta.MotivoOtro, Detalle: "   "}} {
		if _, err := NewService(repo).Cancelar(comando); err == nil || repo.cancelacion != nil {
			t.Fatalf("%+v debe rechazarse: %v", comando, err)
		}
	}
	if _, err := NewService(repo).Cancelar(CancelarComando{MotivoCodigo: venta.MotivoCobroDuplicado}); err != nil || repo.cancelacion.Motivo != "Cobro duplicado" {
		t.Fatalf("motivo de lista: %+v %v", repo.cancelacion, err)
	}
	if _, err := NewService(repo).Cancelar(CancelarComando{MotivoCodigo: venta.MotivoOtro, Detalle: " Cliente sin efectivo "}); err != nil || repo.cancelacion.Motivo != "Otro: Cliente sin efectivo" {
		t.Fatalf("«Otro» guarda la explicación: %+v %v", repo.cancelacion, err)
	}
}

func TestListarPeriodoYPromedio(t *testing.T) {
	if _, err := NewService(&repoFalso{}).Listar(uuid.New(), Filtro{Periodo: "anio", Sucursales: []uuid.UUID{uuid.New()}}); !errors.Is(err, ErrPeriodoInvalido) {
		t.Fatalf("periodo desconocido: %v", err)
	}
	if Promedio(53800, 6) != 8967 || Promedio(0, 0) != 0 || Promedio(100, 3) != 33 {
		t.Fatalf("promedio: %d %d %d", Promedio(53800, 6), Promedio(0, 0), Promedio(100, 3))
	}
}

func TestOcultarCostos(t *testing.T) {
	costo := venta.Centavos(100)
	v := venta.Venta{Lineas: []venta.VentaLinea{{CostoUnitario: &costo}}}
	OcultarCostos(&v)
	if v.Lineas[0].CostoUnitario != nil {
		t.Fatal("el costo debe quitarse")
	}
}
