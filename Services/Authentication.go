package services

import (
	"Aramooz/helperfunc"
	"Aramooz/services"
	db "Aramooz/dataBaseServices"
	"Aramooz/dataModels"
	"Aramooz/services/response"
	"encoding/json"
	"fmt"

	"github.com/kataras/iris"
)

func Authentication(ctx iris.Context, ACL AclVal, dismissAcl bool) response.Response{
	Uid := ctx.GetHeader("X-USER")
	Token := ctx.GetHeader("Authorization-Token")

	//str, _ := json.Marshal(req)
	//return string(str)
	res, logged, uacl := BasicOuth(Uid, Token)
	if !logged {
		return res
	}
	if dismissAcl {
		return res
	}
	acl := Acl()
	if !acl.Allow(helperfunc.UIDNR(uacl), ACL) {
		res := response.Response{
			Error: "Access Denied",
			Code:  services.AccessDenied,
		}
		return res
	}
	return res
}

func NewUserService(uid string, token string) *userService {
	return &userService{
		key:   "poster",
		Uid:   uid,
		Token: token,
	}
}

func GetAcl(uid string) map[string]AclVal {
	myg := db.NewDgraphTrasn()
	q := fmt.Sprintf(`
		{
			user(func: uid(%s)) @filter(eq(key,"user")) {
				acl
			}
		}
		`, uid)

	resb := myg.Query(q)
	var resstrc struct {
		User []datamodels.User `json:"user"`
	}
	json.Unmarshal(resb, &resstrc)
	useraclval, _ := helperfunc.UID(resstrc.User[0].ACL)
	acl := Acl()
	return acl.Privileges(useraclval)

}

//BasicOuth : outhenticate the user with uid and token
func BasicOuth(uid string, token string) (response.Response, bool, string) {
	Uid, err := helperfunc.UIDStrX(uid)
	if err != nil {
		res := response.Response{
			Error:  err.Error(),
			Status: "Error",
			Code:   InvalidUID,
		}
		return res, false, ""
	}
	myg := NewMygraphService()
	q := fmt.Sprintf(`
	 	{
	 		user(func: uid(%s)) @filter(eq(kind,"User")) {
				uid
				token
				
				acl
	 		}
	 	}
		 `, Uid)
	dbresstr := myg.Query(q)
	var dbres struct {
		User []datamodels.User
	}
	//ctx.Write(dbresstr)
	if err := json.Unmarshal(dbresstr, &dbres); err != nil {
		res := response.Response{
			Error:  err.Error(),
			Status: "Error",
			Code:   UserOuthFail,
		}
		return res, false, ""
	}
	if len(dbres.User) < 1 {
		res := response.Response{
			Status: "Error",
			Error:  "Invalid Username Or Password",
			Code:   UserOuthFail,
		}
		return res, false, ""
	}
	dbUser := dbres.User[0]
	if dbUser.Token == token  {
		res := response.Response{
			Status: "OK",
			Code:   UserOuthSucc,
		}
		return res, true, dbUser.ACL
	} else {
		res := response.Response{
			Status: "Error",
			Code:   UserOuthFail,
		}
		return res, false, ""
	}

}

/*
func Owner(ctx iris.Context, OwnerUID uint64, TargetUID uint64) (services.Response, bool) {
	res := services.Response{}
	myg := services.NewMygraphService()
	q := fmt.Sprintf(`
		{
			create(func: uid(%#x)) {
				uid
				creator  @filter(uid(%#x)){
					uid
				}
			}

			own(func:uid(%#x)) {
				uid
				owner @filter(uid(%#x)){
					uid
				}
			}
		}
		`, OwnerUID, TargetUID, OwnerUID, TargetUID)
	var dbres struct {
		Create []struct {
			UID     string `json:"uid,omitempty"`
			Creator []struct {
				UID string `json:"uid,omitempty"`
			} `json:"creator,omitempty"`
		} `json:"create"`
		Own []struct {
			UID   string `json:"uid,omitempty"`
			Owner []struct {
				UID string `json:"uid,omitempty"`
			} `json:"owner,omitempty"`
		} `json:"own"`
	}
	err := json.Unmarshal(myg.Query(q), &dbres)

	if err != nil {
		res.Code = services.Fail
		res.Status = "Fail"
		res.Error = err.Error()
		return res, false
	}
	if len(dbres.Create[0].Creator) > 0 || len(dbres.Own[0].Owner) > 0 {
		res.Code = services.OK
		res.Status = "OK"
		return res, true
	} else {
		res.Code = services.Fail
		res.Status = "Fail"
		return res, false
	}

	return res, false
}
