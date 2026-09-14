package negocio

import (
	"context"
	"errors"
	"strings"
	"time"

	application "tienda/backend/internal/application/negocio"
	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmpleadoRepository struct {
	db *gorm.DB
}

func NewEmpleadoRepository(db *gorm.DB) *EmpleadoRepository {
	return &EmpleadoRepository{db: db}
}

func (r *EmpleadoRepository) ObtenerContextoNegocio(ctx context.Context, usuarioID, negocioID uuid.UUID) (domain.ContextoNegocioSucursal, error) {
	var contexto domain.ContextoNegocioSucursal
	err := r.db.WithContext(ctx).Table("negocios AS n").
		Select("n.estado AS estado_negocio, m.tipo_miembro").
		Joins("JOIN membresias_negocio m ON m.negocio_id = n.id AND m.usuario_id = ? AND m.estado = 'activo'", usuarioID).
		Where("n.id = ?", negocioID).
		Take(&contexto).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ContextoNegocioSucursal{}, application.ErrNegocioNoEncontrado
	}
	return contexto, err
}

func (r *EmpleadoRepository) PermisosEfectivos(ctx context.Context, usuarioID, negocioID uuid.UUID) ([]string, error) {
	return permisosEfectivos(ctx, r.db, usuarioID, negocioID)
}

func (r *EmpleadoRepository) Listar(ctx context.Context, negocioID uuid.UUID, estado, buscar string) ([]domain.EmpleadoResumen, error) {
	items := make([]domain.EmpleadoResumen, 0)
	query := r.db.WithContext(ctx).Table("empleados AS e").
		Select(`e.id, e.negocio_id, e.numero_empleado, e.correo, e.telefono, e.puesto, e.estado,
			(e.membresia_id IS NOT NULL) AS tiene_cuenta,
			e.nombre, e.segundo_nombre, e.primer_apellido, e.segundo_apellido,
			e.creado_en, e.actualizado_en`).
		Where("e.negocio_id = ?", negocioID)
	if estado != "" && estado != "todos" {
		query = query.Where("e.estado = ?", estado)
	}
	if buscar != "" {
		patron := "%" + strings.ToLower(buscar) + "%"
		query = query.Where(`lower(e.nombre) LIKE ? OR lower(e.primer_apellido) LIKE ?
			OR lower(COALESCE(e.numero_empleado, '')) LIKE ? OR lower(COALESCE(e.correo, '')) LIKE ?`,
			patron, patron, patron, patron)
	}

	type fila struct {
		domain.EmpleadoResumen
		Nombre          string
		SegundoNombre   *string
		PrimerApellido  string
		SegundoApellido *string
	}
	filas := make([]fila, 0)
	err := query.Order("e.primer_apellido ASC, e.nombre ASC, e.id ASC").Scan(&filas).Error
	if err != nil {
		return nil, err
	}
	for _, actual := range filas {
		resumen := actual.EmpleadoResumen
		resumen.NombreCompleto = domain.NombreCompletoEmpleado(
			actual.Nombre, actual.SegundoNombre, actual.PrimerApellido, actual.SegundoApellido)
		items = append(items, resumen)
	}
	return items, nil
}

func (r *EmpleadoRepository) Obtener(ctx context.Context, negocioID, empleadoID uuid.UUID) (domain.EmpleadoDetalle, error) {
	var detalle domain.EmpleadoDetalle
	err := r.db.WithContext(ctx).Table("empleados AS e").
		Select(`e.id, e.negocio_id, e.membresia_id, e.numero_empleado, e.nombre, e.segundo_nombre,
			e.primer_apellido, e.segundo_apellido, e.correo, e.telefono, e.puesto, e.estado,
			e.contratado_en, e.terminado_en, (e.membresia_id IS NOT NULL) AS tiene_cuenta,
			e.creado_en, e.actualizado_en`).
		Where("e.id = ? AND e.negocio_id = ?", empleadoID, negocioID).
		Take(&detalle).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.EmpleadoDetalle{}, application.ErrEmpleadoNoEncontrado
	}
	if err != nil {
		return domain.EmpleadoDetalle{}, err
	}
	detalle.NombreCompleto = domain.NombreCompletoEmpleado(
		detalle.Nombre, detalle.SegundoNombre, detalle.PrimerApellido, detalle.SegundoApellido)
	return detalle, nil
}

func (r *EmpleadoRepository) ExisteNumero(ctx context.Context, negocioID uuid.UUID, numero string, excluir *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Table("empleados").
		Where("negocio_id = ? AND upper(numero_empleado) = upper(?)", negocioID, numero)
	if excluir != nil {
		query = query.Where("id <> ?", *excluir)
	}
	var total int64
	err := query.Count(&total).Error
	return total > 0, err
}

func (r *EmpleadoRepository) Crear(ctx context.Context, negocioID, creadoPor uuid.UUID, input domain.CrearEmpleadoInput) (uuid.UUID, error) {
	empleado := domain.Empleado{
		NegocioID: negocioID, NumeroEmpleado: input.NumeroEmpleado,
		Nombre: input.Nombre, SegundoNombre: input.SegundoNombre,
		PrimerApellido: input.PrimerApellido, SegundoApellido: input.SegundoApellido,
		Correo: input.Correo, Telefono: input.Telefono, Puesto: input.Puesto,
		Estado: domain.EstadoEmpleadoPendiente, ContratadoEn: parsearFecha(input.ContratadoEn),
		CreadoPorUsuarioID: creadoPor,
	}
	err := r.db.WithContext(ctx).Create(&empleado).Error
	return empleado.ID, err
}

func (r *EmpleadoRepository) Actualizar(ctx context.Context, negocioID, empleadoID uuid.UUID, input domain.ActualizarEmpleadoInput) error {
	cambios := map[string]any{}
	asignarOpcional(cambios, "numero_empleado", input.NumeroEmpleado)
	asignarOpcional(cambios, "segundo_nombre", input.SegundoNombre)
	asignarOpcional(cambios, "segundo_apellido", input.SegundoApellido)
	asignarOpcional(cambios, "correo", input.Correo)
	asignarOpcional(cambios, "telefono", input.Telefono)
	asignarOpcional(cambios, "puesto", input.Puesto)
	if input.Nombre.Set && input.Nombre.Value != nil {
		cambios["nombre"] = *input.Nombre.Value
	}
	if input.PrimerApellido.Set && input.PrimerApellido.Value != nil {
		cambios["primer_apellido"] = *input.PrimerApellido.Value
	}
	if input.Estado.Set && input.Estado.Value != nil {
		cambios["estado"] = *input.Estado.Value
		// Terminar sin fecha explícita deja constancia del día en que ocurrió.
		if *input.Estado.Value == domain.EstadoEmpleadoTerminado && !input.TerminadoEn.Set {
			cambios["terminado_en"] = time.Now()
		}
	}
	if input.ContratadoEn.Set {
		cambios["contratado_en"] = parsearFecha(input.ContratadoEn.Value)
	}
	if input.TerminadoEn.Set {
		cambios["terminado_en"] = parsearFecha(input.TerminadoEn.Value)
	}
	if len(cambios) == 0 {
		return nil
	}
	cambios["actualizado_en"] = gorm.Expr("now()")
	return r.db.WithContext(ctx).Table("empleados").
		Where("id = ? AND negocio_id = ?", empleadoID, negocioID).Updates(cambios).Error
}

func asignarOpcional(cambios map[string]any, columna string, campo domain.Optional[string]) {
	if !campo.Set {
		return
	}
	if campo.Value == nil {
		cambios[columna] = nil
		return
	}
	cambios[columna] = *campo.Value
}

func parsearFecha(valor *string) *time.Time {
	if valor == nil {
		return nil
	}
	fecha, err := time.Parse("2006-01-02", *valor)
	if err != nil {
		return nil
	}
	return &fecha
}
