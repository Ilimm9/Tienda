package venta

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	ventaapp "tienda/backend/internal/application/venta"
	"tienda/backend/internal/domain"
	ventadomain "tienda/backend/internal/domain/venta"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// vendibleSQL arma productos activos con precio mayor a 0 en la sucursal; la existencia
// se redondea hacia abajo a piezas completas y es baja al llegar al mínimo de la sucursal. Montos en centavos para no usar flotantes.
const vendibleSQL = `SELECT pn.id AS producto_negocio_id,
	COALESCE(p.nombre, vp.nombre) || COALESCE(' · ' || pv.clave, '') AS nombre,
	(SELECT pc.codigo FROM producto_codigos pc
		WHERE pc.negocio_id = pn.negocio_id AND (pc.producto_id = pn.producto_id OR pc.producto_variante_id = pn.producto_variante_id)
		ORDER BY pc.es_principal DESC, pc.codigo LIMIT 1) AS codigo_barras,
	round(ps.precio_venta * 100)::bigint AS precio,
	CASE WHEN ps.costo_vigente > 0 THEN round(ps.costo_vigente * 100)::bigint END AS costo,
	GREATEST(floor(COALESCE(i.stock_actual, 0)), 0)::bigint AS existencia,
	GREATEST(ceil(COALESCE(i.stock_minimo, 0)), 0)::bigint AS stock_minimo
FROM producto_negocio pn
JOIN precios_sucursal ps ON ps.producto_negocio_id = pn.id AND ps.sucursal_id = @sucursal AND ps.negocio_id = pn.negocio_id
LEFT JOIN inventario_sucursal i ON i.producto_negocio_id = pn.id AND i.sucursal_id = ps.sucursal_id
LEFT JOIN productos p ON p.id = pn.producto_id
LEFT JOIN producto_variantes pv ON pv.id = pn.producto_variante_id
LEFT JOIN productos vp ON vp.id = pv.producto_id
WHERE pn.negocio_id = @negocio AND pn.activo = TRUE AND ps.precio_venta > 0
	AND COALESCE(p.activo, vp.activo AND pv.activo) = TRUE`

// nombreNormalizado replica `ventaapp.Normalizar` en SQL sin depender de la extensión unaccent.
const nombreNormalizado = `translate(lower(COALESCE(p.nombre, vp.nombre) || COALESCE(' ' || pv.clave, '')), 'áéíóúüñàèìòùâêîôûäëïöÁÉÍÓÚÜÑ', 'aeiouunaeiouaeiouaeioaeiouun')`

type filaVendible struct {
	ProductoNegocioID uuid.UUID
	Nombre            string
	CodigoBarras      *string
	Precio            int64
	Costo             *int64
	Existencia        int64
	StockMinimo       int64
}

func (f filaVendible) aVendible() ventaapp.ProductoVendible {
	item := ventaapp.ProductoVendible{ProductoNegocioID: f.ProductoNegocioID, Nombre: f.Nombre, CodigoBarras: f.CodigoBarras,
		Precio: ventadomain.Centavos(f.Precio), Existencia: int(f.Existencia), Agotado: f.Existencia <= 0,
		ExistenciaBaja: f.Existencia > 0 && f.Existencia <= f.StockMinimo}
	if f.Costo != nil {
		costo := ventadomain.Centavos(*f.Costo)
		item.Costo = &costo
	}
	return item
}

func aVendibles(filas []filaVendible) []ventaapp.ProductoVendible {
	items := make([]ventaapp.ProductoVendible, 0, len(filas))
	for _, fila := range filas {
		items = append(items, fila.aVendible())
	}
	return items
}

func (r *Repository) BuscarPorCodigo(negocioID, sucursalID uuid.UUID, codigo string) ([]ventaapp.ProductoVendible, error) {
	var filas []filaVendible
	err := r.db.Raw(vendibleSQL+` AND EXISTS (SELECT 1 FROM producto_codigos c WHERE c.negocio_id = pn.negocio_id AND c.codigo = @codigo
		AND (c.producto_id = pn.producto_id OR c.producto_variante_id = pn.producto_variante_id))
		ORDER BY nombre LIMIT 5`,
		map[string]any{"negocio": negocioID, "sucursal": sucursalID, "codigo": codigo}).Scan(&filas).Error
	return aVendibles(filas), err
}

func (r *Repository) BuscarPorNombre(negocioID, sucursalID uuid.UUID, palabras []string, limite int) ([]ventaapp.ProductoVendible, error) {
	argumentos := map[string]any{"negocio": negocioID, "sucursal": sucursalID, "limite": limite, "inicio": escaparLike(palabras[0]) + "%"}
	var condiciones strings.Builder
	for i, palabra := range palabras {
		clave := "p" + strconv.Itoa(i)
		condiciones.WriteString(" AND " + nombreNormalizado + " LIKE @" + clave)
		argumentos[clave] = "%" + escaparLike(palabra) + "%"
	}
	var filas []filaVendible
	err := r.db.Raw(vendibleSQL+condiciones.String()+
		` ORDER BY (`+nombreNormalizado+` LIKE @inicio) DESC, nombre LIMIT @limite`, argumentos).Scan(&filas).Error
	return aVendibles(filas), err
}

func escaparLike(texto string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(texto)
}

func (r *Repository) vendible(tx *gorm.DB, negocioID, sucursalID, productoID uuid.UUID) (filaVendible, bool, error) {
	var filas []filaVendible
	err := tx.Raw(vendibleSQL+` AND pn.id = @producto`, map[string]any{"negocio": negocioID, "sucursal": sucursalID, "producto": productoID}).Scan(&filas).Error
	if err != nil || len(filas) == 0 {
		return filaVendible{}, false, err
	}
	return filas[0], true, nil
}

// Registrar confirma la venta en una sola transacción: revalida existencias con las filas
// bloqueadas, guarda la foto de cada línea, descuenta inventario y reserva el folio al final.
func (r *Repository) Registrar(comando ventaapp.RegistrarComando) (ventadomain.Venta, error) {
	var creada ventadomain.Venta
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var sucursales int64
		if err := tx.Table("sucursales").Where("id = ? AND negocio_id = ? AND activo = TRUE", comando.SucursalID, comando.NegocioID).Count(&sucursales).Error; err != nil {
			return err
		}
		if sucursales == 0 {
			return ventaapp.ErrSucursalInvalida
		}
		// Bloquear en orden fijo evita interbloqueos entre ventas con los mismos productos.
		ordenadas := append([]ventaapp.LineaSolicitada(nil), comando.Lineas...)
		sort.Slice(ordenadas, func(i, j int) bool {
			return ordenadas[i].ProductoNegocioID.String() < ordenadas[j].ProductoNegocioID.String()
		})
		productos := make(map[uuid.UUID]filaVendible, len(ordenadas))
		inventarios := make(map[uuid.UUID]domain.InventarioSucursal, len(ordenadas))
		for _, linea := range ordenadas {
			var filas []domain.InventarioSucursal
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("sucursal_id = ? AND producto_negocio_id = ?", comando.SucursalID, linea.ProductoNegocioID).Limit(1).Find(&filas).Error; err != nil {
				return err
			}
			sinInventario := len(filas) == 0
			var inventario domain.InventarioSucursal
			if !sinInventario {
				inventario = filas[0]
			}
			producto, ok, err := r.vendible(tx, comando.NegocioID, comando.SucursalID, linea.ProductoNegocioID)
			if err != nil {
				return err
			}
			if !ok {
				return ventaapp.ErrProductoNoDisponible
			}
			disponible := 0
			if !sinInventario && inventario.StockActual > 0 {
				disponible = int(inventario.StockActual)
			}
			if disponible < linea.Cantidad {
				return &ventadomain.ExistenciaInsuficienteError{Producto: producto.Nombre, Disponible: disponible, Solicitado: linea.Cantidad}
			}
			productos[linea.ProductoNegocioID] = producto
			inventarios[linea.ProductoNegocioID] = inventario
		}

		lineas := make([]ventadomain.VentaLinea, 0, len(comando.Lineas))
		var total ventadomain.Centavos
		piezas := 0
		for i, linea := range comando.Lineas {
			producto := productos[linea.ProductoNegocioID]
			precio := ventadomain.Centavos(producto.Precio)
			importe := precio * ventadomain.Centavos(linea.Cantidad)
			var costo *ventadomain.Centavos
			if producto.Costo != nil {
				valor := ventadomain.Centavos(*producto.Costo)
				costo = &valor
			}
			lineas = append(lineas, ventadomain.VentaLinea{ProductoNegocioID: linea.ProductoNegocioID, CodigoBarras: producto.CodigoBarras, Nombre: producto.Nombre,
				Cantidad: linea.Cantidad, PrecioUnitario: precio, CostoUnitario: costo, Importe: importe, Orden: i})
			total += importe
			piezas += linea.Cantidad
		}
		recibido, cambio, err := ventadomain.Cobro(comando.FormaPago, total, comando.Recibido)
		if err != nil {
			return err
		}
		subtotal, iva := ventadomain.DesgloseIVA(total)

		var folio struct{ Numero int64 }
		if err := tx.Raw(`INSERT INTO venta_folio_consecutivos (id, negocio_id, siguiente_numero, actualizado_en)
			VALUES (gen_random_uuid(), ?, 2, now())
			ON CONFLICT (negocio_id) DO UPDATE SET siguiente_numero = venta_folio_consecutivos.siguiente_numero + 1, actualizado_en = now()
			RETURNING siguiente_numero - 1 AS numero`, comando.NegocioID).Scan(&folio).Error; err != nil {
			return err
		}
		creada = ventadomain.Venta{NegocioID: comando.NegocioID, SucursalID: comando.SucursalID, Numero: folio.Numero, Folio: ventadomain.FormatoFolio(folio.Numero),
			UsuarioID: comando.UsuarioID, Subtotal: subtotal, IVA: iva, Total: total, Recibido: recibido, Cambio: cambio, Piezas: piezas,
			FormaPago: comando.FormaPago, Referencia: comando.Referencia, Estado: ventadomain.EstadoPagada, Lineas: lineas}
		if err := tx.Create(&creada).Error; err != nil {
			return err
		}
		for _, linea := range ordenadas {
			if err := moverInventario(tx, inventarios[linea.ProductoNegocioID], -linea.Cantidad, ventadomain.MovimientoVentaSalida, productos[linea.ProductoNegocioID].Costo, creada.Folio, nil, comando.UsuarioID); err != nil {
				return err
			}
		}
		return nil
	})
	return creada, err
}

// Cancelar marca la venta como cancelada y regresa sus piezas con movimientos positivos.
func (r *Repository) Cancelar(comando ventaapp.CancelarComando) (ventadomain.Venta, error) {
	var cancelada ventadomain.Venta
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var encontradas []ventadomain.Venta
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND negocio_id = ? AND sucursal_id IN ?", comando.VentaID, comando.NegocioID, comando.Sucursales).Limit(1).Find(&encontradas).Error; err != nil {
			return err
		}
		if len(encontradas) == 0 {
			return ventaapp.ErrVentaNoEncontrada
		}
		cancelada = encontradas[0]
		if cancelada.Estado == ventadomain.EstadoCancelada {
			return ventaapp.ErrVentaYaCancelada
		}
		var mismoDia bool
		if err := tx.Raw(`SELECT (v.creado_en AT TIME ZONE `+zonaNegocio+`)::date = (now() AT TIME ZONE `+zonaNegocio+`)::date
			FROM ventas v JOIN negocios n ON n.id = v.negocio_id WHERE v.id = ?`, cancelada.ID).Scan(&mismoDia).Error; err != nil {
			return err
		}
		if !mismoDia {
			return ventaapp.ErrVentaFueraDePlazo
		}
		var lineas []ventadomain.VentaLinea
		if err := tx.Where("venta_id = ?", cancelada.ID).Order("producto_negocio_id").Find(&lineas).Error; err != nil {
			return err
		}
		for _, linea := range lineas {
			var inventario domain.InventarioSucursal
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("sucursal_id = ? AND producto_negocio_id = ?", cancelada.SucursalID, linea.ProductoNegocioID).First(&inventario).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				inventario = domain.InventarioSucursal{SucursalID: cancelada.SucursalID, ProductoNegocioID: linea.ProductoNegocioID}
				err = tx.Create(&inventario).Error
			}
			if err != nil {
				return err
			}
			var costo *int64
			if linea.CostoUnitario != nil {
				valor := int64(*linea.CostoUnitario)
				costo = &valor
			}
			if err := moverInventario(tx, inventario, linea.Cantidad, ventadomain.MovimientoVentaCancelacion, costo, cancelada.Folio, &comando.Motivo, comando.UsuarioID); err != nil {
				return err
			}
		}
		ahora := time.Now().UTC()
		return tx.Model(&cancelada).Updates(map[string]any{"estado": ventadomain.EstadoCancelada, "motivo_cancelacion": comando.Motivo, "motivo_cancelacion_codigo": comando.MotivoCodigo,
			"cancelada_por_usuario_id": comando.UsuarioID, "cancelada_en": ahora}).Error
	})
	if err != nil {
		return ventadomain.Venta{}, err
	}
	return r.Obtener(comando.NegocioID, comando.VentaID)
}

func moverInventario(tx *gorm.DB, inventario domain.InventarioSucursal, cantidad int, tipo string, costo *int64, folio string, motivo *string, usuarioID uuid.UUID) error {
	anterior := inventario.StockActual
	nuevo := anterior + float64(cantidad)
	if err := tx.Model(&inventario).Update("stock_actual", nuevo).Error; err != nil {
		return err
	}
	var costoPesos *float64
	if costo != nil {
		valor := float64(*costo) / 100
		costoPesos = &valor
	}
	return tx.Create(&domain.MovimientoInventario{SucursalID: inventario.SucursalID, ProductoNegocioID: inventario.ProductoNegocioID, Tipo: tipo,
		Cantidad: float64(cantidad), StockAnterior: anterior, StockNuevo: nuevo, CostoUnitario: costoPesos, Motivo: motivo, Referencia: &folio, UsuarioID: &usuarioID}).Error
}

// zonaNegocio es la zona horaria del negocio (`n` = negocios) para hablar de «hoy» en su calendario.
const zonaNegocio = `COALESCE(NULLIF(n.zona_horaria, ''), 'America/Mexico_City')`

var unidadesPeriodo = map[string]string{"hoy": "day", "semana": "week", "mes": "month"}

func (r *Repository) Listar(negocioID uuid.UUID, filtro ventaapp.Filtro) (ventaapp.ListadoVentas, error) {
	listado := ventaapp.ListadoVentas{Items: []ventadomain.Venta{}}
	if unidad, ok := unidadesPeriodo[filtro.Periodo]; ok {
		var limites struct{ Desde, Hasta time.Time }
		// La semana de PostgreSQL empieza el lunes, como el calendario en México.
		if err := r.db.Raw(`SELECT date_trunc(@unidad, now() AT TIME ZONE `+zonaNegocio+`) AT TIME ZONE `+zonaNegocio+` AS desde,
			(date_trunc(@unidad, now() AT TIME ZONE `+zonaNegocio+`) + ('1 ' || @unidad)::interval) AT TIME ZONE `+zonaNegocio+` AS hasta
			FROM negocios n WHERE n.id = @negocio`, map[string]any{"unidad": unidad, "negocio": negocioID}).Scan(&limites).Error; err != nil {
			return listado, err
		}
		filtro.Desde, filtro.Hasta = &limites.Desde, &limites.Hasta
		listado.Desde, listado.Hasta = &limites.Desde, &limites.Hasta
	}
	// El resumen cubre el periodo y el alcance del usuario; folio, forma de pago y estado solo acotan la tabla.
	alcance := func() *gorm.DB {
		consulta := r.db.Model(&ventadomain.Venta{}).Where("negocio_id = ? AND sucursal_id IN ?", negocioID, filtro.Sucursales)
		if filtro.UsuarioID != nil {
			consulta = consulta.Where("usuario_id = ?", *filtro.UsuarioID)
		}
		if filtro.Desde != nil {
			consulta = consulta.Where("creado_en >= ?", *filtro.Desde)
		}
		if filtro.Hasta != nil {
			consulta = consulta.Where("creado_en < ?", *filtro.Hasta)
		}
		return consulta
	}
	var resumen struct {
		Vendido, Tickets, Efectivo, Cancelados int64
	}
	if err := alcance().Select(`COALESCE(SUM(total_centavos) FILTER (WHERE estado = 'PAGADA'), 0) AS vendido,
		COUNT(*) FILTER (WHERE estado = 'PAGADA') AS tickets,
		COALESCE(SUM(total_centavos) FILTER (WHERE estado = 'PAGADA' AND forma_pago = 'EFECTIVO'), 0) AS efectivo,
		COUNT(*) FILTER (WHERE estado = 'CANCELADA') AS cancelados`).Scan(&resumen).Error; err != nil {
		return listado, err
	}
	listado.Resumen = ventaapp.ResumenVentas{Vendido: ventadomain.Centavos(resumen.Vendido), Tickets: int(resumen.Tickets),
		Efectivo: ventadomain.Centavos(resumen.Efectivo), Cancelados: int(resumen.Cancelados),
		TicketPromedio: ventaapp.Promedio(ventadomain.Centavos(resumen.Vendido), int(resumen.Tickets))}

	consulta := alcance()
	if strings.TrimSpace(filtro.Folio) != "" {
		numero, ok := numeroDeFolio(filtro.Folio)
		if !ok {
			return listado, nil
		}
		consulta = consulta.Where("numero = ?", numero)
	}
	if filtro.FormaPago != "" {
		consulta = consulta.Where("forma_pago = ?", filtro.FormaPago)
	}
	if filtro.Estado != "" {
		consulta = consulta.Where("estado = ?", filtro.Estado)
	}
	if err := consulta.Order("creado_en DESC, numero DESC").Limit(1000).Find(&listado.Items).Error; err != nil {
		return listado, err
	}
	listado.Total = len(listado.Items)
	return listado, r.completar(listado.Items)
}

// completar agrega el nombre de quien cobró o canceló (perfil o, sin él, correo) y si la venta aún se puede cancelar.
func (r *Repository) completar(items []ventadomain.Venta) error {
	if len(items) == 0 {
		return nil
	}
	ventas := make([]uuid.UUID, 0, len(items))
	usuarios := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		ventas = append(ventas, item.ID)
		usuarios = append(usuarios, item.UsuarioID)
		if item.CanceladaPorUsuarioID != nil {
			usuarios = append(usuarios, *item.CanceladaPorUsuarioID)
		}
	}
	var nombres []struct {
		ID     uuid.UUID
		Nombre string
	}
	if err := r.db.Raw(`SELECT u.id, COALESCE(NULLIF(btrim(COALESCE(p.nombres, '') || ' ' || COALESCE(p.primer_apellido, '')), ''), u.correo) AS nombre
		FROM usuarios u LEFT JOIN perfil_usuarios p ON p.usuario_id = u.id WHERE u.id IN ?`, usuarios).Scan(&nombres).Error; err != nil {
		return err
	}
	porUsuario := make(map[uuid.UUID]string, len(nombres))
	for _, fila := range nombres {
		porUsuario[fila.ID] = fila.Nombre
	}
	var cancelables []struct{ ID uuid.UUID }
	if err := r.db.Raw(`SELECT v.id FROM ventas v JOIN negocios n ON n.id = v.negocio_id
		WHERE v.id IN ? AND v.estado = 'PAGADA' AND (v.creado_en AT TIME ZONE `+zonaNegocio+`)::date = (now() AT TIME ZONE `+zonaNegocio+`)::date`, ventas).Scan(&cancelables).Error; err != nil {
		return err
	}
	cancelable := make(map[uuid.UUID]bool, len(cancelables))
	for _, fila := range cancelables {
		cancelable[fila.ID] = true
	}
	for i := range items {
		items[i].Atendio = porUsuario[items[i].UsuarioID]
		items[i].Cancelable = cancelable[items[i].ID]
		if id := items[i].CanceladaPorUsuarioID; id != nil {
			nombre := porUsuario[*id]
			items[i].CanceladaPor = &nombre
		}
	}
	return nil
}

// numeroDeFolio acepta «V-000232», «000232» o «232».
func numeroDeFolio(texto string) (int64, bool) {
	texto = strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(texto)), "V-")
	if texto == "" {
		return 0, false
	}
	numero, err := strconv.ParseInt(texto, 10, 64)
	return numero, err == nil && numero > 0
}

func (r *Repository) Obtener(negocioID, ventaID uuid.UUID) (ventadomain.Venta, error) {
	var item ventadomain.Venta
	err := r.db.Preload("Lineas", func(db *gorm.DB) *gorm.DB { return db.Order("orden") }).Where("id = ? AND negocio_id = ?", ventaID, negocioID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return item, ventaapp.ErrVentaNoEncontrada
	}
	if err != nil {
		return item, err
	}
	items := []ventadomain.Venta{item}
	err = r.completar(items)
	return items[0], err
}
