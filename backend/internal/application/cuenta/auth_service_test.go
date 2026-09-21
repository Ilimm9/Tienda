package cuenta

import (
	"testing"
	"time"

	cuentadomain "tienda/backend/internal/domain/cuenta"

	"golang.org/x/crypto/bcrypt"
)

type userRepositoryStub struct {
	user           *cuentadomain.Usuario
	findErr        error
	createdUser    *cuentadomain.Usuario
	createdProfile *cuentadomain.PerfilUsuario
	saved          bool
}

func (r *userRepositoryStub) FindByEmail(string) (*cuentadomain.Usuario, error) {
	return r.user, r.findErr
}

func (r *userRepositoryStub) Save(*cuentadomain.Usuario) error {
	r.saved = true
	return nil
}

func (r *userRepositoryStub) CreateAccount(user *cuentadomain.Usuario, profile *cuentadomain.PerfilUsuario) error {
	r.createdUser = user
	r.createdProfile = profile
	return nil
}

func TestAuthServiceLoginConservaContrato(t *testing.T) {
	verifiedAt := time.Now()
	hash, err := bcrypt.GenerateFromPassword([]byte("contrasena-segura"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	repository := &userRepositoryStub{user: &cuentadomain.Usuario{
		Correo: "ada@example.com", HashContrasena: string(hash), Estado: "activo", CorreoVerificadoEn: &verifiedAt,
	}}
	service := NewAuthService(repository)

	user, err := service.Login(" ADA@EXAMPLE.COM ", "contrasena-segura")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if user != repository.user || !repository.saved {
		t.Fatal("login no conservó actualización de sesión")
	}
}
