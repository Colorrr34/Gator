package main

import (
	"context"

	"github.com/colorrr34/gator/internal/database"
)

func getPostsForUser(s *state, user database.User, limit int) ([]printPost ,error){
	posts := []printPost{}
	dbPosts, err := s.db.GetPostsForUser(context.Background(),database.GetPostsForUserParams{
		UserID: user.ID,
		Limit: int32(limit),
	})
	if err != nil{
		return nil, err
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

func getPostsForUserAndFeed(s *state, user database.User, limit int, feedName string) ([]printPost ,error){
	posts := []printPost{}
	dbPosts, err := s.db.GetPostsForUserAndFeed(context.Background(),database.GetPostsForUserAndFeedParams{
		UserID: user.ID,
		Limit: int32(limit),
		Name: feedName,
	})
	if err != nil{
		return nil, err
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