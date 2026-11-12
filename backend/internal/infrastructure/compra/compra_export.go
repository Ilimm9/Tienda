package compra

import (
	"fmt"
	"strings"
	"time"

	compraapp "tienda/backend/internal/application/compra"
	compradomain "tienda/backend/internal/domain/compra"
	preciodomain "tienda/backend/internal/domain/precio"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// extraRenglon trae lo que el renglón no guarda: producto, unidad y la foto de la propuesta.
type extraRenglon struct {
	DetalleID        uuid.UUID
	Producto         string
	Contenido        *float64
	UnidadContenido  *string
	Unidad           string
	CostoAnterior    *float64
	PrecioAnterior   *float64
	PrecioSugerido   *float64
	PrecioAutorizado *float64
	Estado           *string
	AutorizadoEn     *time.Time
	Autorizador      *string
}

// ListExport devuelve las compras del filtro con sus renglones vigentes, en orden de recepción.
func (r *Repository) ListExport(businessID uuid.UUID, filtro compraapp.FiltroExportacion) ([]compraapp.CompraExportacion, error) {
	var compras []compradomain.CompraProveedor
	query := r.db.Preload("Detalles", func(db *gorm.DB) *gorm.DB { return db.Where("cancelado = FALSE").Order("creado_en ASC") }).
		Where("negocio_id = ? AND sucursal_id = ?", businessID, filtro.SucursalID)
	if len(filtro.CompraIDs) > 0 {
		query = query.Where("id IN ?", filtro.CompraIDs)
	} else {
		if filtro.ProveedorID != nil {
			query = query.Where("proveedor_id = ?", *filtro.ProveedorID)
		}
		if filtro.Desde != nil {
			query = query.Where("fecha_recepcion >= ?", filtro.Desde.Format("2006-01-02"))
		}
		if filtro.Hasta != nil {
			query = query.Where("fecha_recepcion <= ?", filtro.Hasta.Format("2006-01-02"))
		}
	}
	if !filtro.IncluirSuspendidas {
		query = query.Where("estado = ?", compradomain.EstadoActiva)
	}
	if err := query.Order("fecha_recepcion ASC, creado_en ASC").Find(&compras).Error; err != nil {
		return nil, err
	}
	if len(compras) == 0 {
		return []compraapp.CompraExportacion{}, nil
	}
	proveedores := map[uuid.UUID]string{}
	detalleIDs := []uuid.UUID{}
	for _, compra := range compras {
		proveedores[compra.ProveedorID] = ""
		for _, detalle := range compra.Detalles {
			detalleIDs = append(detalleIDs, detalle.ID)
		}
	}
	var nombres []struct {
		ID     uuid.UUID
		Nombre string
	}
	ids := make([]uuid.UUID, 0, len(proveedores))
	for id := range proveedores {
		ids = append(ids, id)
	}
	if err := r.db.Table("proveedores").Select("id, nombre").Where("id IN ?", ids).Scan(&nombres).Error; err != nil {
		return nil, err
	}
	for _, item := range nombres {
		proveedores[item.ID] = item.Nombre
	}
	extras := map[uuid.UUID]extraRenglon{}
	if len(detalleIDs) > 0 {
		var filas []extraRenglon
		err := r.db.Table("compras_proveedor_detalles d").Select(`d.id AS detalle_id,
			COALESCE(pr.nombre, vp.nombre) || COALESCE(' · ' || pv.clave, '') AS producto,
			COALESCE(pr.contenido, vp.contenido) AS contenido, COALESCE(pr.unidad_contenido, vp.unidad_contenido) AS unidad_contenido,
			COALESCE(u.nombre, '') AS unidad, p.costo_anterior, p.precio_anterior, p.precio_sugerido, p.precio_autorizado, p.estado, p.autorizado_en,
			NULLIF(TRIM(COALESCE(pu.nombres, '') || ' ' || COALESCE(pu.primer_apellido, '')), '') AS autorizador`).
			Joins("JOIN producto_negocio pn ON pn.id = d.producto_negocio_id").Joins("LEFT JOIN productos pr ON pr.id = pn.producto_id").
			Joins("LEFT JOIN producto_variantes pv ON pv.id = pn.producto_variante_id").Joins("LEFT JOIN productos vp ON vp.id = pv.producto_id").
			Joins("LEFT JOIN unidades_medida u ON u.id = d.unidad_medida_id").Joins("LEFT JOIN propuestas_costo_precio p ON p.compra_detalle_id = d.id").
			Joins("LEFT JOIN perfil_usuarios pu ON pu.usuario_id = p.autorizado_por_usuario_id").
			Where("d.id IN ?", detalleIDs).Scan(&filas).Error
		if err != nil {
			return nil, err
		}
		for _, fila := range filas {
			extras[fila.DetalleID] = fila
		}
	}
	result := make([]compraapp.CompraExportacion, 0, len(compras))
	for _, compra := range compras {
		compra.CompletarImpuestosRenglon()
		item := compraapp.CompraExportacion{ID: compra.ID, Ticket: ticket(compra), FechaRecepcion: compra.FechaRecepcion, Proveedor: proveedores[compra.ProveedorID], Estado: compra.Estado, IEPS: compra.IEPS, Total: compra.Total}
		if compra.Observaciones != nil {
			item.Observaciones = *compra.Observaciones
		}
		for _, detalle := range compra.Detalles {
			extra := extras[detalle.ID]
			item.Descuento = money(item.Descuento + detalle.Descuento)
			renglon := compraapp.RenglonExportacion{CompraID: compra.ID, Producto: extra.Producto, TasaIEPS: detalle.TasaIEPS, ImpuestosPorRenglon: detalle.ImpuestosPorRenglon,
				Subtotal: detalle.Subtotal, IEPS: detalle.IEPS, IVA: detalle.IVA, TasaIVA: detalle.TasaIVA, CantidadEmpaques: detalle.CantidadEmpaques, Unidad: extra.Unidad,
				PiezasPorEmpaque: detalle.PiezasPorEmpaque, TotalPiezas: detalle.TotalPiezas, CostoFinal: compradomain.CostoFinal(compra, detalle), CostoAnterior: extra.CostoAnterior,
				PrecioAnterior: extra.PrecioAnterior, ProductoNuevo: detalle.ProductoNuevo, PiezasDanadas: detalle.PiezasDanadas, PiezasFaltantes: detalle.PiezasFaltantes,
				PiezasBuenas: detalle.PiezasBuenas, Descuento: detalle.Descuento, Contenido: contenido(extra)}
			if detalle.CodigoBarras != nil {
				renglon.Codigo = *detalle.CodigoBarras
			}
			renglon.PrecioNuevo, renglon.EstadoValidacion = validacion(extra)
			if detalle.ProductoNuevo && extra.Estado != nil && *extra.Estado == preciodomain.EstadoAutorizada {
				renglon.PrecioAutorizado = extra.PrecioAutorizado
			}
			item.Renglones = append(item.Renglones, renglon)
		}
		result = append(result, item)
	}
	return result, nil
}

func ticket(compra compradomain.CompraProveedor) string {
	return compra.TipoDocumento + " " + compra.FolioDocumento
}

func contenido(extra extraRenglon) string {
	if extra.Contenido == nil {
		return ""
	}
	texto := strings.TrimSuffix(strings.TrimRight(fmt.Sprintf("%.3f", *extra.Contenido), "0"), ".")
	if extra.UnidadContenido != nil {
		texto += " " + *extra.UnidadContenido
	}
	return texto
}

// validacion devuelve el NUEVO P.V (autorizado o sugerido) y el texto del estado.
func validacion(extra extraRenglon) (*float64, string) {
	if extra.Estado == nil {
		return nil, ""
	}
	switch *extra.Estado {
	case preciodomain.EstadoAutorizada:
		texto := "Validado"
		if extra.AutorizadoEn != nil {
			texto += " el " + extra.AutorizadoEn.Format("02/01")
		}
		if extra.Autorizador != nil {
			texto += " por " + *extra.Autorizador
		}
		return extra.PrecioAutorizado, texto
	case preciodomain.EstadoCancelada:
		return extra.PrecioSugerido, "Cancelado"
	default:
		return extra.PrecioSugerido, "Pendiente"
	}
}
