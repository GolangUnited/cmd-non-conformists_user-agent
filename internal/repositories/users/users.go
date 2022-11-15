package users

import (
	"log"
	"strings"
	"time"

	"github.com/golang-unitied-school/useragent/internal/models"
	global "github.com/golang-unitied-school/useragent/internal/pkg/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PGSQL struct {
	dbConn *gorm.DB
}

var User models.User

func (ptr *PGSQL) Init(connectionString string) {
	var err error

	ptr.dbConn, err = gorm.Open(postgres.Open(connectionString), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("error while starting connection: %s", err.Error())
	}

	log.Println("trying to migrate schema...")
	if err = ptr.dbConn.AutoMigrate(&User); err != nil {
		log.Fatalf("error while migrating: %s", err.Error())
	}
}

func (ptr *PGSQL) Create(fname, sname, email, pass, role string) (string, error) {

	hash, err := global.EncodingPassword(pass)
	if err != nil {
		return "", err
	}

	user := models.User{
		Name:      fname,
		Surname:   sname,
		Email:     email,
		Password:  hash,
		Role:      role,
		CreatedAt: time.Now(),
	}

	newRow := ptr.dbConn.Create(&user)

	if newRow.Error != nil {
		return "", newRow.Error
	}

	return user.Id.String(), nil
}
func (ptr *PGSQL) Update(uuid, fname, sname, email, role string) error {

	var row models.User

	res := ptr.dbConn.Model(&User).Where("id = ? and is_deleted = 0", uuid).First(&row)
	if res.Error != nil {
		return res.Error
	}

	changes := 0

	if !strings.EqualFold(row.Name, fname) && fname != "" {
		row.Name = fname
		changes++
	}

	if !strings.EqualFold(row.Surname, sname) && sname != "" {
		row.Surname = sname
		changes++
	}

	if !strings.EqualFold(row.Email, email) && email != "" {
		row.Email = email
		changes++
	}

	if !strings.EqualFold(row.Role, role) && role != "" {
		row.Role = role
		changes++
	}

	if changes == 0 {
		return global.ErrorNoNewData
	}

	res = ptr.dbConn.Model(&row).Updates(&row)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (ptr *PGSQL) Delete(userId string) error {
	var row models.User
	row.IsDeleted = 1

	res := ptr.dbConn.Model(&User).Where("id = ?", userId).Updates(&row)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (ptr *PGSQL) GetById(userId string) (models.User, error) {
	var row models.User
	res := ptr.dbConn.Model(&User).Where("id = ? and is_deleted = 0", userId).First(&row)
	if res.Error != nil {
		if res.Error.Error() == global.ErrorRecordNotFound.Error() {
			return row, global.ErrorUserNotFound
		} else {
			return row, res.Error

		}
	}

	return row, nil
}
func (ptr *PGSQL) GetByEmail(email string) (models.User, error) {
	var row models.User
	res := ptr.dbConn.Model(&User).Where("email = ?", email).First(&row)

	if res.Error != nil {
		if res.Error.Error() == global.ErrorRecordNotFound.Error() {
			return row, global.ErrorUserNotFound
		} else {
			return row, res.Error

		}
	}
	return row, nil
}

func (ptr *PGSQL) GetPassword(userId string) (string, error) {
	var row models.User
	res := ptr.dbConn.Model(&User).Where("id = ? and is_deleted = 0", userId).First(&row)
	if res.Error != nil {
		return "", res.Error
	}

	return row.Password, nil
}

func (ptr *PGSQL) SetPassword(userId, newPass string) error {
	var row models.User

	hash, err := global.EncodingPassword(newPass)
	if err != nil {
		return err
	}

	res := ptr.dbConn.Model(&User).Where("id = ? and is_deleted = 0", userId).First(&row).UpdateColumn("Password", hash)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (ptr *PGSQL) Close() error {
	db, err := ptr.dbConn.DB()
	if err != nil {
		log.Printf("error while getting db object: %s", err.Error())
		return err
	}
	err = db.Close()
	if err != nil {
		log.Printf("error while disconnecting db: %s", err.Error())
		return err
	}
	return nil
}
