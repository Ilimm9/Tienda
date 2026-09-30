package cuenta

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	application "tienda/backend/internal/application/cuenta"
	cuentadomain "tienda/backend/internal/domain/cuenta"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type resetUsersStub struct{}

func (resetUsersStub) FindByEmail(string) (*cuentadomain.Usuario, error) {
	return nil, errors.New("no encontrado")
}
func (resetUsersStub) FindByID(uuid.UUID) (*cuentadomain.Usuario, error) {
	return nil, errors.New("no encontrado")
}

type resetChallengesStub struct{}

func (resetChallengesStub) FindChallenge(uuid.UUID) (*cuentadomain.DesafioAutenticacion, error) {
	return nil, errors.New("no encontrado")
}
func (resetChallengesStub) IssueChallenge(*cuentadomain.DesafioAutenticacion) error { return nil }
func (resetChallengesStub) CountChallengesByUserSince(uuid.UUID, time.Time) (int64, error) {
	return 0, nil
}
func (resetChallengesStub) CountChallengesByIPSince(string, time.Time) (int64, error) {
	return 0, nil
}
func (resetChallengesStub) LastChallengeByUserAndPurpose(uuid.UUID, string) (*cuentadomain.DesafioAutenticacion, error) {
	return nil, errors.New("no encontrado")
}
func (resetChallengesStub) IncrementFailedAttempts(uuid.UUID, int) error { return nil }
func (resetChallengesStub) ResetPasswordWithChallenge(uuid.UUID, uuid.UUID, string, time.Time, int) error {
	return nil
}

type resetMailerStub struct{}

func (resetMailerStub) SendPasswordReset(context.Context, string, string, string, time.Time) error {
	return nil
}
func (resetMailerStub) SendPasswordChanged(context.Context, string) error { return nil }

func resetRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewPasswordResetHandler(application.NewPasswordResetService(resetUsersStub{}, resetChallengesStub{}, resetMailerStub{}, application.PasswordResetConfig{
		HMACSecret: "secreto", TTL: 30 * time.Minute, MaxAttempts: 5, ResendWait: time.Minute, HourlySendMax: 5,
	}))
	router := gin.New()
	router.POST("/solicitar-recuperacion", handler.Request)
	router.POST("/restablecer-contrasena", handler.Reset)
	return router
}

func postJSON(router *gin.Engine, path, body string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestPasswordResetRequestRespondsUniformly(t *testing.T) {
	router := resetRouter()

	response := postJSON(router, "/solicitar-recuperacion", `{"correo":"nadie@ejemplo.com"}`)
	if response.Code != http.StatusAccepted || !strings.Contains(response.Body.String(), "Si existe una cuenta") {
		t.Fatalf("respuesta = %d %s", response.Code, response.Body.String())
	}
	if response := postJSON(router, "/solicitar-recuperacion", `{"correo":"no-es-correo"}`); response.Code != http.StatusBadRequest {
		t.Fatalf("correo inválido: %d", response.Code)
	}
}

func TestPasswordResetConfirmMapsErrors(t *testing.T) {
	router := resetRouter()
	valid := uuid.NewString()

	cases := map[string]struct {
		body   string
		status int
	}{
		"desafío no uuid":     {`{"desafio_id":"x","token":"t","contrasena":"contraseña-larga"}`, http.StatusUnprocessableEntity},
		"campos faltantes":    {`{"desafio_id":"` + valid + `"}`, http.StatusBadRequest},
		"contraseña corta":    {`{"desafio_id":"` + valid + `","token":"t","contrasena":"corta"}`, http.StatusBadRequest},
		"desafío inexistente": {`{"desafio_id":"` + valid + `","token":"t","contrasena":"contraseña-larga"}`, http.StatusUnprocessableEntity},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if response := postJSON(router, "/restablecer-contrasena", tc.body); response.Code != tc.status {
				t.Fatalf("status = %d, se esperaba %d: %s", response.Code, tc.status, response.Body.String())
			}
		})
	}
}
