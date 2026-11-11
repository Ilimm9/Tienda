package infrastructure

import (
	"fmt"
	"strings"
	"tienda/backend/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// El nombre de una categoría es único en el negocio (índice
// idx_categorias_negocio_nombre_ci). Aquí además se ignoran acentos y espacios
// dobles (catalogImportKey), igual que la vista previa del navegador, para
// responder con un mensaje claro en lugar del error de la base.

// validarCategoriaUnica rechaza un nombre ya usado por otra categoría del
// negocio; categoryID es la categoría que se edita (uuid.Nil al crear).
func validarCategoriaUnica(db *gorm.DB, businessID, categoryID uuid.UUID, nombre string) error {
	var categorias []domain.Categoria
	if err := db.Select("id, nombre").Where("negocio_id = ?", businessID).Find(&categorias).Error; err != nil {
		return err
	}
	key := catalogImportKey(nombre)
	for _, categoria := range categorias {
		if categoria.ID != categoryID && catalogImportKey(categoria.Nombre) == key {
			return domain.ErrCategoriaDuplicada
		}
	}
	return nil
}

// categoryTree es el árbol de categorías del negocio en memoria durante una
// importación; crea dentro de la transacción y recuerda lo creado.
type categoryTree struct {
	tx         *gorm.DB
	businessID uuid.UUID
	parentOf   map[uuid.UUID]uuid.UUID // uuid.Nil = raíz
	nameOf     map[uuid.UUID]string
	byKey      map[string]uuid.UUID // nombre normalizado -> categoría
	creadas    map[uuid.UUID]bool   // creadas como renglón en esta importación
	ancestros  map[uuid.UUID]bool   // creadas como ancestro de otro renglón
}

func loadCategoryTree(tx *gorm.DB, businessID uuid.UUID) (*categoryTree, error) {
	var current []domain.Categoria
	if err := tx.Select("id, nombre, categoria_padre_id").Where("negocio_id = ?", businessID).Order("creado_en ASC, id ASC").Find(&current).Error; err != nil {
		return nil, err
	}
	tree := &categoryTree{
		tx: tx, businessID: businessID,
		parentOf: map[uuid.UUID]uuid.UUID{}, nameOf: map[uuid.UUID]string{}, byKey: map[string]uuid.UUID{},
		creadas: map[uuid.UUID]bool{}, ancestros: map[uuid.UUID]bool{},
	}
	for _, category := range current {
		parent := uuid.Nil
		if category.CategoriaPadreID != nil {
			parent = *category.CategoriaPadreID
		}
		tree.register(category.ID, parent, category.Nombre)
	}
	return tree, nil
}

func (t *categoryTree) register(id, parent uuid.UUID, nombre string) {
	t.parentOf[id] = parent
	t.nameOf[id] = nombre
	if _, exists := t.byKey[catalogImportKey(nombre)]; !exists {
		t.byKey[catalogImportKey(nombre)] = id
	}
}

func (t *categoryTree) find(nombre string) (uuid.UUID, bool) {
	id, ok := t.byKey[catalogImportKey(nombre)]
	return id, ok
}

func (t *categoryTree) create(parent uuid.UUID, nombre, descripcion string) (uuid.UUID, error) {
	category := domain.Categoria{ID: uuid.New(), NegocioID: t.businessID, Nombre: strings.TrimSpace(nombre), Activo: true}
	if parent != uuid.Nil {
		parentID := parent
		category.CategoriaPadreID = &parentID
	}
	if descripcion = strings.TrimSpace(descripcion); descripcion != "" {
		category.Descripcion = &descripcion
	}
	if err := t.tx.Create(&category).Error; err != nil {
		return uuid.Nil, err
	}
	t.register(category.ID, parent, category.Nombre)
	return category.ID, nil
}

// path devuelve «Abuela › Madre › Hija» para mensajes.
func (t *categoryTree) path(id uuid.UUID) string {
	names := []string{}
	for current := id; current != uuid.Nil && len(names) < 20; current = t.parentOf[current] {
		names = append([]string{t.nameOf[current]}, names...)
	}
	return strings.Join(names, " › ")
}

// matchPath busca la categoría cuyo nombre y ancestros coinciden con `path`
// («Madre» o «Abuela > Madre»), en cualquier nivel.
func (t *categoryTree) matchPath(path []string) (uuid.UUID, bool) {
	id, ok := t.find(path[len(path)-1])
	current := id
	for i := len(path) - 2; i >= 0 && ok; i-- {
		parent := t.parentOf[current]
		ok = parent != uuid.Nil && catalogImportKey(t.nameOf[parent]) == catalogImportKey(path[i])
		current = parent
	}
	return id, ok
}

// ImportCategoryPaths crea categorías de una lista pegada. Cada renglón trae su
// ruta («Bebidas > Frías > Agua»): cada nombre va dentro del anterior y los
// ancestros que falten se crean en la misma transacción. Se omite lo que ya
// existe o se repite en la lista; un ancestro que ya existe en otro lugar
// invalida el renglón.
func (r *ProductRepository) ImportCategoryPaths(businessID uuid.UUID, rows []domain.CatalogImportCategoryRow) (domain.CatalogImportResult, error) {
	result := domain.CatalogImportResult{Errores: make([]domain.CatalogImportIssue, 0)}
	if len(rows) == 0 {
		return result, nil
	}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		tree, err := loadCategoryTree(tx, businessID)
		if err != nil {
			return err
		}
	renglones:
		for _, row := range rows {
			nombre := row.Ruta[len(row.Ruta)-1]
			ancestros := row.Ruta[:len(row.Ruta)-1]

			// Validar toda la ruta antes de crear, para no dejar ancestros sueltos.
			vistos := map[string]bool{}
			for i, segmento := range row.Ruta {
				key := catalogImportKey(segmento)
				if vistos[key] {
					result.Invalidas++
					result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Campo: "Nombre", Motivo: fmt.Sprintf("«%s» aparece dos veces en la ruta", segmento)})
					continue renglones
				}
				vistos[key] = true
				if i == 0 || i == len(row.Ruta)-1 {
					continue
				}
				id, exists := tree.find(segmento)
				if !exists {
					continue
				}
				padreEsperado, padreExiste := tree.find(row.Ruta[i-1])
				if !padreExiste || tree.parentOf[id] != padreEsperado {
					result.Invalidas++
					result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Campo: "Nombre", Motivo: fmt.Sprintf("«%s» ya existe como «%s»", tree.nameOf[id], tree.path(id))})
					continue renglones
				}
			}

			if id, exists := tree.find(nombre); exists {
				motivo := fmt.Sprintf("La categoría ya existe como «%s»", tree.path(id))
				if tree.creadas[id] || tree.ancestros[id] {
					motivo = "La categoría está repetida en la lista"
				}
				result.Omitidas++
				result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Motivo: motivo})
				continue
			}

			parent := uuid.Nil
			for _, ancestro := range ancestros {
				id, exists := tree.find(ancestro)
				if !exists {
					if id, err = tree.create(parent, ancestro, ""); err != nil {
						return err
					}
					tree.ancestros[id] = true
					result.Creadas++
				}
				parent = id
			}
			id, err := tree.create(parent, nombre, row.Descripcion)
			if err != nil {
				return err
			}
			tree.creadas[id] = true
			result.Creadas++
		}
		return nil
	})
	if err != nil {
		return domain.CatalogImportResult{Errores: make([]domain.CatalogImportIssue, 0)}, err
	}
	return result, nil
}

// ImportCategories crea las categorías del XLSX. La categoría padre es un
// nombre o una ruta «A > B» (para confirmar dónde está), entre las existentes
// y las del mismo archivo, sin importar el orden de los renglones.
func (r *ProductRepository) ImportCategories(businessID uuid.UUID, rows []domain.CatalogImportCategoryRow) (domain.CatalogImportResult, error) {
	result := domain.CatalogImportResult{Errores: make([]domain.CatalogImportIssue, 0)}
	if len(rows) == 0 {
		return result, nil
	}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		tree, err := loadCategoryTree(tx, businessID)
		if err != nil {
			return err
		}
		pendientes := make([]domain.CatalogImportCategoryRow, len(rows))
		copy(pendientes, rows)
		for len(pendientes) > 0 {
			// Un renglón cuyo padre todavía se puede crear en este archivo espera
			// a la siguiente vuelta.
			porCrear := make(map[string]int, len(pendientes))
			for _, row := range pendientes {
				porCrear[catalogImportKey(row.Nombre)]++
			}
			siguientes := pendientes[:0:0]
			// Si un renglón espera, los de abajo con el mismo nombre también, para
			// que el primero del archivo sea el que se cree.
			esperan := map[string]bool{}
			porNombre := map[int]bool{} // filas que solo esperan a una de arriba con su nombre
			for _, row := range pendientes {
				key := catalogImportKey(row.Nombre)
				if esperan[key] {
					siguientes = append(siguientes, row)
					porNombre[row.Fila] = true
					continue
				}
				parent := uuid.Nil
				if row.CategoriaPadre != "" {
					path := strings.Split(row.CategoriaPadre, ">")
					padreKey := catalogImportKey(path[len(path)-1])
					if padreKey == key {
						result.Invalidas++
						result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Campo: "Categoría padre", Motivo: "no puede ser la misma categoría"})
						porCrear[key]--
						continue
					}
					if porCrear[padreKey] > 0 {
						siguientes = append(siguientes, row)
						esperan[key] = true
						continue
					}
					id, ok := tree.matchPath(path)
					if !ok {
						siguientes = append(siguientes, row)
						esperan[key] = true
						continue
					}
					parent = id
				}
				porCrear[key]--
				if id, exists := tree.find(row.Nombre); exists {
					motivo := fmt.Sprintf("La categoría ya existe como «%s»", tree.path(id))
					if tree.creadas[id] {
						motivo = "La categoría está repetida en el archivo"
					}
					result.Omitidas++
					result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Motivo: motivo})
					continue
				}
				id, err := tree.create(parent, row.Nombre, row.Descripcion)
				if err != nil {
					return err
				}
				tree.creadas[id] = true
				result.Creadas++
			}
			if len(siguientes) == len(pendientes) {
				// Nadie avanzó: los que esperan a su padre lo tienen inexistente o
				// circular. Los que solo esperaban a otro renglón siguen.
				restantes := siguientes[:0:0]
				for _, row := range siguientes {
					if porNombre[row.Fila] {
						restantes = append(restantes, row)
						continue
					}
					result.Invalidas++
					result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Campo: "Categoría padre", Motivo: "no existe o forma una jerarquía circular"})
				}
				siguientes = restantes
			}
			pendientes = siguientes
		}
		return nil
	})
	if err != nil {
		return domain.CatalogImportResult{Errores: make([]domain.CatalogImportIssue, 0)}, err
	}
	return result, nil
}
