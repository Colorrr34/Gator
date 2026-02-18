package main

import (
	"context"
	"log"

	"github.com/colorrr34/gator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error{	
	return func (s *state,cmd command)error{
		userName := (*s).cfg.CurrentUserName
		user, err := s.db.GetUser(context.Background(),userName)
		if err != nil{
			log.Fatal("user not exists")
		}
		return handler(s, cmd, user)
	}
}