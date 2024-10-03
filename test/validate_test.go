package test

import (
	"testing"

	"github.com/unstoppablego/framework/logs"
	"github.com/unstoppablego/framework/validation"
)

type ReqLogin struct {
	Passwd   string `validate:"string,6-64,not_empty"`
	UserName string `validate:"not_empty"`
	Phone    string `validate:"not_empty"`
	Email    string `validate:"email,not_empty"`
}

func TestValidate(t *testing.T) {
	var ReqLogina ReqLogin
	ReqLogina.Passwd = "eeeafga"
	logs.Error(validation.ValidateStruct(ReqLogina))
}
