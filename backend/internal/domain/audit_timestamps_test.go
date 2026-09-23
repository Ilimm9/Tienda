package domain

import (
	"sync"
	"testing"

	cuenta "tienda/backend/internal/domain/cuenta"
	negocio "tienda/backend/internal/domain/negocio"

	"gorm.io/gorm/schema"
)

type auditModel struct {
	name         string
	value        any
	createdField string
	updatedAt    bool
}

func TestPersistedModelsUseAutomaticAuditTimestamps(t *testing.T) {
	models := []auditModel{
		{"usuario", &cuenta.Usuario{}, "CreadoEn", true},
		{"perfil_usuario", &cuenta.PerfilUsuario{}, "CreadoEn", true},
		{"desafio_autenticacion", &cuenta.DesafioAutenticacion{}, "CreadoEn", false},
		{"direccion", &negocio.Direccion{}, "CreadoEn", true},
		{"negocio", &negocio.Negocio{}, "CreadoEn", true},
		{"configuracion_negocio", &negocio.ConfiguracionNegocio{}, "CreadoEn", true},
		{"rol", &negocio.Rol{}, "CreadoEn", true},
		{"membresia_negocio", &negocio.MembresiaNegocio{}, "CreadoEn", true},
		{"permiso", &negocio.Permiso{}, "CreadoEn", false},
		{"permiso_rol", &negocio.PermisoRol{}, "CreadoEn", false},
		{"rol_membresia", &negocio.RolMembresia{}, "AsignadoEn", false},
		{"empleado", &negocio.Empleado{}, "CreadoEn", true},
		{"invitacion_negocio", &negocio.InvitacionNegocio{}, "CreadoEn", false},
		{"asignacion_empleado_sucursal", &negocio.AsignacionEmpleadoSucursal{}, "AsignadoEn", false},
		{"sucursal", &negocio.Sucursal{}, "CreadoEn", true},
		{"marca", &Marca{}, "CreadoEn", true},
		{"unidad_medida", &UnidadMedida{}, "CreadoEn", true},
		{"producto", &Producto{}, "CreadoEn", true},
		{"importacion_producto", &ImportacionProducto{}, "CreadoEn", true},
		{"familia_producto", &FamiliaProducto{}, "CreadoEn", true},
		{"producto_variante", &ProductoVariante{}, "CreadoEn", true},
		{"producto_variante_atributo", &ProductoVarianteAtributo{}, "CreadoEn", false},
		{"categoria", &Categoria{}, "CreadoEn", true},
		{"producto_categoria", &ProductoCategoria{}, "CreadoEn", false},
		{"producto_codigo", &ProductoCodigo{}, "CreadoEn", true},
		{"producto_imagen", &ProductoImagen{}, "CreadoEn", true},
		{"producto_unidad", &ProductoUnidad{}, "CreadoEn", true},
		{"producto_negocio", &ProductoNegocio{}, "CreadoEn", true},
		{"producto_sku_consecutivo", &ProductoSKUConsecutivo{}, "CreadoEn", true},
		{"impuesto", &Impuesto{}, "CreadoEn", true},
		{"producto_impuesto", &ProductoImpuesto{}, "CreadoEn", false},
		{"inventario_sucursal", &InventarioSucursal{}, "CreadoEn", true},
		{"proveedor", &Proveedor{}, "CreadoEn", true},
		{"producto_proveedor", &ProductoProveedor{}, "CreadoEn", true},
		{"lote", &Lote{}, "CreadoEn", true},
		{"movimiento_inventario", &MovimientoInventario{}, "CreadoEn", false},
	}

	for _, model := range models {
		t.Run(model.name, func(t *testing.T) {
			parsed, err := schema.Parse(model.value, &sync.Map{}, schema.NamingStrategy{})
			if err != nil {
				t.Fatalf("parsear modelo: %v", err)
			}
			if field := parsed.LookUpField(model.createdField); field == nil || field.AutoCreateTime == 0 {
				t.Fatalf("%s debe usar autoCreateTime", model.createdField)
			}
			if !model.updatedAt {
				return
			}
			if field := parsed.LookUpField("ActualizadoEn"); field == nil || field.AutoUpdateTime == 0 {
				t.Fatal("ActualizadoEn debe usar autoUpdateTime")
			}
		})
	}
}
