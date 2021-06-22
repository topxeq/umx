package main

import (
	"os"
	"runtime"
	"time"

	"net/http"
	"path"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kardianos/service"
	_ "github.com/mattn/go-sqlite3"

	"github.com/topxeq/sqltk"
	"github.com/topxeq/tk"
)

var versionG = "0.99a"
var cfgMapG map[string]string = nil
var basePathG = "."
var portG = ":7492"
var sslPortG = ":7493"
var serviceNameG = "umx"
var serviceModeG = false
var runModeG = ""
var currentOSG = ""
var configFileNameG = serviceNameG + ".cfg"

var defaultConfigG = `
{
	"DBType": "SQLite",
	"DBConnectString": "umx.db",
	"MainSecret": "UMX_easy",
	"TokenSecret": "is_Token",
	"TokenExpire": "1440",
	"TestMode": "true"
}
`

var initSQLs = []string{
	`DROP TABLE APP`,
	`CREATE TABLE APP (ID int(11) NOT NULL AUTO_INCREMENT, CODE VARCHAR(100) DEFAULT NULL, SECRET VARCHAR(100) DEFAULT '', NAME VARCHAR(255) DEFAULT NULL, REMARK varchar(255) DEFAULT NULL, RESV1 varchar(255) DEFAULT NULL, TIME_STAMP datetime NOT NULL DEFAULT NOW(), PRIMARY KEY (ID), UNIQUE KEY APP_UN (CODE)) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8`,
	`INSERT INTO APP (CODE, NAME, SECRET) VALUES('', '', '')`,
	`INSERT INTO APP (CODE, NAME, SECRET) VALUES('COMMON', '', '')`,
	`INSERT INTO APP (CODE, NAME, SECRET) VALUES('TEST', '', '')`,
	`INSERT INTO APP (CODE, NAME, SECRET) VALUES('APP1', '', 'SECRET1')`,

	`DROP TABLE ORG`,
	`CREATE TABLE ORG (ID int(11) NOT NULL AUTO_INCREMENT, UP_ID int(11) DEFAULT NULL, APP_CODE VARCHAR(100) DEFAULT NULL, NAME VARCHAR(1024) NOT NULL, CODE VARCHAR(32) DEFAULT NULL, TYPE VARCHAR(32) DEFAULT NULL, SUB_TYPE VARCHAR(100) DEFAULT NULL, CONTACT VARCHAR(255) DEFAULT NULL, EMAIL VARCHAR(100) DEFAULT NULL, PHONE VARCHAR(50) DEFAULT NULL, MOBILE VARCHAR(32) DEFAULT NULL, ADDRESS VARCHAR(255) DEFAULT NULL, UP_CODE VARCHAR(32) DEFAULT NULL, UP_NAME VARCHAR(1024) DEFAULT NULL, DESCRIPTION VARCHAR(255) DEFAULT NULL, REMARK varchar(255) DEFAULT NULL, RESV1 varchar(255) DEFAULT NULL, RESV2 varchar(255) DEFAULT NULL, RESV3 varchar(255) DEFAULT NULL, REL JSON DEFAULT NULL, TIME_STAMP datetime NOT NULL DEFAULT NOW(), PRIMARY KEY (ID)) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8`,
	`INSERT INTO ORG (NAME, APP_CODE) VALUES('北京理工大学', 'APP1')`,

	`DROP TABLE ORG_GROUP`,
	`CREATE TABLE ORG_GROUP (ID int(11) NOT NULL AUTO_INCREMENT, UP_ID int(11) DEFAULT NULL, APP_CODE VARCHAR(100) DEFAULT NULL, NAME VARCHAR(1024) NOT NULL, CODE VARCHAR(32) DEFAULT NULL, UP_CODE VARCHAR(32) DEFAULT NULL, TYPE VARCHAR(32) DEFAULT NULL, LEVEL VARCHAR(32) DEFAULT NULL, LEVEL_INDEX int(5) DEFAULT NULL, DESCRIPTION VARCHAR(255) DEFAULT NULL, REMARK varchar(255) DEFAULT NULL, RESV1 varchar(255) DEFAULT NULL, REL JSON DEFAULT NULL, TIME_STAMP datetime NOT NULL DEFAULT NOW(), PRIMARY KEY (ID)) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8`,
	`INSERT INTO ORG_GROUP (NAME, APP_CODE, TYPE, LEVEL_INDEX) VALUES('中国', 'APP1', 'area', 3)`,
	`INSERT INTO ORG_GROUP (NAME, APP_CODE, TYPE, LEVEL_INDEX, UP_ID) select '北京', 'APP1', 'area', 2, ID from ORG_GROUP where NAME='中国' AND APP_CODE='APP1'`,
	`INSERT INTO ORG_GROUP (NAME, APP_CODE, TYPE, LEVEL_INDEX, UP_ID) select '海淀区', 'APP1', 'area', 1, ID from ORG_GROUP where NAME='北京' AND APP_CODE='APP1'`,

	`DROP TABLE ORG_GROUP_LINK`,
	`CREATE TABLE ORG_GROUP_LINK (ID int(11) NOT NULL AUTO_INCREMENT, APP_CODE VARCHAR(100) DEFAULT NULL, ORG_GROUP_ID int(11) NOT NULL, ORG_GROUP_CODE VARCHAR(32) DEFAULT '', GROUP_NAME VARCHAR(1024) DEFAULT NULL, ORG_ID int(11) NOT NULL, ORG_NAME VARCHAR(1024) DEFAULT NULL, REMARK varchar(255) DEFAULT NULL, RESV1 varchar(255) DEFAULT NULL, TIME_STAMP datetime NOT NULL DEFAULT NOW(), PRIMARY KEY (ID)) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8`,
	`INSERT INTO ORG_GROUP_LINK (APP_CODE, ORG_GROUP_ID, GROUP_NAME, ORG_ID, ORG_NAME, REMARK) select 'APP1', a.ID, '海淀区', b.ID, "北京理工大学", '' from (SELECT ID from ORG_GROUP where NAME='海淀区' AND APP_CODE='APP1') a, (SELECT ID from ORG where NAME='北京理工大学' AND APP_CODE='APP1') b`,

	`DROP TABLE ORG_GROUP_GROUP_LINK`,
	`CREATE TABLE ORG_GROUP_GROUP_LINK (ID int(11) NOT NULL AUTO_INCREMENT, APP_CODE VARCHAR(100) DEFAULT NULL, GROUP_ID int(11) NOT NULL, GROUP_NAME VARCHAR(1024) DEFAULT NULL, UP_GROUP_ID int(11) NOT NULL, UP_GROUP_CODE VARCHAR(32) DEFAULT '', UP_GROUP_NAME VARCHAR(1024) DEFAULT NULL, REMARK varchar(255) DEFAULT NULL, RESV1 varchar(255) DEFAULT NULL, TIME_STAMP datetime NOT NULL DEFAULT NOW(), PRIMARY KEY (ID)) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8`,
	`INSERT INTO ORG_GROUP_GROUP_LINK (APP_CODE, GROUP_ID, GROUP_NAME, UP_GROUP_ID, UP_GROUP_NAME, REMARK) select 'APP1', a.ID, '海淀区', b.ID, "北京", '' from (SELECT ID from ORG_GROUP where NAME='海淀区' AND APP_CODE='APP1') a, (SELECT ID from ORG_GROUP where NAME='北京' AND APP_CODE='APP1') b`,

	`DROP TABLE USER`,
	`CREATE TABLE USER (ID int(11) NOT NULL AUTO_INCREMENT, APP_CODE VARCHAR(100) DEFAULT NULL, ID_TYPE VARCHAR(32), USER_ID VARCHAR(100) NOT NULL, PASSWORD VARCHAR(100) NOT NULL, NAME VARCHAR(100) DEFAULT NULL, EMAIL VARCHAR(100) DEFAULT NULL, PHONE VARCHAR(50) DEFAULT NULL, FAX VARCHAR(50) DEFAULT NULL, MOBILE VARCHAR(32) DEFAULT NULL, ROLE VARCHAR(100) DEFAULT NULL, RIGHTS VARCHAR(1024) DEFAULT '', USER_STATUS VARCHAR(32) DEFAULT NULL, ORG_ID int(11) DEFAULT NULL, ORG_CODE VARCHAR(32) DEFAULT NULL, AREA_ID int(11) DEFAULT NULL, AREA_CODE VARCHAR(32) DEFAULT NULL, POSITION VARCHAR(100) DEFAULT NULL, TITLE VARCHAR(100) DEFAULT NULL, ADDRESS VARCHAR(1024) DEFAULT NULL, USER_TYPE VARCHAR(100) DEFAULT NULL, INFO_FLAG VARCHAR(32) DEFAULT '', DESCRIPTION VARCHAR(255) DEFAULT NULL, REMARK varchar(255) DEFAULT NULL, RESV1 varchar(255) DEFAULT NULL, RESV2 varchar(255) DEFAULT NULL, RESV3 varchar(255) DEFAULT NULL, LOGIN_TIME datetime DEFAULT NULL, TIME1 datetime DEFAULT NULL, TIME2 datetime DEFAULT NULL, TIME_STAMP datetime NOT NULL DEFAULT NOW(), PRIMARY KEY (ID), UNIQUE KEY USER_UN (APP_CODE,USER_ID)) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8`,
	`INSERT INTO USER (USER_ID, PASSWORD, APP_CODE, ROLE) VALUES('admin', 'admin9768', 'APP1', 'admin')`,

	`DROP TABLE USER_ORG_RIGHT_LINK`,
	`CREATE TABLE USER_ORG_RIGHT_LINK (ID int(11) NOT NULL AUTO_INCREMENT, APP_CODE VARCHAR(100) DEFAULT NULL, REAL_USER_ID int(11) DEFAULT NULL, USER_ID VARCHAR(100) DEFAULT NULL, ORG_ID int(11) DEFAULT NULL, ORG_GROUP_ID int(11) DEFAULT NULL, LINK_TYPE VARCHAR(32) DEFAULT NULL, RIGHT_NAME VARCHAR(100) DEFAULT NULL, ROLE VARCHAR(100) DEFAULT NULL, REMARK varchar(255) DEFAULT NULL, RESV1 varchar(255) DEFAULT NULL, TIME_STAMP datetime NOT NULL DEFAULT NOW(), PRIMARY KEY (ID), UNIQUE KEY USER_UN (APP_CODE,REAL_USER_ID,USER_ID,ORG_ID,ORG_GROUP_ID,RIGHT_NAME)) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8`,
	`INSERT INTO USER_ORG_RIGHT_LINK (APP_CODE, USER_ID, ORG_GROUP_ID, LINK_TYPE, RIGHT_NAME) select 'APP1', 'admin', ID, 'ORG_GROUP', 'all' from ORG_GROUP where NAME='中国' AND APP_CODE='APP1'`,
}

type ConfigType struct {
	DBType          string
	DBConnectString string
	MainSecret      string
	TokenSecret     string
	TokenExpire     string
	TestMode        string
}

var configG ConfigType

func initSystem() {
	var errT error

	tk.Pl("reading configuration...")

	errT = tk.LoadJSONFromFile(filepath.Join(basePathG, "config.json"), &configG)
	if errT != nil {
		tk.Pl("failed to read config: %v", errT)
		tk.Pl("use default config instead")

		errT = tk.LoadJSONFromString(defaultConfigG, &configG)
		if errT != nil {
			tk.Pl("failed to read default config: %v", errT)
			os.Exit(1)
		}
	}

	tk.Plv(configG)

	// cfgMapG = tk.LoadSimpleMapFromString(defaultConfigG)

	dbTypeT := strings.ToLower(configG.DBType)

	var dbTypeStrT string

	switch dbTypeT {
	case "sqlite", "sqlite3":
		dbTypeStrT = "sqlite3"
	case "mysql":
		dbTypeStrT = "mysql"
	default:
		dbTypeStrT = "sqlite3"
	}

	connectStrT := tk.Trim(configG.DBConnectString)

	tk.Pl("DSN: (%v)%v", dbTypeStrT, connectStrT)

	dbT, errT := sqltk.ConnectDBNoPing(dbTypeStrT, connectStrT)
	if errT != nil {
		tk.Pl("failed to connect DB: %v", errT)
		return
	}

	defer dbT.Close()

	errCountT := 0

	for i, v := range initSQLs {
		tk.Pl("[%v] exec SQL: %v", i, v)

		c1, c2, errT := sqltk.ExecV(dbT, v)

		if errT != nil {
			errCountT++
		}

		tk.Pl("SQL result: %v, %v, %v", c1, c2, errT)
	}

	tk.Pl("total errors in SQLs: %v", errCountT)
}

// return "" indicates success, otherwise the fail reason
func checkAuth(appCodeA string, authA string) string {
	dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
	if errT != nil {
		return tk.Spr("failed to connect DB: %v", errT)
	}

	defer dbT.Close()

	secretT, errT := sqltk.QueryDBString(dbT, "select SECRET from APP where CODE='"+tk.Replace(appCodeA, "'", "''")+"'")
	if errT != nil {
		return tk.Spr("APP code not found: %v", appCodeA)
	}

	if authA == "test" {
		if configG.TestMode == "true" {
			return ""
		}
	}

	authTargetT := tk.MD5Encrypt(appCodeA + secretT + tk.GetNowDateString())

	if authA == authTargetT {
		return ""
	}

	return "invalid auth code"

}

var tokenExpireG int = 1

// func generatePrimaryToken(appCodeA string, userIDA string) string {
// 	strT := configG.TokenSecret + "|" + tk.GetNowTimeString()

// 	return tk.EncryptStringByTXDEF(strT, configG.MainSecret)
// }

func checkPrimaryAuth(authA string) string {
	strT := tk.DecryptStringByTXDEF(authA, configG.MainSecret)
	// tk.LogWithTimeCompact("%v, %v\n", tokenA, configG.MainSecret)
	// tk.LogWithTimeCompact("%v, %v\n", strT, configG.TokenSecret)
	// tk.Plvsr(strT, configG.TokenSecret)

	listT := tk.Split(strT, "|")

	if len(listT) < 2 {
		return "invalid auth"
	}

	if configG.TokenSecret != listT[0] {
		return "invalid auth"
	}

	if listT[1] != tk.GetNowDateString() {
		return "invalid auth"
	}

	return ""
}

func generateToken(appCodeA string, userIDA string, roleA string) string {
	strT := appCodeA + "|" + userIDA + "|" + roleA + "|" + tk.GetNowTimeString()

	return tk.EncryptStringByTXDEF(strT, configG.TokenSecret)
}

func getRoleInToken(tokenA string) string {
	strT := tk.DecryptStringByTXDEF(tokenA, configG.TokenSecret)

	listT := tk.Split(strT, "|")

	if len(listT) < 4 {
		return ""
	}

	return listT[2]
}

func checkToken(appCodeA string, userIDA string, tokenA string) string {
	strT := tk.DecryptStringByTXDEF(tokenA, configG.TokenSecret)

	listT := tk.Split(strT, "|")

	if len(listT) < 4 {
		return "invalid token"
	}

	if appCodeA != listT[0] {
		return "invalid token"
	}

	if userIDA != listT[1] {
		return "invalid token"
	}

	timeT, errT := tk.StrToTimeByFormat(listT[3], tk.TimeFormatCompact)
	if errT != nil {
		return "invalid token"
	}

	expectTimeT := timeT.Add(time.Minute * time.Duration(tokenExpireG))

	if expectTimeT.Before(time.Now()) {
		return "token expired"
	}

	return ""
}

func doJapi(res http.ResponseWriter, req *http.Request) string {
	if req != nil {
		req.ParseForm()
	}

	tk.PlNow("REQ: %#v", req)
	tk.Pl("[%v] REQ: %#v", tk.GetNowTimeStringFormal(), req)

	reqT := tk.GetFormValueWithDefaultValue(req, "req", "")

	if res != nil {
		res.Header().Set("Access-Control-Allow-Origin", "*")
		res.Header().Set("Access-Control-Allow-Headers", "*")
		res.Header().Set("Content-Type", "text/json;charset=utf-8")
	}

	res.WriteHeader(http.StatusOK)

	vo := tk.GetFormValueWithDefaultValue(req, "vo", "")

	var paraMapT map[string]string
	var errT error

	if vo == "" {
		paraMapT = tk.FormToMap(req.Form)
	} else {
		paraMapT, errT = tk.MSSFromJSON(vo)

		if errT != nil {
			return tk.GenerateJSONPResponse("success", "invalid vo format", req)
		}
	}

	switch reqT {
	case "debug":
		tk.Pl("%v", req)
		a := make([]int, 3)
		a[2] = 8

		return tk.GenerateJSONPResponse("success", tk.IntToStr(a[5]), req)
	case "requestinfo":
		rs := tk.Spr("%#v", req)

		return tk.GenerateJSONPResponse("success", rs, req)

	case "test":

		return tk.GenerateJSONPResponse("success", "test1", req)

	case "getToken":
		authT := paraMapT["auth"]

		authRsT := checkPrimaryAuth(authT)

		if authRsT != "" {
			appCodeT := paraMapT["appCode"]

			authRsT = checkAuth(appCodeT, authT)

			if authRsT != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}

			dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
			if errT != nil {
				return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
			}

			defer dbT.Close()

			userT := paraMapT["user"]
			if userT == "" {
				return tk.GenerateJSONPResponse("fail", tk.Spr("empty user id"), req)
			}

			passT := paraMapT["password"]
			if passT == "" {
				return tk.GenerateJSONPResponse("fail", tk.Spr("empty password"), req)
			}

			sqlRsT, errT := sqltk.QueryDBNSS(dbT, "select * from USER where APP_CODE='"+tk.Replace(appCodeT, "'", "''")+"' and USER_ID='"+tk.Replace(userT, "'", "''")+"'")
			if errT != nil {
				return tk.GenerateJSONPResponse("fail", tk.Spr("user %v not exists: %v", userT, errT), req)
			}

			tk.Pl("%v", sqlRsT)

			if len(sqlRsT) < 2 {
				return tk.GenerateJSONPResponse("fail", tk.Spr("user %v not exists", userT), req)
			}

			userMapT := sqltk.OneLineRecordToMap(sqlRsT)

			if passT != userMapT["PASSWORD"] {
				tk.Plvsr(passT, userMapT["PASSWORD"])
				return tk.GenerateJSONPResponse("fail", tk.Spr("user id or password not match"), req)
			}

			return tk.GenerateJSONPResponseWithMore("success", generateToken(appCodeT, userT, userMapT["ROLE"]), req)
		}

		appCodeT := paraMapT["appCode"]
		if appCodeT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty app code"), req)
		}

		userT := paraMapT["user"]
		if userT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty user"), req)
		}

		return tk.GenerateJSONPResponse("success", generateToken(appCodeT, userT, "primary"), req)

	case "checkToken":
		appCodeT := paraMapT["appCode"]

		userT := paraMapT["user"]
		if userT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty user id"), req)
		}

		tokenT := paraMapT["token"]
		if tokenT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty tokenT"), req)
		}

		rs := checkToken(appCodeT, userT, tokenT)

		if rs != "" {
			return tk.GenerateJSONPResponse("fail", rs, req)
		}

		return tk.GenerateJSONPResponse("success", "token is valid", req)

	case "refreshToken":
		appCodeT := paraMapT["appCode"]

		userT := paraMapT["user"]
		if userT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty user id"), req)
		}

		tokenT := paraMapT["token"]
		if tokenT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty tokenT"), req)
		}

		rs := checkToken(appCodeT, userT, tokenT)

		if rs != "" {
			return tk.GenerateJSONPResponse("fail", rs, req)
		}

		return tk.GenerateJSONPResponseWithMore("success", "", req, "token", generateToken(appCodeT, userT, getRoleInToken(tokenT)))

	case "clearApp":
		authT := paraMapT["auth"]

		authRsT := checkPrimaryAuth(authT)

		if authRsT != "" {
			return tk.GenerateJSONPResponse("fail", "auth failed: "+authRsT, req)
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		appCodeT := paraMapT["appCode"]
		if appCodeT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty app code"), req)
		}

		totalCountT := int64(0)

		_, c2, errT := sqltk.ExecV(dbT, `DELETE FROM APP WHERE CODE='`+sqltk.FormatSQLValue(appCodeT)+`' `)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action(DELETE APP) failed: %v", errT), req)
		}

		totalCountT += c2

		_, c2, errT = sqltk.ExecV(dbT, `DELETE FROM USER WHERE APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' `)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action(DELETE USER) failed: %v", errT), req)
		}

		totalCountT += c2

		_, c2, errT = sqltk.ExecV(dbT, `DELETE FROM ORG WHERE APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' `)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action(DELETE ORG) failed: %v", errT), req)
		}

		totalCountT += c2

		_, c2, errT = sqltk.ExecV(dbT, `DELETE FROM ORG_GROUP WHERE APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' `)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action(DELETE ORG_GROUP) failed: %v", errT), req)
		}

		totalCountT += c2

		_, c2, errT = sqltk.ExecV(dbT, `DELETE FROM ORG_GROUP_LINK WHERE APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' `)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action(DELETE ORG_GROUP_LINK) failed: %v", errT), req)
		}

		totalCountT += c2

		_, c2, errT = sqltk.ExecV(dbT, `DELETE FROM USER_ORG_RIGHT_LINK WHERE APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' `)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action(DELETE USER_ORG_RIGHT_LINK) failed: %v", errT), req)
		}

		totalCountT += c2

		return tk.GenerateJSONPResponseWithMore("success", tk.Int64ToStr(totalCountT), req)

	case "addApp":
		authT := paraMapT["auth"]

		authRsT := checkPrimaryAuth(authT)

		if authRsT != "" {
			return tk.GenerateJSONPResponse("fail", "auth failed: "+authRsT, req)
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		appCodeT := paraMapT["appCode"]
		if appCodeT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty app code"), req)
		}

		appSecretT := paraMapT["appSecret"]
		if appSecretT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty app secret"), req)
		}

		appNameT := paraMapT["appName"]
		appRemarkT := paraMapT["appRemark"]
		appResv1T := paraMapT["appResv1"]

		c1, c2, errT := sqltk.ExecV(dbT, `INSERT INTO APP (CODE, NAME, SECRET, REMARK, RESV1) VALUES('`+sqltk.FormatSQLValue(appCodeT)+`', '`+sqltk.FormatSQLValue(appNameT)+`', '`+sqltk.FormatSQLValue(appSecretT)+`', '`+sqltk.FormatSQLValue(appRemarkT)+`', '`+sqltk.FormatSQLValue(appResv1T)+`') `)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if c2 < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "no records affected"), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", tk.Int64ToStr(c1), req)

	case "addUser":
		authT := paraMapT["auth"]

		appCodeT := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		userT := tk.Trim(paraMapT["user"])
		if userT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty user id"), req)
		}

		passT := paraMapT["password"]
		if passT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty password"), req)
		}
		if len(passT) != 32 {
			passT = tk.MD5Encrypt(passT)
		}

		userNameT := paraMapT["name"]
		emailT := paraMapT["email"]
		mobileT := paraMapT["mobile"]
		roleT := paraMapT["role"]
		rightsT := paraMapT["rights"]
		statusT := paraMapT["status"]
		descriptionT := paraMapT["description"]
		remarkT := paraMapT["remark"]
		resv1T := paraMapT["resv1"]
		resv2T := paraMapT["resv2"]
		resv3T := paraMapT["resv3"]

		phoneT := paraMapT["phone"]
		faxT := paraMapT["fax"]

		infoFlagT := paraMapT["infoFlag"]

		orgIDT := tk.Trim(paraMapT["orgID"])
		if orgIDT == "" {
			orgIDT = "NULL"
		}

		orgCodeT := tk.Trim(paraMapT["orgCode"])

		if orgCodeT != "" && (orgIDT == "" || orgIDT == "NULL") {
			sqlT := `select ID from ORG where APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `' and CODE='` + orgCodeT + `'`
			sqlRs8T, errT := sqltk.QueryDBString(dbT, sqlT)

			if errT == nil {
				orgIDT = sqlRs8T
			}
		}

		if orgCodeT == "" {
			if orgIDT != "NULL" {
				tmpCodeT, errT := sqltk.QueryDBString(dbT, `SELECT CODE FROM ORG WHERE APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' AND ID=`+orgIDT)

				if errT != nil {
					tk.Pl("DB error: %v", errT)
				} else {
					orgCodeT = tmpCodeT
				}
			}
		}

		// tk.Plvx(orgCodeT)

		areaIDT := tk.Trim(paraMapT["areaID"])
		if areaIDT == "" {
			areaIDT = "NULL"
		}

		areaCodeT := tk.Trim(paraMapT["areaCode"])

		if areaCodeT != "" && (areaIDT == "" || areaIDT == "NULL") {
			sqlT := `select ID from ORG_GROUP where APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `' and CODE='` + areaCodeT + `' and TYPE='area'`
			sqlRs8T, errT := sqltk.QueryDBString(dbT, sqlT)

			if errT == nil {
				areaIDT = sqlRs8T
			}
		}

		positionT := paraMapT["position"]
		addressT := paraMapT["address"]
		userTypeT := paraMapT["userType"]

		c1, c2, errT := sqltk.ExecV(dbT, `INSERT INTO USER (APP_CODE, USER_ID, PASSWORD, NAME, EMAIL, MOBILE, ROLE, RIGHTS, USER_STATUS, DESCRIPTION, REMARK, RESV1, RESV2, RESV3, PHONE, FAX, ORG_ID, ORG_CODE, AREA_ID, AREA_CODE, POSITION, ADDRESS, USER_TYPE, INFO_FLAG) VALUES('`+sqltk.FormatSQLValue(appCodeT)+`', '`+sqltk.FormatSQLValue(userT)+`', '`+sqltk.FormatSQLValue(passT)+`', '`+sqltk.FormatSQLValue(userNameT)+`', '`+sqltk.FormatSQLValue(emailT)+`', '`+sqltk.FormatSQLValue(mobileT)+`', '`+sqltk.FormatSQLValue(roleT)+`', '`+sqltk.FormatSQLValue(rightsT)+`', '`+sqltk.FormatSQLValue(statusT)+`', '`+sqltk.FormatSQLValue(descriptionT)+`', '`+sqltk.FormatSQLValue(remarkT)+`', '`+sqltk.FormatSQLValue(resv1T)+`', '`+sqltk.FormatSQLValue(resv2T)+`', '`+sqltk.FormatSQLValue(resv3T)+`', '`+sqltk.FormatSQLValue(phoneT)+`', '`+sqltk.FormatSQLValue(faxT)+`', `+sqltk.FormatSQLValue(orgIDT)+`, '`+sqltk.FormatSQLValue(orgCodeT)+`', `+sqltk.FormatSQLValue(areaIDT)+`, '`+sqltk.FormatSQLValue(areaCodeT)+`', '`+sqltk.FormatSQLValue(positionT)+`', '`+sqltk.FormatSQLValue(addressT)+`', '`+sqltk.FormatSQLValue(userTypeT)+`', '`+sqltk.FormatSQLValue(infoFlagT)+`') `)

		if errT != nil {
			tk.Plvx(`INSERT INTO USER (APP_CODE, USER_ID, PASSWORD, NAME, EMAIL, MOBILE, ROLE, RIGHTS, USER_STATUS, DESCRIPTION, REMARK, RESV1, RESV2, RESV3, PHONE, FAX, ORG_ID, ORG_CODE, AREA_ID, AREA_CODE, POSITION, ADDRESS, USER_TYPE, INFO_FLAG) VALUES('` + sqltk.FormatSQLValue(appCodeT) + `', '` + sqltk.FormatSQLValue(userT) + `', '` + sqltk.FormatSQLValue(passT) + `', '` + sqltk.FormatSQLValue(userNameT) + `', '` + sqltk.FormatSQLValue(emailT) + `', '` + sqltk.FormatSQLValue(mobileT) + `', '` + sqltk.FormatSQLValue(roleT) + `', '` + sqltk.FormatSQLValue(rightsT) + `', '` + sqltk.FormatSQLValue(statusT) + `', '` + sqltk.FormatSQLValue(descriptionT) + `', '` + sqltk.FormatSQLValue(remarkT) + `', '` + sqltk.FormatSQLValue(resv1T) + `', '` + sqltk.FormatSQLValue(resv2T) + `', '` + sqltk.FormatSQLValue(resv3T) + `', '` + sqltk.FormatSQLValue(phoneT) + `', '` + sqltk.FormatSQLValue(faxT) + `', ` + sqltk.FormatSQLValue(orgIDT) + `, '` + sqltk.FormatSQLValue(orgCodeT) + `', ` + sqltk.FormatSQLValue(areaIDT) + `, '` + sqltk.FormatSQLValue(areaCodeT) + `', '` + sqltk.FormatSQLValue(positionT) + `', '` + sqltk.FormatSQLValue(addressT) + `', '` + sqltk.FormatSQLValue(userTypeT) + `', '` + sqltk.FormatSQLValue(infoFlagT) + `') `)
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if c2 < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "no records affected"), req)
		}

		linksT := tk.Trim(paraMapT["links"])

		var linksAryT []map[string]string = nil

		if linksT != "" {
			linksAryT = tk.JSONToMapStringStringArray(linksT)
		}

		if linksAryT != nil {
			for _, v := range linksAryT {
				vi := tk.StrToInt(tk.Trim(v["ID"]), -1)
				if vi < 0 {
					return tk.GenerateJSONPResponse("fail", tk.Spr("invalid links item: %v", v), req)
				}

				typeT := tk.Trim(v["Type"])

				rightNameT := tk.Trim(v["Right"])

				var sqlT string

				if typeT == "group" {
					sqlT = `INSERT INTO USER_ORG_RIGHT_LINK (APP_CODE, LINK_TYPE, REAL_USER_ID, USER_ID, ORG_GROUP_ID, RIGHT_NAME) VALUES('` + sqltk.FormatSQLValue(appCodeT) + `', 'group', ` + tk.Int64ToStr(c1) + `, '` + userT + `', ` + tk.IntToStr(vi) + `, '` + sqltk.FormatSQLValue(rightNameT) + `')`
				} else {
					sqlT = `INSERT INTO USER_ORG_RIGHT_LINK (APP_CODE, LINK_TYPE, REAL_USER_ID, USER_ID, ORG_ID, RIGHT_NAME) VALUES('` + sqltk.FormatSQLValue(appCodeT) + `', '', ` + tk.Int64ToStr(c1) + `, '` + userT + `', ` + tk.IntToStr(vi) + `, '` + sqltk.FormatSQLValue(rightNameT) + `')`
				}

				_, c2i, errT := sqltk.ExecV(dbT, sqlT)

				if errT != nil {
					return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
				}

				if c2i < 1 {
					return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed, no records affected for links item: %v", v), req)
				}

			}
		}

		return tk.GenerateJSONPResponseWithMore("success", tk.Int64ToStr(c1), req)

	case "modifyUser":
		authT := paraMapT["auth"]

		appCodeT, appCodeOK := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		var bufT strings.Builder

		if appCodeOK {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `'`)
		}

		userIDT, ok := paraMapT["userID"]
		if !ok || userIDT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty user id"), req)
		}

		userT, ok := paraMapT["user"]
		if ok {
			if userT == "" {
				return tk.GenerateJSONPResponse("fail", tk.Spr("empty user account name"), req)
			}

			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`USER_ID='` + sqltk.FormatSQLValue(userT) + `'`)
		}

		passT, ok := paraMapT["password"]
		if ok {
			if passT == "" {
				return tk.GenerateJSONPResponse("fail", tk.Spr("empty password"), req)
			}

			if len(passT) != 32 {
				passT = tk.MD5Encrypt(passT)
			}

			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`PASSWORD='` + sqltk.FormatSQLValue(passT) + `'`)
		}

		userNameT, ok := paraMapT["name"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`NAME='` + sqltk.FormatSQLValue(userNameT) + `'`)
		}

		emailT, ok := paraMapT["email"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`EMAIL='` + sqltk.FormatSQLValue(emailT) + `'`)
		}

		mobileT, ok := paraMapT["mobile"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`MOBILE='` + sqltk.FormatSQLValue(mobileT) + `'`)
		}

		infoFlagT, ok := paraMapT["infoFlag"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`INFO_FLAG='` + sqltk.FormatSQLValue(infoFlagT) + `'`)
		}

		roleT, ok := paraMapT["role"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`ROLE='` + sqltk.FormatSQLValue(roleT) + `'`)
		}

		rightsT, ok := paraMapT["rights"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`RIGHTS='` + sqltk.FormatSQLValue(rightsT) + `'`)
		}

		statusT, ok := paraMapT["status"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`USER_STATUS='` + sqltk.FormatSQLValue(statusT) + `'`)
		}

		descriptionT, ok := paraMapT["description"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`DESCRIPTION='` + sqltk.FormatSQLValue(descriptionT) + `'`)
		}

		remarkT, ok := paraMapT["remark"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`REMARK='` + sqltk.FormatSQLValue(remarkT) + `'`)
		}

		resv1T, ok := paraMapT["resv1"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`RESV1='` + sqltk.FormatSQLValue(resv1T) + `'`)
		}

		resv2T, ok := paraMapT["resv2"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`RESV2='` + sqltk.FormatSQLValue(resv2T) + `'`)
		}

		resv3T, ok := paraMapT["resv3"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`RESV3='` + sqltk.FormatSQLValue(resv3T) + `'`)
		}

		phoneT, ok := paraMapT["phone"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`PHONE='` + sqltk.FormatSQLValue(phoneT) + `'`)
		}

		faxT, ok := paraMapT["fax"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`FAX='` + sqltk.FormatSQLValue(faxT) + `'`)
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		orgCodeT, orgCodeOk := paraMapT["orgCode"]
		if orgCodeOk {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`ORG_CODE='` + sqltk.FormatSQLValue(orgCodeT) + `'`)
		}

		orgIDT, ok := paraMapT["orgID"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`ORG_ID=` + sqltk.FormatSQLValue(orgIDT) + ``)
		} else {
			if orgCodeT != "" {
				sqlT := `select ID from ORG where APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `' and CODE='` + orgCodeT + `'`
				sqlRs8T, errT := sqltk.QueryDBString(dbT, sqlT)

				if errT == nil {
					orgIDT = sqlRs8T

					if bufT.Len() > 0 {
						bufT.WriteString(", ")
					}

					bufT.WriteString(`ORG_ID=` + sqltk.FormatSQLValue(orgIDT) + ``)
				}

			}
		}

		if !orgCodeOk && tk.Trim(orgIDT) != "" {
			tmpCodeT, errT := sqltk.QueryDBString(dbT, `SELECT CODE FROM ORG WHERE APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' AND ID=`+orgIDT)

			if errT != nil {
				tk.Pl("DB error: %v", errT)

			} else {
				orgCodeT = tmpCodeT
				if bufT.Len() > 0 {
					bufT.WriteString(", ")
				}

				bufT.WriteString(`ORG_CODE='` + sqltk.FormatSQLValue(orgCodeT) + `'`)

			}
		}

		areaCodeT, ok := paraMapT["areaCode"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`AREA_CODE='` + sqltk.FormatSQLValue(areaCodeT) + `'`)
		}

		areaIDT, ok := paraMapT["areaID"]
		if ok {
			areaIDT = tk.Trim(areaIDT)
			if areaIDT == "" {
				areaIDT = "NULL"
			}

			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`AREA_ID=` + sqltk.FormatSQLValue(areaIDT) + ``)
		} else {
			if areaCodeT != "" {
				sqlT := `select ID from ORG_GROUP where APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `' and CODE='` + areaCodeT + `'`
				sqlRs8T, errT := sqltk.QueryDBString(dbT, sqlT)

				if errT == nil {
					areaIDT = sqlRs8T

					if bufT.Len() > 0 {
						bufT.WriteString(", ")
					}

					bufT.WriteString(`AREA_ID=` + sqltk.FormatSQLValue(areaIDT) + ``)
				}

			}
		}

		positionT, ok := paraMapT["position"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`POSITION='` + sqltk.FormatSQLValue(positionT) + `'`)
		}

		addressT, ok := paraMapT["address"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`ADDRESS='` + sqltk.FormatSQLValue(addressT) + `'`)
		}

		userTypeT, ok := paraMapT["userType"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`USER_TYPE='` + sqltk.FormatSQLValue(userTypeT) + `'`)
		}

		// c1, c2, errT := sqltk.ExecV(dbT, `INSERT INTO USER (APP_CODE, USER_ID, PASSWORD, NAME, EMAIL, MOBILE, ROLE, USER_STATUS, DESCRIPTION, REMARK, RESV1, RESV2, RESV3) VALUES('`+sqltk.FormatSQLValue(appCodeT)+`', '`+sqltk.FormatSQLValue(userT)+`', '`+sqltk.FormatSQLValue(passT)+`', '`+sqltk.FormatSQLValue(userNameT)+`', '`+sqltk.FormatSQLValue(emailT)+`', '`+sqltk.FormatSQLValue(mobileT)+`', '`+sqltk.FormatSQLValue(roleT)+`', '`+sqltk.FormatSQLValue(statusT)+`', '`+sqltk.FormatSQLValue(descriptionT)+`', '`+sqltk.FormatSQLValue(remarkT)+`', '`+sqltk.FormatSQLValue(resv1T)+`', '`+sqltk.FormatSQLValue(resv2T)+`', '`+sqltk.FormatSQLValue(resv3T)+`') `)

		// if errT != nil {
		// 	return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		// }

		// if c2 < 1 {
		// 	return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "no records affected"), req)
		// }

		allowDuplicateNameT := paraMapT["allowDupName"]
		if allowDuplicateNameT != "true" && userNameT != "" {
			sqlRsT, errT := sqltk.QueryDBCount(dbT, `SELECT COUNT(*) FROM USER WHERE USER_ID='`+sqltk.FormatSQLValue(userNameT)+`' AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' AND ID<>`+userIDT)
			if errT != nil {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
			}

			if sqlRsT > 0 {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "record already exists"), req)
			}
		}

		tk.Pl("%v", `UPDATE USER SET `+bufT.String()+`, TIME_STAMP=NOW() WHERE ID=`+sqltk.FormatSQLValue(userIDT)+`;`)
		_, c2, errT := sqltk.ExecV(dbT, `UPDATE USER SET `+bufT.String()+`, TIME_STAMP=NOW() WHERE ID=`+sqltk.FormatSQLValue(userIDT)+`;`)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if c2 < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "no rows affects"), req)
		}

		_, _, errT = sqltk.ExecV(dbT, `DELETE FROM USER_ORG_RIGHT_LINK WHERE REAL_USER_ID=`+userIDT+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		linksT := tk.Trim(paraMapT["links"])

		var linksAryT []map[string]string = nil

		if linksT != "" {
			linksAryT = tk.JSONToMapStringStringArray(linksT)
		}

		if linksAryT != nil {
			for _, v := range linksAryT {
				vi := tk.StrToInt(tk.Trim(v["ID"]), -1)
				if vi < 0 {
					return tk.GenerateJSONPResponse("fail", tk.Spr("invalid links item: %v", v), req)
				}

				typeT := tk.Trim(v["Type"])

				rightNameT := tk.Trim(v["Right"])

				var sqlT string

				if typeT == "group" {
					sqlT = `INSERT INTO USER_ORG_RIGHT_LINK (APP_CODE, LINK_TYPE, REAL_USER_ID, USER_ID, ORG_GROUP_ID, RIGHT_NAME) VALUES('` + sqltk.FormatSQLValue(appCodeT) + `', 'group', ` + userIDT + `, '` + userT + `', ` + tk.IntToStr(vi) + `, '` + sqltk.FormatSQLValue(rightNameT) + `')`
				} else {
					sqlT = `INSERT INTO USER_ORG_RIGHT_LINK (APP_CODE, LINK_TYPE, REAL_USER_ID, USER_ID, ORG_ID, RIGHT_NAME) VALUES('` + sqltk.FormatSQLValue(appCodeT) + `', '', ` + userIDT + `, '` + userT + `', ` + tk.IntToStr(vi) + `, '` + sqltk.FormatSQLValue(rightNameT) + `')`
				}

				_, c2i, errT := sqltk.ExecV(dbT, sqlT)

				if errT != nil {
					return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
				}

				if c2i < 1 {
					return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed, no records affected for links item: %v", v), req)
				}

			}
		}

		return tk.GenerateJSONPResponseWithMore("success", "", req)

	case "resetPassword":
		authT := paraMapT["auth"]

		appCodeT, appCodeOK := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		var bufT strings.Builder

		if appCodeOK {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `'`)
		}

		userIDT, ok := paraMapT["userID"]
		if !ok || userIDT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty user id"), req)
		}

		// passT, ok := paraMapT["password"]
		// if ok {
		// 	if passT == "" {
		// 		return tk.GenerateJSONPResponse("fail", tk.Spr("empty password"), req)
		// 	}

		// 	if bufT.Len() > 0 {
		// 		bufT.WriteString(", ")
		// 	}

		// 	bufT.WriteString(`PASSWORD='` + sqltk.FormatSQLValue(passT) + `'`)
		// }

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		passwordT := paraMapT["newPass"]
		if passwordT == "" {
			passwordT = tk.GenerateRandomString(6, 12, true, true, true, false, false, false)
		}

		_, c2, errT := sqltk.ExecV(dbT, `UPDATE USER SET PASSWORD='`+tk.MD5Encrypt(passwordT)+`', TIME_STAMP=NOW() WHERE ID=`+sqltk.FormatSQLValue(userIDT)+`;`)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if c2 < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "no rows affects"), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", passwordT, req)

	case "setPassword":
		authT := paraMapT["auth"]

		appCodeT, appCodeOK := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		var bufT strings.Builder

		if appCodeOK {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `'`)
		}

		userIDT, ok := paraMapT["userID"]
		if !ok || userIDT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty user id"), req)
		}

		passT := paraMapT["password"]
		if passT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty password"), req)
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		_, c2, errT := sqltk.ExecV(dbT, `UPDATE USER SET PASSWORD='`+passT+`', TIME_STAMP=NOW() WHERE ID=`+sqltk.FormatSQLValue(userIDT)+`;`)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if c2 < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "no rows affects"), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", "", req)

	case "disableUser":
		authT := paraMapT["auth"]

		appCodeT, appCodeOK := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		var bufT strings.Builder

		if appCodeOK {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `'`)
		}

		userIDT, ok := paraMapT["userID"]
		if !ok || userIDT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty user id"), req)
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		_, c2, errT := sqltk.ExecV(dbT, `UPDATE USER SET USER_STATUS='disabled', TIME_STAMP=NOW() WHERE ID=`+sqltk.FormatSQLValue(userIDT)+`;`)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if c2 < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "no rows affects"), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", "", req)

	case "enableUser":
		authT := paraMapT["auth"]

		appCodeT, appCodeOK := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		var bufT strings.Builder

		if appCodeOK {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `'`)
		}

		userIDT, ok := paraMapT["userID"]
		if !ok || userIDT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty user id"), req)
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		_, c2, errT := sqltk.ExecV(dbT, `UPDATE USER SET USER_STATUS='', TIME_STAMP=NOW() WHERE ID=`+sqltk.FormatSQLValue(userIDT)+`;`)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if c2 < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "no rows affects"), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", "", req)

	case "removeUser":
		authT := paraMapT["auth"]

		appCodeT, appCodeOK := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		var bufT strings.Builder

		if appCodeOK {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `'`)
		}

		userIDT, ok := paraMapT["userID"]
		if !ok || userIDT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty user id"), req)
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		_, c2, errT := sqltk.ExecV(dbT, `DELETE FROM USER WHERE ID=`+sqltk.FormatSQLValue(userIDT)+`;`)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if c2 < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "no rows affects"), req)
		}

		_, c2, errT = sqltk.ExecV(dbT, `DELETE FROM USER_ORG_RIGHT_LINK WHERE REAL_USER_ID=`+sqltk.FormatSQLValue(userIDT)+`;`)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", "", req)

	case "addOrg":
		authT := paraMapT["auth"]

		appCodeT := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		orgNameT := paraMapT["orgName"]
		if orgNameT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty org name"), req)
		}

		orgCodeT := paraMapT["orgCode"]
		remarkT := paraMapT["remark"]
		descriptionT := paraMapT["description"]
		upIDT := paraMapT["upID"]
		if upIDT == "" {
			upIDT = "NULL"
		}

		typeT := paraMapT["type"]
		contactT := paraMapT["contact"]
		addressT := paraMapT["address"]
		subTypeT := paraMapT["subType"]
		emailT := paraMapT["email"]
		phoneT := paraMapT["phone"]
		mobileT := paraMapT["mobile"]
		upCodeT := paraMapT["upCode"]
		upNameT := paraMapT["upName"]

		allowDuplicateNameT := paraMapT["allowDupName"]
		if allowDuplicateNameT != "true" {
			sqlRsT, errT := sqltk.QueryDBCount(dbT, `SELECT COUNT(*) FROM ORG WHERE NAME='`+sqltk.FormatSQLValue(orgNameT)+`' AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
			if errT != nil {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
			}

			if sqlRsT > 0 {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "record already exists"), req)
			}
		}

		allowDuplicateCodeT := paraMapT["allowDupCode"]
		if (allowDuplicateCodeT != "true") || (orgCodeT != "") {
			sqlRsT, errT := sqltk.QueryDBCount(dbT, `SELECT COUNT(*) FROM ORG WHERE CODE='`+sqltk.FormatSQLValue(orgCodeT)+`' AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
			if errT != nil {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
			}

			if sqlRsT > 0 {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "record already exists"), req)
			}
		}

		sqlT := `INSERT INTO ORG (APP_CODE, NAME, CODE, DESCRIPTION, REMARK, UP_ID, TYPE, SUB_TYPE, CONTACT, EMAIL, ADDRESS, PHONE, MOBILE, UP_CODE, UP_NAME) VALUES('` + sqltk.FormatSQLValue(appCodeT) + `', '` + sqltk.FormatSQLValue(orgNameT) + `', '` + sqltk.FormatSQLValue(orgCodeT) + `', '` + sqltk.FormatSQLValue(descriptionT) + `', '` + sqltk.FormatSQLValue(remarkT) + `', ` + upIDT + `, '` + sqltk.FormatSQLValue(typeT) + `', '` + sqltk.FormatSQLValue(subTypeT) + `', '` + sqltk.FormatSQLValue(contactT) + `', '` + sqltk.FormatSQLValue(emailT) + `', '` + sqltk.FormatSQLValue(addressT) + `', '` + sqltk.FormatSQLValue(phoneT) + `', '` + sqltk.FormatSQLValue(mobileT) + `', '` + sqltk.FormatSQLValue(upCodeT) + `', '` + sqltk.FormatSQLValue(upNameT) + `') `
		tk.Pl("%v", sqlT)

		c1, c2, errT := sqltk.ExecV(dbT, sqlT)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if c2 < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "no records affected"), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", tk.Int64ToStr(c1), req)

	case "addOrgWithUpCode":
		authT := paraMapT["auth"]

		appCodeT := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		orgNameT := paraMapT["orgName"]
		if orgNameT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty org name"), req)
		}

		orgCodeT := paraMapT["orgCode"]
		remarkT := paraMapT["remark"]
		descriptionT := paraMapT["description"]
		upIDT := paraMapT["upID"]
		if upIDT == "" {
			upIDT = "NULL"
		}

		typeT := paraMapT["type"]
		contactT := paraMapT["contact"]
		addressT := paraMapT["address"]
		subTypeT := paraMapT["subType"]
		emailT := paraMapT["email"]
		phoneT := paraMapT["phone"]
		mobileT := paraMapT["mobile"]
		upCodeT := tk.Trim(paraMapT["upCode"])
		if upCodeT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty upCode"), req)
		}
		upNameT := paraMapT["upName"]

		allowDuplicateNameT := paraMapT["allowDupName"]
		if allowDuplicateNameT != "true" {
			sqlRsT, errT := sqltk.QueryDBCount(dbT, `SELECT COUNT(*) FROM ORG WHERE NAME='`+sqltk.FormatSQLValue(orgNameT)+`' AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
			if errT != nil {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
			}

			if sqlRsT > 0 {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "record already exists"), req)
			}
		}

		allowDuplicateCodeT := paraMapT["allowDupCode"]
		if (allowDuplicateCodeT != "true") || (orgCodeT != "") {
			sqlRsT, errT := sqltk.QueryDBCount(dbT, `SELECT COUNT(*) FROM ORG WHERE CODE='`+sqltk.FormatSQLValue(orgCodeT)+`' AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
			if errT != nil {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
			}

			if sqlRsT > 0 {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "record already exists"), req)
			}
		}

		sqlRsT, errT := sqltk.QueryDBNSS(dbT, `SELECT ID, NAME FROM ORG_GROUP WHERE APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' AND CODE='`+sqltk.FormatSQLValue(upCodeT)+`'`)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if len(sqlRsT) < 2 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "up code record not exists"), req)
		}

		upIDT = sqlRsT[1][0]
		upNameT = sqlRsT[1][1]

		c1, c2, errT := sqltk.ExecV(dbT, `INSERT INTO ORG (APP_CODE, NAME, CODE, DESCRIPTION, REMARK, UP_ID, TYPE, SUB_TYPE, CONTACT, EMAIL, ADDRESS, PHONE, MOBILE, UP_CODE, UP_NAME) VALUES('`+sqltk.FormatSQLValue(appCodeT)+`', '`+sqltk.FormatSQLValue(orgNameT)+`', '`+sqltk.FormatSQLValue(orgCodeT)+`', '`+sqltk.FormatSQLValue(descriptionT)+`', '`+sqltk.FormatSQLValue(remarkT)+`', `+upIDT+`, '`+sqltk.FormatSQLValue(typeT)+`', '`+sqltk.FormatSQLValue(subTypeT)+`', '`+sqltk.FormatSQLValue(contactT)+`', '`+sqltk.FormatSQLValue(emailT)+`', '`+sqltk.FormatSQLValue(addressT)+`', '`+sqltk.FormatSQLValue(phoneT)+`', '`+sqltk.FormatSQLValue(mobileT)+`', '`+sqltk.FormatSQLValue(upCodeT)+`', '`+sqltk.FormatSQLValue(upNameT)+`') `)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if c2 < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "no records affected"), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", tk.Int64ToStr(c1), req)

	case "addOrgGroup":
		authT := paraMapT["auth"]

		appCodeT := tk.Trim(paraMapT["appCode"])

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		groupNameT := tk.Trim(paraMapT["groupName"])
		if groupNameT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty org group name"), req)
		}

		groupCodeT := tk.Trim(paraMapT["groupCode"])

		typeT := tk.Trim(paraMapT["type"])
		levelT := tk.Trim(paraMapT["level"])
		levelIndexT := tk.Trim(paraMapT["levelIndex"])
		if levelIndexT == "" {
			levelIndexT = "NULL"
		}
		upIDT := tk.Trim(paraMapT["upID"])
		if upIDT == "" {
			upIDT = "NULL"
		}
		upCodeT := tk.Trim(paraMapT["upCode"])

		remarkT := paraMapT["remark"]
		descriptionT := paraMapT["description"]
		relT := paraMapT["rel"]
		if relT == "" {
			relT = "NULL"
		} else {
			relT = `'` + sqltk.FormatSQLValue(relT) + `'`
		}

		linksT := tk.Trim(paraMapT["links"])
		linksTypeT := strings.ToLower(tk.Trim(paraMapT["linksType"]))

		allowDuplicateNameT := paraMapT["allowDupName"]
		if allowDuplicateNameT != "true" {
			sqlRsT, errT := sqltk.QueryDBCount(dbT, `SELECT COUNT(*) FROM ORG_GROUP WHERE NAME='`+sqltk.FormatSQLValue(groupNameT)+`' AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
			if errT != nil {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
			}

			if sqlRsT > 0 {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "record with same name already exists"), req)
			}
		}

		allowDuplicateCodeT := paraMapT["allowDupCode"]
		if (allowDuplicateCodeT != "true") && (groupCodeT != "") {
			sqlRsT, errT := sqltk.QueryDBCount(dbT, `SELECT COUNT(*) FROM ORG_GROUP WHERE CODE='`+sqltk.FormatSQLValue(groupCodeT)+`' AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
			if errT != nil {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
			}

			if sqlRsT > 0 {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "record with same code already exists"), req)
			}
		}

		c1, c2, errT := sqltk.ExecV(dbT, `INSERT INTO ORG_GROUP (APP_CODE, NAME, CODE, TYPE, UP_ID, UP_CODE, LEVEL, LEVEL_INDEX, DESCRIPTION, REMARK, REL) VALUES('`+sqltk.FormatSQLValue(appCodeT)+`', '`+sqltk.FormatSQLValue(groupNameT)+`', '`+sqltk.FormatSQLValue(groupCodeT)+`', '`+sqltk.FormatSQLValue(typeT)+`', `+upIDT+`, '`+sqltk.FormatSQLValue(upCodeT)+`', '`+sqltk.FormatSQLValue(levelT)+`', `+levelIndexT+`, '`+sqltk.FormatSQLValue(descriptionT)+`', '`+sqltk.FormatSQLValue(remarkT)+`', `+relT+`) `)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if c2 < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "no records affected"), req)
		}

		var linksAryT []string = nil

		if linksT != "" {
			linksAryT = tk.JSONToStringArray(linksT)
		}

		if linksAryT != nil {
			for _, v := range linksAryT {
				vi := tk.StrToInt(v, -1)
				if vi < 0 {
					return tk.GenerateJSONPResponse("fail", tk.Spr("invalid links item: %v", v), req)
				}

				var sqlT string

				if linksTypeT == "group-group" {
					sqlT = `INSERT INTO ORG_GROUP_GROUP_LINK (APP_CODE, UP_GROUP_ID, GROUP_ID) VALUES('` + sqltk.FormatSQLValue(appCodeT) + `', ` + tk.Int64ToStr(c1) + `, ` + tk.IntToStr(vi) + `)`
				} else {
					sqlT = `INSERT INTO ORG_GROUP_LINK (APP_CODE, ORG_GROUP_ID, ORG_ID) VALUES('` + sqltk.FormatSQLValue(appCodeT) + `', ` + tk.Int64ToStr(c1) + `, ` + tk.IntToStr(vi) + `)`
				}
				_, c2i, errT := sqltk.ExecV(dbT, sqlT)

				if errT != nil {
					return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
				}

				if c2i < 1 {
					return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed, no records affected for links item: %v", v), req)
				}

			}
		}

		return tk.GenerateJSONPResponseWithMore("success", tk.Int64ToStr(c1), req)

	case "removeOrgs":
		authT := paraMapT["auth"]

		appCodeT := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		orgListStrT := paraMapT["orgList"]
		if orgListStrT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty org list"), req)
		}

		var orgListT []string

		errT = tk.LoadJSONFromString(orgListStrT, &orgListT)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("invalid org id list"), req)
		}

		failListT := make([]string, 0)

		for _, v := range orgListT {
			_, c2, errT := sqltk.ExecV(dbT, `DELETE FROM ORG WHERE ID=`+tk.Trim(v)+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)

			if errT != nil {
				tk.Pl("failed to remove org: %v(%v)", errT, `DELETE FROM ORG WHERE ID=`+tk.Trim(v)+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
				failListT = append(failListT, v)
				continue
			}

			if c2 < 1 {
				tk.Pl("failed to remove org: %v(%v)", "no rows infected", `DELETE FROM ORG WHERE ID=`+tk.Trim(v)+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
				failListT = append(failListT, v)
				continue
			}

			_, _, errT = sqltk.ExecV(dbT, `DELETE FROM ORG_GROUP_LINK WHERE ORG_ID=`+tk.Trim(v)+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)

			if errT != nil {
				tk.Pl("failed to remove org: %v(%v)", errT, `DELETE FROM ORG_GROUP_LINK WHERE ORG_ID=`+tk.Trim(v)+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
				failListT = append(failListT, v)
				continue
			}

		}

		if len(failListT) > 0 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed for the following ORG(s): %v", strings.Join(failListT, ",")), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", tk.IntToStr(len(orgListT)), req)

	case "removeOrgGroups":
		authT := paraMapT["auth"]

		appCodeT := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		groupListStrT := paraMapT["groupList"]
		if groupListStrT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty org group list"), req)
		}

		var groupListT []string

		errT = tk.LoadJSONFromString(groupListStrT, &groupListT)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("invalid org group id list"), req)
		}

		failListT := make([]string, 0)

		for _, v := range groupListT {
			_, c2, errT := sqltk.ExecV(dbT, `DELETE FROM ORG_GROUP WHERE ID=`+tk.Trim(v)+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)

			if errT != nil {
				tk.Pl("failed to remove org: %v(%v)", errT, `DELETE FROM ORG WHERE ID=`+tk.Trim(v)+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
				failListT = append(failListT, v)
				continue
			}

			if c2 < 1 {
				tk.Pl("failed to remove org: %v(%v)", "no rows infected", `DELETE FROM ORG_GROUP WHERE ID=`+tk.Trim(v)+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
				failListT = append(failListT, v)
				continue
			}

			_, c2, errT = sqltk.ExecV(dbT, `DELETE FROM ORG_GROUP_LINK WHERE ORG_GROUP_ID=`+tk.Trim(v)+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)

			if errT != nil {
				tk.Pl("failed to remove org: %v(%v)", errT, `DELETE FROM ORG_GROUP_LINK WHERE ORG_GROUP_ID=`+tk.Trim(v)+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
				failListT = append(failListT, v)
				continue
			}

			_, c2, errT = sqltk.ExecV(dbT, `DELETE FROM ORG_GROUP_GROUP_LINK WHERE GROUP_ID=`+tk.Trim(v)+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)

			if errT != nil {
				tk.Pl("failed to remove org: %v(%v)", errT, `DELETE FROM ORG_GROUP_GROUP_LINK WHERE GROUP_ID=`+tk.Trim(v)+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
				failListT = append(failListT, v)
				continue
			}

			_, c2, errT = sqltk.ExecV(dbT, `DELETE FROM ORG_GROUP_GROUP_LINK WHERE UP_GROUP_ID=`+tk.Trim(v)+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)

			if errT != nil {
				tk.Pl("failed to remove org: %v(%v)", errT, `DELETE FROM ORG_GROUP_GROUP_LINK WHERE UP_GROUP_ID=`+tk.Trim(v)+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
				failListT = append(failListT, v)
				continue
			}

			// if c2 < 1 {
			// 	tk.Pl("failed to remove org: %v(%v)", "no rows infected", `DELETE FROM ORG_GROUP_LINK WHERE ORG_GROUP_ID=`+tk.Trim(v)+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
			// 	failListT = append(failListT, v)
			// 	continue
			// }
		}

		if len(failListT) > 0 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed for the following ORG group(s): %v", strings.Join(failListT, ",")), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", tk.IntToStr(len(groupListT)), req)

	case "removeOrgGroup":
		authT := paraMapT["auth"]

		appCodeT := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		groupIDT := tk.Trim(paraMapT["groupID"])
		if groupIDT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty org group id"), req)
		}

		_, c2, errT := sqltk.ExecV(dbT, `DELETE FROM ORG_GROUP WHERE ID=`+groupIDT+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to remove org group: %v(%v)", groupIDT, errT), req)
		}

		if c2 < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to remove org group: %v(%v)", groupIDT, "no row affected"), req)
		}

		_, c2, errT = sqltk.ExecV(dbT, `DELETE FROM ORG_GROUP_LINK WHERE ORG_GROUP_ID=`+groupIDT+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to remove org group: %v(%v)", groupIDT, errT), req)
		}

		_, c2, errT = sqltk.ExecV(dbT, `DELETE FROM ORG_GROUP_GROUP_LINK WHERE GROUP_ID=`+groupIDT+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to remove org group link: %v(%v)", groupIDT, errT), req)
		}

		_, c2, errT = sqltk.ExecV(dbT, `DELETE FROM ORG_GROUP_GROUP_LINK WHERE UP_GROUP_ID=`+groupIDT+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to remove org group link: %v(%v)", groupIDT, errT), req)
		}

		// if c2 < 1 {
		// 	return tk.GenerateJSONPResponse("fail", tk.Spr("failed to remove org group: %v(%v)", groupIDT, "no records affected"), req)
		// }

		return tk.GenerateJSONPResponseWithMore("success", "1", req)

	case "removeOrg":
		authT := paraMapT["auth"]

		appCodeT := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		orgIDT := tk.Trim(paraMapT["orgID"])
		if orgIDT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty org id"), req)
		}

		_, c2, errT := sqltk.ExecV(dbT, `DELETE FROM ORG WHERE ID=`+orgIDT+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to remove org: %v(%v)", orgIDT, errT), req)
		}

		if c2 < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to remove org: %v(%v)", orgIDT, "no row affected"), req)
		}

		_, _, errT = sqltk.ExecV(dbT, `DELETE FROM ORG_GROUP_LINK WHERE ORG_ID=`+orgIDT+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to remove org link: %v(%v)", orgIDT, errT), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", "1", req)

	case "clearOrgs":
		authT := paraMapT["auth"]

		appCodeT := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		typeT := tk.Trim(paraMapT["type"])

		var sqlT string

		if typeT != "" {
			tk.Plvx(typeT)
			sqlT = `DELETE FROM ORG WHERE APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `' AND TYPE='` + sqltk.FormatSQLValue(typeT) + `'`
		} else {
			sqlT = `DELETE FROM ORG WHERE APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `'`
		}

		_, _, errT = sqltk.ExecV(dbT, sqlT)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		// if c2 < 1 {
		// 	return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "no rows affects"), req)
		// }

		_, _, errT = sqltk.ExecV(dbT, `DELETE FROM ORG_GROUP_LINK WHERE APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' AND ORG_ID not exists(select ID from ORG) `)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		// if c2 < 1 {
		// 	return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "no rows affects"), req)
		// }

		return tk.GenerateJSONPResponseWithMore("success", "", req)

	case "clearOrgGroups":
		authT := paraMapT["auth"]

		appCodeT := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		levelIndexT := tk.Trim(paraMapT["levelIndex"])

		var sqlT string
		var sql2T string

		if levelIndexT != "" {
			sqlT = `DELETE FROM ORG_GROUP WHERE APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `' AND LEVEL_INDEX=` + sqltk.FormatSQLValue(levelIndexT) + ``
			sql2T = `SELECT ID FROM ORG_GROUP WHERE APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `' AND LEVEL_INDEX=` + sqltk.FormatSQLValue(levelIndexT) + ``
		} else {
			sqlT = `DELETE FROM ORG_GROUP WHERE APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `'`
			sql2T = `SELECT ID FROM ORG_GROUP WHERE APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `'`
		}

		sql3T := `DELETE FROM ORG_GROUP_GROUP_LINK WHERE APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `' AND (GROUP_ID IN (` + sql2T + `) OR UP_GROUP_ID IN (` + sql2T + `))`

		_, _, errT = sqltk.ExecV(dbT, sql3T)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		_, _, errT = sqltk.ExecV(dbT, sqlT)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		// if c2 < 1 {
		// 	return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "no rows affects"), req)
		// }

		// if c2 < 1 {
		// 	return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "no rows affects"), req)
		// }

		return tk.GenerateJSONPResponseWithMore("success", "", req)

	case "modifyOrg":
		authT := paraMapT["auth"]

		appCodeT, appCodeOK := paraMapT["appCode"]
		appCodeT = tk.Trim(appCodeT)

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		orgIDT := tk.Trim(paraMapT["orgID"])
		if orgIDT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty org id"), req)
		}

		var bufT strings.Builder

		if appCodeOK {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `'`)
		}

		orgNameT, ok := paraMapT["orgName"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`NAME='` + sqltk.FormatSQLValue(orgNameT) + `'`)
		}

		orgCodeT, ok := paraMapT["orgCode"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`CODE='` + sqltk.FormatSQLValue(orgCodeT) + `'`)
		}

		remarkT, ok := paraMapT["remark"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`REMARK='` + sqltk.FormatSQLValue(remarkT) + `'`)
		}

		descriptionT, ok := paraMapT["description"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`DESCRIPTION='` + sqltk.FormatSQLValue(descriptionT) + `'`)
		}

		typeT, ok := paraMapT["type"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`TYPE='` + sqltk.FormatSQLValue(typeT) + `'`)
		}

		subTypeT, ok := paraMapT["subType"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`SUB_TYPE='` + sqltk.FormatSQLValue(subTypeT) + `'`)
		}

		contactT, ok := paraMapT["contact"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`CONTACT='` + sqltk.FormatSQLValue(contactT) + `'`)
		}

		emailT, ok := paraMapT["email"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`EMAIL='` + sqltk.FormatSQLValue(emailT) + `'`)
		}

		phoneT, ok := paraMapT["phone"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`PHONE='` + sqltk.FormatSQLValue(phoneT) + `'`)
		}

		mobileT, ok := paraMapT["mobile"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`MOBILE='` + sqltk.FormatSQLValue(mobileT) + `'`)
		}

		addressT, ok := paraMapT["address"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`ADDRESS='` + sqltk.FormatSQLValue(addressT) + `'`)
		}

		upCodeT, ok := paraMapT["upCode"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`UP_CODE='` + sqltk.FormatSQLValue(upCodeT) + `'`)
		}

		upNameT, ok := paraMapT["upName"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`UP_NAME='` + sqltk.FormatSQLValue(upNameT) + `'`)
		}

		upIDT, ok := paraMapT["upID"]
		if upIDT == "" {
			upIDT = "NULL"
		}
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`UP_ID=` + sqltk.FormatSQLValue(upIDT) + ``)
		}

		relT, ok := paraMapT["rel"]
		if relT == "" {
			relT = "NULL"
		} else {
			relT = `'` + sqltk.FormatSQLValue(relT) + `'`
		}
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`REL='` + sqltk.FormatSQLValue(relT) + `'`)
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		allowDuplicateNameT := paraMapT["allowDupName"]
		if allowDuplicateNameT != "true" {
			sqlRsT, errT := sqltk.QueryDBCount(dbT, `SELECT COUNT(*) FROM ORG WHERE NAME='`+sqltk.FormatSQLValue(orgNameT)+`' AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' AND ID<>`+orgIDT)
			if errT != nil {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
			}

			if sqlRsT > 0 {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "record already exists"), req)
			}
		}

		allowDuplicateCodeT := paraMapT["allowDupCode"]
		if (allowDuplicateCodeT != "true") || (orgCodeT != "") {
			sqlRsT, errT := sqltk.QueryDBCount(dbT, `SELECT COUNT(*) FROM ORG WHERE CODE='`+sqltk.FormatSQLValue(orgCodeT)+`' AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' AND ID<>`+orgIDT)
			if errT != nil {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
			}

			if sqlRsT > 0 {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "record already exists"), req)
			}
		}

		_, c2, errT := sqltk.ExecV(dbT, `UPDATE ORG SET `+bufT.String()+`, TIME_STAMP=NOW() WHERE ID=`+sqltk.FormatSQLValue(orgIDT)+``)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if c2 < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "no rows affects"), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", "", req)

	case "modifyOrgGroup":
		authT := paraMapT["auth"]

		appCodeT, appCodeOK := paraMapT["appCode"]
		appCodeT = tk.Trim(appCodeT)

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		groupIDT := tk.Trim(paraMapT["groupID"])
		if groupIDT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty org group id"), req)
		}

		var bufT strings.Builder

		if appCodeOK {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `'`)
		}

		groupNameT, ok := paraMapT["groupName"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`NAME='` + sqltk.FormatSQLValue(groupNameT) + `'`)
		}

		groupCodeT, ok := paraMapT["groupCode"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`CODE='` + sqltk.FormatSQLValue(groupCodeT) + `'`)
		}

		remarkT, ok := paraMapT["remark"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`REMARK='` + sqltk.FormatSQLValue(remarkT) + `'`)
		}

		descriptionT, ok := paraMapT["description"]
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`DESCRIPTION='` + sqltk.FormatSQLValue(descriptionT) + `'`)
		}

		relT, ok := paraMapT["rel"]
		if relT == "" {
			relT = "NULL"
		} else {
			relT = `'` + sqltk.FormatSQLValue(relT) + `'`
		}
		if ok {
			if bufT.Len() > 0 {
				bufT.WriteString(", ")
			}

			bufT.WriteString(`REL='` + sqltk.FormatSQLValue(relT) + `'`)
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		allowDuplicateNameT := paraMapT["allowDupName"]
		if allowDuplicateNameT != "true" {
			sqlRsT, errT := sqltk.QueryDBCount(dbT, `SELECT COUNT(*) FROM ORG_GROUP WHERE NAME='`+sqltk.FormatSQLValue(groupNameT)+`' AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' AND ID<>`+groupIDT)
			if errT != nil {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
			}

			if sqlRsT > 0 {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "record already exists"), req)
			}
		}

		allowDuplicateCodeT := paraMapT["allowDupCode"]
		if (allowDuplicateCodeT != "true") && (groupCodeT != "") {
			sqlRsT, errT := sqltk.QueryDBCount(dbT, `SELECT COUNT(*) FROM ORG_GROUP WHERE CODE='`+sqltk.FormatSQLValue(groupCodeT)+`' AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' AND ID<>`+groupIDT)
			if errT != nil {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
			}

			if sqlRsT > 0 {
				return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", "record already exists"), req)
			}
		}

		c1, c2, errT := sqltk.ExecV(dbT, `UPDATE ORG_GROUP SET `+bufT.String()+`, TIME_STAMP=NOW() WHERE ID=`+sqltk.FormatSQLValue(groupIDT)+`;`)

		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if c2 < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v (%v)", "no rows affects"+tk.Int64ToStr(c1), `UPDATE ORG_GROUP SET `+bufT.String()+`, TIME_STAMP=NOW() WHERE ID=`+sqltk.FormatSQLValue(groupIDT)+`;`), req)
		}

		_, _, errT = sqltk.ExecV(dbT, `DELETE FROM ORG_GROUP_LINK WHERE ORG_GROUP_ID=`+groupIDT+` AND APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		linksT := tk.Trim(paraMapT["links"])
		linksTypeT := strings.ToLower(tk.Trim(paraMapT["linksType"]))

		var linksAryT []string = nil

		if linksT != "" {
			linksAryT = tk.JSONToStringArray(linksT)
		}

		if linksAryT != nil {
			if linksTypeT == "group-group" {
				_, _, errT := sqltk.ExecV(dbT, `DELETE FROM ORG_GROUP_GROUP_LINK WHERE APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' AND UP_GROUP_ID='`+groupIDT+`'`)
				if errT != nil {
					return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
				}

			} else {
				_, _, errT := sqltk.ExecV(dbT, `DELETE FROM ORG_GROUP_LINK WHERE APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' AND ORG_GROUP_ID='`+groupIDT+`'`)
				if errT != nil {
					return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
				}

			}

			for _, v := range linksAryT {
				vi := tk.StrToInt(v, -1)
				if vi < 0 {
					return tk.GenerateJSONPResponse("fail", tk.Spr("invalid links item: %v", v), req)
				}

				var sqlT string

				if linksTypeT == "group-group" {
					sqlT = `INSERT INTO ORG_GROUP_GROUP_LINK (APP_CODE, UP_GROUP_ID, GROUP_ID) VALUES('` + sqltk.FormatSQLValue(appCodeT) + `', ` + groupIDT + `, ` + tk.IntToStr(vi) + `)`
				} else {
					sqlT = `INSERT INTO ORG_GROUP_LINK (APP_CODE, ORG_GROUP_ID, ORG_ID) VALUES('` + sqltk.FormatSQLValue(appCodeT) + `', ` + groupIDT + `, ` + tk.IntToStr(vi) + `)`
				}
				_, c2i, errT := sqltk.ExecV(dbT, sqlT)

				if errT != nil {
					return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
				}

				if c2i < 1 {
					return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed, no records affected for links item: %v", v), req)
				}

			}
		}

		return tk.GenerateJSONPResponseWithMore("success", "", req)

	case "getOrgList":
		authT := paraMapT["auth"]

		appCodeT := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		typeT := tk.Trim(paraMapT["type"])

		var sqlT string

		if typeT != "" {
			sqlT = `SELECT * FROM ORG WHERE APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `' WHERE TYPE='` + sqltk.FormatSQLValue(typeT) + `'`
		} else {
			sqlT = `SELECT * FROM ORG WHERE APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `'`
		}

		sqlRsT, errT := sqltk.QueryDBNSS(dbT, sqlT)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if len(sqlRsT) < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action length failed: %v", len(sqlRsT)), req)
		}

		if tk.ToLower(paraMapT["format"]) == "mss" {
			return tk.GenerateJSONPResponseWithMore("success", tk.TableToMSSJSON(sqlRsT), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", tk.ObjectToJSON(sqlRsT), req)

	case "getOrgListByUpID":
		authT := paraMapT["auth"]

		appCodeT := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		upIDT := tk.Trim(paraMapT["upID"])
		if upIDT == "" {
			return tk.GenerateJSONPResponse("fail", "empty up id", req)
		}

		typeT := tk.Trim(paraMapT["type"])

		var sqlT string

		if typeT != "" {
			sqlT = `SELECT * FROM ORG WHERE APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `' WHERE TYPE='` + sqltk.FormatSQLValue(typeT) + `'` + ` AND UP_ID=` + upIDT
		} else {
			sqlT = `SELECT * FROM ORG WHERE APP_CODE='` + sqltk.FormatSQLValue(appCodeT) + `' AND UP_ID=` + upIDT
		}

		sqlRsT, errT := sqltk.QueryDBNSS(dbT, sqlT)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if len(sqlRsT) < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action length failed: %v", len(sqlRsT)), req)
		}

		if tk.ToLower(paraMapT["format"]) == "mss" {
			return tk.GenerateJSONPResponseWithMore("success", tk.TableToMSSJSON(sqlRsT), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", tk.ObjectToJSON(sqlRsT), req)

	case "getUserList":
		authT := paraMapT["auth"]

		appCodeT := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		sqlRsT, errT := sqltk.QueryDBNSS(dbT, `SELECT * FROM USER WHERE APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if len(sqlRsT) < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action length failed: %v", len(sqlRsT)), req)
		}

		if tk.ToLower(paraMapT["format"]) == "mss" {
			return tk.GenerateJSONPResponseWithMore("success", tk.TableToMSSJSON(sqlRsT), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", tk.ObjectToJSON(sqlRsT), req)

	case "getOrgGroupList":
		authT := paraMapT["auth"]

		appCodeT := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		sqlRsT, errT := sqltk.QueryDBNSS(dbT, `SELECT * FROM ORG_GROUP WHERE APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if len(sqlRsT) < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action length failed: %v", len(sqlRsT)), req)
		}

		if tk.ToLower(paraMapT["format"]) == "mss" {
			return tk.GenerateJSONPResponseWithMore("success", tk.TableToMSSJSON(sqlRsT), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", tk.ObjectToJSON(sqlRsT), req)

	case "getOrgGroupLinks":
		authT := paraMapT["auth"]

		appCodeT := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		sqlRsT, errT := sqltk.QueryDBNSS(dbT, `SELECT * FROM ORG_GROUP_LINK WHERE APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`'`)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if len(sqlRsT) < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action length failed: %v", len(sqlRsT)), req)
		}

		if tk.ToLower(paraMapT["format"]) == "mss" {
			return tk.GenerateJSONPResponseWithMore("success", tk.TableToMSSJSON(sqlRsT), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", tk.ObjectToJSON(sqlRsT), req)

	case "getOrgGroupLinksGroup":
		authT := paraMapT["auth"]

		appCodeT := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		sqlRsT, errT := sqltk.QueryDBNSS(dbT, `SELECT c.APP_CODE, a.ORG_GROUP_ID, c.NAME as ORG_GROUP_NAME, c.CODE as ORG_GROUP_CODE, c.REMARK as REMARK, group_concat(b.NAME order by b.NAME DESC SEPARATOR ',') as ORGS from ORG_GROUP_LINK a, ORG b, ORG_GROUP c WHERE c.APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' and a.ORG_ID=b.ID and a.ORG_GROUP_ID=c.ID group by a.ORG_GROUP_ID`)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if len(sqlRsT) < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action length failed: %v", len(sqlRsT)), req)
		}

		if tk.ToLower(paraMapT["format"]) == "mss" {
			return tk.GenerateJSONPResponseWithMore("success", tk.TableToMSSJSON(sqlRsT), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", tk.ObjectToJSON(sqlRsT), req)

	case "getOrg":
		authT := paraMapT["auth"]

		appCodeT := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		orgIDT := tk.Trim(paraMapT["orgID"])
		if orgIDT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty org id"), req)
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		sqlRsT, errT := sqltk.QueryDBNSS(dbT, `SELECT * FROM ORG WHERE ID='`+sqltk.FormatSQLValue(orgIDT)+`'`)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if len(sqlRsT) < 2 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("no record"), req)
		}

		if tk.ToLower(paraMapT["format"]) == "mss" {
			return tk.GenerateJSONPResponseWithMore("success", tk.ToJSONX(sqltk.OneLineRecordToMap(sqlRsT), "-default=", "-sort"), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", tk.ObjectToJSON(sqlRsT), req)

	case "getOrgGroup":
		authT := paraMapT["auth"]

		appCodeT := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		groupIDT := tk.Trim(paraMapT["groupID"])
		if groupIDT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty group id"), req)
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		sqlRsT, errT := sqltk.QueryDBNSS(dbT, `SELECT a.APP_CODE, c.ID, c.NAME as ORG_GROUP_NAME, c.CODE as ORG_GROUP_CODE, c.REMARK as REMARK, group_concat(b.ID order by b.ID DESC SEPARATOR ',') as SUB_GROUPS, group_concat(b.NAME order by b.NAME DESC SEPARATOR ',') as SUB_GROUP_NAMES from ORG_GROUP_GROUP_LINK a, ORG_GROUP b, ORG_GROUP c WHERE a.APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' and a.UP_GROUP_ID=`+sqltk.FormatSQLValue(groupIDT)+` and a.GROUP_ID=b.ID and a.UP_GROUP_ID=c.ID and c.LEVEL_INDEX=2 group by c.ID`)

		// sqlRsT, errT := sqltk.QueryDBNSS(dbT, `SELECT c.APP_CODE, a.ORG_GROUP_ID, c.NAME as ORG_GROUP_NAME, c.CODE as ORG_GROUP_CODE, c.REMARK as REMARK, group_concat(b.NAME order by b.NAME DESC SEPARATOR ',') as ORGS, group_concat(b.ID order by b.ID SEPARATOR ',') as ORGIDS from ORG_GROUP_LINK a, ORG b, ORG_GROUP c WHERE c.APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' and a.ORG_GROUP_ID=`+sqltk.FormatSQLValue(groupIDT)+` and a.ORG_ID=b.ID and a.ORG_GROUP_ID=c.ID group by a.ORG_GROUP_ID`)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if len(sqlRsT) < 2 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("no record"), req)
		}

		if tk.ToLower(paraMapT["format"]) == "mss" {
			return tk.GenerateJSONPResponseWithMore("success", tk.ToJSONX(sqltk.OneLineRecordToMap(sqlRsT), "-default=", "-sort"), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", tk.ObjectToJSON(sqlRsT), req)

	case "getUserInfo":
		authT := paraMapT["auth"]

		appCodeT := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		userIDT := tk.Trim(paraMapT["userID"])
		if userIDT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty user id"), req)
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		sqlRsT, errT := sqltk.QueryDBNSS(dbT, `SELECT ID, APP_CODE, ID_TYPE, USER_ID, NAME, EMAIL, MOBILE, ROLE, RIGHTS, USER_STATUS, REMARK from USER WHERE APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' AND ID=`+userIDT+``)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if len(sqlRsT) < 2 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("no record"), req)
		}

		if tk.ToLower(paraMapT["format"]) == "mss" {
			return tk.GenerateJSONPResponseWithMore("success", tk.ToJSONX(sqltk.OneLineRecordToMap(sqlRsT), "-default=", "-sort"), req)
		}

		return tk.GenerateJSONPResponseWithMore("success", tk.ObjectToJSON(sqlRsT), req)

	case "getUserInfoX":
		authT := paraMapT["auth"]

		appCodeT := paraMapT["appCode"]

		if checkPrimaryAuth(authT) != "" {
			if checkAuth(appCodeT, authT) != "" {
				return tk.GenerateJSONPResponse("fail", "auth failed", req)
			}
		}

		userIDT := tk.Trim(paraMapT["userID"])
		if userIDT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty user id"), req)
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		sqlRsT, errT := sqltk.QueryDBNSS(dbT, `SELECT ID, APP_CODE, ID_TYPE, USER_ID, NAME, EMAIL, MOBILE, ROLE, RIGHTS, USER_STATUS, REMARK from USER WHERE APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' AND ID=`+userIDT+``)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if len(sqlRsT) < 2 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("no record"), req)
		}

		mapT := sqltk.OneLineRecordToMap(sqlRsT)

		sqlRsT, errT = sqltk.QueryDBNSS(dbT, `SELECT * from USER_ORG_RIGHT_LINK WHERE APP_CODE='`+sqltk.FormatSQLValue(appCodeT)+`' AND REAL_USER_ID=`+userIDT+``)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("DB action failed: %v", errT), req)
		}

		if len(sqlRsT) < 1 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("no record"), req)
		}

		mapT["Links"] = tk.ToJSONX(sqltk.RecordsToMapArray(sqlRsT))

		return tk.GenerateJSONPResponseWithMore("success", tk.ToJSONX(mapT, "-default=", "-sort"), req)

	case "login":
		authT := paraMapT["auth"]
		appCodeT := paraMapT["appCode"]

		authRsT := checkAuth(appCodeT, authT)

		if authRsT != "" {
			return tk.GenerateJSONPResponse("fail", "auth failed: "+authRsT, req)
		}

		dbT, errT := sqltk.ConnectDBNoPing(configG.DBType, configG.DBConnectString)
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("failed to connect DB: %v", errT), req)
		}

		defer dbT.Close()

		userT := paraMapT["user"]
		if userT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty user id"), req)
		}

		passT := paraMapT["password"]
		if passT == "" {
			return tk.GenerateJSONPResponse("fail", tk.Spr("empty password"), req)
		}

		sqlRsT, errT := sqltk.QueryDBNSS(dbT, "select * from USER where APP_CODE='"+tk.Replace(appCodeT, "'", "''")+"' and USER_ID='"+tk.Replace(userT, "'", "''")+"'")
		if errT != nil {
			return tk.GenerateJSONPResponse("fail", tk.Spr("user %v not exists: %v", userT, errT), req)
		}

		tk.Pl("%v", sqlRsT)

		if len(sqlRsT) < 2 {
			return tk.GenerateJSONPResponse("fail", tk.Spr("user %v not exists", userT), req)
		}

		userMapT := sqltk.OneLineRecordToMap(sqlRsT)

		if passT != userMapT["PASSWORD"] {
			tk.Plvsr(passT, userMapT["PASSWORD"])
			return tk.GenerateJSONPResponse("fail", tk.Spr("user id or password not match"), req)
		}

		delete(userMapT, "PASSWORD")

		sqlRs2T, errT := sqltk.QueryDBString(dbT, "select NAME from ORG where APP_CODE='"+tk.Replace(appCodeT, "'", "''")+"' and ID="+userMapT["ORG_ID"])
		if errT != nil {
			sqlRs2T = ""
		}

		userMapT["ORG_NAME"] = sqlRs2T

		return tk.GenerateJSONPResponseWithMore("success", tk.ToJSONWithDefault(userMapT, ""), req, "Token", generateToken(appCodeT, userT, userMapT["ROLE"]))

	default:
		return tk.GenerateJSONPResponse("fail", tk.Spr("unknown request: %v", req), req)
	}
}

func japiHandler(w http.ResponseWriter, req *http.Request) {
	rs := doJapi(w, req) //"abc"

	// println("resp: %v", rs)

	w.Write([]byte(rs))

}

var muxG *http.ServeMux

func startHttpsServer(portA string) {
	if !tk.StartsWith(portA, ":") {
		portA = ":" + portA
	}
	err := http.ListenAndServeTLS(portA, filepath.Join(basePathG, "server.crt"), filepath.Join(basePathG, "server.key"), muxG)
	if err != nil {
		tk.PlNow("failed to start https: %v", err)
	}

}

var staticFS http.Handler = nil

func serveStaticDirHandler(w http.ResponseWriter, r *http.Request) {
	if staticFS == nil {
		hdl := http.FileServer(http.Dir(filepath.Join(basePathG, "w")))
		staticFS = hdl
	}

	old := r.URL.Path

	name := filepath.Join(basePathG, "w", path.Clean(old))

	info, err := os.Lstat(name)
	if err == nil {
		if !info.IsDir() {
			staticFS.ServeHTTP(w, r)
			// http.ServeFile(w, r, name)
		} else {
			if tk.IfFileExists(filepath.Join(name, "index.html")) {
				staticFS.ServeHTTP(w, r)
			} else {
				http.NotFound(w, r)
			}
		}
	} else {
		http.NotFound(w, r)
	}

}

func cleanPlaceholders(strA string) string {
	strA = tk.RegReplace(strA, `TX_.*?_XT`, "")

	return strA
}

func replaceHtml(strA string, mapA map[string]string) string {
	if mapA == nil {
		return strA
	}

	for k, v := range mapA {
		strA = tk.Replace(strA, "TX_"+k+"_XT", v)
	}

	return strA
}

func doHttp(res http.ResponseWriter, req *http.Request) {
	if req != nil {
		req.ParseForm()
	}

	reqT := tk.GetFormValueWithDefaultValue(req, "req", "")

	if res != nil {
		res.Header().Set("Access-Control-Allow-Origin", "*")
		res.Header().Set("Access-Control-Allow-Headers", "*")
	}

	if reqT == "" {
		if tk.StartsWith(req.RequestURI, "/dp") {
			reqT = req.RequestURI[3:]
		}
	}

	tmps := tk.Split(reqT, "?")
	if len(tmps) > 1 {
		reqT = tmps[0]
	}

	tk.Pl("reqT: %v", reqT)

	toWriteT := ""

	switch reqT {

	case "test":
		{
			res.Write([]byte("test"))
			return
		}
	case "hub", "/hub":
		vo := tk.GetFormValueWithDefaultValue(req, "vo", "")

		var mapT map[string]string
		var errT error

		if vo == "" {
			mapT = tk.FormToMap(req.Form)
		} else {
			mapT, errT = tk.MSSFromJSON(vo)

			if errT != nil {
				toWriteT = tk.Spr("action failed: %v", errT.Error())
				break
			}
		}

		tmplT := tk.LoadStringFromFile(filepath.Join(basePathG, "tmpl", mapT["dest"]+".html"))

		tmplT = replaceHtml(tmplT, mapT)

		tmplT = cleanPlaceholders(tmplT)

		toWriteT = tmplT
	}

	res.Header().Set("Content-Type", "text/html; charset=utf-8")

	res.Write([]byte(toWriteT))

}

func initService() {
	tk.Pl("reading configuration...")

	errT := tk.LoadJSONFromFile(filepath.Join(basePathG, "config.json"), &configG)
	if errT != nil {
		tk.Pl("failed to read config: %v", errT)
		tk.Pl("use default config instead")

		errT = tk.LoadJSONFromString(defaultConfigG, &configG)
		if errT != nil {
			tk.Pl("failed to read default config: %v", errT)
			os.Exit(1)
		}
	}

	tokenExpireG = tk.StrToInt(configG.TokenExpire, 1440)

	tk.Plv(configG)

}

func startService() {
	initService()

	portG = tk.GetSwitch(os.Args, "-port=", portG)

	if !tk.StartsWith(portG, ":") {
		portG = ":" + portG
	}

	sslPortG = tk.GetSwitchWithDefaultValue(os.Args, "-sslPort=", sslPortG)

	if !tk.StartsWith(sslPortG, ":") {
		sslPortG = ":" + sslPortG
	}

	tk.EnsureMakeDirs(filepath.Join(basePathG, "logs"))
	tk.EnsureMakeDirs(filepath.Join(basePathG, "w"))

	tk.SetLogFile(filepath.Join(basePathG, "logs", "umx.log"))

	muxG = http.NewServeMux()

	muxG.Handle("/404", http.NotFoundHandler())

	muxG.HandleFunc("/japi", japiHandler)

	muxG.HandleFunc("/dp/", doHttp)

	muxG.HandleFunc("/", serveStaticDirHandler)

	tk.PlNow("try starting ssl server on %v...", sslPortG)

	go startHttpsServer(sslPortG)

	tk.PlNow("try starting server on %v...", portG)

	err := http.ListenAndServe(portG, muxG)

	if err != nil {
		tk.PlNow("failed to start: %v", err)
	}

}

// func runCmd(cmdA string) string {
// 	// tk.Pl("run cmd: %v", cmdA)

// 	// var errT error

// 	switch cmdA {
// 	case "version":
// 		{
// 			tk.Pl("umx " + versionG)
// 		}

// 	case "init":
// 		{
// 			initSystem()
// 		}

// 	case "run":
// 		{
// 			startService()
// 		}

// 	default:
// 		tk.Pl("unknown command: %v", cmdA)
// 	}

// 	return "exit"
// }

type program struct {
	BasePath string
}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	// basePathG = p.BasePath
	// logWithTime("basePath: %v", basePathG)
	serviceModeG = true

	go p.run()

	return nil
}

func (p *program) run() {
	go doWork()
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	return nil
}

func plByMode(formatA string, argsA ...interface{}) {
	if runModeG == "cmd" {
		tk.Pl(formatA, argsA...)
	} else {
		tk.AddDebugF(formatA, argsA...)
	}
}

func initSvc() *service.Service {
	svcConfigT := &service.Config{
		Name:        serviceNameG,
		DisplayName: serviceNameG,
		Description: serviceNameG + " V" + versionG,
	}

	prgT := &program{BasePath: basePathG}
	var s, err = service.New(prgT, svcConfigT)

	if err != nil {
		tk.LogWithTimeCompact("%s unable to start: %s\n", svcConfigT.DisplayName, err)
		return nil
	}

	return &s
}

func mainHandler(w http.ResponseWriter, req *http.Request) {
	if req != nil {
		req.ParseForm()
	}

	// reqT := tk.GetFormValueWithDefaultValue(req, "prms", "")

	plByMode("req: %+v", req)

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Test."))
	// fmt.Fprintf(w, "This is an example server.\n")
	// io.WriteString(w, "This is an example server.\n")
}

func Svc() {
	tk.SetLogFile(filepath.Join(basePathG, serviceNameG+".log"))

	defer func() {
		if v := recover(); v != nil {
			tk.LogWithTimeCompact("panic in svc %v", v)
		}
	}()

	if runModeG != "cmd" {
		runModeG = "service"
	}

	plByMode("runModeG: %v", runModeG)

	tk.DebugModeG = true

	tk.LogWithTimeCompact("%v V%v", serviceNameG, versionG)
	tk.LogWithTimeCompact("os: %v, basePathG: %v, configFileNameG: %v", runtime.GOOS, basePathG, configFileNameG)

	if tk.GetOSName() == "windows" {
		plByMode("Windows mode")
		currentOSG = "win"
		if tk.Trim(basePathG) == "" {
			basePathG = "c:\\" + serviceNameG
		}
		configFileNameG = serviceNameG + "win.cfg"
	} else {
		plByMode("Linux mode")
		currentOSG = "linux"
		if tk.Trim(basePathG) == "" {
			basePathG = "/" + serviceNameG
		}
		configFileNameG = serviceNameG + "linux.cfg"
	}

	if !tk.IfFileExists(basePathG) {
		os.MkdirAll(basePathG, 0777)
	}

	tk.SetLogFile(filepath.Join(basePathG, serviceNameG+".log"))

	cfgFileNameT := filepath.Join(basePathG, configFileNameG)
	if tk.IfFileExists(cfgFileNameT) {
		plByMode("Process config file: %v", cfgFileNameT)
		fileContentT := tk.LoadSimpleMapFromFile(cfgFileNameT)

		if fileContentT != nil {
			portG = fileContentT["port"]
			sslPortG = fileContentT["sslPort"]
			basePathG = fileContentT["crmBasePath"]
		}
	}

	plByMode("portG: %v, sslPortG: %v, basePathG: %v", portG, sslPortG, basePathG)

	tk.LogWithTimeCompact("portG: %v, sslPortG: %v, basePathG: %v", portG, sslPortG, basePathG)

	tk.LogWithTimeCompact("Service started.")
	tk.LogWithTimeCompact("Using config file: %v", cfgFileNameT)

	go startService()
}

var exitG = make(chan struct{})

func doWork() {

	go Svc()

	for {
		select {
		case <-exitG:
			os.Exit(0)
			return
		}
	}
}

func runCmd(cmdLineA []string) {
	cmdT := ""

	for _, v := range cmdLineA {
		if !strings.HasPrefix(v, "-") {
			cmdT = v
			break
		}
	}

	// if cmdT == "" {
	// 	fmt.Println("empty command")
	// 	return
	// }

	var errT error

	basePathG = tk.GetSwitchWithDefaultValue(cmdLineA, "-base=", basePathG)

	tk.EnsureMakeDirs(basePathG)

	if !tk.IfFileExists(basePathG) {
		tk.Pl("base path not exists: %v, use current directory instead", basePathG)
		basePathG, errT = filepath.Abs(".")

		if errT != nil {
			tk.Pl("failed to analyze base path")
			return
		}
		// return
	}

	if !tk.IsDirectory(basePathG) {
		tk.Pl("base path not exists: %v", basePathG)
		return
	}

	// tk.Pl("base path: %v", basePathG)

	// testPortG = tk.GetSwitchWithDefaultIntValue(cmdLineA, "-port=", 0)
	// if testPortG > 0 {
	// 	tk.Pl("test port: %v", testPortG)
	// }

	switch cmdT {
	case "version":
		tk.Pl(serviceNameG+" V%v", versionG)
		break
	case "go": // run in cmd mode
		runModeG = "cmd"
		doWork()
		break
	case "test":
		{

		}
		break
	case "init":
		{
			initSystem()
		}
	case "", "run":
		s := initSvc()

		if s == nil {
			tk.LogWithTimeCompact("Failed to init service")
			break
		}

		errT = (*s).Run()
		if errT != nil {
			tk.LogWithTimeCompact("Service \"%s\" failed to run: %v.", (*s).String(), errT)
		}
		break
	case "installonly":
		s := initSvc()

		if s == nil {
			tk.Pl("Failed to install")
			break
		}

		errT = (*s).Install()
		if errT != nil {
			tk.Pl("Failed to install: %v", errT)
			return
		}

		tk.Pl("Service \"%s\" installed.", (*s).String())

	case "install":
		s := initSvc()

		if s == nil {
			tk.Pl("Failed to install")
			break
		}

		tk.Pl("Installing service \"%v\"...", (*s).String())

		errT = (*s).Install()
		if errT != nil {
			tk.Pl("Failed to install: %v", errT)
			return
		}

		tk.Pl("Service \"%s\" installed.", (*s).String())

		tk.Pl("Starting service \"%v\"...", (*s).String())

		errT = (*s).Start()
		if errT != nil {
			tk.Pl("Failed to start: %v", errT)
			return
		}

		tk.Pl("Service \"%s\" started.", (*s).String())
	case "uninstall":
		s := initSvc()

		if s == nil {
			tk.Pl("Failed to install")
			break
		}

		errT = (*s).Stop()
		if errT != nil {
			tk.Pl("Failed to stop: %s", errT)
		} else {
			tk.Pl("Service \"%s\" stopped.", (*s).String())
		}

		errT = (*s).Uninstall()
		if errT != nil {
			tk.Pl("Failed to remove: %v", errT)
			return
		}

		tk.Pl("Service \"%s\" removed.", (*s).String())
		break
	case "reinstall":
		s := initSvc()

		if s == nil {
			tk.Pl("Failed to install")
			break
		}

		errT = (*s).Stop()
		if errT != nil {
			tk.Pl("Failed to stop: %s", errT)
		} else {
			tk.Pl("Service \"%s\" stopped.", (*s).String())
		}

		errT = (*s).Uninstall()
		if errT != nil {
			tk.Pl("Failed to remove: %v", errT)
			return
		}

		tk.Pl("Service \"%s\" removed.", (*s).String())

		errT = (*s).Install()
		if errT != nil {
			tk.Pl("Failed to install: %v", errT)
			return
		}

		tk.Pl("Service \"%s\" installed.", (*s).String())

		errT = (*s).Start()
		if errT != nil {
			tk.Pl("Failed to start: %v", errT)
			return
		}

		tk.Pl("Service \"%s\" started.", (*s).String())
	case "start":
		s := initSvc()

		if s == nil {
			tk.Pl("Failed to install")
			break
		}

		errT = (*s).Start()
		if errT != nil {
			tk.Pl("Failed to start: %v", errT)
			return
		}

		tk.Pl("Service \"%s\" started.", (*s).String())
		break
	case "stop":
		s := initSvc()

		if s == nil {
			tk.Pl("Failed to install")
			break
		}

		errT = (*s).Stop()
		if errT != nil {
			tk.Pl("Failed to stop: %v", errT)
			return
		}

		tk.Pl("Service \"%s\" stopped.", (*s).String())
		break
	default:
		tk.Pl("unknown command")
		break
	}

}

func main() {
	// var rs string

	argsT := os.Args

	if strings.HasPrefix(runtime.GOOS, "win") {
		basePathG = "c:\\" + serviceNameG
	} else {
		basePathG = "/" + serviceNameG
	}

	basePathG = tk.GetSwitch(argsT, "-base=", basePathG)

	// cmdT := tk.GetParameter(argsT, 1)

	// if !tk.IsErrStr(cmdT) {
	// 	rs = runCmd(cmdT)
	// }

	// if rs == "exit" {
	// 	os.Exit(0)
	// }

	if len(os.Args) < 2 {
		tk.Pl("%v V%v is in service(server) mode. Running the application without any arguments will cause it in service mode.\n", serviceNameG, versionG)
		serviceModeG = true

		s := initSvc()

		if s == nil {
			tk.LogWithTimeCompact("Failed to init service")
			return
		}

		err := (*s).Run()
		if err != nil {
			tk.LogWithTimeCompact("Service \"%s\" failed to run.", (*s).String())
		}

		return
	}

	if tk.GetOSName() == "windows" {
		plByMode("Windows mode")
		currentOSG = "win"
		basePathG = "c:\\" + serviceNameG
		configFileNameG = serviceNameG + "win.cfg"
	} else {
		plByMode("Linux mode")
		currentOSG = "linux"
		basePathG = "/" + serviceNameG
		configFileNameG = serviceNameG + "linux.cfg"
	}

	if !tk.IfFileExists(basePathG) {
		os.MkdirAll(basePathG, 0777)
	}

	tk.SetLogFile(filepath.Join(basePathG, serviceNameG+".log"))

	cfgFileNameT := filepath.Join(basePathG, configFileNameG)
	if tk.IfFileExists(cfgFileNameT) {
		plByMode("Process config file: %v", cfgFileNameT)
		fileContentT := tk.LoadSimpleMapFromFile(cfgFileNameT)

		if fileContentT != nil {
			portG = fileContentT["port"]
			sslPortG = fileContentT["sslPort"]
			basePathG = fileContentT["basePath"]
		}
	}

	plByMode("portG: %v, sslPortG: %v, basePathG: %v", portG, sslPortG, basePathG)

	runCmd(os.Args[1:])

}
