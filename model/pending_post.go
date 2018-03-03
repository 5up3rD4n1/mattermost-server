package model

import (
	"encoding/json"
	"io"
	"net/http"
)

type PendingPost struct {
	Id       	string    `json:"id"`
	PostId		string    `json:"post_id"`
	UserId 		string 	  `json:"user_id"`
	ChannelId 	string 	  `json:"channel_id"`
	Props  		StringMap `json:"props"`
	CreateAt 	int64     `json:"create_at"`
	UpdateAt 	int64     `json:"update_at"`
	DeleteAt 	int64     `json:"delete_at"`
}

func (o *PendingPost) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	if o.CreateAt == 0 {
		o.CreateAt = GetMillis()
	}

	o.UpdateAt = o.CreateAt
}

func (o *PendingPost) ToJson() string {
	b, err := json.Marshal(o)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func (o *PendingPost) String() string {
	return "PendingPost: " + o.PostId + "for User: " + o.UserId + "with Props: " + o.Props.String()
}

func PendingPostFromJson(data io.Reader) *PendingPost {
	decoder := json.NewDecoder(data)
	var o PendingPost
	err := decoder.Decode(&o)
	if err == nil {
		return &o
	} else {
		return nil
	}
}

func (o *PendingPost) IsValid() *AppError {
	if o.UserId == "" && o.PostId == "" && o.ChannelId == "" {
		return NewAppError("PendingStore.IsValid", "camino.pending_post.missing_data", nil, "", http.StatusBadRequest)
	}

	return nil
}


func (o *PendingPost) SoftDelete() {
	currentTime := GetMillis()
	o.UpdateAt = currentTime
	o.DeleteAt = currentTime
}
