NO MODIFICAR

Table direcciones [headercolor: #64748B] {
id uuid [pk, default: `gen_random_uuid()`]
codigo_pais varchar(2) [not null, default: 'MX']
estado varchar(120)
municipio varchar(120)
ciudad varchar(120)
colonia varchar(150)
codigo_postal varchar(12)
calle varchar(180)
numero_exterior varchar(30)
numero_interior varchar(30)
referencias text
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]
}

Table usuarios [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
correo varchar(254) [not null, unique]
hash_contrasena varchar(255) [not null]
estado estado_usuario [not null, default: 'pendiente_verificacion']
correo_verificado_en timestamptz
intentos_inicio_sesion_fallidos int [not null, default: 0]
bloqueado_hasta timestamptz
ultimo_inicio_sesion_en timestamptz
contrasena_cambiada_en timestamptz
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]
deshabilitado_en timestamptz

Note: 'Cuenta global. No contiene datos de negocio. El correo debe compararse normalizado/case-insensitive en la implementacion PostgreSQL.'
}

Table perfiles_usuario [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
usuario_id uuid [not null, unique, ref: > usuarios.id]
nombre varchar(100)
segundo_nombre varchar(100)
primer_apellido varchar(100)
segundo_apellido varchar(100)
telefono varchar(30)
direccion_id uuid [ref: > direcciones.id]
archivo_avatar_id uuid
zona_horaria varchar(80) [not null, default: 'America/Mexico_City']
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]
}

Table tokens_autenticacion [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
usuario_id uuid [not null, ref: > usuarios.id]
tipo_token tipo_token_autenticacion [not null]
hash_token varchar(255) [not null, unique]
correo_destino varchar(254)
expira_en timestamptz [not null]
usado_en timestamptz
creado_en timestamptz [not null, default: `now()`]

indexes {
(usuario_id, tipo_token)
expira_en
}
}

Table dispositivos_usuario [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
usuario_id uuid [not null, ref: > usuarios.id]
nombre_dispositivo varchar(120)
hash_huella_dispositivo varchar(255)
agente_usuario text
visto_primera_vez_en timestamptz [not null, default: `now()`]
visto_ultima_vez_en timestamptz [not null, default: `now()`]
confiable_desde timestamptz
revocado_en timestamptz
}

Table sesiones_usuario [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
usuario_id uuid [not null, ref: > usuarios.id]
dispositivo_id uuid [ref: > dispositivos_usuario.id]
hash_token_actualizacion varchar(255) [not null, unique]
direccion_ip varchar(64)
agente_usuario text
emitido_en timestamptz [not null, default: `now()`]
expira_en timestamptz [not null]
ultima_actividad_en timestamptz
revocado_en timestamptz
motivo_revocacion varchar(255)

indexes {
usuario_id
expira_en
}
}

Table eventos_inicio_sesion [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
usuario_id uuid [ref: > usuarios.id]
dispositivo_id uuid [ref: > dispositivos_usuario.id]
correo_intentado varchar(254)
exitoso boolean [not null]
direccion_ip varchar(64)
agente_usuario text
motivo_fallo varchar(255)
creado_en timestamptz [not null, default: `now()`]

indexes {
usuario_id
creado_en
}
}

Table versiones_terminos [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
version varchar(40) [not null, unique]
titulo varchar(180) [not null]
url_contenido text
vigente_desde timestamptz [not null]
activo boolean [not null, default: true]
creado_en timestamptz [not null, default: `now()`]
}

Table aceptaciones_terminos_usuario [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
usuario_id uuid [not null, ref: > usuarios.id]
version_terminos_id uuid [not null, ref: > versiones_terminos.id]
aceptado_en timestamptz [not null, default: `now()`]
direccion_ip varchar(64)

indexes {
(usuario_id, version_terminos_id) [unique]
}
}

Table planes_servicio [headercolor: #475569] {
id uuid [pk, default: `gen_random_uuid()`]
codigo varchar(50) [not null, unique]
nombre varchar(120) [not null]
descripcion text
maximo_usuarios_negocio int
funcionalidades jsonb
activo boolean [not null, default: true]
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]
}

Table negocios [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
slug varchar(120) [not null, unique]
nombre_comercial varchar(180) [not null]
razon_social varchar(220)
rfc varchar(20)
telefono varchar(30)
correo varchar(254)
direccion_id uuid [ref: > direcciones.id]
archivo_logo_id uuid
codigo_moneda varchar(3) [not null, default: 'MXN']
zona_horaria varchar(80) [not null, default: 'America/Mexico_City']
estado estado_negocio [not null, default: 'activo']
creado_por_usuario_id uuid [not null, ref: > usuarios.id]
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]
archivado_en timestamptz

indexes {
rfc
}
}

Table suscripciones_negocio [headercolor: #475569] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
plan_servicio_id uuid [not null, ref: > planes_servicio.id]
estado estado_suscripcion [not null]
inicia_en timestamptz [not null]
termina_en timestamptz
prueba_termina_en timestamptz
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]

indexes {
(negocio_id, estado)
}
}

Table configuraciones_negocio [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, unique, ref: > negocios.id]
metodo_costeo metodo_costeo_inventario [not null, default: 'promedio_ponderado']
permitir_inventario_negativo boolean [not null, default: false]
requerir_autorizacion_cambio_precio boolean [not null, default: true]
requerir_autorizacion_eliminar_partida_venta boolean [not null, default: false]
permitir_multiples_cajas_abiertas_por_empleado boolean [not null, default: false]
permitir_cambio_caja_con_autorizacion boolean [not null, default: true]
habilitar_cortes_parciales_caja boolean [not null, default: false]
monto_corte_parcial_caja numeric(18,2)
hora_corte_parcial_caja time
precios_mayoreo_habilitados boolean [not null, default: true]
dias_devolucion_envase_predeterminados int
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]

checks {
`permitir_inventario_negativo = false` [name: 'ck_negocio_sin_inventario_negativo_alcance_inicial']
}
}

Table configuraciones_ticket [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, unique, ref: > negocios.id]
imprimir_solo_despues_pago boolean [not null, default: true]
marcar_reimpresiones_como_copia boolean [not null, default: true]
qr_habilitado boolean [not null, default: false]
mostrar_cajero boolean [not null, default: true]
mostrar_cliente_cuando_disponible boolean [not null, default: true]
ancho_papel_mm int [not null, default: 80]
texto_encabezado text
texto_pie text
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]
}

Table sucursales [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
codigo varchar(40) [not null]
nombre varchar(160) [not null]
telefono varchar(30)
direccion_id uuid [ref: > direcciones.id]
es_principal boolean [not null, default: false]
activo boolean [not null, default: true]
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]
eliminado_en timestamptz

indexes {
(negocio_id, codigo) [unique]
(negocio_id, nombre)
}
}

Table membresias_negocio [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
usuario_id uuid [not null, ref: > usuarios.id]
tipo_miembro tipo_membresia [not null, default: 'miembro']
estado estado_membresia [not null, default: 'activo']
se_unio_en timestamptz [default: `now()`]
suspendido_en timestamptz
revocado_en timestamptz
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]

indexes {
(negocio_id, usuario_id) [unique]
(negocio_id, tipo_miembro)
}

Note: 'Relacion de acceso entre una cuenta global y un negocio. Un mismo usuario puede tener una membresia en varios negocios.'
}

Table empleados [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
membresia_id uuid [unique, ref: > membresias_negocio.id]
numero_empleado varchar(40)
nombre varchar(100) [not null]
segundo_nombre varchar(100)
primer_apellido varchar(100) [not null]
segundo_apellido varchar(100)
correo varchar(254)
telefono varchar(30)
puesto varchar(120)
estado estado_empleado [not null, default: 'pendiente']
contratado_en date
terminado_en date
creado_por_usuario_id uuid [not null, ref: > usuarios.id]
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]

indexes {
(negocio_id, numero_empleado) [unique]
(negocio_id, correo)
}

Note: 'Empleado del negocio. Puede existir antes de tener cuenta; cuando acepta una invitacion se vincula a membresias_negocio.'
}

Table asignaciones_empleado_sucursal [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
empleado_id uuid [not null, ref: > empleados.id]
sucursal_id uuid [not null, ref: > sucursales.id]
es_principal boolean [not null, default: false]
activo boolean [not null, default: true]
asignado_en timestamptz [not null, default: `now()`]
finalizado_en timestamptz

indexes {
(empleado_id, sucursal_id) [unique]
(negocio_id, sucursal_id)
}
}

Table permisos [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
codigo varchar(120) [not null, unique]
codigo_modulo varchar(30) [not null]
nombre varchar(160) [not null]
descripcion text
creado_en timestamptz [not null, default: `now()`]
}

Table roles [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
codigo varchar(60) [not null]
nombre varchar(120) [not null]
descripcion text
es_rol_sistema boolean [not null, default: false]
activo boolean [not null, default: true]
creado_por_usuario_id uuid [ref: > usuarios.id]
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]

indexes {
(negocio_id, codigo) [unique]
}
}

Table permisos_rol [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
rol_id uuid [not null, ref: > roles.id]
permiso_id uuid [not null, ref: > permisos.id]
creado_en timestamptz [not null, default: `now()`]

indexes {
(rol_id, permiso_id) [unique]
}
}

Table roles_membresia [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
membresia_negocio_id uuid [not null, ref: > membresias_negocio.id]
rol_id uuid [not null, ref: > roles.id]
asignado_por_usuario_id uuid [ref: > usuarios.id]
asignado_en timestamptz [not null, default: `now()`]

indexes {
(membresia_negocio_id, rol_id) [unique]
}
}

Table invitaciones_negocio [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
empleado_id uuid [ref: > empleados.id]
correo varchar(254) [not null]
rol_predeterminado_id uuid [ref: > roles.id]
hash_token varchar(255) [not null, unique]
estado estado_invitacion [not null, default: 'pendiente']
invitado_por_usuario_id uuid [not null, ref: > usuarios.id]
expira_en timestamptz [not null]
aceptado_por_usuario_id uuid [ref: > usuarios.id]
aceptado_en timestamptz
creado_en timestamptz [not null, default: `now()`]

indexes {
(negocio_id, correo)
expira_en
}
}

Table secuencias_documentos [headercolor: #2463EB] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
sucursal_id uuid [ref: > sucursales.id]
tipo_documento varchar(50) [not null]
prefijo varchar(20)
siguiente_numero bigint [not null, default: 1]
relleno int [not null, default: 6]
activo boolean [not null, default: true]
actualizado_en timestamptz [not null, default: `now()`]

indexes {
(negocio_id, sucursal_id, tipo_documento) [unique]
}

Note: 'Generacion transaccional de folios por negocio/sucursal. next_number debe incrementarse con bloqueo de fila/operacion atomica.'
}

Table personal_plataforma [headercolor: #475569] {
id uuid [pk, default: `gen_random_uuid()`]
usuario_id uuid [not null, unique, ref: > usuarios.id]
rol_personal rol_personal_plataforma [not null]
activo boolean [not null, default: true]
creado_en timestamptz [not null, default: `now()`]
}

// ============================================================================
// 02. ARCHIVOS R2 Y CATALOGO GLOBAL
// ============================================================================

Table archivos_multimedia [headercolor: #0891B2] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [ref: > negocios.id, note: 'Null para catalogo global o avatar fuera de contexto de negocio']
usuario_propietario_id uuid [ref: > usuarios.id]
proveedor_almacenamiento varchar(30) [not null, default: 'cloudflare_r2']
nombre_bucket varchar(120) [not null]
clave_objeto varchar(500) [not null]
visibilidad visibilidad_multimedia [not null, default: 'privado']
tipo_origen tipo_origen_multimedia [not null]
nombre_archivo_original varchar(255)
tipo_mime varchar(120) [not null]
tamano_archivo_bytes bigint [not null]
ancho_px int
alto_px int
hash_contenido_sha256 varchar(64)
url_fuente_original text
nombre_licencia varchar(160)
url_licencia text
texto_atribucion text
subido_por_usuario_id uuid [ref: > usuarios.id]
creado_en timestamptz [not null, default: `now()`]
eliminado_en timestamptz

indexes {
(nombre_bucket, clave_objeto) [unique]
negocio_id
hash_contenido_sha256
}

Note: 'Solo metadatos. El binario vive en Cloudflare R2. No guardar URLs firmadas temporales.'
}

Table variantes_archivo [headercolor: #0891B2] {
id uuid [pk, default: `gen_random_uuid()`]
archivo_multimedia_id uuid [not null, ref: > archivos_multimedia.id]
nombre_variante varchar(30) [not null, note: 'Ej. medium, thumbnail']
nombre_bucket varchar(120) [not null]
clave_objeto varchar(500) [not null]
tipo_mime varchar(120) [not null]
tamano_archivo_bytes bigint [not null]
ancho_px int
alto_px int
creado_en timestamptz [not null, default: `now()`]

indexes {
(archivo_multimedia_id, nombre_variante) [unique]
(nombre_bucket, clave_objeto) [unique]
}
}

Table fuentes_catalogo [headercolor: #0891B2] {
id uuid [pk, default: `gen_random_uuid()`]
codigo varchar(50) [not null, unique, note: 'internal, open_food_facts, manual']
nombre varchar(120) [not null]
url_base text
activo boolean [not null, default: true]
creado_en timestamptz [not null, default: `now()`]
}

Table articulos_catalogo_global [headercolor: #0891B2] {
id uuid [pk, default: `gen_random_uuid()`]
nombre_canonico varchar(220) [not null]
nombre_marca varchar(160)
descripcion text
empaque_descripcion varchar(180)
valor_contenido_neto numeric(18,6)
unidad_contenido_neto varchar(30)
estado estado_articulo_catalogo [not null, default: 'no_verificado']
creado_por_usuario_id uuid [ref: > usuarios.id]
verificado_por_usuario_plataforma_id uuid [ref: > usuarios.id]
verificado_en timestamptz
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]

Note: 'Identidad comercial global reutilizable entre negocios. No contiene costo, precio, inventario ni proveedor del negocio.'
}

Table identificadores_catalogo_global [headercolor: #0891B2] {
id uuid [pk, default: `gen_random_uuid()`]
articulo_catalogo_global_id uuid [not null, ref: > articulos_catalogo_global.id]
tipo_identificador tipo_identificador [not null]
valor varchar(80) [not null]
valor_normalizado varchar(80) [not null]
es_principal boolean [not null, default: false]
creado_en timestamptz [not null, default: `now()`]

indexes {
(tipo_identificador, valor_normalizado) [unique]
articulo_catalogo_global_id
}
}

Table enlaces_fuente_catalogo_global [headercolor: #0891B2] {
id uuid [pk, default: `gen_random_uuid()`]
articulo_catalogo_global_id uuid [not null, ref: > articulos_catalogo_global.id]
fuente_catalogo_id uuid [not null, ref: > fuentes_catalogo.id]
id_externo varchar(180)
url_fuente text
carga_cruda jsonb
ultima_sincronizacion_en timestamptz
creado_en timestamptz [not null, default: `now()`]

indexes {
(fuente_catalogo_id, id_externo) [unique]
}
}

Table multimedia_catalogo_global [headercolor: #0891B2] {
id uuid [pk, default: `gen_random_uuid()`]
articulo_catalogo_global_id uuid [not null, ref: > articulos_catalogo_global.id]
archivo_multimedia_id uuid [not null, ref: > archivos_multimedia.id]
es_principal boolean [not null, default: false]
orden int [not null, default: 0]
creado_en timestamptz [not null, default: `now()`]

indexes {
(articulo_catalogo_global_id, archivo_multimedia_id) [unique]
}
}

// ============================================================================
// 03. CATALOGO LOCAL: PRODUCTOS, VARIANTES, PRESENTACIONES, PRECIOS E IMPUESTOS
// ============================================================================

Table marcas [headercolor: #7C3AED] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
nombre varchar(160) [not null]
activo boolean [not null, default: true]
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]
eliminado_en timestamptz

indexes {
(negocio_id, nombre) [unique]
}
}

Table categorias [headercolor: #7C3AED] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
categoria_padre_id uuid [ref: > categorias.id]
nombre varchar(160) [not null]
descripcion text
activo boolean [not null, default: true]
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]
eliminado_en timestamptz

indexes {
(negocio_id, categoria_padre_id, nombre) [unique]
}
}

Table etiquetas_producto [headercolor: #7C3AED] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
nombre varchar(80) [not null]
creado_en timestamptz [not null, default: `now()`]

indexes {
(negocio_id, nombre) [unique]
}
}

Table unidades_medida [headercolor: #7C3AED] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
codigo varchar(30) [not null]
nombre varchar(80) [not null]
simbolo varchar(20)
tipo_unidad varchar(40) [not null, note: 'piece, mass, volume, length, other']
precision_decimal int [not null, default: 0]
activo boolean [not null, default: true]
creado_en timestamptz [not null, default: `now()`]

indexes {
(negocio_id, codigo) [unique]
}
}

Table productos [headercolor: #7C3AED] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
categoria_id uuid [ref: > categorias.id]
marca_id uuid [ref: > marcas.id]
nombre varchar(220) [not null]
descripcion text
estado estado_producto [not null, default: 'activo']
creado_por_usuario_id uuid [not null, ref: > usuarios.id]
actualizado_por_usuario_id uuid [ref: > usuarios.id]
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]
eliminado_en timestamptz

indexes {
(negocio_id, nombre)
categoria_id
marca_id
}

Note: 'Familia o producto conceptual, por ejemplo Cheetos. La variante es la unidad comercial/inventariable real.'
}

Table variantes_producto [headercolor: #7C3AED] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
producto_id uuid [not null, ref: > productos.id]
sku varchar(80)
codigo_interno varchar(80)
nombre_variante varchar(220) [not null]
unidad_medida_base_id uuid [not null, ref: > unidades_medida.id]
valor_contenido numeric(18,6)
unidad_medida_contenido_id uuid [ref: > unidades_medida.id]
es_variante_predeterminada boolean [not null, default: false]
controlar_inventario boolean [not null, default: true]
es_producto_retornable boolean [not null, default: false]
es_vendible boolean [not null, default: true]
es_comprable boolean [not null, default: true]
estado estado_producto [not null, default: 'activo']
creado_por_usuario_id uuid [not null, ref: > usuarios.id]
actualizado_por_usuario_id uuid [ref: > usuarios.id]
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]
eliminado_en timestamptz

indexes {
(negocio_id, sku) [unique]
(negocio_id, codigo_interno) [unique]
producto_id
}
}

Table presentaciones_producto [headercolor: #7C3AED] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
variante_producto_id uuid [not null, ref: > variantes_producto.id]
articulo_catalogo_global_id uuid [ref: > articulos_catalogo_global.id]
nombre varchar(120) [not null, note: 'Pieza, six, caja, carton, paquete...']
conversion_a_base numeric(18,6) [not null, default: 1]
es_presentacion_base boolean [not null, default: false]
puede_comprarse boolean [not null, default: true]
puede_venderse boolean [not null, default: true]
puede_desarmarse boolean [not null, default: false]
es_compra_predeterminada boolean [not null, default: false]
es_venta_predeterminada boolean [not null, default: false]
activo boolean [not null, default: true]
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]
eliminado_en timestamptz

indexes {
(variante_producto_id, nombre) [unique]
articulo_catalogo_global_id
}

checks {
`conversion_a_base > 0` [name: 'ck_presentacion_conversion_positiva']
}

Note: 'Cada presentacion se convierte a la unidad base del producto. Ej.: carton=24 piezas. Puede enlazarse a un item del catalogo global por GTIN/EAN/UPC.'
}

Table codigos_barras_producto [headercolor: #7C3AED] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
presentacion_producto_id uuid [not null, ref: > presentaciones_producto.id]
tipo_codigo_barras tipo_codigo_barras [not null]
codigo varchar(100) [not null]
normalizado_codigo varchar(100) [not null]
es_principal boolean [not null, default: false]
activo boolean [not null, default: true]
creado_en timestamptz [not null, default: `now()`]

indexes {
(negocio_id, normalizado_codigo) [unique]
presentacion_producto_id
}
}

Table multimedia_producto [headercolor: #7C3AED] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
variante_producto_id uuid [not null, ref: > variantes_producto.id]
presentacion_producto_id uuid [ref: > presentaciones_producto.id]
archivo_multimedia_id uuid [not null, ref: > archivos_multimedia.id]
es_principal boolean [not null, default: false]
orden int [not null, default: 0]
creado_en timestamptz [not null, default: `now()`]

indexes {
(variante_producto_id, archivo_multimedia_id) [unique]
}

Note: 'Imagen propia del negocio. Si no existe, la aplicacion puede mostrar la imagen del global_catalog_item asociado a la presentacion.'
}

Table relaciones_etiqueta_producto [headercolor: #7C3AED] {
id uuid [pk, default: `gen_random_uuid()`]
producto_id uuid [not null, ref: > productos.id]
etiqueta_producto_id uuid [not null, ref: > etiquetas_producto.id]
creado_en timestamptz [not null, default: `now()`]

indexes {
(producto_id, etiqueta_producto_id) [unique]
}
}

Table definiciones_impuesto [headercolor: #7C3AED] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
codigo varchar(40) [not null]
nombre varchar(120) [not null]
tipo_impuesto tipo_impuesto [not null]
porcentaje_tasa numeric(9,6) [not null]
incluido_por_defecto boolean [not null, default: false]
activo boolean [not null, default: true]
valido_desde date
valido_hasta date
creado_en timestamptz [not null, default: `now()`]

indexes {
(negocio_id, codigo) [unique]
}

checks {
`porcentaje_tasa >= 0` [name: 'ck_tasa_impuesto_no_negativa']
}
}

Table impuestos_predeterminados_producto [headercolor: #7C3AED] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
variante_producto_id uuid [not null, ref: > variantes_producto.id]
definicion_impuesto_id uuid [not null, ref: > definiciones_impuesto.id]
aplica_por_defecto boolean [not null, default: true]
creado_en timestamptz [not null, default: `now()`]

indexes {
(variante_producto_id, definicion_impuesto_id) [unique]
}
}

Table reglas_precio [headercolor: #7C3AED] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
sucursal_id uuid [ref: > sucursales.id]
categoria_id uuid [ref: > categorias.id]
variante_producto_id uuid [ref: > variantes_producto.id]
nombre varchar(140) [not null]
metodo metodo_fijacion_precio [not null, default: 'margen']
porcentaje_objetivo numeric(9,4)
incremento_fijo numeric(18,6)
metodo_redondeo metodo_redondeo [not null, default: 'mas_cercano']
incremento_redondeo numeric(18,2) [default: 1]
prioridad int [not null, default: 100]
activo boolean [not null, default: true]
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]

Note: 'Reglas para sugerir precio al detectar nuevo costo. Se puede aplicar por negocio, sucursal, categoria o variante; la aplicacion resuelve la regla de mayor prioridad/especificidad.'
}

Table precios_producto [headercolor: #7C3AED] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
sucursal_id uuid [ref: > sucursales.id, note: 'Null = precio general del negocio; no null = override de sucursal']
presentacion_producto_id uuid [not null, ref: > presentaciones_producto.id]
tipo_precio tipo_precio [not null]
cantidad_minima numeric(18,6) [not null, default: 1]
monto numeric(18,2) [not null]
valido_desde timestamptz [not null, default: `now()`]
valido_hasta timestamptz
solicitud_autorizacion_id uuid
creado_por_usuario_id uuid [not null, ref: > usuarios.id]
aprobado_por_usuario_id uuid [ref: > usuarios.id]
creado_en timestamptz [not null, default: `now()`]

indexes {
(negocio_id, presentacion_producto_id, sucursal_id, tipo_precio, valido_desde)
(negocio_id, sucursal_id, tipo_precio)
}

checks {
`monto >= 0` [name: 'ck_precio_producto_no_negativo']
`cantidad_minima > 0` [name: 'ck_precio_producto_cantidad_minima_positiva']
}

Note: 'Historial temporal de precios. En PostgreSQL conviene agregar un indice unico parcial para impedir mas de un precio vigente por presentacion/sucursal/tipo/escalon.'
}

Table historial_costos_producto [headercolor: #7C3AED] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
sucursal_id uuid [not null, ref: > sucursales.id]
variante_producto_id uuid [not null, ref: > variantes_producto.id]
proveedor_id uuid
partida_recepcion_compra_id uuid
costo_por_unidad_base numeric(18,6) [not null]
costo_promedio_ponderado_posterior numeric(18,6)
fuente_tipo varchar(40) [not null, note: 'purchase, import, manual, adjustment']
creado_por_usuario_id uuid [ref: > usuarios.id]
vigente_desde timestamptz [not null, default: `now()`]
creado_en timestamptz [not null, default: `now()`]

indexes {
(negocio_id, sucursal_id, variante_producto_id, vigente_desde)
}

checks {
`costo_por_unidad_base >= 0` [name: 'ck_costo_producto_no_negativo']
}
}

// ============================================================================
// 04. PROVEEDORES, RECEPCION DE MERCANCIA Y ANALISIS DE COSTOS/PRECIOS
// ============================================================================

Table proveedores [headercolor: #0F766E] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
codigo varchar(50)
nombre_comercial varchar(180) [not null]
razon_social varchar(220)
rfc varchar(20)
nombre_contacto varchar(160)
telefono varchar(30)
correo varchar(254)
direccion_id uuid [ref: > direcciones.id]
condiciones_pago text
descripcion_dias_entrega varchar(180)
estado estado_proveedor [not null, default: 'activo']
notas text
creado_por_usuario_id uuid [not null, ref: > usuarios.id]
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]
eliminado_en timestamptz

indexes {
(negocio_id, codigo) [unique]
(negocio_id, nombre_comercial)
rfc
}
}

Table productos_proveedor [headercolor: #0F766E] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
proveedor_id uuid [not null, ref: > proveedores.id]
presentacion_producto_id uuid [not null, ref: > presentaciones_producto.id]
sku_proveedor varchar(100)
nombre_producto_proveedor varchar(220)
preferido boolean [not null, default: false]
ultimo_costo_empaque_conocido numeric(18,6)
activo boolean [not null, default: true]
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]

indexes {
(proveedor_id, presentacion_producto_id) [unique]
(proveedor_id, sku_proveedor)
}
}

Table reglas_impuesto_producto_proveedor [headercolor: #0F766E] {
id uuid [pk, default: `gen_random_uuid()`]
producto_proveedor_id uuid [not null, ref: > productos_proveedor.id]
definicion_impuesto_id uuid [not null, ref: > definiciones_impuesto.id]
aplica_por_defecto boolean [not null, default: true]
creado_en timestamptz [not null, default: `now()`]

indexes {
(producto_proveedor_id, definicion_impuesto_id) [unique]
}
}

Table solicitudes_autorizacion [headercolor: #B45309] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
sucursal_id uuid [ref: > sucursales.id]
tipo_solicitud tipo_solicitud_autorizacion [not null]
estado estado_autorizacion [not null, default: 'pendiente']
tipo_sujeto varchar(80) [not null]
sujeto_id uuid [not null]
solicitado_por_usuario_id uuid [not null, ref: > usuarios.id]
motivo text
carga_capturada jsonb
solicitado_en timestamptz [not null, default: `now()`]
resuelto_en timestamptz

indexes {
(negocio_id, estado, tipo_solicitud)
(tipo_sujeto, sujeto_id)
}

Note: 'Bandeja transversal de autorizaciones. La tabla de dominio tambien conserva approval_request_id cuando requiere integridad relacional.'
}

Table acciones_autorizacion [headercolor: #B45309] {
id uuid [pk, default: `gen_random_uuid()`]
solicitud_autorizacion_id uuid [not null, ref: > solicitudes_autorizacion.id]
tipo_accion tipo_accion_autorizacion [not null]
realizado_por_usuario_id uuid [not null, ref: > usuarios.id]
comentarios text
creado_en timestamptz [not null, default: `now()`]

indexes {
solicitud_autorizacion_id
}
}

Table recepciones_compra [headercolor: #0F766E] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
sucursal_id uuid [not null, ref: > sucursales.id]
proveedor_id uuid [not null, ref: > proveedores.id]
folio varchar(60) [not null]
tipo_documento tipo_documento_compra [not null]
numero_documento_proveedor varchar(100)
fecha_documento date
recibido_en timestamptz [not null]
modo_captura modo_captura_compra [not null, default: 'manual']
estado estado_recepcion_compra [not null, default: 'borrador']
metodo_pago_proveedor varchar(40)
subtotal_declarado numeric(18,2)
iva_total_declarado numeric(18,2)
ieps_total_declarado numeric(18,2)
impuestos_totales_declarados numeric(18,2)
total_declarado numeric(18,2)
subtotal_calculado numeric(18,2)
impuestos_totales_calculados numeric(18,2)
total_calculado numeric(18,2)
diferencia_conciliacion numeric(18,2)
notas text
capturado_por_usuario_id uuid [not null, ref: > usuarios.id]
recibido_por_empleado_id uuid [ref: > empleados.id]
aprobado_por_usuario_id uuid [ref: > usuarios.id]
aprobado_en timestamptz
operacion_cliente_id uuid [note: 'Idempotencia cuando la captura proviene de reintento/sin conexion']
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]

indexes {
(negocio_id, sucursal_id, folio) [unique]
(negocio_id, proveedor_id, recibido_en)
(negocio_id, operacion_cliente_id) [unique]
estado
}

Note: 'Documento de recepcion prioritaria. Solo al aprobarse debe impactar costo/inventario conforme a la etapa habilitada.'
}

Table partidas_recepcion_compra [headercolor: #0F766E] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
recepcion_compra_id uuid [not null, ref: > recepciones_compra.id]
numero_partida int [not null]
presentacion_producto_id uuid [ref: > presentaciones_producto.id, note: 'Puede ser null mientras el producto sea nuevo/no relacionado']
codigo_barras_escaneado varchar(100)
nombre_producto_capturado varchar(220) [not null]
nombre_presentacion_capturada varchar(120)
cantidad_empaques numeric(18,6) [not null]
unidades_por_empaque numeric(18,6) [not null, default: 1]
cantidad_base numeric(18,6) [not null]
costo_empaque numeric(18,6) [not null]
costo_unidad_base numeric(18,6) [not null]
monto_descuento numeric(18,2) [not null, default: 0]
monto_subtotal numeric(18,2) [not null]
monto_impuesto numeric(18,2) [not null, default: 0]
monto_total numeric(18,2) [not null]
ieps_marcado_en_documento boolean
es_producto_nuevo boolean [not null, default: false]
es_producto_sin_vincular boolean [not null, default: false]
presentacion_cambiada boolean [not null, default: false]
tiene_datos_faltantes boolean [not null, default: false]
notas text
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]

indexes {
(recepcion_compra_id, numero_partida) [unique]
presentacion_producto_id
codigo_barras_escaneado
}

checks {
`cantidad_empaques > 0` [name: 'ck_partida_compra_cantidad_empaque_positiva']
`unidades_por_empaque > 0` [name: 'ck_partida_compra_unidades_positivas']
`cantidad_base > 0` [name: 'ck_partida_compra_cantidad_base_positiva']
`costo_empaque >= 0` [name: 'ck_partida_compra_costo_empaque_no_negativo']
`costo_unidad_base >= 0` [name: 'ck_partida_compra_costo_unitario_no_negativo']
}
}

Table impuestos_partida_recepcion [headercolor: #0F766E] {
id uuid [pk, default: `gen_random_uuid()`]
partida_recepcion_compra_id uuid [not null, ref: > partidas_recepcion_compra.id]
definicion_impuesto_id uuid [ref: > definiciones_impuesto.id]
tipo_impuesto tipo_impuesto [not null]
porcentaje_tasa numeric(9,6) [not null]
base_gravable numeric(18,2) [not null]
monto_impuesto numeric(18,2) [not null]
incluido_en_precio boolean [not null, default: false]
indicado_en_documento_proveedor boolean
creado_en timestamptz [not null, default: `now()`]

indexes {
partida_recepcion_compra_id
}
}

Table incidencias_recepcion_compra [headercolor: #0F766E] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
recepcion_compra_id uuid [not null, ref: > recepciones_compra.id]
partida_recepcion_compra_id uuid [ref: > partidas_recepcion_compra.id]
tipo_incidencia tipo_incidencia_compra [not null]
cantidad_afectada numeric(18,6)
estado estado_incidencia_compra [not null, default: 'abierto']
notas text
reportado_por_usuario_id uuid [not null, ref: > usuarios.id]
reportado_en timestamptz [not null, default: `now()`]
reconocido_por_usuario_id uuid [ref: > usuarios.id]
reconocido_en timestamptz
resuelto_por_usuario_id uuid [ref: > usuarios.id]
resuelto_en timestamptz

indexes {
(negocio_id, estado)
recepcion_compra_id
}

Note: 'Faltantes, danados o rechazados deben disparar notificaciones a los responsables configurados.'
}

Table adjuntos_recepcion_compra [headercolor: #0F766E] {
id uuid [pk, default: `gen_random_uuid()`]
recepcion_compra_id uuid [not null, ref: > recepciones_compra.id]
archivo_multimedia_id uuid [not null, ref: > archivos_multimedia.id]
tipo_adjunto varchar(40) [not null, note: 'ticket, invoice, evidence, other']
creado_por_usuario_id uuid [not null, ref: > usuarios.id]
creado_en timestamptz [not null, default: `now()`]

indexes {
(recepcion_compra_id, archivo_multimedia_id) [unique]
}

Note: 'La primera version captura manualmente; esta relacion queda lista para adjuntar evidencia cuando se habilite.'
}

Table revisiones_precio [headercolor: #B45309] {
id uuid [pk, default: `gen_random_uuid()`]
negocio_id uuid [not null, ref: > negocios.id]
sucursal_id uuid [not null, ref: > sucursales.id]
partida_recepcion_compra_id uuid [not null, unique, ref: > partidas_recepcion_compra.id]
presentacion_producto_id uuid [ref: > presentaciones_producto.id]
regla_precio_id uuid [ref: > reglas_precio.id]
resultado_analisis resultado_analisis_precio [not null]
costo_unidad_base_anterior numeric(18,6)
costo_unidad_base_nuevo numeric(18,6) [not null]
diferencia_costo numeric(18,6)
porcentaje_cambio_costo numeric(12,6)
precio_venta_actual numeric(18,2)
porcentaje_margen_actual numeric(12,6)
precio_venta_sugerido numeric(18,2)
estado estado_revision_precio [not null, default: 'pendiente']
solicitud_autorizacion_id uuid [ref: > solicitudes_autorizacion.id]
revisado_por_usuario_id uuid [ref: > usuarios.id]
revisado_en timestamptz
notas_revision text
creado_en timestamptz [not null, default: `now()`]
actualizado_en timestamptz [not null, default: `now()`]

indexes {
(negocio_id, sucursal_id, estado)
presentacion_producto_id
}

Note: 'Conserva comparacion y sugerencia. Aprobar este registro permite crear el nuevo precios_producto; nunca cambiar el precio silenciosamente.'
}
