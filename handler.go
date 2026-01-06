package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/colorrr34/gator/internal/database"
	"github.com/google/uuid"
)


func handlerLogin(s *state, cmd command)error{
	if len(cmd.arg) ==0{
		return errors.New("insufficient arguments")
	}
	
	user,err := s.db.GetUser(context.Background(),cmd.arg[0])
	if err != nil{
		return errors.New("User doesn't exist")
	}

	s.cfg.SetUser(user.Name)
	fmt.Printf("User %s has been set\n",s.cfg.CurrentUserName)
	return nil
}

func handlerRegister(s *state, cmd command)error{
	if len(cmd.arg) ==0{
		log.Fatal("username required")
	}

	_,err := s.db.GetUser(context.Background(),cmd.arg[0])
	if err == nil{
		return errors.New("User already exists")
	}

	user, err := s.db.CreateUser(context.Background(),database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:cmd.arg[0],
	})
	if err != nil{
		return err
	}
	msgCreated := fmt.Sprintf("user %s has been created", user.Name)
	fmt.Println(msgCreated)
	s.cfg.SetUser(user.Name)
	fmt.Printf("Now logged in as user %s\n", user.Name)

	return nil
}

func handlerReset(s *state, _ command)error{
	err := s.db.DeleteUsers(context.Background())
	if err != nil{
		return err
	}

	return nil
}

func handlerUsers(s *state, _ command)error{
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	
	for _,user := range users{
		if user.Name == (*s).cfg.CurrentUserName{
			fmt.Printf("* %s (current)\n", user.Name)
		}
		fmt.Printf("* %s\n", user.Name)
	}
	return nil
}

func handlerAgg(s *state, cmd command)error{
	if len(cmd.arg) <1 {
		return errors.New("Missing time argument")
	}
	time_between_reqs := cmd.arg[0]
	dur,err := time.ParseDuration(time_between_reqs)
	if err != nil{
		return err
	}
	fmt.Printf("Collecting feeds every %s\n", dur)
	ticker := time.NewTicker(dur)
	for ;;<-ticker.C{
		err := scrapeFeeds(s)
		if err != nil{
			return err
		}
	}
}

func handlerAddFeed(s *state, cmd command,user database.User)error{
	if len(cmd.arg)<2{
		return errors.New("insufficient arguments")
	}
	feed, err := s.db.CreateFeed(context.Background(),database.CreateFeedParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: cmd.arg[0],
		Url: cmd.arg[1],
		UserID: user.ID,
	})
	if err != nil{
		return err
	}
	s.db.CreateFeedFollow(context.Background(),database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feed.ID,
	})
	fmt.Println(feed)
	return nil
}

func handlerFeeds(s *state, _ command)error{
	feeds,err := s.db.GetFeeds(context.Background())
	if err !=nil{
		return err
	}
	for _,feed := range feeds{
		fmt.Println(feed)
	}

	return nil
}

func handlerFollow(s *state, cmd command,user database.User)error{
	if len(cmd.arg) < 1 {
		return errors.New("Insufficient arguments")
	}
	feed, err := s.db.GetFeed(context.Background(),cmd.arg[0])
	if err != nil{
		errorStr := fmt.Sprintf("URL does not exists in feed list: %s",err)
		return errors.New(errorStr)
	}
	feedFollowed, err := s.db.CreateFeedFollow(context.Background(),database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil{
		return err
	}
	fmt.Println(feedFollowed)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User)error{
	feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(),user.ID)
	if err != nil{
		errFeedFollows := fmt.Sprintf("Error getting feed follows: %s",err)
		return errors.New(errFeedFollows)
	}
	fmt.Println(feedFollows)
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User)error{
	feed, err:= s.db.GetFeed(context.Background(),cmd.arg[0])
	if err != nil{
		errStr := fmt.Sprintf("Failed to get feed, error: %s", err)
		return errors.New(errStr)
	}
	err = s.db.DeleteFeedFollow(context.Background(),database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil{
		return err
	}
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User)error{
	limit := 2
	if len(cmd.arg)==1{
		i, err := strconv.Atoi(cmd.arg[0])
		if err != nil{
			return err
		}
		limit = i
	}
	if len(cmd.arg)>1{
		return errors.New("Too many arguments")
	}
	posts, err := s.db.GetPostsForUser(context.Background(),database.GetPostsForUserParams{
		UserID: user.ID,
		Limit: int32(limit),
	})
	if err != nil{
		return err
	}
	for _, post := range posts{
		fmt.Printf("Title: %s\nLink: %s\nDescription: %s\nPublished at: %v\n",post.Title,post.Url,post.Description.String,post.PublishedAt.Time)
	}
	return nil
}