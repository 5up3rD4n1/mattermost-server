package jobs

import (
	"sync"
	"time"
	"github.com/mattermost/mattermost-server/store"
	l4g "github.com/alecthomas/log4go"
	"github.com/mattermost/mattermost-server/model"
)

type EsisJobsServer struct {
	stop      chan bool
	stopped   chan bool
	Store     store.Store
	startOnce sync.Once
}

func NewEsisJobsServer(store store.Store) *EsisJobsServer {
	return &EsisJobsServer{
		stop:    make(chan bool),
		stopped: make(chan bool),
		Store:   store,
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
					// TODO: Update user status online
					result := <-s.Store.User().GetEsisApiAvailable(now)
					users := result.Data.([]*model.User)

					// for user := users
					for user := range users {
						l4g.Info(user)
					}
					l4g.Info("Running Esis messages task", now.UTC(), users)
				}
			}
		})
	}()
}

func (s *EsisJobsServer) Stop() {
	close(s.stop)
	<-s.stopped
}
