package cuenta

import (
	"context"
	"errors"
	"testing"
	"time"

	cuentadomain "tienda/backend/internal/domain/cuenta"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type verificationUsersStub struct {
	byEmail map[string]*cuentadomain.Usuario
	byID    map[uuid.UUID]*cuentadomain.Usuario
}

func (r *verificationUsersStub) FindByEmail(email string) (*cuentadomain.Usuario, error) {
	user, ok := r.byEmail[email]
	if !ok {
		return nil, errors.New("no encontrado")
	}
	return user, nil
}
func (r *verificationUsersStub) FindByID(id uuid.UUID) (*cuentadomain.Usuario, error) {
	user, ok := r.byID[id]
	if !ok {
		return nil, errors.New("no encontrado")
	}
	return user, nil
}
func (r *verificationUsersStub) CreateAccount(user *cuentadomain.Usuario, _ *cuentadomain.PerfilUsuario) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	r.byEmail[user.Correo] = user
	r.byID[user.ID] = user
	return nil
}

type verificationChallengesStub struct {
	items      map[uuid.UUID]*cuentadomain.DesafioAutenticacion
	userCounts map[uuid.UUID]int64
	ipCounts   map[string]int64
	last       *cuentadomain.DesafioAutenticacion
}

func (r *verificationChallengesStub) FindChallenge(id uuid.UUID) (*cuentadomain.DesafioAutenticacion, error) {
	challenge, ok := r.items[id]
	if !ok {
		return nil, errors.New("no encontrado")
	}
	return challenge, nil
}
func (r *verificationChallengesStub) IssueChallenge(challenge *cuentadomain.DesafioAutenticacion) error {
	for _, existing := range r.items {
		if existing.UsuarioID == challenge.UsuarioID && existing.UsadoEn == nil {
			usedAt := challenge.UltimoEnvioEn
			existing.UsadoEn = &usedAt
		}
	}
	r.items[challenge.ID] = challenge
	r.userCounts[challenge.UsuarioID]++
	r.ipCounts[challenge.DireccionIP]++
	r.last = challenge
	return nil
}
func (r *verificationChallengesStub) IssueChallengeForPendingAccount(user *cuentadomain.Usuario, _ *cuentadomain.PerfilUsuario, challenge *cuentadomain.DesafioAutenticacion) error {
	return r.IssueChallenge(challenge)
}
func (r *verificationChallengesStub) CountChallengesByUserSince(id uuid.UUID, _ time.Time) (int64, error) {
	return r.userCounts[id], nil
}
func (r *verificationChallengesStub) CountChallengesByIPSince(ip string, _ time.Time) (int64, error) {
	return r.ipCounts[ip], nil
}
func (r *verificationChallengesStub) LastChallengeByUser(uuid.UUID) (*cuentadomain.DesafioAutenticacion, error) {
	if r.last == nil {
		return nil, errors.New("no encontrado")
	}
	return r.last, nil
}
func (r *verificationChallengesStub) IncrementFailedAttempts(id uuid.UUID, _ int) error {
	r.items[id].IntentosFallidos++
	return nil
}
func (r *verificationChallengesStub) ActivateUserWithChallenge(id, userID uuid.UUID, now time.Time, max int) error {
	challenge := r.items[id]
	if challenge == nil || challenge.UsadoEn != nil || !challenge.ExpiraEn.After(now) || challenge.IntentosFallidos >= max {
		return errors.New("inválido")
	}
	challenge.UsadoEn = &now
	return nil
}

type verificationMailerStub struct {
	recipient string
	code      string
	err       error
}

func (m *verificationMailerStub) SendVerificationOTP(_ context.Context, recipient, code string, _ time.Time) error {
	m.recipient, m.code = recipient, code
	return m.err
}

func newVerificationServiceForTest() (*VerificationService, *verificationUsersStub, *verificationChallengesStub, *verificationMailerStub) {
	users := &verificationUsersStub{byEmail: map[string]*cuentadomain.Usuario{}, byID: map[uuid.UUID]*cuentadomain.Usuario{}}
	challenges := &verificationChallengesStub{items: map[uuid.UUID]*cuentadomain.DesafioAutenticacion{}, userCounts: map[uuid.UUID]int64{}, ipCounts: map[string]int64{}}
	mailer := &verificationMailerStub{}
	service := NewVerificationService(users, challenges, mailer, VerificationConfig{
		HMACSecret: "unit-test-secret-with-at-least-32-characters", TTL: 10 * time.Minute,
		MaxAttempts: 5, ResendWait: time.Minute, HourlySendMax: 5,
	})
	service.now = func() time.Time { return time.Date(2026, 9, 20, 20, 0, 0, 0, time.UTC) }
	return service, users, challenges, mailer
}

func TestRegisterCreatesPendingAccountAndStoresOnlyOTPHMAC(t *testing.T) {
	service, users, challenges, mailer := newVerificationServiceForTest()
	result, err := service.Register(context.Background(), "Ada Lovelace", " ADA@EXAMPLE.COM ", "+52 55", "contrasena-segura", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	user := users.byEmail["ada@example.com"]
	if user == nil || user.Estado != "pendiente_verificacion" || user.CorreoVerificadoEn != nil {
		t.Fatalf("usuario pendiente inválido: %#v", user)
	}
	challenge := challenges.items[result.ChallengeID]
	if challenge == nil || len(challenge.HashOTP) != 64 || challenge.HashOTP == mailer.code || mailer.recipient != user.Correo {
		t.Fatalf("desafío o correo inválido: %#v", challenge)
	}
}

func TestVerifyRejectsWrongCodeThenActivatesAndConsumesCorrectCode(t *testing.T) {
	service, users, challenges, mailer := newVerificationServiceForTest()
	result, err := service.Register(context.Background(), "Ada Lovelace", "ada@example.com", "", "contrasena-segura", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Verify(context.Background(), result.ChallengeID, "000000"); !errors.Is(err, ErrVerificationInvalid) {
		t.Fatalf("código incorrecto: %v", err)
	}
	if challenges.items[result.ChallengeID].IntentosFallidos != 1 {
		t.Fatal("el intento fallido no fue contabilizado")
	}
	user, err := service.Verify(context.Background(), result.ChallengeID, mailer.code)
	if err != nil {
		t.Fatal(err)
	}
	user.Estado = "activo"
	now := service.now()
	user.CorreoVerificadoEn = &now
	if challenges.items[result.ChallengeID].UsadoEn == nil || users.byID[user.ID] == nil {
		t.Fatal("el desafío no quedó consumido")
	}
	if _, err := service.Verify(context.Background(), result.ChallengeID, mailer.code); !errors.Is(err, ErrVerificationInvalid) {
		t.Fatal("un código consumido no puede reutilizarse")
	}
}

func TestResendEnforcesWaitAndHourlyLimit(t *testing.T) {
	service, _, challenges, _ := newVerificationServiceForTest()
	result, err := service.Register(context.Background(), "Ada Lovelace", "ada@example.com", "", "contrasena-segura", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Resend(context.Background(), result.ChallengeID, "127.0.0.1"); !errors.Is(err, ErrVerificationTooSoon) {
		t.Fatalf("reenvío inmediato: %v", err)
	}
	service.now = func() time.Time { return time.Date(2026, 9, 20, 20, 2, 0, 0, time.UTC) }
	challenges.userCounts[challenges.last.UsuarioID] = 5
	if _, err := service.Resend(context.Background(), result.ChallengeID, "127.0.0.1"); !errors.Is(err, ErrVerificationLimited) {
		t.Fatalf("límite horario: %v", err)
	}
}

func TestResendInvalidatesPreviousChallenge(t *testing.T) {
	service, _, challenges, _ := newVerificationServiceForTest()
	first, err := service.Register(context.Background(), "Ada Lovelace", "ada@example.com", "", "contrasena-segura", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	service.now = func() time.Time { return time.Date(2026, 9, 20, 20, 2, 0, 0, time.UTC) }
	second, err := service.Resend(context.Background(), first.ChallengeID, "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if first.ChallengeID == second.ChallengeID || challenges.items[first.ChallengeID].UsadoEn == nil {
		t.Fatal("el reenvío no invalidó el desafío anterior")
	}
}

func TestVerifyRejectsExpiredAndExhaustedChallenges(t *testing.T) {
	service, _, challenges, mailer := newVerificationServiceForTest()
	result, err := service.Register(context.Background(), "Ada Lovelace", "ada@example.com", "", "contrasena-segura", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	challenge := challenges.items[result.ChallengeID]
	challenge.ExpiraEn = service.now().Add(-time.Second)
	if _, err := service.Verify(context.Background(), result.ChallengeID, mailer.code); !errors.Is(err, ErrVerificationInvalid) {
		t.Fatal("un desafío expirado fue aceptado")
	}
	challenge.ExpiraEn = service.now().Add(time.Minute)
	challenge.IntentosFallidos = 5
	if _, err := service.Verify(context.Background(), result.ChallengeID, mailer.code); !errors.Is(err, ErrVerificationInvalid) {
		t.Fatal("un desafío agotado fue aceptado")
	}
}

func TestRegisterDoesNotRevealExistingActiveEmail(t *testing.T) {
	service, users, challenges, mailer := newVerificationServiceForTest()
	verifiedAt := service.now()
	active := &cuentadomain.Usuario{ID: uuid.New(), Correo: "ada@example.com", Estado: "activo", CorreoVerificadoEn: &verifiedAt}
	users.byEmail[active.Correo], users.byID[active.ID] = active, active
	result, err := service.Register(context.Background(), "Otra Persona", active.Correo, "", "contrasena-segura", "127.0.0.1")
	if err != nil || result.ChallengeID == uuid.Nil || len(challenges.items) != 0 || mailer.code != "" {
		t.Fatalf("registro enumerable: result=%#v err=%v", result, err)
	}
}

func TestRegisterAgainReplacesPendingPasswordAndInvalidatesOldCode(t *testing.T) {
	service, users, challenges, _ := newVerificationServiceForTest()
	first, err := service.Register(context.Background(), "Ada Lovelace", "ada@example.com", "", "primera-contrasena", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	service.now = func() time.Time { return time.Date(2026, 9, 20, 20, 2, 0, 0, time.UTC) }
	second, err := service.Register(context.Background(), "Ada Actualizada", "ada@example.com", "", "segunda-contrasena", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	user := users.byEmail["ada@example.com"]
	if bcrypt.CompareHashAndPassword([]byte(user.HashContrasena), []byte("segunda-contrasena")) != nil {
		t.Fatal("el registro pendiente conservó una contraseña anterior")
	}
	if challenges.items[first.ChallengeID].UsadoEn == nil || first.ChallengeID == second.ChallengeID {
		t.Fatal("el desafío anterior no fue invalidado")
	}
}
