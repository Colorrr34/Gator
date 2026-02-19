package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/colorrr34/gator/internal/database"
	"github.com/google/uuid"
)

func scrapeFeeds(s *state,ch chan error){
	ctx:=context.Background()
	feedNext, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil{
		errStr := fmt.Sprintf("Get next feed error: %s", err)
		ch <- errors.New(errStr)
	}
	if err := s.db.MarkFeedFetched(context.Background(),feedNext.ID); err != nil{
		errStr := fmt.Sprintf("Mark feed error: %s", err)
		ch <- errors.New(errStr)
	}
	feed,err := fetchFeed(context.Background(),feedNext.Url)
	if err != nil{
		errStr := fmt.Sprintf("Fetch feed error: %s", err)
		ch <- errors.New(errStr)
	}
	for _, item := range feed.Channel.Item{
		var description sql.NullString
		if item.Description != ""{
			description = sql.NullString{
				String:item.Description,
				Valid:true,
			}
		}
		var publishedAt sql.NullTime
		if item.PubDate != ""{
			timeParsed,err := time.Parse(time.RFC1123Z,item.PubDate)
			if err !=nil{
				ch <- err
			}
			publishedAt = sql.NullTime{
				Time: timeParsed,
				Valid: true,
			}
		}
		post, err := s.db.CreatePost(ctx,database.CreatePostParams{
			ID: uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Title: item.Title,
			Url: item.Link,
			Description: description,
			PublishedAt: publishedAt,
			FeedID: feedNext.ID,
		})
		if err != nil{
			if strings.Contains(err.Error(),"posts_url_key"){
				continue
			}
			ch <- err
		}
		fmt.Println(post)
	}
	fmt.Printf("Finished Saving %s\n", feed.Channel.Title)
	close(ch)
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error){
	req, err := http.NewRequestWithContext(ctx,"GET",feedURL,nil)
	if err != nil{
		return nil, err
	}
	(*req).Header.Set("User-Agent","gator")
	client := &http.Client{}
	
	res, err:= client.Do(req)
	if err !=nil{
		return nil, err
	}
	defer res.Body.Close()
	resBodyJson, err := io.ReadAll(res.Body)
	if err != nil{
		return nil,err
	}
	
	var feed RSSFeed
	if err:= xml.Unmarshal(resBodyJson,&feed);err!=nil{
		return nil,err
	}

	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	for _,item := range feed.Channel.Item{
		item.Description = html.UnescapeString(item.Description)
		item.Title = html.UnescapeString(item.Title)
	}

	return &feed,nil
}	