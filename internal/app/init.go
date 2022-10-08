package app

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/loebfly/ezgin/internal/config"
	"github.com/loebfly/ezgin/internal/dblite"
	"github.com/loebfly/ezgin/internal/dblite/mongo"
	"github.com/loebfly/ezgin/internal/dblite/mysql"
	"github.com/loebfly/ezgin/internal/dblite/redis"
	"github.com/loebfly/ezgin/internal/engine"
	"github.com/loebfly/ezgin/internal/logs"
	"github.com/loebfly/ezgin/internal/nacos"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// getLocalYml 获取yml配置文件路径
func (receiver enter) getYml() string {
	var fileName string
	flag.StringVar(&fileName, "f", os.Args[0]+".yml", "yml配置文件名")
	flag.Parse()
	path, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	if strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {
		return fileName
	}
	return path + "/" + fileName
}

// initPath 初始化所有组件入口
func (receiver enter) initEZGin(ymlPath string, ginEngine *gin.Engine) {
	receiver.initConfig(ymlPath)
	receiver.initLogs()
	receiver.initServer()
	receiver.initNacos()
	receiver.initDBLite()
	receiver.initEngine(ginEngine)
}

func (receiver enter) initConfig(ymlPath string) {
	if ymlPath == "" {
		ymlPath = receiver.getYml()
	}
	config.InitPath(ymlPath)
}

// initLogs 初始化日志模块
func (receiver enter) initLogs() {
	out := config.EZGin().Logs.Out
	file := config.EZGin().Logs.File
	if file == "" {
		file = "/opt/logs/" + config.EZGin().App.Name
	}
	yml := logs.Yml{
		Out:  out,
		File: file,
	}
	logs.InitObj(yml)
}

// initServer 初始化服务
func (receiver enter) initServer() {
	ez := config.EZGin()

	if ez.App.Port > 0 {
		// HTTP 端口
		servers = append(servers, &http.Server{
			Addr:    ":" + strconv.Itoa(ez.App.Port),
			Handler: engine.Enter.GetOriEngine(),
		})
		go func() {
			listenErr := servers[0].ListenAndServe()
			logs.Enter.CWarn("APP", "ListenAndServe:{}:{}", ez.App.Port, listenErr.Error())
		}()
	}
	if ez.App.PortSsl > 0 {
		// HTTPS 端口
		servers = append(servers, &http.Server{
			Addr:    ":" + strconv.Itoa(ez.App.PortSsl),
			Handler: engine.Enter.GetOriEngine(),
		})
		go func() {
			path, _ := filepath.Abs(filepath.Dir(os.Args[0]))
			listenErr := servers[1].ListenAndServeTLS(path+"/"+ez.App.Cert, path+"/"+ez.App.Key)
			logs.Enter.CWarn("APP", "ListenAndServeTLS:{}:{}", ez.App.PortSsl, listenErr.Error())
		}()
	}
}

// initNacos 初始化nacos
func (receiver enter) initNacos() {
	ez := config.EZGin()
	if ez.Nacos.Server != "" && ez.Nacos.Yml.Nacos != "" {
		nacosPrefix := ez.Nacos.Yml.Nacos
		if nacosPrefix != "" {
			nacosUrl := ez.GetNacosUrl(nacosPrefix)
			var yml nacos.Yml
			err := config.Enter.GetYmlObj(nacosUrl, &yml)
			if err != nil {
				panic(fmt.Errorf("nacos配置文件获取失败: %s", err.Error()))
			}
			yml.EZGinNacos.App = nacos.App{
				Name:    ez.App.Name,
				Ip:      ez.App.Ip,
				Port:    ez.App.Port,
				PortSsl: ez.App.PortSsl,
				Debug:   ez.App.Debug,
			}
			nacos.InitObj(yml.EZGinNacos)
		}
	}
}

func (receiver enter) initDBLite() {
	ez := config.EZGin()

	var mongoObjs []mongo.EZGinMongo
	if ez.Nacos.Yml.Mongo != "" {
		names := strings.Split(ez.Nacos.Yml.Mongo, ",")
		mongoObjs = make([]mongo.EZGinMongo, len(names))
		for _, name := range names {
			var yml mongo.Yml
			nacosUrl := ez.GetNacosUrl(name)
			err := config.Enter.GetYmlObj(nacosUrl, &yml)
			if err != nil {
				panic(fmt.Errorf("mysql配置文件获取失败: %s", err.Error()))
			}
			mongoObjs = append(mongoObjs, yml.EZGinMongo)
		}
	}

	var mysqlObjs []mysql.EZGinMysql
	if ez.Nacos.Yml.Mysql != "" {
		names := strings.Split(ez.Nacos.Yml.Mysql, ",")
		mysqlObjs = make([]mysql.EZGinMysql, len(names))
		for _, name := range names {
			var yml mysql.Yml
			nacosUrl := ez.GetNacosUrl(name)
			err := config.Enter.GetYmlObj(nacosUrl, &yml)
			if err != nil {
				panic(fmt.Errorf("mysql配置文件获取失败: %s", err.Error()))
			}
			mysqlObjs = append(mysqlObjs, yml.EZGinMysql)
		}
	}
	var redisObjs []redis.EZGinRedis
	if ez.Nacos.Yml.Redis != "" {
		names := strings.Split(ez.Nacos.Yml.Redis, ",")
		redisObjs = make([]redis.EZGinRedis, len(names))
		for _, name := range names {
			var yml redis.Yml
			nacosUrl := ez.GetNacosUrl(name)
			err := config.Enter.GetYmlObj(nacosUrl, &yml)
			if err != nil {
				panic(fmt.Errorf("mysql配置文件获取失败: %s", err.Error()))
			}
			redisObjs = append(redisObjs, yml.EZGinRedis)
		}
	}
	dblite.InitDB(mongoObjs, mysqlObjs, redisObjs)
}

// initEngine 初始化gin引擎
func (receiver enter) initEngine(ginEngine *gin.Engine) {
	ez := config.EZGin()

	mongoTag := ez.Gin.MwLogs.MongoTag
	if mongoTag != "-" && mongoTag != "" {
		if !dblite.IsExistMongoTag(mongoTag) {
			panic(fmt.Errorf("mongo_tag:%s不存在", mongoTag))
		}
	}
	table := ez.Gin.MwLogs.Table
	if table == "" {
		table = ez.App.Name + "APIRequestLogs"
	}

	yml := engine.Yml{
		Mode:       ez.Gin.Mode,
		Middleware: ez.Gin.Middleware,
		Engine:     ginEngine,
		MwLogs:     engine.MwLogs{MongoTag: mongoTag, Table: table},
	}
	engine.InitObj(yml)
}
