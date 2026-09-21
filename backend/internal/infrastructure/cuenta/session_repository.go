package cuenta

import (
	"time"

	cuentadomain "tienda/backend/internal/domain/cuenta"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionRepository struct{ db *gorm.DB }

func NewSessionRepository(db *gorm.DB) *SessionRepository { return &SessionRepository{db: db} }

func (r *SessionRepository) CreateSession(session *cuentadomain.SesionUsuario) error {
	return r.db.Create(session).Error
}

func (r *SessionRepository) FindSessionByHash(hash string) (*cuentadomain.SesionUsuario, *cuentadomain.Usuario, error) {
	var session cuentadomain.SesionUsuario
	if err := r.db.Where("hash_token = ?", hash).First(&session).Error; err != nil {
		return nil, nil, err
	}
	var user cuentadomain.Usuario
	if err := r.db.Where("id = ?", session.UsuarioID).First(&user).Error; err != nil {
		return nil, nil, err
	}
	return &session, &user, nil
}

func (r *SessionRepository) TouchSession(id uuid.UUID, activityAt time.Time) error {
	return r.db.Model(&cuentadomain.SesionUsuario{}).
		Where("id = ? AND revocado_en IS NULL", id).
		Update("ultima_actividad_en", activityAt).Error
}

func (r *SessionRepository) RevokeSession(hash string, revokedAt time.Time, reason string) error {
	return r.db.Model(&cuentadomain.SesionUsuario{}).
		Where("hash_token = ? AND revocado_en IS NULL", hash).
		Updates(map[string]interface{}{"revocado_en": revokedAt, "motivo_revocacion": reason}).Error
}

func (r *SessionRepository) RevokeAllSessions(userID uuid.UUID, revokedAt time.Time, reason string) error {
	return r.db.Model(&cuentadomain.SesionUsuario{}).
		Where("usuario_id = ? AND revocado_en IS NULL", userID).
		Updates(map[string]interface{}{"revocado_en": revokedAt, "motivo_revocacion": reason}).Error
}

func (r *SessionRepository) DeleteInactiveSessions(before time.Time) error {
	return r.db.Where("expira_en < ? OR (revocado_en IS NOT NULL AND revocado_en < ?)", before, before).
		Delete(&cuentadomain.SesionUsuario{}).Error
}
