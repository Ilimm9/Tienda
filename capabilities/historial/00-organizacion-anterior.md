# Capabilities

Esta carpeta contiene planes funcionales ejecutables para Tienda. Una capability define qué construir y cómo comprobarlo; no contiene código de producción.

## Convención

Cada funcionalidad usa número consecutivo y nombre en minúsculas con guiones:

```text
capabilities/01-nombre-funcionalidad/CAPABILITY.md
capabilities/02-otra-funcionalidad/CAPABILITY.md
```

Cuando una capability crece por fases, `CAPABILITY.md` funciona como índice y conserva reglas transversales. El detalle puede dividirse sin duplicarlo:

```text
capabilities/NN-nombre/CAPABILITY.md
capabilities/NN-nombre/fases/01-nombre.md
capabilities/NN-nombre/fases/02-nombre.md
capabilities/NN-nombre/SEGUIMIENTO.md
```

Para trabajar una fase se carga el índice, el archivo de esa fase y solo las fuentes que dicho archivo indique. Estados, contratos e historial deben tener una única ubicación canónica.

Copiar `_template/CAPABILITY.md` al crear una nueva capability. No eliminar secciones; escribir `No aplica` cuando corresponda.

## Ciclo

1. Crear documento en estado `Borrador`.
2. Completar decisiones y pasarlo a `En revisión`.
3. Esperar aprobación explícita del usuario.
4. Registrar aprobación y usar estado `Aprobada`.
5. Implementar usando estado `En implementación`.
6. Registrar verificaciones y usar estado `Verificada`.
7. Cerrar únicamente tras aceptación final del usuario.

## Límites

- Especificaciones y decisiones viven aquí.
- Código Go vive en `backend/`.
- Código Angular vive en `frontend/`.
- Scripts auxiliares deben vivir junto al subsistema que los utiliza, salvo decisión distinta documentada y aprobada.
- `schema.json` y `prompt_hints.md` solo se agregan si una capability futura introduce Tool Calling o agentes ejecutables.
