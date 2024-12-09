package useragent

// UserAgent is the interface for the user agent.
type UserAgent interface {
	String() string
	IsMobile() bool
	IsTablet() bool
	IsDesktop() bool
	IsBot() bool
	IsWeChat() bool
}

// New creates a user agent.
func New(userAgentRaw string) UserAgent {
	return userAgent(userAgentRaw)
}

// userAgent is the user agent.
type userAgent string

// String returns the string of the user agent.
func (u userAgent) String() string {
	return string(u)
}

// IsMobile checks if the request is from mobile.
func (u userAgent) IsMobile() bool {
	return UserAgentMobileRegExp.Match([]byte(u))
}

// IsTablet checks if the request is from tablet.
func (u userAgent) IsTablet() bool {
	return UserAgentTabletRegExp.Match([]byte(u))
}

// IsDesktop checks if the request is from desktop.
func (u userAgent) IsDesktop() bool {
	return !u.IsMobile() && !u.IsTablet()
}

// IsBot checks if the request is from bot.
func (u userAgent) IsBot() bool {
	return UserAgentBotRegExp.Match([]byte(u))
}

// IsWeChat checks if the request is from wechat.
func (u userAgent) IsWeChat() bool {
	return UserAgentWechatRegExp.Match([]byte(u))
}
