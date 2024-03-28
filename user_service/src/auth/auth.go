package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"regexp"
	"time"
	"userservice/src/database"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	loginMinLen = 3
	loginMaxLen = 30

	passwordMinLen = 8
	passwordMaxLen = 100
)

var tokenTTL time.Duration = 30 * time.Minute

type AuthService struct {
	db     *database.DataBase
	secret *rsa.PrivateKey
}

func New(db *database.DataBase) *AuthService {
	jwtPrivate, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	return &AuthService{db: db, secret: jwtPrivate}
}

func (a *AuthService) ValidateLogin(login string) error {
	err := validation.Validate(login,
		validation.Required,
		validation.Length(loginMinLen, loginMaxLen),
		is.Alphanumeric,
	)
	if err != nil {
		return err
	}

	ok, err := a.db.UserExist(login)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("user with this login already exist")
	}

	return nil
}

func (a *AuthService) ValidatePassword(password string) error {
	err := validation.Validate(password,
		validation.Required,
		validation.Length(passwordMinLen, passwordMaxLen),
		validation.By(requiredLetters),
	)
	return err
}

func (a *AuthService) ValidateUserData(userData *database.UserData) error {
	if userData == nil {
		return fmt.Errorf("user data not found")
	}

	err := validation.Errors{
		"name":         validation.Validate(userData.Name, validation.Length(2, 20)),
		"surname":      validation.Validate(userData.Surname, validation.Length(2, 20)),
		"birth_day":    validation.Validate(userData.BirthDay, validation.By(dateFormat)),
		"mail":         validation.Validate(userData.Mail, is.Email),
		"phone_number": validation.Validate(userData.PhoneNumber, is.E164),
	}.Filter()

	return err
}

func dateFormat(value interface{}) error {
	date, _ := value.(string)
	if date == "" {
		return nil
	}
	_, err := time.Parse("02/01/2006", date)
	return err
}

func requiredLetters(value interface{}) error {
	s, _ := value.(string)

	if !regexp.MustCompile(`\d`).MatchString(s) {
		return fmt.Errorf("password must include at least one digit")
	}
	if !regexp.MustCompile(`[A-Z]`).MatchString(s) {
		return fmt.Errorf("password must include at least one uppercase letter")
	}
	if !regexp.MustCompile(`[a-z]`).MatchString(s) {
		return fmt.Errorf("password must include at least one lowercase letter")
	}
	return nil
}

func (a *AuthService) CreateUser(login, password string) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return a.db.CreateUser(login, passwordHash)
}

func (a *AuthService) CheckPassword(login, password string) error {
	passwordHash, err := a.db.GetPasswordHash(login)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword(passwordHash, []byte(password)); err != nil {
		return err
	}
	return nil
}

func (a *AuthService) CreateToken(login string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"login": login,
		"exp":   time.Now().Add(tokenTTL).Unix(),
	})
	tokenString, err := token.SignedString(a.secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *AuthService) CheckAuth(w http.ResponseWriter, r *http.Request) (login string, authorized bool) {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "No JWT token: %v", err)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Can not parse JWT cookie: %v", err)
		return
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		return &a.secret.PublicKey, nil
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Can not parse JWT cookie: %v", err)
		return
	}

	if !token.Valid {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "JWT token not valid")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Can not cast JWT token's claims")
		return
	}

	login = claims["login"].(string)
	ok, err = a.db.UserExist(login)
	if !ok || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Can not find user by JWT token")
		return
	}
	authorized = true
	return
}
