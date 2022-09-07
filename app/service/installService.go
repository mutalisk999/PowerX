package service

import (
	"bytes"
	"errors"
	fmt2 "fmt"
	"github.com/ArtisanCloud/PowerLibs/v2/corountine/locker"
	"github.com/ArtisanCloud/PowerLibs/v2/fmt"
	"github.com/ArtisanCloud/PowerWeChat/v2/src/work"
	"github.com/ArtisanCloud/PowerX/boostrap/cache"
	globalBootstrap "github.com/ArtisanCloud/PowerX/boostrap/cache/global"
	"github.com/ArtisanCloud/PowerX/config"
	"github.com/ArtisanCloud/PowerX/database"
	logger "github.com/ArtisanCloud/PowerX/loggerManager"
	"github.com/gin-gonic/gin"
	"os/exec"
	"strings"
	"sync"
)

type InstallService struct {
	*Service
}

type ResponseTask struct {
	Name   string
	Status string
	ErrMsg string
}

func NewResponseTask() *ResponseTask {
	return &ResponseTask{
		Name:   "task",
		Status: "success",
		ErrMsg: "",
	}
}

//var _InstallWG sync.WaitGroup
var _InstallMux sync.Mutex
var _InstallTasks []func(taskChannel chan error, appConfig *config.AppConfig, rsTask *ResponseTask)

func init() {

	// 压栈安装任务
	_InstallTasks = append(_InstallTasks, TaskInstallApp())
	_InstallTasks = append(_InstallTasks, TaskInstallDatabase())
	_InstallTasks = append(_InstallTasks, TaskInstallCache())
	_InstallTasks = append(_InstallTasks, TaskInstallLog())
	_InstallTasks = append(_InstallTasks, TaskInstallJWT())
	_InstallTasks = append(_InstallTasks, TaskInstallWechat())

}

func NewInstallService(ctx *gin.Context) (r *InstallService) {
	r = &InstallService{
		Service: NewService(ctx),
	}
	return r
}

func (srv *InstallService) InstallSystem(config *config.AppConfig) (installStatusList []*ResponseTask, err error) {

	//fmt.Dump(config)
	installStatusList = []*ResponseTask{}

	// 检查是否安装任务是否被锁
	if locker.MutexLocked(&_InstallMux) {
		return nil, errors.New("无法执行安装任务，系统被锁定，已有其他安装任务在执行，请过后尝试。")
	}

	// 先锁定当前安装任务
	_InstallMux.Lock()

	// 完成安装任务后，解锁
	defer _InstallMux.Unlock()

	// 任务数量
	n := len(_InstallTasks)

	// 创建安装任务通道
	taskChannel := make(chan error, n)
	fmt2.Printf("length:%d, cap:%d \r\n", len(taskChannel), cap(taskChannel))

	for i, task := range _InstallTasks {

		installStatusList = append(installStatusList, NewResponseTask())

		//协程方式去并发检测安装执行
		go task(taskChannel, config, installStatusList[i])

	}

	// 如果当前的任务完成数量，没有达到任务总数，阻赛等待
	//time.Sleep(1 * time.Second)
	for len(taskChannel) < n {
		//fmt2.Printf("waiting taskChannel length:%d, cap:%d \r\n", len(taskChannel), cap(taskChannel))
	}

	// 如果任务都完成了，关闭任务通道
	// 必须要关闭通道，否则下方的遍历会阻赛
	//fmt2.Printf("length:%d, cap:%d \r\n", len(taskChannel), cap(taskChannel))
	close(taskChannel)

	// 遍历任务通道，检查通道返回值
	for chError := range taskChannel {
		if chError != nil {
			fmt.Dump(chError.Error())
		}
	}

	fmt.Dump("finish tasks")

	return installStatusList, err

}

func (srv *InstallService) CheckSystemInstallation() (installStatusList []*ResponseTask, err error) {

	installStatusList = []*ResponseTask{}

	return installStatusList, err

}

func TaskInstallApp() func(taskChannel chan error, appConfig *config.AppConfig, rsTask *ResponseTask) {
	return func(taskChannel chan error, appConfig *config.AppConfig, rsTask *ResponseTask) {
		var err error
		fmt.Dump("run task app")
		rsTask.Name = "app"
		rsTask.Status = "failed"

		rsTask.Status = "success"
		taskChannel <- err
		return
	}
}
func TaskInstallDatabase() func(taskChannel chan error, appConfig *config.AppConfig, rsTask *ResponseTask) {

	return func(taskChannel chan error, appConfig *config.AppConfig, rsTask *ResponseTask) {
		var err error
		fmt.Dump("run task database")
		rsTask.Name = "database"
		rsTask.Status = "failed"

		// 链接数据库
		err = database.SetupDatabase(&appConfig.DatabaseConfig.DatabaseConnections.PostgresConfig)
		if err != nil {
			rsTask.ErrMsg = err.Error()
			taskChannel <- err
			return
		}

		// migrate 数据库
		cmd := exec.Command("make", "./Makefile", "migrate-tables")
		cmd.Stdin = strings.NewReader("and old falcon")

		var out bytes.Buffer
		cmd.Stdout = &out
		err = cmd.Run()
		if err != nil {
			rsTask.ErrMsg = err.Error()
			taskChannel <- err
			return
		}

		// 导入原始系统数据
		cmd = exec.Command("make", "./Makefile", "import-rbac-data")
		cmd.Stdin = strings.NewReader("and old falcon")

		cmd.Stdout = &out
		err = cmd.Run()
		if err != nil {
			rsTask.ErrMsg = err.Error()
			taskChannel <- err
			return
		}

		rsTask.Status = "success"
		taskChannel <- err
		return
	}
}

func TaskInstallCache() func(taskChannel chan error, appConfig *config.AppConfig, rsTask *ResponseTask) {

	return func(taskChannel chan error, appConfig *config.AppConfig, rsTask *ResponseTask) {
		var err error
		rsTask.Name = "cache"
		rsTask.Status = "failed"

		fmt.Dump("run task cache")

		err = cache.SetupCache(&appConfig.CacheConfig.CacheConnections.RedisConfig)
		if err != nil {
			rsTask.ErrMsg = err.Error()
			taskChannel <- err
			return
		}

		_, err = globalBootstrap.G_CacheConnection.Keys()
		if err != nil {
			rsTask.ErrMsg = err.Error()
			taskChannel <- err
			return
		}

		rsTask.Status = "success"
		taskChannel <- err
		return
	}
}

func TaskInstallLog() func(taskChannel chan error, appConfig *config.AppConfig, rsTask *ResponseTask) {

	return func(taskChannel chan error, appConfig *config.AppConfig, rsTask *ResponseTask) {
		var err error
		rsTask.Name = "log"
		rsTask.Status = "failed"

		fmt.Dump("run task log")

		//fmt.Dump(appConfig.LogConfig.LogPath)
		err = logger.SetupLog(&appConfig.LogConfig)
		if err != nil {
			rsTask.ErrMsg = err.Error()
			taskChannel <- err
			return
		}

		rsTask.Status = "success"
		taskChannel <- err
		return
	}
}

func TaskInstallJWT() func(taskChannel chan error, appConfig *config.AppConfig, rsTask *ResponseTask) {

	return func(taskChannel chan error, appConfig *config.AppConfig, rsTask *ResponseTask) {
		var err error
		rsTask.Name = "jwt"
		rsTask.Status = "failed"

		fmt.Dump("run task jwt")

		err = SetupJWTKeyPairs(&appConfig.JWTConfig)
		if err != nil {
			rsTask.ErrMsg = err.Error()
			taskChannel <- err
			return
		}

		rsTask.Status = "success"
		taskChannel <- err
		return
	}
}

func TaskInstallWechat() func(taskChannel chan error, appConfig *config.AppConfig, rsTask *ResponseTask) {

	return func(taskChannel chan error, appConfig *config.AppConfig, rsTask *ResponseTask) {
		var err error
		rsTask.Name = "weCom"
		rsTask.Status = "failed"

		fmt.Dump("run task wechat")

		fmt.Dump(appConfig.WecomConfig)
		weComApp, err := work.NewWork(&work.UserConfig{
			CorpID:      appConfig.WecomConfig.CorpID,                // 企业微信的corp id，所有企业微信共用一个。
			AgentID:     appConfig.WecomConfig.WecomAgentID,          // 内部应用的app id
			Secret:      appConfig.WecomConfig.WecomSecret,           // 默认内部应用的app secret
			CallbackURL: appConfig.WecomConfig.AppMessageCallbackURL, // 内部应用的场景回调设置
			OAuth: work.OAuth{
				Callback: appConfig.WecomConfig.AppOauthCallbackURL, // 内部应用的app oauth url
				Scopes:   []string{"snsapi_base"},
			},
			HttpDebug: true,
		})
		if err != nil {
			rsTask.ErrMsg = err.Error()
			taskChannel <- err
			return
		}

		rs, err := weComApp.Base.GetCallbackIP()
		if err != nil {
			rsTask.ErrMsg = err.Error()
			taskChannel <- err
			return
		}

		if rs.ErrCode != 0 {
			err = errors.New(rs.ErrMSG)
			rsTask.ErrMsg = err.Error()
			taskChannel <- err
			return
		}

		rsTask.Status = "success"
		taskChannel <- err
		return
	}
}