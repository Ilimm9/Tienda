package venta

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	EstadoPagada    = "PAGADA"
	EstadoCancelada = "CANCELADA"

	PagoEfectivo      = "EFECTIVO"
	PagoTarjeta       = "TARJETA"
	PagoTransferencia = "TRANSFERENCIA"

	MovimientoVentaSalida      = "VENTA_SALIDA"
	MovimientoVentaCancelacion = "VENTA_CANCELACION"

	MotivoErrorCaptura        = "ERROR_CAPTURA"
	MotivoCobroDuplicado      = "COBRO_DUPLICADO"
	MotivoDevolucionInmediata = "DEVOLUCION_INMEDIATA"
	MotivoOtro                = "OTRO"

	NivelMargenNegativo = "negativo"
	NivelMargenBajo     = "bajo"
	NivelMargenOK       = "ok"
)

// Venta guarda el ticket cobrado con importes en centavos y la foto de cada línea.
type Venta struct {
	ID                      uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID               uuid.UUID    `gorm:"type:uuid;not null;index;uniqueIndex:idx_venta_negocio_numero,priority:1" json:"negocio_id"`
	SucursalID              uuid.UUID    `gorm:"type:uuid;not null;index" json:"sucursal_id"`
	Numero                  int64        `gorm:"not null;uniqueIndex:idx_venta_negocio_numero,priority:2" json:"numero"`
	Folio                   string       `gorm:"type:varchar(20);not null" json:"folio"`
	UsuarioID               uuid.UUID    `gorm:"type:uuid;not null;index" json:"usuario_id"`
	Subtotal                Centavos     `gorm:"column:subtotal_centavos;type:bigint;not null" json:"subtotal"`
	IVA                     Centavos     `gorm:"column:iva_centavos;type:bigint;not null" json:"iva"`
	Total                   Centavos     `gorm:"column:total_centavos;type:bigint;not null" json:"total"`
	Recibido                Centavos     `gorm:"column:recibido_centavos;type:bigint;not null" json:"recibido"`
	Cambio                  Centavos     `gorm:"column:cambio_centavos;type:bigint;not null" json:"cambio"`
	Piezas                  int          `gorm:"not null" json:"piezas"`
	FormaPago               string       `gorm:"type:varchar(20);not null" json:"forma_pago"`
	Referencia              *string      `gorm:"type:varchar(120)" json:"referencia,omitempty"`
	Estado                  string       `gorm:"type:varchar(20);not null;default:PAGADA" json:"estado"`
	MotivoCancelacion       *string      `gorm:"type:text" json:"motivo_cancelacion,omitempty"`
	MotivoCancelacionCodigo *string      `gorm:"type:varchar(30)" json:"motivo_cancelacion_codigo,omitempty"`
	CanceladaPorUsuarioID   *uuid.UUID   `gorm:"type:uuid" json:"cancelada_por_usuario_id,omitempty"`
	CanceladaEn             *time.Time   `json:"cancelada_en,omitempty"`
	CreadoEn                time.Time    `gorm:"autoCreateTime;index" json:"creado_en"`
	Lineas                  []VentaLinea `gorm:"foreignKey:VentaID" json:"lineas,omitempty"`

	// Datos de presentación que arma la consulta; no se guardan.
	Atendio      string  `gorm:"-" json:"atendio"`
	CanceladaPor *string `gorm:"-" json:"cancelada_por,omitempty"`
	Cancelable   bool    `gorm:"-" json:"cancelable"`
}

// VentaLinea conserva nombre, código, precio y costo del momento del cobro.
type VentaLinea struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	VentaID           uuid.UUID `gorm:"type:uuid;not null;index" json:"venta_id"`
	ProductoNegocioID uuid.UUID `gorm:"type:uuid;not null;index" json:"producto_negocio_id"`
	CodigoBarras      *string   `gorm:"type:varchar(120)" json:"codigo_barras,omitempty"`
	Nombre            string    `gorm:"type:varchar(300);not null" json:"nombre"`
	Cantidad          int       `gorm:"not null" json:"cantidad"`
	PrecioUnitario    Centavos  `gorm:"column:precio_unitario_centavos;type:bigint;not null" json:"precio_unitario"`
	CostoUnitario     *Centavos `gorm:"column:costo_unitario_centavos;type:bigint" json:"costo_unitario,omitempty"`
	Importe           Centavos  `gorm:"column:importe_centavos;type:bigint;not null" json:"importe"`
	Orden             int       `gorm:"not null" json:"-"`
}

// VentaFolioConsecutivo reserva el siguiente número de folio por negocio.
type VentaFolioConsecutivo struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	NegocioID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	SiguienteNumero int64     `gorm:"not null"`
	ActualizadoEn   time.Time `gorm:"autoUpdateTime"`
}

func (Venta) TableName() string                 { return "ventas" }
func (VentaLinea) TableName() string            { return "venta_lineas" }
func (VentaFolioConsecutivo) TableName() string { return "venta_folio_consecutivos" }

func (v *Venta) BeforeCreate(*gorm.DB) error                 { setID(&v.ID); return nil }
func (l *VentaLinea) BeforeCreate(*gorm.DB) error            { setID(&l.ID); return nil }
func (f *VentaFolioConsecutivo) BeforeCreate(*gorm.DB) error { setID(&f.ID); return nil }

func setID(id *uuid.UUID) {
	if *id == uuid.Nil {
		*id = uuid.New()
	}
}

// FormatoFolio presenta el número consecutivo como V-000232.
func FormatoFolio(numero int64) string { return fmt.Sprintf("V-%06d", numero) }

// DesgloseIVA separa el IVA del 16 % incluido en el total: subtotal = round(total ÷ 1.16).
func DesgloseIVA(total Centavos) (subtotal, iva Centavos) {
	subtotal = Centavos((int64(total)*100 + 58) / 116)
	return subtotal, total - subtotal
}

var (
	ErrFormaPagoInvalida = errors.New("la forma de pago no es válida")
	ErrPagoInsuficiente  = errors.New("el importe recibido es menor al total")
)

// Cobro devuelve lo recibido y el cambio según la forma de pago.
// Tarjeta y transferencia cobran exactamente el total.
func Cobro(formaPago string, total Centavos, recibido *Centavos) (Centavos, Centavos, error) {
	switch formaPago {
	case PagoEfectivo:
		if recibido == nil || *recibido < total {
			return 0, 0, ErrPagoInsuficiente
		}
		return *recibido, *recibido - total, nil
	case PagoTarjeta, PagoTransferencia:
		return total, 0, nil
	default:
		return 0, 0, ErrFormaPagoInvalida
	}
}

// Margen calcula (precio − costo) ÷ precio con cuatro decimales; sin precio no hay margen.
func Margen(precio, costo Centavos) (float64, bool) {
	if precio <= 0 {
		return 0, false
	}
	diezmilesimas := (int64(precio-costo)*10000*2 + sign(int64(precio-costo))*int64(precio)) / (int64(precio) * 2)
	return float64(diezmilesimas) / 10000, true
}

// NivelMargen clasifica el margen para la alerta de «Verificar costos».
func NivelMargen(precio, costo Centavos) string {
	switch {
	case costo > precio:
		return NivelMargenNegativo
	case int64(precio-costo)*100 < int64(precio)*20:
		return NivelMargenBajo
	default:
		return NivelMargenOK
	}
}

func sign(v int64) int64 {
	if v < 0 {
		return -1
	}
	return 1
}

var (
	ErrMotivoInvalido   = errors.New("elige el motivo de la cancelación")
	ErrDetalleRequerido = errors.New("explica el motivo de la cancelación")
	etiquetasMotivo     = map[string]string{MotivoErrorCaptura: "Error de captura", MotivoCobroDuplicado: "Cobro duplicado", MotivoDevolucionInmediata: "Devolución inmediata del cliente", MotivoOtro: "Otro"}
	MaxLongitudDetalle  = 300
)

// MotivoCancelacion valida el motivo de lista y arma el texto guardado: la etiqueta,
// seguida de la explicación cuando la hay. «Otro» exige explicación.
func MotivoCancelacion(codigo, detalle string) (string, error) {
	etiqueta, ok := etiquetasMotivo[codigo]
	if !ok {
		return "", ErrMotivoInvalido
	}
	detalle = strings.TrimSpace(detalle)
	if len([]rune(detalle)) > MaxLongitudDetalle {
		return "", ErrDetalleRequerido
	}
	if detalle == "" {
		if codigo == MotivoOtro {
			return "", ErrDetalleRequerido
		}
		return etiqueta, nil
	}
	return etiqueta + ": " + detalle, nil
}

// ExistenciaInsuficienteError nombra el producto que ya no alcanza al confirmar.
type ExistenciaInsuficienteError struct {
	Producto   string
	Disponible int
	Solicitado int
}

func (e *ExistenciaInsuficienteError) Error() string {
	return fmt.Sprintf("Ya no hay existencias suficientes de %s: quedan %d y pediste %d", e.Producto, e.Disponible, e.Solicitado)
}
