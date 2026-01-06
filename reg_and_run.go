package main

import (
	"errors"
	"log"
)

func (c *commands) register(name string, f func(*state, command) error){
	if _,exists := (*c).cmdMap[name];exists{
		log.Fatal("Command already exists")
	}
	(*c).cmdMap[name] = f
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
