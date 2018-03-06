// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package sqlstore

import (
	// "fmt"
	"net/http"
	// "regexp"
	// "strconv"
	// "strings"

	// "bytes"

	// l4g "github.com/alecthomas/log4go"
	"github.com/mattermost/mattermost-server/einterfaces"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/store"
	// "github.com/mattermost/mattermost-server/utils"
)

type SqlPendingPostStore struct {
	SqlStore
	metrics einterfaces.MetricsInterface
}

func NewSqlPendingPostStore(sqlStore SqlStore, metrics einterfaces.MetricsInterface) store.PendingPostStore {
	s := &SqlPendingPostStore{
		SqlStore: sqlStore,
		metrics:  metrics,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.PendingPost{}, "PendingPosts").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("UserId").SetMaxSize(26)
		table.ColMap("PostId").SetMaxSize(26)
		table.ColMap("ChannelId").SetMaxSize(26)
		table.ColMap("Props").SetMaxSize(2000)
	}

	return s
}

func (s SqlPendingPostStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_pending_posts_update_at", "PendingPosts", "UpdateAt")
	s.CreateIndexIfNotExists("idx_pending_posts_create_at", "PendingPosts", "CreateAt")
	s.CreateIndexIfNotExists("idx_pending_posts_delete_at", "PendingPosts", "DeleteAt")
	s.CreateIndexIfNotExists("idx_pending_posts_channel_id", "PendingPosts", "ChannelId")
	s.CreateIndexIfNotExists("idx_pending_posts_user_id", "PendingPosts", "UserId")
	s.CreateIndexIfNotExists("idx_pending_posts_post_id", "PendingPosts", "PostId")

	//s.CreateCompositeIndexIfNotExists("idx_pending_posts_channel_id_update_at", "Posts", []string{"ChannelId", "UpdateAt"})
	//s.CreateCompositeIndexIfNotExists("idx_pending_posts_channel_id_delete_at_create_at", "Posts", []string{"ChannelId", "DeleteAt", "CreateAt"})
}

func (s SqlPendingPostStore) Save(pendingPost *model.PendingPost) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if len(pendingPost.Id) > 0 {
			result.Err = model.NewAppError("SqlPendingPostStore.Save", "store.sql_pending_post.save.existing.app_error", nil, "id="+pendingPost.Id, http.StatusBadRequest)
			return
		}

		pendingPost.PreSave()
		if result.Err = pendingPost.IsValid(); result.Err != nil {
			return
		}

		if err := s.GetMaster().Insert(pendingPost); err != nil {
			result.Err = model.NewAppError("SqlPendingPostStore.Save", "store.sql_pending_post.save.app_error", nil, "id="+pendingPost.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = pendingPost
		}
	})
}

func (s SqlPendingPostStore) PendingPostsForUser(userId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		var pendingPosts []*model.PendingPost
		if _, err := s.GetMaster().Select(&pendingPosts,
			`SELECT 
				*
			FROM 
				PendingPosts
			WHERE 
				UserId = :UserId
			AND 
				DeleteAt = 0
			ORDER BY CreateAt DESC`,
			map[string]interface{}{"UserId": userId}); err != nil {
			result.Err = model.NewAppError("SqlPendingPostStore.PendingPostForUser",
				"store.sql_pending_post.pending_posts_for_user.app_error", nil, err.Error(),
				http.StatusInternalServerError)
		} else {
			result.Data = pendingPosts
		}
	})
}

func (s SqlPendingPostStore) Delete(pendingPost *model.PendingPost) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		pendingPost.SoftDelete()

		_, err := s.GetMaster().Exec(
			`UPDATE 
				PendingPosts
			SET
				DeleteAt = :DeleteAt, UpdateAt = :UpdateAt
			WHERE 
				Id = :Id`,
			map[string]interface{}{"Id": pendingPost.Id, "DeleteAt": pendingPost.DeleteAt, "UpdateAt": pendingPost.UpdateAt})

		if err != nil {
			result.Err = model.NewAppError("SqlPendingPostStore.Delete", "store.sql_pending_post.delete.app_error", nil, "id="+pendingPost.Id+", err="+err.Error(), http.StatusInternalServerError)
		}
	})
}