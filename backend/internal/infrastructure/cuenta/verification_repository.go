package cuenta

import (
	"errors"
	"time"

	cuentadomain "tienda/backend/internal/domain/cuenta"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type VerificationRepository struct{ db *gorm.DB }

func NewVerificationRepository(db *gorm.DB) *VerificationRepository {
	return &VerificationRepository{db: db}
}

func (r *VerificationRepository) FindChallenge(id uuid.UUID) (*cuentadomain.DesafioAutenticacion, error) {
	var challenge cuentadomain.DesafioAutenticacion
	if err := r.db.First(&challenge, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (r *VerificationRepository) IssueChallenge(challenge *cuentadomain.DesafioAutenticacion) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return issueChallenge(tx, challenge)
	})
}

func (r *VerificationRepository) IssueChallengeForPendingAccount(user *cuentadomain.Usuario, profile *cuentadomain.PerfilUsuario, challenge *cuentadomain.DesafioAutenticacion) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var locked cuentadomain.Usuario
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&locked, "id = ?", user.ID).Error; err != nil {
			return err
		}
		if locked.Estado != "pendiente_verificacion" || locked.CorreoVerificadoEn != nil {
			return errors.New("usuario no disponible")
		}
		if err := tx.Model(&locked).Update("hash_contrasena", user.HashContrasena).Error; err != nil {
			return err
		}
		if err := tx.Model(&cuentadomain.PerfilUsuario{}).Where("usuario_id = ?", user.ID).
			Updates(map[string]any{"nombres": profile.Nombres, "apellidos": profile.Apellidos, "telefono": profile.Telefono}).Error; err != nil {
			return err
		}
		return issueChallenge(tx, challenge)
	})
}

func issueChallenge(tx *gorm.DB, challenge *cuentadomain.DesafioAutenticacion) error {
	if err := tx.Model(&cuentadomain.DesafioAutenticacion{}).
		Where("usuario_id = ? AND proposito = ? AND usado_en IS NULL", challenge.UsuarioID, challenge.Proposito).
		Update("usado_en", challenge.UltimoEnvioEn).Error; err != nil {
		return err
	}
	return tx.Create(challenge).Error
}

func (r *VerificationRepository) CountChallengesByUserSince(userID uuid.UUID, since time.Time) (int64, error) {
	var total int64
	err := r.db.Model(&cuentadomain.DesafioAutenticacion{}).
		Where("usuario_id = ? AND creado_en >= ?", userID, since).Count(&total).Error
	return total, err
}

func (r *VerificationRepository) CountChallengesByIPSince(ip string, since time.Time) (int64, error) {
	var total int64
	err := r.db.Model(&cuentadomain.DesafioAutenticacion{}).
		Where("direccion_ip = ? AND creado_en >= ?", ip, since).Count(&total).Error
	return total, err
}

func (r *VerificationRepository) LastChallengeByUser(userID uuid.UUID) (*cuentadomain.DesafioAutenticacion, error) {
	var challenge cuentadomain.DesafioAutenticacion
	err := r.db.Where("usuario_id = ? AND proposito = ?", userID, cuentadomain.PropositoVerificacionCorreo).
		Order("ultimo_envio_en DESC").First(&challenge).Error
	if err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (r *VerificationRepository) IncrementFailedAttempts(id uuid.UUID, maxAttempts int) error {
	result := r.db.Model(&cuentadomain.DesafioAutenticacion{}).
		Where("id = ? AND usado_en IS NULL AND intentos_fallidos < ?", id, maxAttempts).
		UpdateColumn("intentos_fallidos", gorm.Expr("intentos_fallidos + 1"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *VerificationRepository) ActivateUserWithChallenge(challengeID, userID uuid.UUID, now time.Time, maxAttempts int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var challenge cuentadomain.DesafioAutenticacion
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&challenge, "id = ?", challengeID).Error; err != nil {
			return err
		}
		if challenge.UsuarioID != userID || challenge.Proposito != cuentadomain.PropositoVerificacionCorreo || challenge.UsadoEn != nil || !challenge.ExpiraEn.After(now) || challenge.IntentosFallidos >= maxAttempts {
			return errors.New("desafío inválido")
		}
		result := tx.Model(&cuentadomain.DesafioAutenticacion{}).
			Where("id = ? AND usado_en IS NULL", challengeID).Update("usado_en", now)
		if result.Error != nil || result.RowsAffected != 1 {
			if result.Error != nil {
				return result.Error
			}
			return errors.New("desafío ya consumido")
		}
		result = tx.Model(&cuentadomain.Usuario{}).
			Where("id = ? AND estado = ? AND correo_verificado_en IS NULL", userID, "pendiente_verificacion").
			Updates(map[string]any{"estado": "activo", "correo_verificado_en": now})
		if result.Error != nil || result.RowsAffected != 1 {
			if result.Error != nil {
				return result.Error
			}
			return errors.New("usuario no disponible")
		}
		return tx.Model(&cuentadomain.DesafioAutenticacion{}).
			Where("usuario_id = ? AND proposito = ? AND usado_en IS NULL", userID, challenge.Proposito).
			Update("usado_en", now).Error
	})
}
