package cuenta

import (
	"errors"
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

func TestAuthServiceRegisterConservaContrato(t *testing.T) {
	repository := &userRepositoryStub{findErr: errors.New("no encontrado")}
	service := NewAuthService(repository)

	if err := service.Register("Ada Lovelace", " ADA@EXAMPLE.COM ", "+52 55 1234", "contrasena-segura"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if repository.createdUser == nil || repository.createdUser.Correo != "ada@example.com" {
		t.Fatalf("usuario creado = %#v", repository.createdUser)
	}
	if repository.createdUser.CorreoVerificadoEn == nil {
		t.Fatal("durante la transición previa a OTP el registro debe conservar el acceso existente")
	}
	if repository.createdProfile == nil || repository.createdProfile.Nombres != "Ada" || repository.createdProfile.Apellidos != "Lovelace" {
		t.Fatalf("perfil creado = %#v", repository.createdProfile)
	}
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
