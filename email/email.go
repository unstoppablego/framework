package email

import (
	"fmt"
	"net/url"

	hermes "github.com/matcornic/hermes/v2"
	"github.com/unstoppablego/framework/config"
)

// 用户登录名称 mailout@5761912992379058.onaliyun.com
// AccessKey ID LTAI4GGJv3t3T1RVDmqWgRN2
// AccessKey Secret rwN9Jcc3UCjiCcQVA7xz4BCbtpmd6h
// CreateConfrimEmailContent 创建 认证EMAIL 发送内容
func CreateConfrimEmailContent(lang, websiteName, authTokenURL string) string {

	urla, err := url.Parse(authTokenURL)
	// fmt.Println(urla, err)
	website := urla.Scheme + "://" + urla.Host

	if lang == "zh-CN" {
		h := hermes.Hermes{
			// Optional Theme
			// Theme: new(Default)
			Product: hermes.Product{
				// Appears in header & footer of e-mails
				Name:      "" + websiteName + "",
				Link:      website,
				Copyright: "Copyright © 2019 " + websiteName + ". All rights reserved.",
				// Optional product logo
				// Logo: "http://www.duchess-france.org/wp-content/uploads/2016/01/gopher.png",
			},
		}
		email := hermes.Email{
			Body: hermes.Body{
				Name: "Jon Snow",
				Intros: []string{
					"Welcome to " + websiteName + "! We're very excited to have you on board.",
				},
				Actions: []hermes.Action{
					{
						Instructions: "To get started with " + websiteName + ", please click here:",
						Button: hermes.Button{
							Color: "#22BC66", // Optional action button color
							Text:  "Confirm your account",
							Link:  authTokenURL,
						},
					},
				},
				Outros: []string{
					"Need help, or have questions? Just reply to this email, we'd love to help.",
				},
			},
		}

		// Generate an HTML email with the provided contents (for modern clients)
		emailBody, err := h.GenerateHTML(email)
		if err != nil {
			panic(err) // Tip: Handle error with something else than a panic ;)
		}

		return emailBody
	}

	h := hermes.Hermes{
		// Optional Theme
		// Theme: new(Default)
		Product: hermes.Product{
			// Appears in header & footer of e-mails
			Name:      "" + websiteName + "",
			Link:      website,
			Copyright: "Copyright © 2019 " + websiteName + ". All rights reserved.",
			// Optional product logo
			// Logo: "http://www.duchess-france.org/wp-content/uploads/2016/01/gopher.png",
		},
	}
	email := hermes.Email{
		Body: hermes.Body{
			Name: "Jon Snow",
			Intros: []string{
				"Welcome to " + websiteName + "! We're very excited to have you on board.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "To get started with " + websiteName + ", please click here:",
					Button: hermes.Button{
						Color: "#22BC66", // Optional action button color
						Text:  "Confirm your account",
						Link:  authTokenURL,
					},
				},
			},
			Outros: []string{
				"Need help, or have questions? Just reply to this email, we'd love to help.",
			},
		},
	}

	// Generate an HTML email with the provided contents (for modern clients)
	emailBody, err := h.GenerateHTML(email)
	if err != nil {
		panic(err) // Tip: Handle error with something else than a panic ;)
	}

	return emailBody

}

const baseURL = "https://dm.ap-southeast-1.aliyuncs.com/"

// 发送 邮件认证
func SendTokenURLEmail(lang, ToAddress string, tokenURL string) {
	//dm.ap-southeast-1.aliyuncs.com
	//https://dm.aliyuncs.com/?Format=xml
	// &Version=2015-11-23
	// &Signature=Pc5WB8gokVn0xfeu%2FZV%2BiNM1dgI%3D
	// &SignatureMethod=HMAC-SHA1
	// &SignatureNonce=e1b44502-6d13-4433-9493-69eeb068e955
	// &SignatureVersion=1.0
	// &AccessKeyId=key-test
	// &Timestamp=2015-11-23T12:00:00Z
	req := MailRequest{
		Action:         "SingleSendMail",
		AccountName:    "",
		ReplyToAddress: true,
		ToAddress:      "",
		FromAlias:      "",
		Subject:        "",
		HtmlBody:       CreateConfrimEmailContent(lang, "", tokenURL),
	}

	accessKeyID := config.Cfg.Aliyunemail.AccessKeyID
	accessKeySecret := config.Cfg.Aliyunemail.AccessKeySecret
	c := NewClient(accessKeyID, accessKeySecret)
	body, _ := c.SendRequest(&req)
	fmt.Println(body)
}
