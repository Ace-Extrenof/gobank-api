package main

import "math/rand/v2"

type Account struct {
    ID int `json:"id"`
    Firstname string `json:"first_name"`
    LastName string `json:"last_name"`
    Number int64 `json:"number"`
    Balance int64 `json:"balance"`
}

func NewAccount(firstName, lastName string) *Account {
    return &Account{
        ID: rand.IntN(10000),
        Firstname: firstName,
        LastName: lastName,
    }
}
