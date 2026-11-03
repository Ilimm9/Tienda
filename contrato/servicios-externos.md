# Servicios externos: opciones y recomendación

- **Para:** thrs e Ilian.
- **Consultado:** 2026-10-31. Precios en USD para almacenamiento y MXN para cobros. Confirmar con cada proveedor antes de contratar.
- **Alcance:** comparar y recomendar. No autoriza contratar ni implementar.

**Revisión del 2026-11-02:** máximo cuatro meses y 20 h semanales por persona. Solo almacenamiento y terminal independiente pertenecen a la entrega prevista; los plus permanecen fuera. Los precios del 2026-10-31 no se han vuelto a verificar en esta revisión y no deben tratarse como cotización.

## Resumen

| Necesidad | Recomendación | Cuándo |
| --- | --- | --- |
| Respaldos y archivos | **Cloudflare R2** (decidido), con API S3 para poder usar Amazon S3 | Fase 1, antes de E1-02 y E1-09 |
| Tarjeta | **Terminal independiente con referencia manual** (decidido). Después, **Mercado Pago Point** integrado si la Lic. lo pide | Fase 2 (E2-05); integración como plus |
| Recargas y servicios | **Taecel**, solo en efectivo | Posible plus |
| Pasarela y planes de uso | **Stripe** para planes de uso; comparar Mercado Pago para cobros del negocio | Posible plus |

## 1. Almacenamiento

Hoy no hay almacenamiento de objetos: sin SDK de S3 y sin subida de archivos en el backend (`ImagenURL` solo guarda URLs). Hay dos necesidades: respaldos fuera del servidor (E1-02) y fotos de tickets (E1-09).

| Opción | Almacenamiento | Salida a internet | Inmutabilidad | Valoración |
| --- | --- | --- | --- | --- |
| **Cloudflare R2** | $0.015/GB-mes; 10 GB gratis | Gratis | Bucket locks | **Recomendada.** Ya usamos Cloudflare, API compatible con S3 y costo predecible |
| Amazon S3 | ~$0.023/GB-mes (fuente secundaria) | ~$0.09/GB; 100 GB/mes gratis entre servicios AWS | Object Lock (fuente secundaria) | La más completa, pero mismo proveedor que el servidor y salida con costo |
| Lightsail Object Storage | $1 (5 GB), $3 (100 GB), $5 (250 GB) al mes | 25, 250 y 500 GB incluidos | No aparece en precios | Sencilla, pero también dentro de AWS |
| Backblaze B2 | $6.95/TB-mes; 10 GB gratis | Gratis hasta 3 veces lo almacenado | Object Lock (fuente secundaria) | La más barata con volumen; mejor como segunda copia futura |

**Por qué R2:**

- **Respaldo en otro proveedor.** Si se pierde el servidor o la cuenta de AWS, el respaldo sobrevive.
- **Restaurar no cuesta salida,** justo cuando más urge.
- **Portabilidad de operaciones básicas.** La interfaz puede abstraer subir, descargar y firmar URLs; cambiar de proveedor exige verificar compatibilidad y configuración de seguridad, retención y credenciales.

**Y Amazon S3:** R2 implementa parte de la API S3 y puede usar `aws-sdk-go-v2`. La retención no es intercambiable automáticamente: R2 administra bucket locks por sus propios mecanismos y no implementa toda la API S3 Object Lock. Validar portabilidad mediante pruebas, sin prometer que todo cambia solo por configuración. Fuentes oficiales: [compatibilidad S3](https://developers.cloudflare.com/r2/api/s3/api/) y [bucket locks](https://developers.cloudflare.com/r2/buckets/bucket-locks/). S3 puede ser una segunda copia futura; decidirlo en F3-06.

**Costo esperado:** cerca de $0 a $1 USD al mes durante el primer año. Hipótesis: 40 fotos al día de 500 KB (~7 GB al año) y 30 respaldos de 50 MB (1.5 GB). Medir con datos reales.

**Condiciones de implementación:**

- Dos buckets privados: `respaldos` y `archivos`, con credenciales distintas.
- Respaldos cifrados antes de salir del servidor, con bucket lock. La credencial del servidor no debe poder borrarlos.
- Archivos con ruta por negocio y sucursal, URLs firmadas de corta duración, validación de tipo y tamaño, y sin nombres originales.
- Interfaz `AlmacenArchivos` en `application`, implementada en `infrastructure`.
- Decidir si un RPO de 1 h justifica `pgBackRest` o `WAL-G`; si no, `pg_dump` diario cubre el mínimo de 24 h.
- Definir retención, custodia y recuperación de llaves, restauración de fotos y permisos de administración. No guardar la única llave de descifrado junto al servidor respaldado. RPO/RTO quedan pendientes en D-21 y deben demostrarse con restauración.

## 2. Pagos con tarjeta

**Regla común:** la tarjeta se lee solo en la terminal. Stockion nunca recibe ni guarda número de tarjeta. Así evitamos la mayor parte de la carga de cumplimiento PCI.

| Opción | Comisión publicada | Integración | Limitante | Valoración |
| --- | --- | --- | --- | --- |
| **Terminal independiente + referencia manual** | La del banco actual | Ninguna: el cajero captura la referencia | Error de captura; cuadre manual; el cobro depende de la conexión y autorización de la terminal | **Recomendada para Fase 2.** Stockion registra sin internet; verificar el cobro bancario por separado |
| **Mercado Pago Point** | 3.5 % + IVA; sin renta | API pública (`/v1/orders`) y Webhooks | Modelos compatibles por confirmar | **Recomendada si se integra** |
| Clip Empresas | 2.99 % + $1 + IVA, con condición | API, SDK y Webhooks; sin documentación pública | Exige volumen: su página dice $60 mil y también $100 mil al mes | Opción si la Lic. supera ese volumen |
| Stripe Terminal | 3.6 % + $3 (texto ambiguo) | Muy buena | Vista previa pública en México; sin meses sin intereses presenciales | No todavía |
| Kushki | No publicada | Cloud Terminal API | Beta; solo terminales Ultra y Ultra P | No todavía |

**Comparación por ticket:**

| Ticket | Mercado Pago | Clip (si aplica) |
| --- | --- | --- |
| $100 | $4.06 | $4.63 |
| $500 | $20.30 | $18.50 |

Clip resulta más barato arriba de unos $196 por ticket. Son cálculos propios con las tasas publicadas.

**Por qué no integrar primero:**

- Una terminal integrada necesita internet en la caja para recibir la orden. El respaldo manual se construye de todos modos.
- `origin/ventas` ya acepta `TARJETA` y `TRANSFERENCIA` en backend; falta la interfaz, los pagos mixtos y la referencia.
- Conviene modelar `venta_pagos` con proveedor y referencia desde E1-05. Así, integrar después no cambia el esquema.

**Preguntar a la Lic.:** banco y terminal actuales, comisión que paga y monto mensual con tarjeta.

## 3. Recargas y pago de servicios

**Taecel sí sirve para lo que pensabas:** ofrece API REST para vender tiempo aire, cobrar recibos y vender pines desde sistema propio.

| Opción | API | Comisión | Valoración |
| --- | --- | --- | --- |
| **Taecel** | Sí, con datos de prueba al aprobar el levantamiento tecnológico | Hay que preguntarla | **Recomendada para iniciar.** Cuenta gratis, sin contrato forzoso |
| SIPREL | Sí, con códigos de prueba y verificación previa | No publicada | Alternativa. Productos y proceso muy parecidos a Taecel: confirmar si es la misma plataforma |
| MTCenter | No encontrada | No publicada | Más de 75 servicios, pero sin integración visible |

**Modelo de Taecel:**

- **Saldo prepagado por depósito o transferencia,** validado hasta en 30 minutos.
- **Tiempo aire:** primera compra desde $50; después, $300. La comisión llega como saldo extra.
- **Pago de servicios:** saldo separado, desde $1,000. La comisión es la que se cobra al cliente.
- **No se transfiere** saldo de servicios a tiempo aire.

**Riesgos antes de decidir:**

- **Requiere internet.** No puede sumarse a la venta sin conexión.
- **Margen muy delgado con tarjeta.** Referencia de SICAR: Telcel 5 % y $5 por recibo de CFE. En una recarga de $100 se ganan $5, y cobrarla con Mercado Pago cuesta $4.06. **Vender solo en efectivo.**
- **Resultado incierto.** Si la red se cae durante la petición, no se sabe si se aplicó. Exige estado pendiente, consulta y conciliación diaria.
- **No diferencia.** SICAR ya lo incluye en su POS. Atrae clientes a la tienda, pero no es razón para elegir Stockion.
- **Saldo del negocio.** La Lic. inmoviliza dinero y asume el riesgo operativo; no conviene manejarlo con cuenta del equipo.

**Antes de implementar:** pedir a Taecel comisiones por escrito y documentación técnica, y preguntar a la Lic. si hoy vende recargas y cuántas.

## 4. Pasarela de pagos y planes de uso

Son dos usos distintos:

- **Planes de uso:** cada negocio nos paga una suscripción. **Uso interno; no se presenta a la Lic. Mari.**
- **Cobros en línea:** un negocio cobra a sus clientes por liga o tarjeta. Sí puede ofrecerse a la Lic. Mari como plus.

Comisiones de fuentes secundarias, sin confirmar en páginas oficiales:

| Opción | Comisión reportada | Recurrencia | Valoración |
| --- | --- | --- | --- |
| **Stripe** | 3.6 % + $3 + IVA | Suscripciones y facturación muy completas | **Recomendada para planes de uso** |
| Mercado Pago | ~3.49 % + IVA | API de suscripciones; admite CLABE y OXXO | **Comparar para cobros del negocio.** Conocida por comercios y dueños |
| Conekta | ~2.9 % + $2.50 | API de planes y suscripciones | Alternativa barata y local; validar soporte |
| Openpay | ~2.9 % + $2.50 | No confirmada | Por investigar |

**Reglas:**

- Usar formularios del proveedor: Stockion nunca toca datos de tarjeta.
- Webhooks firmados e idempotentes. El estado de un pago solo cambia por confirmación del proveedor.
- Cada negocio cobra con su propia cuenta; no pasar dinero de clientes por nuestra cuenta.
- Interfaz `PasarelaPago` en `application`, para poder cambiar de proveedor.
- Antes de cobrar planes: pentest externo y términos legales (E3-08, E3-11, E3-13).

## Fuentes

**Almacenamiento**
- [Precios de Cloudflare R2](https://developers.cloudflare.com/r2/pricing/)
- [Bucket locks de R2](https://developers.cloudflare.com/r2/buckets/bucket-locks/)
- [Precios de Lightsail](https://aws.amazon.com/lightsail/pricing/)
- [Precios de Backblaze B2](https://www.backblaze.com/cloud-storage/pricing)
- [Precios de Amazon S3](https://aws.amazon.com/s3/pricing/): no se pudo leer la tabla de almacenamiento.
- [Comparativa R2, S3 y B2](https://infratally.com/articles/cloudflare-r2-vs-aws-s3-vs-backblaze-b2-2026/): fuente secundaria.

**Pagos**
- [Mercado Pago Point para desarrolladores](https://www.mercadopago.com.mx/developers/es/docs/mp-point/payment-processing)
- [Terminales y comisiones de Mercado Pago Point](https://www.mercadopago.com.mx/herramientas-para-vender/lectores-point)
- [Clip Empresas](https://www.clip.mx/clip-para-empresas)
- [Comisiones de Clip, octubre de 2026](https://www.eluniversal.com.mx/descubre-y-compra/clip-comisiones-cuanto-cobra-clip-mexico-y-que-terminal-elegir/)
- [Stripe Terminal en México](https://support.stripe.com/questions/terminal-in-mexico)
- [Precios de Stripe México](https://stripe.com/mx/pricing)
- [Kushki Cloud Terminal API](https://docs.kushki.com/mx/card-present-payments/cloud-terminal-api)

**Recargas**
- [Integración API de Taecel (PDF)](https://cdn.taecel.com/src/web/taecel/descargas/integrador-api-taecel.pdf)
- [Preguntas frecuentes de Taecel](https://taecel.com/portal/preguntas-frecuentes-venta-de-tiempo-aire-telcel)
- [Integración de SIPREL](https://siprel.mx/integracion-web-services/)
- [MTCenter](https://www.mtcenter.com.mx/)
- [Recargas en SICAR](https://www.sicar.mx/sistema-de-recargas-electronicas)

**Pasarela** (fuentes secundarias, salvo Mercado Pago)
- [Suscripciones de Mercado Pago](https://www.mercadopago.com.mx/developers/en/docs/subscriptions/overview)
- [Comparativa de pasarelas en México](https://www.guiadesoftware.com/blog/mejor-pasarela-pago-mexico)
- [Stripe y Mercado Pago](https://www.guiadesoftware.com/blog/stripe-vs-mercado-pago)
- [Opciones de pago, Platzi](https://platzi.com/cursos/pagos-electronicos/opciones-de-pago-internacionales/)
- [Conekta: guía de API](https://jentic.com/apis/conekta.com/conekta)
