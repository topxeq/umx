package main

import (
	"os"
	"time"

	"net/http"
	"path"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"

	"github.com/topxeq/sqltk"
	"github.com/topxeq/tk"
)

var versionG = "0.99a"
var cfgMapG map[string]string = nil
var basePathG = "."
var portG = ":7492"
var sslPortG = ":7493"

var defaultConfigG = `
{
	"DBType": "SQLite",
	"DBConnectString": "umx.db",
	"MainSecret": "UMX_easy",
	"TokenSecret": "is_Token",
	"TestMode": "true"
}
`

var initSQLs = []string{
	`DROP TABLE APP`,
	`CREATE TABLE APP (ID int(11) NOT NULL AUTO_INCREMENT, CODE VARCHAR(100) DEFAULT NULL, SECRET VARCHAR(100) DEFAULT '', NAME VARCHAR(255) DEFAULT NULL, REMARK varchar(255) DEFAULT NULL, RESV1 varchar(255) DEFAULT NULL, PRIMARY KEY (ID), UNIQUE KEY APP_UN (CODE)) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8`,
	`INSERT INTO APP (CODE, NAME, SECRET) VALUES('', '', '')`,
	`INSERT INTO APP (CODE, NAME, SECRET) VALUES('COMMON', '', '')`,
	`INSERT INTO APP (CODE, NAME, SECRET) VALUES('TEST', '', '')`,
	`INSERT INTO APP (CODE, NAME, SECRET) VALUES('APP1', '', 'SECRET1')`,

	`DROP TABLE ORG`,
	`CREATE TABLE ORG (ID int(11) NOT NULL AUTO_INCREMENT, UP_ID int(11) DEFAULT NULL, APP_CODE VARCHAR(100) DEFAULT NULL, NAME VARCHAR(1024) NOT NULL, CODE VARCHAR(32) DEFAULT NULL, DESCRIPTION VARCHAR(255) DEFAULT NULL, REMARK varchar(255) DEFAULT NULL, PRIMARY KEY (ID)) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8`,
	`INSERT INTO ORG (NAME, APP_CODE) VALUES('中国', 'APP1')`,
	`INSERT INTO ORG (NAME, APP_CODE, UP_ID) select '北京', 'APP1', ID from ORG where NAME='中国'`,
	`INSERT INTO ORG (NAME, APP_CODE, UP_ID) select '海淀区', 'APP1', ID from ORG where NAME='北京'`,

	`DROP TABLE ORG_GROUP`,
	`CREATE TABLE ORG_GROUP (ID int(11) NOT NULL AUTO_INCREMENT, APP_CODE VARCHAR(100) DEFAULT NULL, NAME VARCHAR(1024) NOT NULL, CODE VARCHAR(32) DEFAULT NULL, DESCRIPTION VARCHAR(255) DEFAULT NULL, REMARK varchar(255) DEFAULT NULL, PRIMARY KEY (ID)) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8`,
	`INSERT INTO ORG_GROUP (NAME, APP_CODE, DESCRIPTION, REMARK) VALUES('华北大区', 'APP1', '', '')`,
	`INSERT INTO ORG_GROUP (NAME, APP_CODE, DESCRIPTION, REMARK) VALUES('东北大区', 'APP1', '', '')`,

	`DROP TABLE ORG_GROUP_LINK`,
	`CREATE TABLE ORG_GROUP_LINK (ID int(11) NOT NULL AUTO_INCREMENT, APP_CODE VARCHAR(100) DEFAULT NULL, ORG_GROUP_ID int(11) NOT NULL, ORG_ID int(11) NOT NULL, REMARK varchar(255) DEFAULT NULL, PRIMARY KEY (ID)) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8`,
	`INSERT INTO ORG_GROUP_LINK (APP_CODE, ORG_GROUP_ID, ORG_ID, REMARK) select 'APP1', a.ID, b.ID, '' from (SELECT ID from ORG_GROUP where NAME='华北大区') a, (SELECT ID from ORG where NAME='北京') b`,

	`DROP TABLE USER`,
	`CREATE TABLE USER (ID int(11) NOT NULL AUTO_INCREMENT, APP_CODE VARCHAR(100) DEFAULT NULL, ID_TYPE VARCHAR(32), USER_ID VARCHAR(100) NOT NULL, PASSWORD VARCHAR(100) NOT NULL, NAME VARCHAR(100) DEFAULT NULL, EMAIL VARCHAR(100) DEFAULT NULL, MOBILE VARCHAR(32) DEFAULT NULL, ROLE VARCHAR(100) DEFAULT NULL, DESCRIPTION VARCHAR(255) DEFAULT NULL, REMARK varchar(255) DEFAULT NULL, RESV1 varchar(255) DEFAULT NULL, RESV2 varchar(255) DEFAULT NULL, RESV3 varchar(255) DEFAULT NULL, PRIMARY KEY (ID), UNIQUE KEY USER_UN (APP_CODE,USER_ID)) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8`,
	`INSERT INTO USER (USER_ID, PASSWORD, APP_CODE, ROLE) VALUES('admin', 'admin9768', 'APP1', 'admin')`,
}

type ConfigType struct {
	DBType          string
	DBConnectString string
	MainSecret      string
	TokenSecret     string
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
	tk.Plvsr(appCodeA, authA)
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

func generateToken(appCodeA string, userIDA string) string {
	strT := appCodeA + "|" + userIDA + "|" + tk.GetNowTimeString()

	return tk.EncryptStringByTXDEF(strT, configG.TokenSecret)
}

func checkToken(appCodeA string, userIDA string, tokenA string) string {
	strT := tk.DecryptStringByTXDEF(tokenA, configG.TokenSecret)

	listT := tk.Split(strT, "|")

	if len(listT) < 3 {
		return "invalid token"
	}

	if appCodeA != listT[0] {
		return "invalid token"
	}

	if userIDA != listT[1] {
		return "invalid token"
	}

	timeT, errT := tk.StrToTimeByFormat(listT[2], tk.TimeFormatCompact)
	if errT != nil {
		return "invalid token"
	}

	expectTimeT := timeT.Add(time.Minute)

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

		return tk.GenerateJSONPResponseWithMore("success", "", req, "token", generateToken(appCodeT, userT))

	case "login":
		authT := paraMapT["auth"]
		appCodeT := paraMapT["appCode"]

		authRsT := checkAuth(appCodeT, authT)

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

		delete(userMapT, "PASSWORD")

		return tk.GenerateJSONPResponseWithMore("success", tk.ToJSONWithDefault(userMapT, ""), req, "Token", generateToken(appCodeT, userT))

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

func runCmd(cmdA string) string {
	// tk.Pl("run cmd: %v", cmdA)

	// var errT error

	switch cmdA {
	case "version":
		{
			tk.Pl("umx " + versionG)
		}

	case "init":
		{
			initSystem()
		}

	case "run":
		{
			startService()
		}

	default:
		tk.Pl("unknown command: %v", cmdA)
	}

	return "exit"
}

func main() {
	var rs string

	argsT := os.Args

	basePathG = tk.GetSwitch(argsT, "-base=", basePathG)

	cmdT := tk.GetParameter(argsT, 1)

	if !tk.IsErrStr(cmdT) {
		rs = runCmd(cmdT)
	}

	if rs == "exit" {
		os.Exit(0)
	}

	// tk.Pl("Initialzing...")
}
