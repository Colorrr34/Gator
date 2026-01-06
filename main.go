package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/colorrr34/gator/internal/config"
	"github.com/colorrr34/gator/internal/database"
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
