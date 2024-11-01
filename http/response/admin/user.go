package admin

import "Gwen/model"

type LoginPayload struct {
	Username   string   `json:"username"`
	Token      string   `json:"token"`
	RouteNames []string `json:"route_names"`
	Nickname   string   `json:"nickname"`
}

var UserRouteNames = []string{
	"MyTagList", "MyAddressBookList", "MyInfo", "MyAddressBookCollection",
}
var AdminRouteNames = []string{"*"}

type UserOauthItem struct {
	ThirdType string `json:"third_type"`
	Status    int    `json:"status"`
}

type GroupUsersPayload struct {
	Id       uint   `json:"id"`
	Username string `json:"username"`
	Status   int    `json:"status"`
}

func (g *GroupUsersPayload) FromUser(user *model.User) {
	g.Id = user.Id
	g.Username = user.Username
	g.Status = 1
}
