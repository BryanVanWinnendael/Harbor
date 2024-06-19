package services

import (
	"github.com/BryanVanWinnendael/Harbor/db"
	"golang.org/x/crypto/bcrypt"
)

func NewUserServices(u User, uStore db.Store) *UserServices {

	return &UserServices{
		User:      u,
		UserStore: uStore,
	}
}

type User struct {
	ID              int    `json:"id"`
	Password        string `json:"password"`
	Username        string `json:"username"`
	ChangedPassword bool   `json:"changed_password"`
}

type UserServices struct {
	User      User
	UserStore db.Store
}

func (us *UserServices) CheckUsername(username string) (User, error) {

	query := `SELECT id, password, username, changed_password FROM users
		WHERE username = ?`

	stmt, err := us.UserStore.Db.Prepare(query)
	if err != nil {
		return User{}, err
	}

	defer stmt.Close()

	us.User.Username = username
	err = stmt.QueryRow(
		us.User.Username,
	).Scan(
		&us.User.ID,
		&us.User.Password,
		&us.User.Username,
		&us.User.ChangedPassword,
	)

	if err != nil {
		return User{}, err
	}

	return us.User, nil
}

func (us *UserServices) ChangePassword(username, password string) error {

	query := `UPDATE users
		SET password = ?, changed_password = ?
		WHERE username = ?`

	stmt, err := us.UserStore.Db.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(
		hashedPassword,
		true,
		username,
	)

	if err != nil {
		return err
	}

	return nil
}
