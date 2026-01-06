package main

import (
	"context"

	"github.com/colorrr34/gator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error{	
	return func (s *state,cmd command)error{
		userName := (*s).cfg.CurrentUserName
		user, _ := s.db.GetUser(context.Background(),userName)
		return handler(s, cmd, user)
	}
}