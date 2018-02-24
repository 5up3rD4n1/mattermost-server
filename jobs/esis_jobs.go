package jobs

import (
	"sync"
	"time"
	l4g "github.com/alecthomas/log4go"
	"github.com/mattermost/mattermost-server/store"
	"github.com/mattermost/mattermost-server/model"

)

type NotificationService interface {
	SendPostNotification(post *model.Post, receiverId string, properties model.StringMap)
}

type EsisJobsServer struct {
	stop      		chan bool
	stopped   		chan bool
	Store     		store.Store
	NotifyService 	NotificationService
	startOnce 		sync.Once
}

func NewEsisJobsServer(notificationService NotificationService, store store.Store) *EsisJobsServer {
	return &EsisJobsServer{
		stop:    make(chan bool),
		stopped: make(chan bool),
		Store:   store,
		NotifyService: notificationService,
	}
}

func (s *EsisJobsServer) Start() {
	l4g.Info("Initializing ESIS Message Delivery Task.")
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
				case now = <-time.After(1 * time.Minute):
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
		// l4g.Debug("Pending posts len: ", len(pPosts))
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
		l4g.Debug("Notifying Pending Post", pendingPost)
		s.NotifyService.SendPostNotification(post, pendingPost.UserId, pendingPost.Props)
		s.Store.PendingPost().Delete(pendingPost)
	}
}
