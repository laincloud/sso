package iuser

type UserProfile interface{}

type User interface {
	GetMobile() string

	// sso v2 中 不同后端对应不同的 sso 实例，所以暂时不需要大一统的 sub
	GetId() int

	// 尽量全局唯一，可以作为 <id, backend> 的单射
	GetSub() string

	GetProfile() UserProfile

	// 本接口的实现中，应该有一个属性用来区分不同的后端, 具体逻辑待定
	// 另一方面，由于在逻辑上，属于不同后端的同一个人在 sso 中被看待为不同的用户，
	// 所以从逻辑上，我们的 user 的实现是一个 “用户身份” 而非 “用户”
	SetBackend(back UserBackend)

	// 下面的设计来源于旧的 sso，为了兼容 sso v1 api
	GetName() string
}
