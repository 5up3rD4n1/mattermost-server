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
	Config() *model.Config
	AddConfigListener(func(old, current *model.Config)) string
	RemoveConfigListener(string)
}

type CaminoJobsServer struct {
	stop          chan bool
	stopped       chan bool
	configChanged chan *model.Config
	listenerId    string
	startOnce     sync.Once
	Store         store.Store
	AppService    AppService
}

func NewCaminoJobsServer(app AppService, store store.Store) *CaminoJobsServer {
	return &CaminoJobsServer{
		stop:       make(chan bool),
		stopped:    make(chan bool),
		Store:      store,
		AppService: app,
	}
}

func (s *CaminoJobsServer) Start() {

	caminoConfig := s.AppService.Config().CaminoSettings

	if *caminoConfig.Enable {

		s.listenerId = s.AppService.AddConfigListener(s.handleConfigChange)

		l4g.Info("Initializing CAMINO Message Delivery Task.")

		cfgTimeWindowMinutes := *caminoConfig.MessageDeliveryTimeWindowMinutes

		go func() {
			s.startOnce.Do(func() {
				l4g.Info("Starting CAMINO Message Delivery Task.")

				defer func() {
					l4g.Info("CAMINO Message Delivery Task Stopped")
					close(s.stopped)
				}()
				now := time.Now()
				for {
					select {
					case <-s.stop:
						l4g.Debug("CAMINO Message Delivery received stop signal.")
						s.AppService.RemoveConfigListener(s.listenerId)
						return
					case newCfg := <-s.configChanged:
						caminoConfig = newCfg.CaminoSettings
						cfgTimeWindowMinutes = *caminoConfig.MessageDeliveryTimeWindowMinutes
					case now = <-time.After(time.Duration(cfgTimeWindowMinutes) * time.Minute):
						result := <-s.Store.User().GetCaminoApiAvailable(now)
						users := result.Data.([]*model.User)

						l4g.Info("Running Camino messages task", now.UTC(), users)
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

func (s *CaminoJobsServer) Stop() {
	close(s.stop)
	<-s.stopped
}

func (s *CaminoJobsServer) processPostsForUser(uid string) {
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

func (s *CaminoJobsServer) notifyPendingPostToUser(pendingPost *model.PendingPost) {
	result := <- s.Store.Post().GetSingle(pendingPost.PostId)
	if result.Err != nil {
		l4g.Warn("Error getting post from pending post postId", result.Err.Error())
	} else {
		post := result.Data.(*model.Post)
		s.AppService.SendPostNotification(post, pendingPost.UserId, pendingPost.Props)
		s.Store.PendingPost().Delete(pendingPost)
	}
}

func (s *CaminoJobsServer) handleConfigChange(oldConfig *model.Config, newConfig *model.Config) {
	l4g.Debug("CAMINO Message Delivery received config change.")
	s.configChanged <- newConfig
}
