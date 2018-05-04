package common

type (
	//IUserInfoGetter - common interface for classes which can be used to obtain information of validity and user info of the third party tokens
	IUserInfoGetter interface {
		//GetUserInfo - get user info or error (d = NotFound if token is invalid or obsolete)
		GetUserInfo() (string, string, error) //social id, name, error
	}
)
