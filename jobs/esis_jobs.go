package jobs

import (
	"sync"
	"time"
	l4g "github.com/alecthomas/log4go"
	"github.com/mattermost/mattermost-server/store"
	"github.com/mattermost/mattermost-server/model"

)

type AppService interface {
	SendPostNotification(post *model.Post, receiverId string, properties model.StringMap)
}

type EsisJobsServer struct {
	stop          chan bool
	stopped       chan bool
	configChanged chan *model.Config
	listenerId    string
	startOnce     sync.Once
	Store         store.Store
	ConfigService ConfigService
	NotifyService AppService
}

func NewEsisJobsServer(app AppService, configService ConfigService, store store.Store) *EsisJobsServer {
	return &EsisJobsServer{
		stop:    make(chan bool),
		stopped: make(chan bool),
		Store:   store,
		NotifyService: app,
		ConfigService: configService,
	}
}

func (s *EsisJobsServer) Start() {

	esisConfig := s.ConfigService.Config().EsisSettings

	if *esisConfig.Enable {

		s.listenerId = s.ConfigService.AddConfigListener(s.handleConfigChange)

		l4g.Info("Initializing ESIS Message Delivery Task.")

		cfgTimeWindowMinutes := *esisConfig.MessageDeliveryTimeWindowMinutes

		go func() {
			s.startOnce.Do(func() {
				l4g.Info("Starting ESIS Message Delivery Task.")

				defer func() {
					l4g.Info("ESIS Message Delivery Task Stopped")
					close(s.stopped)
				}()
				now := time.Now()
				for {
					select {
					case <-s.stop:
						l4g.Debug("ESIS Message Delivery received stop signal.")
						return
					case newCfg := <-s.configChanged:
						esisConfig = newCfg.EsisSettings
						cfgTimeWindowMinutes = *esisConfig.MessageDeliveryTimeWindowMinutes
					case now = <-time.After(time.Duration(cfgTimeWindowMinutes) * time.Minute):
						result := <-s.Store.User().GetEsisApiAvailable(now)
						users := result.Data.([]*model.User)

						// for user := users
						l4g.Info("Running Esis messages task", now.UTC(), users)
						for _, user := range users {
							s.processPostsForUser(user.Id)
							// l4g.Debug(user)
						}
					}
				}
			})
		}()
	}
}

func (s *EsisJobsServer) Stop() {
	close(s.stop)
	<-s.stopped
}

func (s *EsisJobsServer) processPostsForUser(uid string) {
	pPostsResult := <- s.Store.PendingPost().PendingPostsForUser(uid)

	if pPostsResult.Err != nil {
		l4g.Error("Error getting pending posts", pPostsResult.Err.Error())
	} else {
		pPosts := pPostsResult.Data.([]*model.PendingPost)
		for _, pendingPost := range pPosts {
			s.notifyPendingPostToUser(pendingPost)
		}
	}
}


func (s *EsisJobsServer) notifyPendingPostToUser(pendingPost *model.PendingPost) {
	result := <- s.Store.Post().GetSingle(pendingPost.PostId)
	if result.Err != nil {
		l4g.Warn("Error getting post from pending post postId", result.Err.Error())
	} else {
		post := result.Data.(*model.Post)
		s.NotifyService.SendPostNotification(post, pendingPost.UserId, pendingPost.Props)
		s.Store.PendingPost().Delete(pendingPost)
	}
}

func (s *EsisJobsServer) handleConfigChange(oldConfig *model.Config, newConfig *model.Config) {
	l4g.Debug("ESIS Message Delivery received config change.")
	s.configChanged <- newConfig
}
