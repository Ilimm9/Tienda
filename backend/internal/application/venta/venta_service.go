package venta

import (
	"errors"
	"strings"
	"time"
	"unicode"

	"tienda/backend/internal/domain/venta"

	"github.com/google/uuid"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	MaxLineas           = 200
	MaxCantidadLinea    = 9999
	MaxResultados       = 20
	MinDigitosCodigo    = 6
	MaxLongitudBusqueda = 120
)

var (
	ErrVentaInvalida        = errors.New("revisa los productos de la venta")
	ErrProductoNoDisponible = errors.New("uno de los productos ya no está disponible para venta en esta sucursal")
	ErrVentaNoEncontrada    = errors.New("venta no encontrada")
	ErrVentaYaCancelada     = errors.New("la venta ya está cancelada")
	ErrMotivoRequerido      = errors.New("escribe el motivo de la cancelación")
	ErrVentaFueraDePlazo    = errors.New("solo se pueden cancelar ventas del mismo día")
	ErrPeriodoInvalido      = errors.New("el periodo no es válido")
	ErrSucursalInvalida     = errors.New("la sucursal no existe o está inactiva")
	ErrReferenciaLarga      = errors.New("la referencia es demasiado larga")
)

// ProductoVendible es una coincidencia de búsqueda con precio y existencia de la sucursal.
type ProductoVendible struct {
	ProductoNegocioID uuid.UUID       `json:"producto_negocio_id"`
	Nombre            string          `json:"nombre"`
	CodigoBarras      *string         `json:"codigo_barras"`
	Precio            venta.Centavos  `json:"precio"`
	Existencia        int             `json:"existencia"`
	Agotado           bool            `json:"agotado"`
	ExistenciaBaja    bool            `json:"existencia_baja"`
	Costo             *venta.Centavos `json:"costo,omitempty"`
	Margen            *float64        `json:"margen,omitempty"`
	NivelMargen       *string         `json:"nivel_margen,omitempty"`
}

type ResultadoBusqueda struct {
	Items     []ProductoVendible `json:"items"`
	PorCodigo bool               `json:"por_codigo"`
}

type LineaInput struct {
	ProductoNegocioID string `json:"producto_negocio_id"`
	Cantidad          int    `json:"cantidad"`
}

type RegistrarInput struct {
	SucursalID string          `json:"sucursal_id"`
	FormaPago  string          `json:"forma_pago"`
	Recibido   *venta.Centavos `json:"recibido"`
	Referencia *string         `json:"referencia"`
	Lineas     []LineaInput    `json:"lineas"`
}

// LineaSolicitada es una línea validada y agrupada por producto.
type LineaSolicitada struct {
	ProductoNegocioID uuid.UUID
	Cantidad          int
}

type RegistrarComando struct {
	NegocioID  uuid.UUID
	SucursalID uuid.UUID
	UsuarioID  uuid.UUID
	FormaPago  string
	Recibido   *venta.Centavos
	Referencia *string
	Lineas     []LineaSolicitada
}

type CancelarComando struct {
	NegocioID    uuid.UUID
	VentaID      uuid.UUID
	UsuarioID    uuid.UUID
	MotivoCodigo string
	Detalle      string
	// Motivo es el texto que se guarda; lo arma el servicio a partir del código y el detalle.
	Motivo string
	// Sucursales limita la cancelación a las sucursales accesibles del usuario.
	Sucursales []uuid.UUID
}

// Filtro acota el listado; Sucursales siempre llega resuelto por el acceso del usuario.
type Filtro struct {
	Sucursales []uuid.UUID
	UsuarioID  *uuid.UUID
	Folio      string
	FormaPago  string
	Estado     string
	Desde      *time.Time
	Hasta      *time.Time
	// Periodo `hoy`, `semana` o `mes` en el calendario del negocio; reemplaza Desde y Hasta.
	Periodo string
}

// ResumenVentas acumula las ventas pagadas del periodo; las canceladas solo se cuentan.
type ResumenVentas struct {
	Vendido        venta.Centavos `json:"vendido"`
	Tickets        int            `json:"tickets"`
	TicketPromedio venta.Centavos `json:"ticket_promedio"`
	Efectivo       venta.Centavos `json:"efectivo"`
	Cancelados     int            `json:"cancelados"`
}

type ListadoVentas struct {
	Items   []venta.Venta `json:"items"`
	Total   int           `json:"total"`
	Resumen ResumenVentas `json:"resumen"`
	Desde   *time.Time    `json:"desde,omitempty"`
	Hasta   *time.Time    `json:"hasta,omitempty"`
}

type Repository interface {
	BuscarPorCodigo(negocioID, sucursalID uuid.UUID, codigo string) ([]ProductoVendible, error)
	BuscarPorNombre(negocioID, sucursalID uuid.UUID, palabras []string, limite int) ([]ProductoVendible, error)
	Registrar(RegistrarComando) (venta.Venta, error)
	Cancelar(CancelarComando) (venta.Venta, error)
	Listar(negocioID uuid.UUID, filtro Filtro) (ListadoVentas, error)
	Obtener(negocioID, ventaID uuid.UUID) (venta.Venta, error)
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

// Buscar intenta primero el código exacto; un texto de solo dígitos largo se trata como escaneo y no cae a nombre.
func (s *Service) Buscar(negocioID, sucursalID uuid.UUID, texto string, verCostos bool) (ResultadoBusqueda, error) {
	texto = strings.TrimSpace(texto)
	if texto == "" || len([]rune(texto)) > MaxLongitudBusqueda {
		return ResultadoBusqueda{Items: []ProductoVendible{}}, nil
	}
	items, err := s.repo.BuscarPorCodigo(negocioID, sucursalID, texto)
	if err != nil {
		return ResultadoBusqueda{}, err
	}
	if len(items) > 0 {
		return ResultadoBusqueda{Items: conCostos(items, verCostos), PorCodigo: true}, nil
	}
	if EsCodigoEscaneado(texto) {
		return ResultadoBusqueda{Items: []ProductoVendible{}, PorCodigo: true}, nil
	}
	palabras := strings.Fields(Normalizar(texto))
	if len(palabras) == 0 {
		return ResultadoBusqueda{Items: []ProductoVendible{}}, nil
	}
	items, err = s.repo.BuscarPorNombre(negocioID, sucursalID, palabras, MaxResultados)
	if err != nil {
		return ResultadoBusqueda{}, err
	}
	return ResultadoBusqueda{Items: conCostos(items, verCostos)}, nil
}

func (s *Service) Registrar(negocioID, usuarioID uuid.UUID, input RegistrarInput) (venta.Venta, error) {
	comando, err := ValidarRegistro(negocioID, usuarioID, input)
	if err != nil {
		return venta.Venta{}, err
	}
	return s.repo.Registrar(comando)
}

func (s *Service) Cancelar(comando CancelarComando) (venta.Venta, error) {
	motivo, err := venta.MotivoCancelacion(comando.MotivoCodigo, comando.Detalle)
	if err != nil {
		return venta.Venta{}, err
	}
	comando.Motivo = motivo
	return s.repo.Cancelar(comando)
}

func (s *Service) Listar(negocioID uuid.UUID, filtro Filtro) (ListadoVentas, error) {
	switch filtro.Periodo {
	case "", "hoy", "semana", "mes":
	default:
		return ListadoVentas{}, ErrPeriodoInvalido
	}
	if len(filtro.Sucursales) == 0 {
		return ListadoVentas{Items: []venta.Venta{}}, nil
	}
	return s.repo.Listar(negocioID, filtro)
}

// Promedio divide el vendido entre los tickets redondeando al centavo.
func Promedio(vendido venta.Centavos, tickets int) venta.Centavos {
	if tickets <= 0 {
		return 0
	}
	return (vendido*2 + venta.Centavos(tickets)) / venta.Centavos(2*tickets)
}

func (s *Service) Obtener(negocioID, ventaID uuid.UUID) (venta.Venta, error) {
	return s.repo.Obtener(negocioID, ventaID)
}

// ValidarRegistro revisa forma de pago, referencia y líneas, y suma las líneas repetidas del mismo producto.
func ValidarRegistro(negocioID, usuarioID uuid.UUID, input RegistrarInput) (RegistrarComando, error) {
	sucursalID, err := uuid.Parse(input.SucursalID)
	if err != nil {
		return RegistrarComando{}, ErrSucursalInvalida
	}
	switch input.FormaPago {
	case venta.PagoEfectivo, venta.PagoTarjeta, venta.PagoTransferencia:
	default:
		return RegistrarComando{}, venta.ErrFormaPagoInvalida
	}
	if input.Recibido != nil && *input.Recibido < 0 {
		return RegistrarComando{}, venta.ErrImporteInvalido
	}
	if len(input.Lineas) == 0 || len(input.Lineas) > MaxLineas {
		return RegistrarComando{}, ErrVentaInvalida
	}
	lineas := make([]LineaSolicitada, 0, len(input.Lineas))
	posicion := map[uuid.UUID]int{}
	for _, linea := range input.Lineas {
		productoID, err := uuid.Parse(linea.ProductoNegocioID)
		if err != nil || linea.Cantidad <= 0 || linea.Cantidad > MaxCantidadLinea {
			return RegistrarComando{}, ErrVentaInvalida
		}
		if i, ok := posicion[productoID]; ok {
			lineas[i].Cantidad += linea.Cantidad
			if lineas[i].Cantidad > MaxCantidadLinea {
				return RegistrarComando{}, ErrVentaInvalida
			}
			continue
		}
		posicion[productoID] = len(lineas)
		lineas = append(lineas, LineaSolicitada{ProductoNegocioID: productoID, Cantidad: linea.Cantidad})
	}
	var referencia *string
	if input.FormaPago != venta.PagoEfectivo && input.Referencia != nil {
		if texto := strings.TrimSpace(*input.Referencia); texto != "" {
			if len([]rune(texto)) > 120 {
				return RegistrarComando{}, ErrReferenciaLarga
			}
			referencia = &texto
		}
	}
	return RegistrarComando{NegocioID: negocioID, SucursalID: sucursalID, UsuarioID: usuarioID, FormaPago: input.FormaPago, Recibido: input.Recibido, Referencia: referencia, Lineas: lineas}, nil
}

// EsCodigoEscaneado reconoce la lectura de un lector de códigos: solo dígitos, 6 o más.
func EsCodigoEscaneado(texto string) bool {
	if len(texto) < MinDigitosCodigo {
		return false
	}
	for _, r := range texto {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// Normalizar quita acentos y mayúsculas; el SQL aplica la misma regla con translate().
func Normalizar(texto string) string {
	sinAcentos, _, err := transform.String(transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC), texto)
	if err != nil {
		sinAcentos = texto
	}
	return strings.ToLower(sinAcentos)
}

// conCostos agrega margen y nivel cuando el usuario puede ver costos; si no, los quita.
func conCostos(items []ProductoVendible, verCostos bool) []ProductoVendible {
	for i := range items {
		if !verCostos || items[i].Costo == nil {
			items[i].Costo, items[i].Margen, items[i].NivelMargen = nil, nil, nil
			continue
		}
		if margen, ok := venta.Margen(items[i].Precio, *items[i].Costo); ok {
			nivel := venta.NivelMargen(items[i].Precio, *items[i].Costo)
			items[i].Margen, items[i].NivelMargen = &margen, &nivel
		}
	}
	return items
}

// OcultarCostos quita el costo de cada línea para usuarios sin `ventas.ver_costos`.
func OcultarCostos(v *venta.Venta) {
	for i := range v.Lineas {
		v.Lineas[i].CostoUnitario = nil
	}
}
