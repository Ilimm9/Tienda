package http

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// LimitarPorIP aplica una ventana fija en memoria por IP; protege rutas públicas sin sesión.
//
// Es por proceso: con varias réplicas el límite efectivo se multiplica.
func LimitarPorIP(maximo int, ventana time.Duration) gin.HandlerFunc {
	type cuenta struct {
		inicio time.Time
		total  int
	}
	var mu sync.Mutex
	cuentas := map[string]*cuenta{}
	return func(c *gin.Context) {
		ahora := time.Now()
		ip := c.ClientIP()
		mu.Lock()
		// La limpieza oportunista evita que el mapa crezca con IPs que ya no vuelven.
		if len(cuentas) > 10_000 {
			for clave, actual := range cuentas {
				if ahora.Sub(actual.inicio) >= ventana {
					delete(cuentas, clave)
				}
			}
		}
		actual, ok := cuentas[ip]
		if !ok || ahora.Sub(actual.inicio) >= ventana {
			actual = &cuenta{inicio: ahora}
			cuentas[ip] = actual
		}
		actual.total++
		excedido := actual.total > maximo
		mu.Unlock()
		if excedido {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"codigo": "LIMITE_EXCEDIDO", "mensaje": "Demasiados intentos. Espera un momento.", "campos": gin.H{},
			})
			return
		}
		c.Next()
	}
}
