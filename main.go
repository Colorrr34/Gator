package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/colorrr34/gator/internal/config"
	"github.com/colorrr34/gator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type state struct{
	db *database.Queries
	cfg *config.Config
}

type command struct{
	name string
	arg []string
}

type commands struct{
	cmdMap map[string]func(*state, command) error
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func main(){
	
	cfg := config.Read()
	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil{
		log.Fatal(err)
	}
	dbQueries := database.New(db)
	cfgState := state{
		cfg:&cfg,
		db:dbQueries,
	}
	cmdMap := make(map[string]func(*state, command) error)
	c:= commands{
		cmdMap: cmdMap,
	}
	

	(&c).register("login",handlerLogin)
	(&c).register("register",handlerRegister)
	(&c).register("reset", handlerReset)
	(&c).register("users", handlerUsers)
	(&c).register("agg", handlerAgg)
	(&c).register("addfeed",middlewareLoggedIn(handlerAddFeed))
	(&c).register("feeds",handlerFeeds)
	(&c).register("follow",middlewareLoggedIn(handlerFollow))
	(&c).register("following",middlewareLoggedIn(handlerFollowing))
	(&c).register("unfollow",middlewareLoggedIn(handlerUnfollow))
	(&c).register("browse",middlewareLoggedIn(handlerBrowse))
	args := os.Args
	if len(args)<2{
		log.Fatal("Insufficient arguments")
	}
	if (cfg)==(config.Config{}){
		log.Fatal("Empty Config")
	}
	if err:= c.run(&cfgState,command{name:args[1],arg:args[2:]});err!=nil{
		log.Fatal(err)
	}
}

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
		fmt.Printf("Title: %s\nLink: %s\nDescription: %s\nPublished at: %v",post.Title,post.Url,post.Description.String,post.PublishedAt.Time)
	}
	return nil
}

func scrapeFeeds(s *state)error{
	ctx:=context.Background()
	feedNext, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil{
		errStr := fmt.Sprintf("Get next feed error: %s", err)
		return errors.New(errStr)
	}
	if err := s.db.MarkFeedFetched(context.Background(),feedNext.ID); err != nil{
		errStr := fmt.Sprintf("Mark feed error: %s", err)
		return errors.New(errStr)
	}
	feed,err := fetchFeed(context.Background(),feedNext.Url)
	if err != nil{
		errStr := fmt.Sprintf("Fetch feed error: %s", err)
		return errors.New(errStr)
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
				return err
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
			return err
		}
		fmt.Println(post)
	}
	fmt.Printf("Finished Saving %s\n", feed.Channel.Title)
	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error{	
	return func (s *state,cmd command)error{
		userName := (*s).cfg.CurrentUserName
		user, _ := s.db.GetUser(context.Background(),userName)
		return handler(s, cmd, user)
	}
}

func (c *commands)run(s *state, cmd command) error{	
	f, exists := (*c).cmdMap[cmd.name]
	if !exists{
		return errors.New("command does not exist")
	}
	err := f(s,cmd)
	if err!= nil{
		return err
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error){
	if _,exists := (*c).cmdMap[name];exists{
		log.Fatal("Command already exists")
	}
	(*c).cmdMap[name] = f
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