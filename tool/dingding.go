package tool

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/unstoppablego/framework/logs"
)

func NewRobot(token, secret string) *Robot {
	return &Robot{
		token:  token,
		secret: secret,
	}
}

func sign(t int64, secret string) string {
	strToHash := fmt.Sprintf("%d\n%s", t, secret)
	hmac256 := hmac.New(sha256.New, []byte(secret))
	hmac256.Write([]byte(strToHash))
	data := hmac256.Sum(nil)
	return base64.StdEncoding.EncodeToString(data)
}

type Robot struct {
	token, secret string
}

func (robot *Robot) SendMessage(msg interface{}) error {
	body := bytes.NewBuffer(nil)
	err := json.NewEncoder(body).Encode(msg)
	if err != nil {
		return fmt.Errorf("msg json failed, msg: %v, err: %v", msg, err.Error())
	}

	value := url.Values{}
	value.Set("access_token", robot.token)
	if robot.secret != "" {
		t := time.Now().UnixNano() / 1e6
		value.Set("timestamp", fmt.Sprintf("%d", t))
		value.Set("sign", sign(t, robot.secret))
	}

	request, err := http.NewRequest(http.MethodPost, "https://oapi.dingtalk.com/robot/send", body)
	if err != nil {
		return fmt.Errorf("error request: %v", err.Error())
	}
	request.URL.RawQuery = value.Encode()
	request.Header.Add("Content-Type", "application/json;charset=utf-8")
	res, err := (&http.Client{}).Do(request)
	if err != nil {
		return fmt.Errorf("send dingTalk message failed, error: %v", err.Error())
	}
	defer func() { _ = res.Body.Close() }()
	result, err := ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		return fmt.Errorf("send dingTalk message failed, %s", httpError(request, res, result, "http code is not 200"))
	}
	if err != nil {
		return fmt.Errorf("send dingTalk message failed, %s", httpError(request, res, result, err.Error()))
	}

	type response struct {
		ErrCode int `json:"errcode"`
	}
	var ret response

	if err := json.Unmarshal(result, &ret); err != nil {
		return fmt.Errorf("send dingTalk message failed, %s", httpError(request, res, result, err.Error()))
	}

	if ret.ErrCode != 0 {
		return fmt.Errorf("send dingTalk message failed, %s", httpError(request, res, result, "errcode is not 0"))
	}

	return nil
}

func httpError(request *http.Request, response *http.Response, body []byte, error string) string {
	return fmt.Sprintf(
		"http request failure, error: %s, status code: %d, %s %s, body:\n%s",
		error,
		response.StatusCode,
		request.Method,
		request.URL.String(),
		string(body),
	)
}

func (robot *Robot) SendTextMessage(content string, atMobiles []string, isAtAll bool) error {
	msg := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": content,
		},
		"at": map[string]interface{}{
			"atMobiles": atMobiles,
			"isAtAll":   isAtAll,
		},
	}

	return robot.SendMessage(msg)
}

func (robot *Robot) SendMarkdownMessage(title string, text string, atMobiles []string, isAtAll bool) error {
	msg := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": title,
			"text":  text,
		},
		"at": map[string]interface{}{
			"atMobiles": atMobiles,
			"isAtAll":   isAtAll,
		},
	}

	return robot.SendMessage(msg)
}

func (robot *Robot) SendLinkMessage(title string, text string, messageUrl string, picUrl string) error {
	msg := map[string]interface{}{
		"msgtype": "link",
		"link": map[string]string{
			"title":      title,
			"text":       text,
			"messageUrl": messageUrl,
			"picUrl":     picUrl,
		},
	}

	return robot.SendMessage(msg)
}

// 使用多个token 发送消息
func SendMessage(text string) {
	// if config.NoSendMessage {
	// 	return
	// }
	// config.IsSuccess = true
	var dt = []DinToke{
		{Token: "57dfb989a48076e46bdca28e60074a803f8996cdc691bd804345db3c9f14c752", Sec: "SECfd22d13b0e1e952cc85b3a27da5499c3ff8cb070bf42648b6c412f2777b34745"},
		{Token: "418bb31ce086b86b5e0bd84d39703131087a4a6497242d86ef8fb29daa0bcd1f", Sec: "SEC37c1208797c78a59a8eaa9918012f2214c37b544c5efd61082782581a5c41d88"},
		{Token: "39b1b6e9543ae118f7a7506b267a600d432de2d3dc9bc577d50a17a27d5237bb", Sec: "SEC499c53363429efefcd6097376bfac936437f0643de45ae5900728474f538136a"},
		{Token: "c2a53d4bfc0d541eaf130e854f4b182625e0a3547e4fc7d64fd0efc3f1f0871a", Sec: "SEC7aa4c17932ac819952e622a30b895ff07b9a4db8d0b9c49ae212027ebfb5f2f1"},
		{Token: "312f97f25a514141867614244e7631de59f1198d5eecdd8ab21baeb2f2f4e3f8", Sec: "SEC8f20df0b25a8c6ae4b789e69cf4a223f827779b7519206436e8bab5ba83e159c"},
		{Token: "3b92cf8e9d8381431ed8d42b0b2e62af9214d34a25f00fa4373bc35ce33710b3", Sec: "SECc9d3305809d864eba1e2bf28bcfb445126452d0e5a81854c4af63d3133138754"},
		{Token: "6c13d1fc6ac588bfd202c2b82d33064a2f831bcfff22f4e6861c8355691a7d29", Sec: "SEC4907b6ae1e576638e29f15e40b369280f8d4d023ecdc60583c8b3cefdddfffc6"},
		{Token: "e8409af427a5223928cf4edefe16f290fdd0ef533567370d041f7046edc0c9ae", Sec: "SECbdcd96ec873183a9a799cb40d2e3dbdfbccc24c9b7dd1bfc070f4ecec2d1af1d"},
		{Token: "dc404e32e07a6d28ef24220b57bf8932e582684b8ba23a85f3e35ee4882d1e53", Sec: "SEC74513dca98edcd16c6e837a49c0c60d1057e6ed037045e08914d649245ffe66d"},
		{Token: "e161636c9c1898aedd8e34e8f09787067f0db7393566f041be119a8fa3283512", Sec: "SEC59b0add62178d9327a8f325fa6581da960417c451e2db0a9429379709c72e77f"},
	}

	for _, v := range dt {
		rb := NewRobot(v.Token, v.Sec)
		logs.Info(text)
		err := rb.SendTextMessage(text, nil, false)
		if err != nil {
			fmt.Println(err)
			continue
		}
		break
	}
}

type DinToke struct {
	Token string
	Sec   string
	Used  bool
}

//step 将cookie 弄过来
//填充接口数据

type DingMsg struct {
	Msgtype string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}

type SendMessageManager struct {
	SystemAutoPayToken   []DinToke //余额自动支付
	SystemAlipayPayToken []DinToke
	UserToken            []DinToke
}

var SendMessageManagerG SendMessageManager

func init() {
	SendMessageManagerG.SystemAutoPayToken = []DinToke{
		{Token: "a91d6cf9b105502ea2ed44f2c160b298412a305c5f2efa86a48584e9d2982e0b", Sec: "SEC7b5d3a703ec033c5e03a43f42d91513b0adc6c3b822a18bae0f2c7f3df94421e"},
		{Token: "db424ec74f269bfd82e765d996e3630e86df42bc1b70714f4267ed2c4b4ac064", Sec: "SECcf0450b49f7f11bd8100bd16498632ee88c28e0bbf8a745380cfff03906d6dbe"},
		{Token: "df5dcbccc4061f1443204ffee0eab2f72a6ca3fb6ebd13066367b2055597d9ba", Sec: "SEC489ff3f3813c651a14ddae50d1c583ec84024bab35098d40fdecaa923984eb09"},
		{Token: "f4510b010da0370ef9ab39d413fc72fbee5393e3994b56c2a736d486f42424e7", Sec: "SEC02c108caf0c9529d3b10e3278e1179e8426b1b1b2c6b93b631b5a25c2a9490e8"},
		{Token: "de610e0ba9ea8d53f51edaaa971ec7cedaea0e1ebecb5ff42b66685fd780cee7", Sec: "SEC05bd7034b6f5968257415de26c14893bd5e0e1e2dda074e263746a6e8969797e"},
		{Token: "e17ae628b3d2b7b4223d1260edcb4f54fe60ec36001b912e0c88957e211bee12", Sec: "SEC58c8267951574330eea06cba5cc2395ad14eb492b364cb1d86275388623f99e4"},
		{Token: "429a8d891bbae22ce61bcba32863276fa3057fc7901cd75ab8b7b7d609a1b8ea", Sec: "SEC4483580e94a0661fe8ae8e69c80a0643f0c28048206823329a863d7972303795"},
		{Token: "c30b591fa1f934e3cbb6d39674a704f7e29b336ae3ee9f2816f3bd59410600ae", Sec: "SEC354e8000c952ab32e284fdbeea3dd1ca5d8c968e0ddf30965e93b3b120b34729"},
		{Token: "8def999f4c381705979e4741672e704da02fa3caaba18c7ba16aeeece436f76f", Sec: "SECba0d9437484a8e0999d9b08194a0b5a2107c7b96ca1f69b0a6820a53f396296f"},
		{Token: "7730a818f6cb5e29680f48e4bcc54336971a0c92b89746c02d0a37bc201edc42", Sec: "SEC2f0e9db349c6014aa6734429619b5004507a09c6460a038779397990c88ab3e3"},
		{Token: "411bc2ad0be609c3c252609d0b7a42617cbb5856986cc384b125979bb5191a2a", Sec: "SECbb5fa883c2dd9bd78933ae58ad9cd2d177e9cbf864880bf38e4acb5bc79b51b1"},
		{Token: "f1e47f9033f3a7b3ecaea61e627d1f281c9a719bbca00218bd295afa568f077c", Sec: "SEC2bb48f4e418b4fc445fad6d1ff343bf51f7d01d09b90c4cde4a24525b812eb46"},
		{Token: "fca111c7dcc623984d3ac707cdc324a9869eb9439572875b349f4134745290e1", Sec: "SECc9f15387b683063e72e0e20f2afa362077884eea5e092b681c68e6d3e14e3988"},
		{Token: "180a63a7618a0d29d21caf8e84f2324e895f4a5b648e31e2e04ccbbc1d677c24", Sec: "SEC6010f4c8a80efc10a8aa6700fef4661fd6eb4e628015aac3d260e8aa35f5b99f"},
		{Token: "34055e2dcda7cef2e7521e8d36502c73cca7902cb6077c889be205941016f953", Sec: "SECefbbf821fc705486458491fbc660fbb547afc056a096fc983fbfd59564325370"},
		{Token: "6745cc32232476702539507dd0802a342f3619dfbea22b22fc88fe705c48ce66", Sec: "SEC730b93acd5b26ee537fabe278a42691807f8c76d02d78ec5cd558bacb01305ad"},
		{Token: "6e60ea828c99275e2c2bbe8e471c007b92297c202d7f9208c079ee84c51ec4e7", Sec: "SEC79a46475626dd63f4ee647990d5e2c9e592cd2e9c629d633495f5429ca021b44"},
		{Token: "f3c85d35c7705a9ed5a0fbc82f31059c5befb4741e811aa501a94776737bc3c4", Sec: "SEC6c66e8fb331f266ebd4824688b86eebdbd66b5c0ac3e89b9296c2ab9baddb792"},
		{Token: "5c469a4f56beeee5d9e18e95f0be645cc1c77eeeb0d11ae7d2a87635536f02e8", Sec: "SEC2f9ac28cfcf09bc043d3c209d720757e966fd358df60c9f5cd3e0462ffd05632"},
		{Token: "dfe7d5079ef2c62f8e1e6d12a41beaf4866e17a4a569da41e63e8f88fafa2f9e", Sec: "SECcb537cac7aa0c109922ea4b9674ee15317c28cfaecf0606f1cbca6768d96c140"},
		{Token: "35b88e7790b666f7fe6c84d794e3f76037c7cb70d336c153bdbcf32716c90205", Sec: "SEC072eff096789d0141a41d2239ac3ab617c32b037748ccece8e229a8c6dd7e007"},
		{Token: "2bd5d89513d2afa5dd0839c56abe21129ee86a4be6a1a33bed0cf2a8f1bc1287", Sec: "SECe7e88da4ef57a96baf03dcc5c7f4b3eb127fd85b6bfc55f2f290c8203513fd21"},
		{Token: "be669a2253d687b60a81a9410e9e5867f85cc989b0d3e3ae54de280dca390242", Sec: "SEC0cee37b28322ff1abeadcc2ae2d4e72490b5cb9dc9ef473178a834ac03e4a691"},
		{Token: "610a39069fa5a08cb3f8b5081078dfe310910cb62b40713feb7ff62a7cf3b143", Sec: "SECa43deecbbfa01f215dbbb7983ab84039138004e359c479e73666705c7db4ed52"},
		{Token: "b3fb3de37484c6dc6312e7878c5ec63c2f585376efce3c846575d83dc9ab9323", Sec: "SECb66afe8d76af735aede2e9b74fa09dec53998773093229bb85f883ae8d36c7d1"},
		{Token: "1ae0cb237d6550f0b6b6ecabb5c4ae55a7115571a117bb261abea7f8e7c09849", Sec: "SECab9394d39c7fe22c8967785a34a447962c7b236fed0a0c91eed485675b9a3244"},
		{Token: "1e9966a2f90364debb4fef3a71769fb10d1a20dcf56686df280ec295165eed07", Sec: "SEC76a6d61b7d2e019592358cf65795a6f0b3e403715dc6193fe1a1551e72c7b934"},
	}

	SendMessageManagerG.SystemAlipayPayToken = []DinToke{
		{Token: "aae3662d129defd8eba45ab1818be7e66cfd1539e93722cd2db6584f02b647c3", Sec: "SEC28dcdcde805f48207905aa349cf0da85691762dd3fafd487e9d7271eb36734f7"},
		{Token: "a61a5286b62adb6a28875bcd1a5b46d44630d45be0640ff42cada73453ac5319", Sec: "SECd55154a90c30eb27b8a7f30f9f23455eeb0cd0f27f72435240d61fa72a86420a"},
		{Token: "6fb38a57c5d41178dc86463a5ea6f219bfd90be0c0b13a2c241fa2cea2f475f1", Sec: "SEC8028d474781f92d83bfca27af9c01ddda80361891e60276a3e676222b518fdf8"},
		{Token: "6cd3ed2fe00b0712deca20a45f4d63449765548f3b770c61231cfaa7ef249312", Sec: "SEC28dc31a7af4ec6847a8f7b6751a8d9dfc8da2c38af7be893aa3aea87a6daec15"},
		{Token: "ad891986b90a2e749e6784c1f134b18faaba3c765c5180516cc83fa894fb75c6", Sec: "SECbe7765a274671107747d0854c51bba9dc528a7502f49691f77cfae20f91af247"},
		{Token: "6a79f895083a84443f0d326cf227c21a0094f626bb76262c82f1e78e053e91ce", Sec: "SECcb21c95bdf6d27ad6a39d388192591454f53786c50072bc52d87c652e4071ac9"},
		{Token: "e3fa8e5a1a02f583f7404e1b7ee7cef9b8f41fb241ebfa357c31480c1fb5ba34", Sec: "SEC2c84f58667056e4f59b1b973b6fb8a2bcfab8b00a981ec5bbd068f56dcfd7094"},
		{Token: "f1cf7b3d85c867162c17a15335fdc17bff25c91ee7f3085fb3c3ed81113ae6c6", Sec: "SEC9918975861fa449abc3fac9f2ff1afd5932060f96a1a95c7452ceb1511184cf9"},
		{Token: "450350059ebc8b590d09fc97e71440cddad50cc8dcc1ea8ed425901a4861a3d7", Sec: "SEC557318768ea981dabf0b732d636e67417836e0192cffc03c835ab24dc9a2ebc0"},
		// 411bc2ad0be609c3c252609d0b7a42617cbb5856986cc384b125979bb5191a2a SECbb5fa883c2dd9bd78933ae58ad9cd2d177e9cbf864880bf38e4acb5bc79b51b1
		// f1e47f9033f3a7b3ecaea61e627d1f281c9a719bbca00218bd295afa568f077c SEC2bb48f4e418b4fc445fad6d1ff343bf51f7d01d09b90c4cde4a24525b812eb46
		// fca111c7dcc623984d3ac707cdc324a9869eb9439572875b349f4134745290e1 SECc9f15387b683063e72e0e20f2afa362077884eea5e092b681c68e6d3e14e3988
		// 180a63a7618a0d29d21caf8e84f2324e895f4a5b648e31e2e04ccbbc1d677c24 SEC6010f4c8a80efc10a8aa6700fef4661fd6eb4e628015aac3d260e8aa35f5b99f
		// 34055e2dcda7cef2e7521e8d36502c73cca7902cb6077c889be205941016f953 SECefbbf821fc705486458491fbc660fbb547afc056a096fc983fbfd59564325370
		// 6745cc32232476702539507dd0802a342f3619dfbea22b22fc88fe705c48ce66 SEC730b93acd5b26ee537fabe278a42691807f8c76d02d78ec5cd558bacb01305ad
		// 6e60ea828c99275e2c2bbe8e471c007b92297c202d7f9208c079ee84c51ec4e7 SEC79a46475626dd63f4ee647990d5e2c9e592cd2e9c629d633495f5429ca021b44
		// f3c85d35c7705a9ed5a0fbc82f31059c5befb4741e811aa501a94776737bc3c4 SEC6c66e8fb331f266ebd4824688b86eebdbd66b5c0ac3e89b9296c2ab9baddb792
		// 5c469a4f56beeee5d9e18e95f0be645cc1c77eeeb0d11ae7d2a87635536f02e8 SEC2f9ac28cfcf09bc043d3c209d720757e966fd358df60c9f5cd3e0462ffd05632
		// dfe7d5079ef2c62f8e1e6d12a41beaf4866e17a4a569da41e63e8f88fafa2f9e SECcb537cac7aa0c109922ea4b9674ee15317c28cfaecf0606f1cbca6768d96c140
		// 35b88e7790b666f7fe6c84d794e3f76037c7cb70d336c153bdbcf32716c90205 SEC072eff096789d0141a41d2239ac3ab617c32b037748ccece8e229a8c6dd7e007
		// 2bd5d89513d2afa5dd0839c56abe21129ee86a4be6a1a33bed0cf2a8f1bc1287 SECe7e88da4ef57a96baf03dcc5c7f4b3eb127fd85b6bfc55f2f290c8203513fd21
		// be669a2253d687b60a81a9410e9e5867f85cc989b0d3e3ae54de280dca390242 SEC0cee37b28322ff1abeadcc2ae2d4e72490b5cb9dc9ef473178a834ac03e4a691
		// 610a39069fa5a08cb3f8b5081078dfe310910cb62b40713feb7ff62a7cf3b143 SECa43deecbbfa01f215dbbb7983ab84039138004e359c479e73666705c7db4ed52
		// b3fb3de37484c6dc6312e7878c5ec63c2f585376efce3c846575d83dc9ab9323 SECb66afe8d76af735aede2e9b74fa09dec53998773093229bb85f883ae8d36c7d1
		// 1ae0cb237d6550f0b6b6ecabb5c4ae55a7115571a117bb261abea7f8e7c09849 SECab9394d39c7fe22c8967785a34a447962c7b236fed0a0c91eed485675b9a3244
		// 1e9966a2f90364debb4fef3a71769fb10d1a20dcf56686df280ec295165eed07 SEC76a6d61b7d2e019592358cf65795a6f0b3e403715dc6193fe1a1551e72c7b934
		//

	}

}

// 添加用户token
func (smm *SendMessageManager) AddUserToken(sec string, msgurl string) {
	urla, err := url.Parse(msgurl)
	if err != nil {
		logs.Info(err)
		return
	}
	access_token := urla.Query().Get("access_token")
	smm.UserToken = append(smm.UserToken, DinToke{Token: access_token, Sec: sec})
}

func (smm *SendMessageManager) SendMessage(text string, msgtype uint) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered. Error:\n", r)
		}
	}()
	var dt []DinToke
	if len(smm.UserToken) > 0 {
		//是用user token 发送
		dt = smm.UserToken
	} else {
		if msgtype == MessageSystemAlipayPayToken {
			dt = smm.SystemAlipayPayToken
		}
		if msgtype == MessageSystemAutoPayToken {
			dt = smm.SystemAutoPayToken
		}
	}
	for _, v := range dt {
		rb := NewRobot(v.Token, v.Sec)

		logs.Info(text)
		err := rb.SendTextMessage(text, nil, false)
		if err != nil {
			logs.Info(err)
			continue
		}
		break
	}
}

const (
	MessageSystemAlipayPayToken = iota
	MessageSystemAutoPayToken
	MessageUserToken
)
