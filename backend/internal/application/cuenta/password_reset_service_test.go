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

type resetChallengesStub struct {
	items      map[uuid.UUID]*cuentadomain.DesafioAutenticacion
	userCounts map[uuid.UUID]int64
	ipCounts   map[string]int64
	resetHash  string
	resets     int
	users      *verificationUsersStub
}

func newResetChallengesStub(users *verificationUsersStub) *resetChallengesStub {
	return &resetChallengesStub{items: map[uuid.UUID]*cuentadomain.DesafioAutenticacion{}, userCounts: map[uuid.UUID]int64{}, ipCounts: map[string]int64{}, users: users}
}

func (r *resetChallengesStub) FindChallenge(id uuid.UUID) (*cuentadomain.DesafioAutenticacion, error) {
	challenge, ok := r.items[id]
	if !ok {
		return nil, errors.New("no encontrado")
	}
	return challenge, nil
}
func (r *resetChallengesStub) IssueChallenge(challenge *cuentadomain.DesafioAutenticacion) error {
	for _, existing := range r.items {
		if existing.UsuarioID == challenge.UsuarioID && existing.Proposito == challenge.Proposito && existing.UsadoEn == nil {
			usedAt := challenge.UltimoEnvioEn
			existing.UsadoEn = &usedAt
		}
	}
	r.items[challenge.ID] = challenge
	r.userCounts[challenge.UsuarioID]++
	r.ipCounts[challenge.DireccionIP]++
	return nil
}
func (r *resetChallengesStub) CountChallengesByUserSince(id uuid.UUID, _ time.Time) (int64, error) {
	return r.userCounts[id], nil
}
func (r *resetChallengesStub) CountChallengesByIPSince(ip string, _ time.Time) (int64, error) {
	return r.ipCounts[ip], nil
}
func (r *resetChallengesStub) LastChallengeByUserAndPurpose(userID uuid.UUID, purpose string) (*cuentadomain.DesafioAutenticacion, error) {
	var last *cuentadomain.DesafioAutenticacion
	for _, item := range r.items {
		if item.UsuarioID == userID && item.Proposito == purpose && (last == nil || item.UltimoEnvioEn.After(last.UltimoEnvioEn)) {
			last = item
		}
	}
	if last == nil {
		return nil, errors.New("no encontrado")
	}
	return last, nil
}
func (r *resetChallengesStub) IncrementFailedAttempts(id uuid.UUID, _ int) error {
	r.items[id].IntentosFallidos++
	return nil
}
func (r *resetChallengesStub) ResetPasswordWithChallenge(challengeID, userID uuid.UUID, hash string, now time.Time, _ int) error {
	challenge := r.items[challengeID]
	if challenge.UsadoEn != nil {
		return errors.New("consumido")
	}
	challenge.UsadoEn = &now
	r.users.byID[userID].HashContrasena = hash
	r.resetHash = hash
	r.resets++
	return nil
}

type resetMailerStub struct {
	challengeID, token, recipient string
	changed                       []string
	err                           error
}

func (m *resetMailerStub) SendPasswordReset(_ context.Context, recipient, challengeID, token string, _ time.Time) error {
	m.recipient, m.challengeID, m.token = recipient, challengeID, token
	return m.err
}
func (m *resetMailerStub) SendPasswordChanged(_ context.Context, recipient string) error {
	m.changed = append(m.changed, recipient)
	return nil
}

func newResetFixture(t *testing.T) (*PasswordResetService, *verificationUsersStub, *resetChallengesStub, *resetMailerStub, *cuentadomain.Usuario) {
	t.Helper()
	verified := time.Now().Add(-time.Hour)
	user := &cuentadomain.Usuario{ID: uuid.New(), Correo: "ana@ejemplo.com", Estado: "activo", CorreoVerificadoEn: &verified}
	users := &verificationUsersStub{byEmail: map[string]*cuentadomain.Usuario{user.Correo: user}, byID: map[uuid.UUID]*cuentadomain.Usuario{user.ID: user}}
	challenges := newResetChallengesStub(users)
	mailer := &resetMailerStub{}
	service := NewPasswordResetService(users, challenges, mailer, PasswordResetConfig{
		HMACSecret: "secreto-de-prueba", TTL: 30 * time.Minute, MaxAttempts: 5, ResendWait: time.Minute, HourlySendMax: 5,
	})
	return service, users, challenges, mailer, user
}

func TestPasswordResetRequestIssuesHashedSingleUseLink(t *testing.T) {
	service, _, challenges, mailer, user := newResetFixture(t)

	if err := service.Request(context.Background(), "  ANA@ejemplo.com ", "10.0.0.1"); err != nil {
		t.Fatal(err)
	}
	if mailer.recipient != user.Correo || len(mailer.token) < 40 {
		t.Fatalf("correo no enviado correctamente: %#v", mailer)
	}
	challenge := challenges.items[uuid.MustParse(mailer.challengeID)]
	if challenge.Proposito != cuentadomain.PropositoRecuperacionContrasena || challenge.HashOTP == mailer.token || len(challenge.HashOTP) != 64 {
		t.Fatalf("desafío inseguro o incorrecto: %#v", challenge)
	}
	if ttl := challenge.ExpiraEn.Sub(challenge.UltimoEnvioEn); ttl != 30*time.Minute {
		t.Fatalf("vigencia = %s", ttl)
	}
}

func TestPasswordResetRequestIsUniformForUnknownOrUnavailableAccounts(t *testing.T) {
	service, _, challenges, mailer, user := newResetFixture(t)

	if err := service.Request(context.Background(), "nadie@ejemplo.com", "10.0.0.1"); err != nil {
		t.Fatal(err)
	}
	user.Estado = "pendiente_verificacion"
	if err := service.Request(context.Background(), user.Correo, "10.0.0.1"); err != nil {
		t.Fatal(err)
	}
	if len(challenges.items) != 0 || mailer.recipient != "" {
		t.Fatal("no debía emitir desafíos para cuentas inexistentes o no activas")
	}
}

func TestPasswordResetRequestRespectsWaitAndHourlyLimits(t *testing.T) {
	service, _, challenges, _, _ := newResetFixture(t)
	ctx := context.Background()

	_ = service.Request(ctx, "ana@ejemplo.com", "10.0.0.1")
	_ = service.Request(ctx, "ana@ejemplo.com", "10.0.0.1")
	if len(challenges.items) != 1 {
		t.Fatalf("la espera mínima no se respetó: %d desafíos", len(challenges.items))
	}

	service.now = func() time.Time { return time.Now().Add(2 * time.Minute) }
	challenges.ipCounts["10.0.0.2"] = 5
	if err := service.Request(ctx, "ana@ejemplo.com", "10.0.0.2"); err != nil {
		t.Fatal(err)
	}
	if len(challenges.items) != 1 {
		t.Fatal("el límite por IP no se respetó")
	}
}

func TestPasswordResetRequestIgnoresDeliveryFailure(t *testing.T) {
	service, _, challenges, mailer, _ := newResetFixture(t)
	mailer.err = errors.New("smtp caído")

	if err := service.Request(context.Background(), "ana@ejemplo.com", "10.0.0.1"); err != nil {
		t.Fatalf("la respuesta debía ser uniforme aunque falle el correo: %v", err)
	}
	if len(challenges.items) != 1 {
		t.Fatal("el desafío debía quedar emitido para permitir reintento")
	}
}

func TestPasswordResetConsumesTokenOnceAndNotifies(t *testing.T) {
	service, users, challenges, mailer, user := newResetFixture(t)
	ctx := context.Background()
	_ = service.Request(ctx, user.Correo, "10.0.0.1")
	challengeID := uuid.MustParse(mailer.challengeID)

	if err := service.Reset(ctx, challengeID, mailer.token, "nueva-contraseña-segura"); err != nil {
		t.Fatal(err)
	}
	if bcrypt.CompareHashAndPassword([]byte(users.byID[user.ID].HashContrasena), []byte("nueva-contraseña-segura")) != nil {
		t.Fatal("la contraseña no se actualizó con bcrypt")
	}
	if len(mailer.changed) != 1 || mailer.changed[0] != user.Correo {
		t.Fatalf("no se notificó el cambio: %#v", mailer.changed)
	}
	if err := service.Reset(ctx, challengeID, mailer.token, "otra-contraseña-segura"); !errors.Is(err, ErrPasswordResetInvalid) {
		t.Fatalf("el enlace debía ser de un solo uso; error=%v", err)
	}
	if challenges.resets != 1 {
		t.Fatalf("resets = %d", challenges.resets)
	}
}

func TestPasswordResetRejectsWrongExpiredOrForeignChallenges(t *testing.T) {
	service, _, challenges, mailer, user := newResetFixture(t)
	ctx := context.Background()
	_ = service.Request(ctx, user.Correo, "10.0.0.1")
	challengeID := uuid.MustParse(mailer.challengeID)

	if err := service.Reset(ctx, challengeID, "token-incorrecto", "nueva-contraseña"); !errors.Is(err, ErrPasswordResetInvalid) {
		t.Fatalf("token incorrecto: %v", err)
	}
	if challenges.items[challengeID].IntentosFallidos != 1 {
		t.Fatal("el intento fallido no se contabilizó")
	}

	challenges.items[challengeID].IntentosFallidos = 5
	if err := service.Reset(ctx, challengeID, mailer.token, "nueva-contraseña"); !errors.Is(err, ErrPasswordResetInvalid) {
		t.Fatalf("desafío agotado: %v", err)
	}
	challenges.items[challengeID].IntentosFallidos = 0

	service.now = func() time.Time { return time.Now().Add(31 * time.Minute) }
	if err := service.Reset(ctx, challengeID, mailer.token, "nueva-contraseña"); !errors.Is(err, ErrPasswordResetInvalid) {
		t.Fatalf("desafío expirado: %v", err)
	}
	service.now = time.Now

	challenges.items[challengeID].Proposito = cuentadomain.PropositoVerificacionCorreo
	if err := service.Reset(ctx, challengeID, mailer.token, "nueva-contraseña"); !errors.Is(err, ErrPasswordResetInvalid) {
		t.Fatalf("un desafío de verificación no debe servir para recuperar: %v", err)
	}
	if challenges.resets != 0 {
		t.Fatal("ninguna contraseña debía cambiar")
	}
}

func TestPasswordResetEnforcesPasswordPolicy(t *testing.T) {
	service, _, _, _, _ := newResetFixture(t)
	for _, password := range []string{"corta", string(make([]byte, 73))} {
		if err := service.Reset(context.Background(), uuid.New(), "token", password); !errors.Is(err, ErrPasswordPolicy) {
			t.Fatalf("contraseña de %d bytes: error=%v", len(password), err)
		}
	}
}
