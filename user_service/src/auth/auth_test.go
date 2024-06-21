package auth

import (
	"testing"
	"userservice/src/database"
)

func TestValidatePassword(t *testing.T) {
	as := &AuthService{}

	t.Run("Good password", func(t *testing.T) {
		if err := as.ValidatePassword("GoodPassword1337"); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("Short password", func(t *testing.T) {
		if err := as.ValidatePassword("T1ny"); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("No digits", func(t *testing.T) {
		if err := as.ValidatePassword("NoDigitsPassword"); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("No uppercases", func(t *testing.T) {
		if err := as.ValidatePassword("lower_case_password123"); err == nil {
			t.Error("expected error")
		}
	})
}

func TestValidateUserData(t *testing.T) {
	as := &AuthService{}

	t.Run("Good data", func(t *testing.T) {
		if err := as.ValidateUserData(&database.UserData{
			Name:        "Name",
			Surname:     "Surname",
			BirthDay:    "02/01/2006",
			Mail:        "mail@mail.com",
			PhoneNumber: "88005553535",
		}); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("Bad name", func(t *testing.T) {
		if err := as.ValidateUserData(&database.UserData{
			Name: "x",
		}); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Bad surname", func(t *testing.T) {
		if err := as.ValidateUserData(&database.UserData{
			Surname: "y",
		}); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Bad birth day", func(t *testing.T) {
		if err := as.ValidateUserData(&database.UserData{
			BirthDay: "32/33/2222",
		}); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Bad mail", func(t *testing.T) {
		if err := as.ValidateUserData(&database.UserData{
			Mail: "aaaa@bbbb23",
		}); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Bad phone number", func(t *testing.T) {
		if err := as.ValidateUserData(&database.UserData{
			PhoneNumber: "0",
		}); err == nil {
			t.Error("expected error")
		}
	})
}
