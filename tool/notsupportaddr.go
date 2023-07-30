package tool

import (
	"regexp"
)

// 不支持Proxy访问的网站
// https://www.91yunbbs.com/discussion/comment/1376/
func IsNotSupportAddr(addr string) bool {
	var rule = []string{
		`(.*\.||)(netvigator|torproject)\.(com|cn|net|org)`,
		`(.*\.||)(dafahao|minghui|dongtaiwang|epochtimes|ntdtv|falundafa|wujieliulan|zhengjian)\.(org|com|net)`,
		`(api|ps|sv|offnavi|newvector|ulog\.imap|newloc)(\.map|)\.(baidu|n\.shifen)\.com`,
		`(.+\.|A)(360|so)\.(cn|com)`,
		`(.*\.||)(360|360safe|yunpan|so|qihoo|360totalsecurity)\.(cn|com)`,
		`(torrent|\.torrent|peerJd=|info_hash|get_peers|find_node|Bit7brrent|announce_peer|announce\.php\?passkey=)`, //这条是URL规则
		`(A.*\@)(guerrillamail|guerrillamailblock|sharklasers|grr|pokemail|spam4|bccto|chacuo|027168)\.(info|biz|com|cn|cc|de|net|org|me|la)`,
		`(.?)(xunlei|sandai|Thunder|XLLiveUD)(.)`,
	}

	for _, v := range rule {
		r := regexp.MustCompile(v)
		if r.Match([]byte(addr)) {
			return true
		}
	}
	//轮子网
	return false
}

func IsNotSupportEmail(b []byte) bool {
	r := regexp.MustCompile(`(?i)(Subject|HELO|SMTP)`)
	return r.Match(b)
	// return false
}

//禁止乱发邮件
//(Subject|HELO|SMTP)
// 具体规则

// 带TLS的SMTP端口为465，普通SMTP端口为25，如果标准端口的话，阻止这两个端口应该足够。
// SMTP有通信协议的模式，理论上可以识别，但开发者愿不愿意弄是一回事，opensource自己解决了。
// 带TLS的，都带TLS的，就别指望了。

// 禁用 BT 防止版权争议
// BitTorrent protocol
// 数据包明文匹配

// 禁止 百度高精度定位 防止IP与客户端地理位置被记录
// (api|ps|sv|offnavi|newvector|ulog\.imap|newloc)(\.map|)\.(baidu|n\.shifen)\.com
// 数据包明文匹配

// 禁止360有毒服务 屏蔽360
// (.+\.|^)(360|so)\.(cn|com)
// 数据包明文匹配

// 禁止 邮件滥发 防止垃圾邮件滥用
// (Subject|HELO|SMTP)
// 数据包明文匹配

// 屏蔽轮子网站
// 感谢 烟雨阁 前辈网站
// 感谢 Daniel Jack
// (.*\.||)(dafahao|minghui|dongtaiwang|epochtimes|ntdtv|falundafa|wujieliulan)\.(org|com|net)
// 数据包明文匹配

// 屏蔽 BT（2）
// 感谢 91yunbbs 用户 Hina
// (torrent|\.torrent|peer_id=|info_hash|get_peers|find_node|BitTorrent|announce_peer|announce\.php\?passkey=)
// 数据包明文匹配

// 屏蔽Spam邮箱
// 感谢 91yunbbs 用户 liangzhukun
// (JasonLee修改完善)
// (^.*\@)(guerrillamail|guerrillamailblock|sharklasers|grr|pokemail|spam4|bccto|chacuo|027168)\.(info|biz|com|de|net|org|me|la)
// 数据包明文匹配

// 屏蔽迅雷
// 感谢Ashe 提供
// (.?)(xunlei|sandai|Thunder|XLLiveUD)(.)
// 数据包明文匹配
// 可基本屏蔽掉迅雷的网站，并且在全局模式下会影响迅雷的下载能力
// 带TLS的SMTP端口为465，普通SMTP端口为25，如果标准端口的话，阻止这两个端口应该足够。
// SMTP有通信协议的模式，理论上可以识别，但开发者愿不愿意弄是一回事，opensource自己解决了。
// 带TLS的，都带TLS的，就别指望了。
