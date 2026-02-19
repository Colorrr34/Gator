package main

import (
	"context"
	"database/sql"

	"github.com/colorrr34/gator/internal/database"
)

type printPost struct{
	Post database.Post
	FeedName string
}

func getPosts(s *state, user database.User, limit int, feedName string, sort string,page int, isDesc bool) ([]printPost ,error){
	var name sql.NullString
	if feedName != "null"{
		name = sql.NullString{
			Valid: true,
			String: feedName,
		}
	}
	offset := (page-1)*limit
	
	posts := []printPost{}

	dbPosts,err := s.db.GetPosts(context.Background(),database.GetPostsParams{
		UserID: user.ID,
		Limit: int32(limit),
		Name: name,
		Sort: sort,
		IsDesc: isDesc,
		Offset: int32(offset),
	})
	if err != nil{
		return nil,err
	}
	for _,dbPost := range dbPosts{
		post := printPost{
			Post: database.Post{
			ID: dbPost.ID,
			CreatedAt: dbPost.CreatedAt,
			UpdatedAt: dbPost.UpdatedAt,
			Title: dbPost.Title,
			Url: dbPost.Url,
			Description: dbPost.Description,
			PublishedAt: dbPost.PublishedAt,
			FeedID: dbPost.FeedID,
		},
		FeedName: dbPost.Name.String,
		}
		posts = append(posts, post)
	}

	return posts,nil
}

func searchPosts(s *state, user database.User, limit int, feedName string, sort string, page int, isDesc bool, searchWord string)([]printPost ,error){
	var name sql.NullString
	if feedName != "null"{
		name = sql.NullString{
			Valid: true,
			String: feedName,
		}
	}
	offset := (page-1)*limit

	posts := []printPost{}

	dbPosts,err := s.db.SearchPosts(context.Background(),database.SearchPostsParams{
		UserID: user.ID,
		Name: name,
		Limit: int32(limit),
		Offset: int32(offset),
		IsDesc: isDesc,
		Sort: sort,
		Title:searchWord,
	})
	if err != nil{
		return nil,err
	}
	for _,dbPost := range dbPosts{
		post := printPost{
			Post: database.Post{
			ID: dbPost.ID,
			CreatedAt: dbPost.CreatedAt,
			UpdatedAt: dbPost.UpdatedAt,
			Title: dbPost.Title,
			Url: dbPost.Url,
			Description: dbPost.Description,
			PublishedAt: dbPost.PublishedAt,
			FeedID: dbPost.FeedID,
		},
		FeedName: dbPost.Name.String,
		}
		posts = append(posts, post)
	}

	return posts,nil
}