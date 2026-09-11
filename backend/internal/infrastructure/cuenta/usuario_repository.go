package cuenta

import (
	"gorm.io/gorm"
	cuentadomain "tienda/backend/internal/domain/cuenta"
)

type UserRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }
func (r *UserRepository) FindByEmail(email string) (*cuentadomain.Usuario, error) {
	var user cuentadomain.Usuario
	if err := r.db.Where("correo = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
func (r *UserRepository) Save(user *cuentadomain.Usuario) error { return r.db.Save(user).Error }

func (r *UserRepository) CreateAccount(user *cuentadomain.Usuario, profile *cuentadomain.PerfilUsuario) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		profile.UsuarioID = user.ID
		return tx.Create(profile).Error
	})
}
