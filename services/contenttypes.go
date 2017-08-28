package services

import (
	"database/sql"

	"github.com/dennisbappert/fileharbor/common"
	"github.com/dennisbappert/fileharbor/models"
	"github.com/jmoiron/sqlx"
)

type ContentTypeService struct {
	Service
}

type ContentType struct {
	ID          string
	ParentID    *string
	Name        string
	Description string
	Group       string
	Sealed      bool
	Columns     []ColumnMapping
}

type ColumnMapping struct {
	ID       string
	Required bool
	Visible  bool
	Default  string
}

func NewContentTypeService(configuration *common.Configuration, database *sqlx.DB, services *ServiceContext) *ContentTypeService {
	service := &ContentTypeService{Service{database: database, configuration: configuration, ServiceContext: services}}
	return service
}

func (service *ContentTypeService) Exists(id string, collectionID string, tx *sqlx.Tx) (bool, error) {
	contentType := models.ContentTypeEntity{}
	service.log.Println("looking up contenttype", id)

	commit := false
	if tx == nil {
		tx = service.database.MustBegin()
		commit = true
	}

	err := tx.Get(&contentType, "SELECT id FROM contenttypes where id=$1 and collection_id=$2", id, collectionID)

	// TODO: thhis looks like bullshit, there should be a better way
	if err != nil && err == sql.ErrNoRows {
		service.log.Println("column is not existing")
		return false, nil
	} else if err != nil {
		return true, err
	}

	if commit == true {
		err := tx.Commit()

		if err != nil {
			service.log.Println("error while executing transaction", err)
			return false, err
		}
	} else {
		service.log.Println("transaction passed to function - skipping commit")
	}

	service.log.Println("contenttype is existing", contentType)
	return true, nil
}

func (service *ContentTypeService) Create(contentType *ContentType, collectionID string, tx *sqlx.Tx) error {
	service.log.Println("creating contentType", contentType)
	service.log.Println("in target collection", collectionID)

	// check if the collection exists
	if exists, err := service.CollectionService.Exists(collectionID, tx); err != nil {
		service.log.Println("unable to check if collection exists - aborting...", err)
		return err
	} else if !exists {
		service.log.Println("collection is not existing - aborting...", collectionID)
		return common.NewApplicationError("Collection is not existing", common.ErrNotFound)
	}

	// check if the column may already exists
	if exists, err := service.Exists(contentType.ID, collectionID, tx); err != nil {
		service.log.Println("unable to check if contentType exists - aborting...", err)
		return err
	} else if exists {
		service.log.Println("contentType is existing - aborting...", contentType.ID)
		return common.NewApplicationError("The contentType is already existing", common.ErrContentTypeAlreadyExists)
	}

	commit := false
	if tx == nil {
		tx = service.database.MustBegin()
		commit = true
	}

	_, err := tx.Exec("INSERT INTO contenttypes (id, collection_id, parent_id, name, description, \"group\", sealed) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		contentType.ID, collectionID, contentType.ParentID, contentType.Name, contentType.Description, contentType.Group, contentType.Sealed)

	if err != nil {
		service.log.Println("unexpected error while creating contentType", err)

		if commit {
			tx.Rollback()
		}

		return err
	}

	if commit == true {
		err := tx.Commit()

		if err != nil {
			service.log.Println("error while executing transaction", err)
			return err
		}
	} else {
		service.log.Println("transaction passed to function - skipping commit")
	}

	return nil
}

func (service *ContentTypeService) AddColumn(collectionID string, contentTypeId string, columnId string, visible bool, required bool, isDefault string, tx *sqlx.Tx) error {
	service.log.Println("adding column to contentType", contentTypeId)
	service.log.Println("with id", columnId)
	service.log.Println("in collection", collectionID)

	// check if the collection exists
	if exists, err := service.CollectionService.Exists(collectionID, tx); err != nil {
		service.log.Println("unable to check if collection exists - aborting...", err)
		return err
	} else if !exists {
		service.log.Println("collection is not existing - aborting...", collectionID)
		return common.NewApplicationError("Collection is not existing", common.ErrNotFound)
	}

	// check if the column may already exists
	if exists, err := service.Exists(contentTypeId, collectionID, tx); err != nil {
		service.log.Println("unable to check if contentType exists - aborting...", err)
		return err
	} else if !exists {
		service.log.Println("contentType is not existing - aborting...", contentTypeId)
		return common.NewApplicationError("The contentType is not existing", common.ErrContentTypeNotFound)
	}

	// TODO: handle existing mapping

	commit := false
	if tx == nil {
		tx = service.database.MustBegin()
		commit = true
	}

	_, err := tx.Exec("INSERT INTO contenttype_column_mappings (contenttype_id, column_id, collection_id, required, visible, \"default\") VALUES ($1, $2, $3, $4, $5, $6)",
		contentTypeId, columnId, collectionID, required, visible, isDefault)

	if err != nil {
		service.log.Println("unexpected error while adding column to content type", err)

		if commit {
			tx.Rollback()
		}

		return err
	}

	if commit == true {
		err := tx.Commit()

		if err != nil {
			service.log.Println("error while executing transaction", err)
			return err
		}
	} else {
		service.log.Println("transaction passed to function - skipping commit")
	}

	return nil
}
