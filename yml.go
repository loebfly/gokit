package gokit

import dbYml "github.com/loebfly/dblite/yml"

/*
# 应用程序配置
app:
  name: xxxx
  port: 0000 # http端口
  port_ssl: 0000 # https端口
  env: test # 环境
  debug: true # 是否开启debug模式
  db_config: local or nacos # 数据库配置来源
  is_register_nacos: true # 是否注册到nacos

# nacos 配置
nacos:
  server: http:// # nacos 服务地址
    yml:
      mysql: xxxx # mysql 配置文件名称，不带.yml后缀
      mongo: xxxx # mongo 配置文件名称，不带.yml后缀
      redis: xxxx # redis 配置文件名称，不带.yml后缀

# log 配置
log:
  out: console,file # log output
  file: /opt/logs/xxxx # log file path
  mongo: xxxx # mongodb log collection name

# 本地mysql配置
db_local:
  mysql:
    url:
    pool:
      max: 20
      idle: 10
      timeout:
        idle: 60
        life: 60

  # 本地redis配置
  redis:
    host:
    port:
    password:
    database: 5
    timeout: 1000
    pool:
      min: 3
      max: 20
      idle: 10
      timeout: 300

  # 本地mongodb配置
  mongo:
    url:
    database:
    pool_max: 20
*/

type GoYml struct {
	App struct {
		Name            string `yaml:"name"`
		Port            int    `yaml:"port"`
		PortSsl         int    `yaml:"port_ssl"`
		Env             string `yaml:"env"`
		Debug           bool   `yaml:"debug"`
		DbConfig        string `yaml:"db_config"`
		IsRegisterNacos bool   `yaml:"is_register_nacos"`
	} `yaml:"app"`

	Nacos struct {
		Server string `yaml:"server"`
		Yml    struct {
			Mysql string `yaml:"mysql"`
			Mongo string `yaml:"mongo"`
			Redis string `yaml:"redis"`
		} `yaml:"yml"`
	} `yaml:"nacos"`

	Log struct {
		Out   string `yaml:"out"`
		File  string `yaml:"file"`
		Mongo string `yaml:"mongo"`
	} `yaml:"log"`

	DBLocal struct {
		MySql dbYml.Mysql `yaml:"mysql"`
		Mongo dbYml.Mongo `yaml:"mongo"`
		Redis dbYml.Redis `yaml:"redis"`
	} `yaml:"db_local"`
}
