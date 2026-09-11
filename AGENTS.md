# Reglas de trabajo del proyecto Tienda

Estas reglas aplican a cualquier agente o colaborador automatizado que trabaje en este repositorio.

## Planeación obligatoria

- No implementar una funcionalidad nueva sin un archivo `capabilities/NN-nombre/CAPABILITY.md`.
- La capability debe describir alcance, reglas, contratos, pruebas y criterios de aceptación.
- No comenzar la implementación mientras su estado no sea `Aprobada`.
- Solo una aprobación explícita del usuario permite cambiar el estado a `Aprobada`.
- Durante la implementación, mantener actualizadas las tareas y decisiones de la capability.
- Si el trabajo requerido se desvía de la capability aprobada, detener esa parte, explicar la desviación y solicitar revisión.

## Protección del repositorio

- Preservar cambios existentes del usuario y evitar sobrescribir trabajo no relacionado.
- Antes de modificar archivos, revisar el estado actual cuando sea necesario para evitar conflictos.
- Antes de la confirmación final del usuario, Git queda limitado a operaciones de lectura: `git status`, `git diff`, `git log` y `git show`.
- Sin confirmación final, no ejecutar `git add`, `git commit`, `git push`, `git pull`, `git fetch`, `git merge`, `git rebase`, `git checkout`, `git switch`, `git stash`, creación o borrado de ramas o tags, ni modificar `.git` directamente.
- No desplegar, publicar ni escribir en servicios externos sin autorización explícita.

## Calidad y seguridad

- Ejecutar pruebas y verificaciones relacionadas antes de declarar una implementación terminada.
- Registrar en la capability pruebas ejecutadas, resultados y pendientes conocidos.
- No asumir que una autorización conversacional elimina límites técnicos. `sudo`, acceso de red y escritura fuera del workspace siguen sujetos al sandbox y a sus aprobaciones.
- Si una petición contradice estas reglas, detener la acción conflictiva, identificar la regla y pedir confirmación explícita.

## Organización

- Usar `PLAN_IMPLEMENTACION.md` como índice y estado general.
- Mantener especificaciones funcionales dentro de `capabilities/`.
- Mantener código ejecutable dentro de `backend/` y `frontend/`; no duplicarlo en `capabilities/`.
- Organizar backend por dominio dentro de cada capa. Un dominio aprobado debe usar el mismo nombre de carpeta en `domain`, `application`, `infrastructure` e `interfaces/http` cuando tenga responsabilidades en esas capas.
- Usar nombres de dominio singulares, descriptivos y en minúsculas, por ejemplo `cuenta`, `negocio`, `catalogo` o `inventario`.
- Mantener dirección de dependencias: `interfaces/http` depende de `application`; `infrastructure` implementa interfaces de `application`; `application` depende de `domain`; `domain` no depende de capas externas del proyecto.
- Evitar imports entre implementaciones de dominios. Una integración entre dominios usa identificadores, contratos de aplicación o interfaces explícitas; nunca repositories o handlers de otro dominio.
- No crear paquetes genéricos `shared`, `common`, `helpers` o `utils` como depósito. Código transversal permanece en raíz solo cuando no pertenece a dominio y tiene responsabilidad concreta.
- No mover ni modificar archivos de otro dominio o de trabajo activo de un colaborador sin incluirlo expresamente en capability aprobada.
- No crear carpetas `agents/`, `tools/` o `knowledge/` hasta que una capability aprobada demuestre su necesidad.
