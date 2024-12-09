package useragent

import "regexp"

// UserAgentRegExp is the regular expression for user agent.
var (
	// UserAgentMobileRegExp is the regular expression for mobile user agent.
	UserAgentMobileRegExp = regexp.MustCompile(`(MIDP)|(WAP)|(UP.Browser)|(Smartphone)|(Obigo)|(Mobile)|(AU.Browser)|(wxd.Mms)|(WxdB.Browser)|(CLDC)|(UP.Link)|(KM.Browser)|(UCWEB)|(SEMC\-Browser)|(Mini)|(Symbian)|(Palm)|(Nokia)|(Panasonic)|(MOT\-)|(SonyEricsson)|(NEC\-)|(Alcatel)|(Ericsson)|(BENQ)|(BenQ)|(Amoisonic)|(Amoi\-)|(Capitel)|(PHILIPS)|(SAMSUNG)|(Lenovo)|(Mitsu)|(Motorola)|(SHARP)|(WAPPER)|(LG\-)|(LG/)|(EG900)|(CECT)|(Compal)|(kejian)|(Bird)|(BIRD)|(G900/V1.0)|(Arima)|(CTL)|(TDG)|(Daxian)|(DAXIAN)|(DBTEL)|(Eastcom)|(EASTCOM)|(PANTECH)|(Dopod)|(Haier)|(HAIER)|(KONKA)|(KEJIAN)|(LENOVO)|(Soutec)|(SOUTEC)|(SAGEM)|(SEC\-)|(SED\-)|(EMOL\-)|(INNO55)|(ZTE)|(iPhone)|(Android)|(Windows CE)|(Wget)|(Java)|(curl)|(Opera)`)

	// UserAgentTabletRegExp is the regular expression for tablet user agent.
	UserAgentTabletRegExp = regexp.MustCompile(`(iPad)|(PlayBook)|(BB10)|(Tablet)|(Kindle)|(Silk)|(Xoom)|(SM\-T)|(GT\-P)|(Nexus 7)|(Nexus 10)|(KFAPWI)`)

	// UserAgentBotRegExp is the regular expression for bot user agent.
	UserAgentBotRegExp = regexp.MustCompile(`(Googlebot)|(Baiduspider)|(bingbot)|(Slurp)|(DuckDuckBot)|(YandexBot)|(Sogou)|(Exabot)|(facebot)|(ia_archiver)`)

	// UserAgentWechatRegExp is the regular expression for WeChat user agent.
	UserAgentWechatRegExp = regexp.MustCompile(`MicroMessenger`)
)
